package ts3

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
	"ts3-go/ts3/models"
)

// Config 客户端配置
type Config struct {
	Host            string
	Port            int
	Timeout         time.Duration
	KeepAlivePeriod time.Duration // 自动发送心跳的间隔
}

type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner
	mu      sync.Mutex // 确保一次只发送一个命令

	// cmdResChan 用于在 Exec 内部接收 readLoop 读取到的命令响应行
	cmdResChan chan string

	// errorChan 用于通知连接级错误
	errorChan chan error

	// notifications 存储事件回调
	notifications map[string][]func(string)
	notifyMu      sync.RWMutex

	// close 信号
	quit chan struct{}

	logger Logger
}

// NewClient 创建并连接到 TS3 服务器
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
		conn:          conn,
		scanner:       bufio.NewScanner(conn),
		cmdResChan:    make(chan string), // 无缓冲，确保同步
		errorChan:     make(chan error, 1),
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
	}

	// 处理初始握手信息
	if err := c.handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	// 启动后台读取循环
	go c.readLoop()

	// 启动心跳 (KeepAlive)
	if cfg.KeepAlivePeriod > 0 {
		go c.keepAliveLoop(cfg.KeepAlivePeriod)
	}

	return c, nil
}

func (c *Client) handshake() error {
	scanner := bufio.NewScanner(c.conn)
	for i := 0; i < 2; i++ {
		if !scanner.Scan() {
			return errors.New("connection closed during handshake")
		}
	}
	c.scanner = scanner
	return nil
}

// Close 关闭连接
func (c *Client) Close() error {
	close(c.quit)
	return c.conn.Close()
}

// Exec 执行原始命令并返回结果字符串
func (c *Client) Exec(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 发送命令 (添加换行符)
	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		return "", err
	}

	var responseBuilder strings.Builder

	// 循环等待 readLoop 发来的数据
	for {
		select {
		case line := <-c.cmdResChan:
			if strings.HasPrefix(line, "error id=") {
				if line != "error id=0 msg=ok" {
					return responseBuilder.String(), fmt.Errorf("ts3 error: %s", line)
				}
				return strings.TrimSpace(responseBuilder.String()), nil
			}

			if responseBuilder.Len() > 0 {
				responseBuilder.WriteString("|")
			}
			responseBuilder.WriteString(line)

		case err := <-c.errorChan:
			return "", fmt.Errorf("connection error: %v", err)
		case <-c.quit:
			return "", errors.New("client closed")
		case <-time.After(10 * time.Second): // 命令超时防死锁
			return "", errors.New("command timeout")
		}
	}
}

func (c *Client) readLoop() {
	defer close(c.errorChan)

	for c.scanner.Scan() {
		text := c.scanner.Text()

		// 1. 忽略空行
		if len(text) == 0 {
			continue
		}

		// 2. 处理通知/事件 (Notification)
		if strings.HasPrefix(text, "notify") {
			go c.dispatchNotify(text)
			continue
		}

		// 3. 处理命令响应
		select {
		case c.cmdResChan <- text:
		case <-c.quit:
			return
		default:
		}
	}

	if err := c.scanner.Err(); err != nil {
		c.errorChan <- err
	}
}

// keepAliveLoop 定时发送命令防止断开
func (c *Client) keepAliveLoop(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, _ = c.Exec("whoami")
		case <-c.quit:
			return
		}
	}
}
func (c *Client) SetLogger(l Logger) {
	if l == nil {
		c.logger = &NopLogger{}
	} else {
		c.logger = l
	}
}
func (c *Client) ServerGroupList() ([]models.ServerGroup, error) {
	resp, err := c.Exec("servergrouplist")
	if err != nil {
		return nil, err
	}
	var groups []models.ServerGroup
	if err := NewDecoder().Decode(resp, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}
