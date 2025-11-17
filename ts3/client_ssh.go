package ts3

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

type sshConnWrapper struct {
	stdin   io.WriteCloser
	stdout  io.Reader
	session *ssh.Session
	client  *ssh.Client
}

func (w *sshConnWrapper) Read(p []byte) (n int, err error) {
	return w.stdout.Read(p)
}

func (w *sshConnWrapper) Write(p []byte) (n int, err error) {
	return w.stdin.Write(p)
}

func (w *sshConnWrapper) Close() error {
	// 关闭顺序很重要
	_ = w.stdin.Close()
	_ = w.session.Close()
	return w.client.Close()
}

// NewSSHClient 创建基于 SSH 的连接
func NewSSHClient(host string, port int, user, password string) (*Client, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
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

	if err := session.Shell(); err != nil {
		session.Close()
		sshClient.Close()
		return nil, fmt.Errorf("ssh shell failed: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// 构造 Wrapper
	wrapper := &sshConnWrapper{
		stdin:   stdin,
		stdout:  stdout,
		session: session,
		client:  sshClient,
	}

	c := &Client{
		conn:          wrapper, // 现在可以赋值了
		scanner:       bufio.NewScanner(stdout),
		cmdResChan:    make(chan string),
		errorChan:     make(chan error, 1),
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
		logger:        &NopLogger{},
	}

	go c.readLoop()

	return c, nil
}
