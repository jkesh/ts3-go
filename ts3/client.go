package ts3

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	defaultQueryPort   = 10011
	defaultDialTimeout = 10 * time.Second
	defaultMaxLineSize = 1024 * 1024
	defaultCmdBufSize  = 256
)

// Config holds connection and runtime options for a TS3 ServerQuery client.
type Config struct {
	Host            string
	Port            int
	Timeout         time.Duration
	KeepAlivePeriod time.Duration
	MaxLineSize     int
}

// Client is a TS3 ServerQuery client.
//
// Commands are executed sequentially because ServerQuery replies do not include
// per-request IDs. Use one Client instance per connection/session.
type Client struct {
	conn    io.ReadWriteCloser
	scanner *bufio.Scanner

	mu         sync.Mutex
	cmdResChan chan string
	errorChan  chan error

	notifications map[string][]func(string)
	notifyMu      sync.RWMutex

	quit      chan struct{}
	closeOnce sync.Once

	logger   Logger
	loggerMu sync.RWMutex
}

// NewClient creates a TCP-based TS3 ServerQuery client.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("ts3: host is required")
	}

	port := cfg.Port
	if port == 0 {
		port = defaultQueryPort
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultDialTimeout
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", cfg.Host, port), timeout)
	if err != nil {
		return nil, fmt.Errorf("ts3: dial failed: %w", err)
	}

	cfg.Port = port
	cfg.Timeout = timeout
	return newClientFromConn(conn, cfg, true)
}

// NewClientFromConn creates a client from an existing connection.
//
// It is useful for tests or custom transports.
func NewClientFromConn(conn io.ReadWriteCloser, cfg Config) (*Client, error) {
	return newClientFromConn(conn, cfg, true)
}

func newClientFromConn(conn io.ReadWriteCloser, cfg Config, doHandshake bool) (*Client, error) {
	if conn == nil {
		return nil, errors.New("ts3: nil connection")
	}

	maxLineSize := cfg.MaxLineSize
	if maxLineSize <= 0 {
		maxLineSize = defaultMaxLineSize
	}

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	c := &Client{
		conn:          conn,
		scanner:       scanner,
		cmdResChan:    make(chan string, defaultCmdBufSize),
		errorChan:     make(chan error, 1),
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
		logger:        &NopLogger{},
	}

	if doHandshake {
		if err := c.handshake(); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}

	go c.readLoop()

	if cfg.KeepAlivePeriod > 0 {
		go c.keepAliveLoop(cfg.KeepAlivePeriod)
	}

	return c, nil
}

// handshake reads and validates the initial ServerQuery banner.
func (c *Client) handshake() error {
	lines := make([]string, 0, 2)
	for len(lines) < 2 {
		if !c.scanner.Scan() {
			return errors.New("ts3: connection closed during handshake")
		}
		text := strings.TrimSpace(c.scanner.Text())
		if text == "" {
			continue
		}
		lines = append(lines, text)
	}

	if !strings.Contains(strings.ToUpper(lines[0]), "TS3") {
		return fmt.Errorf("ts3: unexpected handshake banner: %q", lines[0])
	}

	return nil
}

// Close closes the client connection and stops background loops.
func (c *Client) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		close(c.quit)
		closeErr = c.conn.Close()
	})
	return closeErr
}

// Exec sends a raw ServerQuery command and returns the data part of response.
//
// The returned string contains one or multiple response rows joined by "|" and
// excludes the final "error id=..." line.
func (c *Client) Exec(ctx context.Context, cmd string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.quit:
		return "", errors.New("ts3: client closed")
	default:
	}

	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		return "", fmt.Errorf("ts3: write failed: %w", err)
	}
	c.debugf("-> %s", cmd)

	ctxDone := ctx.Done()
	var ctxErr error
	var responseLines []string
	cmdCh := c.cmdResChan
	errCh := c.errorChan

	for {
		if cmdCh == nil && errCh == nil {
			if ctxErr != nil {
				return strings.Join(responseLines, "|"), ctxErr
			}
			return strings.Join(responseLines, "|"), errors.New("ts3: connection closed")
		}

		select {
		case <-ctxDone:
			// Keep draining until the command terminator ("error id=...") arrives.
			// This preserves protocol sync for subsequent commands.
			ctxErr = ctx.Err()
			ctxDone = nil

		case line, ok := <-cmdCh:
			if !ok {
				cmdCh = nil
				continue
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "error id=") {
				c.debugf("<- %s", line)

				var ts3Err Error
				if err := NewDecoder().Decode(line, &ts3Err); err != nil {
					return strings.Join(responseLines, "|"), fmt.Errorf("ts3: failed to decode error line: %w", err)
				}

				if ctxErr != nil {
					return strings.Join(responseLines, "|"), ctxErr
				}

				if ts3Err.ID != ErrOK {
					return strings.Join(responseLines, "|"), &ts3Err
				}
				return strings.Join(responseLines, "|"), nil
			}

			responseLines = append(responseLines, line)

		case err, ok := <-errCh:
			if !ok {
				errCh = nil
				continue
			}
			if err != nil {
				return strings.Join(responseLines, "|"), fmt.Errorf("ts3: connection error: %w", err)
			}

		case <-c.quit:
			if ctxErr != nil {
				return strings.Join(responseLines, "|"), ctxErr
			}
			return strings.Join(responseLines, "|"), errors.New("ts3: client closed")
		}
	}
}

// readLoop continuously reads lines from the connection.
func (c *Client) readLoop() {
	defer func() {
		if r := recover(); r != nil {
			c.logf("panic in readLoop: %v", r)
			c.sendConnErr(fmt.Errorf("ts3: panic in readLoop: %v", r))
		}
		close(c.cmdResChan)
		close(c.errorChan)
	}()

	for c.scanner.Scan() {
		text := strings.TrimSpace(c.scanner.Text())
		if text == "" {
			continue
		}

		if strings.HasPrefix(text, "notify") {
			go c.dispatchNotify(text)
			continue
		}

		select {
		case c.cmdResChan <- text:
		case <-c.quit:
			return
		}
	}

	if err := c.scanner.Err(); err != nil {
		c.sendConnErr(err)
	}
}

func (c *Client) sendConnErr(err error) {
	if err == nil {
		return
	}
	select {
	case c.errorChan <- err:
	default:
	}
}

// keepAliveLoop periodically executes "whoami" to keep long-lived sessions alive.
func (c *Client) keepAliveLoop(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, _ = c.Exec(ctx, "whoami")
			cancel()
		case <-c.quit:
			return
		}
	}
}

// SetLogger sets the logger used by the client.
func (c *Client) SetLogger(l Logger) {
	if l == nil {
		l = &NopLogger{}
	}
	c.loggerMu.Lock()
	c.logger = l
	c.loggerMu.Unlock()
}

func (c *Client) getLogger() Logger {
	c.loggerMu.RLock()
	l := c.logger
	c.loggerMu.RUnlock()
	if l == nil {
		return &NopLogger{}
	}
	return l
}

func (c *Client) logf(format string, v ...interface{}) {
	c.getLogger().Printf(format, v...)
}

func (c *Client) debugf(format string, v ...interface{}) {
	c.getLogger().Debugf(format, v...)
}
