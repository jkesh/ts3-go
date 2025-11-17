package ts3

import (
	"bufio"
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// NewSSHClient 创建基于 SSH 的连接
func NewSSHClient(host string, port int, user, password string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// 在生产环境中，应该使用 ssh.FixedHostKey(key) 来验证服务器
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed: %w", err)
	}

	session, err := sshClient.NewSession()
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("ssh session failed: %w", err)
	}

	// 请求 Shell 以便进行交互
	if err := session.Shell(); err != nil {
		session.Close()
		sshClient.Close()
		return nil, fmt.Errorf("ssh shell failed: %w", err)
	}

	// 获取 stdin 和 stdout 管道
	_, err = session.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// 构造 Client
	// 注意：这里我们需要对 Client 结构体做微调，
	// 因为 ssh session 不是 net.Conn。
	// 建议：将 Client.conn 类型改为 io.ReadWriteCloser 接口

	// 临时适配方案（为了不破坏现有结构体）：
	// 我们可以创建一个实现了 net.Conn 的 wrapper，或者直接重构 Client.conn
	// 下面演示使用 net.Pipe 进行适配是比较复杂的，
	// 最好的方法是修改 Client 结构体中的 `conn net.Conn` 为 `rw io.ReadWriteCloser`

	// 假设您已将 Client.conn 改为 io.ReadWriteCloser：
	c := &Client{
		// conn: wrapper{stdin, stdout, session}, // 需要自定义 wrapper
		scanner:       bufio.NewScanner(stdout),
		cmdResChan:    make(chan string),
		errorChan:     make(chan error, 1),
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
	}

	// SSH 模式不需要像 Raw TCP 那样先发送 "TS3" 握手，直接开始 readLoop
	go c.readLoop()

	// SSH 连接后通常不需要 login 命令（因为 SSH 握手时已验证），
	// 但需要 use sid=1

	return c, nil
}
