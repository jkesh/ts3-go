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

	// 确保引用路径正确
	"github.com/jkesh/ts3-go/ts3/models"
)

// Config 客户端配置
type Config struct {
	Host            string
	Port            int
	Timeout         time.Duration
	KeepAlivePeriod time.Duration
}

// Client TS3 ServerQuery 客户端
type Client struct {
	conn    io.ReadWriteCloser
	scanner *bufio.Scanner
	mu      sync.Mutex

	cmdResChan    chan string
	errorChan     chan error
	notifications map[string][]func(string)
	notifyMu      sync.RWMutex
	quit          chan struct{}
	logger        Logger
}

// NewClient 创建连接
func NewClient(cfg Config) (*Client, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn: conn,
		// [重要] Scanner 在这里创建一次并复用，防止缓冲区数据丢失
		scanner: bufio.NewScanner(conn),
		// [重要] 缓冲区设为 100，防止阻塞
		cmdResChan:    make(chan string, 100),
		errorChan:     make(chan error, 1),
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
		logger:        &NopLogger{},
	}

	// 执行握手
	if err := c.handshake(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 启动后台读取
	go c.readLoop()

	// 启动心跳
	if cfg.KeepAlivePeriod > 0 {
		go c.keepAliveLoop(cfg.KeepAlivePeriod)
	}

	return c, nil
}

// handshake 读取前两行欢迎信息
func (c *Client) handshake() error {
	for i := 0; i < 2; i++ {
		if !c.scanner.Scan() {
			return errors.New("connection closed during handshake")
		}
		// 可以在这里校验 text 是否包含 "TS3"
	}
	return nil
}

// Close 关闭连接
func (c *Client) Close() error {
	select {
	case <-c.quit:
		return nil
	default:
		close(c.quit)
		return c.conn.Close()
	}
}

// Exec 执行命令
func (c *Client) Exec(ctx context.Context, cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 快速失败检查
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// 发送命令
	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		return "", err
	}

	c.logger.Debugf("-> %s", cmd)

	var responseBuilder strings.Builder

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()

		case line, ok := <-c.cmdResChan:
			if !ok {
				return "", errors.New("connection closed")
			}

			// [核心修复] 去除首尾空白字符（包括 \r 和 \n），防止判断失败
			trimmedLine := strings.TrimSpace(line)

			// 检查是否是 error 行
			if strings.HasPrefix(trimmedLine, "error id=") {
				c.logger.Debugf("<- %s", trimmedLine)

				var ts3Err Error
				// 解析错误信息
				if err := NewDecoder().Decode(trimmedLine, &ts3Err); err != nil {
					return "", fmt.Errorf("decode error: %w", err)
				}

				if ts3Err.ID != 0 {
					return responseBuilder.String(), &ts3Err
				}
				// 成功，返回之前累积的数据
				return strings.TrimSpace(responseBuilder.String()), nil
			}

			// 拼接数据行
			if responseBuilder.Len() > 0 {
				responseBuilder.WriteString("|")
			}
			responseBuilder.WriteString(line)

		case err := <-c.errorChan:
			return "", fmt.Errorf("connection error: %w", err)

		case <-c.quit:
			return "", errors.New("client closed")
		}
	}
}

// readLoop 后台读取循环
func (c *Client) readLoop() {
	defer close(c.errorChan)

	// Panic 恢复，防止程序崩溃
	defer func() {
		if r := recover(); r != nil {
			c.logger.Printf("Panic in readLoop: %v", r)
			c.errorChan <- fmt.Errorf("panic: %v", r)
		}
	}()

	for c.scanner.Scan() {
		text := c.scanner.Text()

		if len(text) == 0 {
			continue
		}

		// 处理异步通知
		if strings.HasPrefix(text, "notify") {
			go c.dispatchNotify(text)
			continue
		}

		// 将数据发给 Exec
		// [重要] 移除了 default 分支，确保数据一定会进入通道（除非连接关闭）
		select {
		case c.cmdResChan <- text:
		case <-c.quit:
			return
		}
	}

	if err := c.scanner.Err(); err != nil {
		c.errorChan <- err
	}
}

// keepAliveLoop 心跳保活
func (c *Client) keepAliveLoop(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送 whoami 作为心跳，设置 5 秒超时
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, _ = c.Exec(ctx, "whoami")
			cancel()
		case <-c.quit:
			return
		}
	}
}

// SetLogger 设置日志
func (c *Client) SetLogger(l Logger) {
	if l == nil {
		c.logger = &NopLogger{}
	} else {
		c.logger = l
	}
}

// 兼容性 Helper 方法（建议在 methods.go 中实现，这里保留是为了防止报错）
func (c *Client) ServerGroupList(ctx context.Context) ([]models.ServerGroup, error) {
	resp, err := c.Exec(ctx, "servergrouplist")
	if err != nil {
		return nil, err
	}
	var groups []models.ServerGroup
	if err := NewDecoder().Decode(resp, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}
