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

// Config 客户端配置
type Config struct {
	Host            string
	Port            int
	Timeout         time.Duration // 连接超时
	KeepAlivePeriod time.Duration // 自动发送心跳的间隔 (建议 1-5 分钟)
}

// Client TS3 ServerQuery 客户端核心结构
type Client struct {
	conn    io.ReadWriteCloser
	scanner *bufio.Scanner
	mu      sync.Mutex // 互斥锁，确保 Request-Response 的原子性

	// cmdResChan 用于将 readLoop 读取到的命令响应数据传递给 Exec
	cmdResChan chan string

	// errorChan 用于通知连接级的致命错误（如断网）
	errorChan chan error

	// notifications 存储事件回调 (key: eventName, value: callbacks)
	notifications map[string][]func(string)
	notifyMu      sync.RWMutex

	// quit 用于关闭信号
	quit chan struct{}

	// logger 日志接口
	logger Logger
}

// NewClient 创建并连接到 TS3 服务器
func NewClient(cfg Config) (*Client, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// 默认超时设置
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
		cmdResChan:    make(chan string, 100), // 无缓冲 channel，确保同步读写
		errorChan:     make(chan error, 1),
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
		logger:        &NopLogger{}, // 默认不打印日志
	}

	// 1. 处理初始握手信息 ("TS3", "Welcome...")
	if err := c.handshake(); err != nil {
		_ = conn.Close()
		return nil, err
	}

	// 2. 启动后台读取循环
	go c.readLoop()

	// 3. 启动心跳保活 (KeepAlive)
	if cfg.KeepAlivePeriod > 0 {
		go c.keepAliveLoop(cfg.KeepAlivePeriod)
	}

	return c, nil
}

// handshake 读取连接后的前两行欢迎信息
func (c *Client) handshake() error {
	// TS3 Server 连接后会立即发送两行:
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
	select {
	case <-c.quit:
		return nil // 已经关闭
	default:
		close(c.quit)
		return c.conn.Close()
	}
}

// Exec 执行原始命令并返回结果字符串
// ctx: 用于控制超时或取消
// cmd: 原始命令字符串 (无需换行符)
func (c *Client) Exec(ctx context.Context, cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. 检查 Context 是否已取消
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// 2. 发送命令 (追加换行符)
	if _, err := c.conn.Write([]byte(cmd + "\n")); err != nil {
		return "", err
	}

	// 记录调试日志
	c.logger.Debugf("-> %s", cmd)

	var responseBuilder strings.Builder

	// 3. 循环等待响应
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()

		case line, ok := <-c.cmdResChan:
			if !ok {
				return "", errors.New("connection closed")
			}

			// 检查是否为命令结束标志 "error id=..."
			if strings.HasPrefix(line, "error id=") {
				c.logger.Debugf("<- %s", line)

				// 使用 Decoder 解析错误信息
				var ts3Err Error
				// 注意：NewDecoder() 和 Error 结构体已经在您的项目中定义
				if err := NewDecoder().Decode(line, &ts3Err); err != nil {
					return "", fmt.Errorf("failed to decode error line: %w", err)
				}

				// 如果 ID != 0，返回具体的 TS3 错误
				if ts3Err.ID != 0 {
					return responseBuilder.String(), &ts3Err
				}

				// ID == 0，表示成功，返回累积的数据字符串
				return strings.TrimSpace(responseBuilder.String()), nil
			}

			// 累积数据行 (多行数据用 "|" 分隔)
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

// readLoop 后台读取循环，负责分流 Notify 事件和 Command 响应
func (c *Client) readLoop() {
	defer close(c.errorChan)

	// 增加 Panic 恢复机制，提高健壮性
	defer func() {
		if r := recover(); r != nil {
			c.logger.Printf("Panic in readLoop: %v", r)
			// 触发错误通知，避免死锁
			c.errorChan <- fmt.Errorf("panic: %v", r)
		}
	}()

	for c.scanner.Scan() {
		text := c.scanner.Text()

		// 忽略空行
		if len(text) == 0 {
			continue
		}

		// 1. 处理异步事件 (Notification)
		if strings.HasPrefix(text, "notify") {
			// 异步分发，避免阻塞读取循环
			go c.dispatchNotify(text)
			continue
		}

		// 2. 处理命令响应
		// 尝试发送给 Exec，如果 Exec 没在等 (超时或无命令)，则丢弃或记录
		select {
		case c.cmdResChan <- text:
		case <-c.quit:
			return
		default:
			// 收到非预期的消息 (可能是上一个超时命令的残留数据)
			c.logger.Debugf("Ignored unexpected line: %s", text)
		}
	}

	if err := c.scanner.Err(); err != nil {
		c.errorChan <- err
	}
}

// keepAliveLoop 定时发送无副作用命令，防止服务器断开连接
func (c *Client) keepAliveLoop(period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送 "whoami" 作为心跳
			// 设置短暂的超时上下文
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_, err := c.Exec(ctx, "whoami")
			cancel()

			if err != nil {
				c.logger.Printf("KeepAlive failed: %v", err)
				// 如果心跳连续失败，可以在这里考虑触发重连逻辑
			}
		case <-c.quit:
			return
		}
	}
}

// SetLogger 设置日志记录器
func (c *Client) SetLogger(l Logger) {
	if l == nil {
		c.logger = &NopLogger{}
	} else {
		c.logger = l
	}
}
