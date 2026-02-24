package ts3

import (
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"
)

const defaultQuerySSHPort = 10022

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
	_ = w.stdin.Close()
	_ = w.session.Close()
	return w.client.Close()
}

// NewSSHClient creates a TS3 ServerQuery client over SSH.
//
// TS3 SSH ServerQuery usually listens on port 10022.
func NewSSHClient(host string, port int, user, password string) (*Client, error) {
	return NewSSHClientWithConfig(host, port, user, password, Config{})
}

// NewSSHClientWithConfig creates an SSH client with custom runtime options.
func NewSSHClientWithConfig(host string, port int, user, password string, cfg Config) (*Client, error) {
	if port == 0 {
		port = defaultQuerySSHPort
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultDialTimeout
	}

	sshCfg := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("ts3: ssh dial failed: %w", err)
	}

	session, err := sshClient.NewSession()
	if err != nil {
		_ = sshClient.Close()
		return nil, fmt.Errorf("ts3: ssh new session failed: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		_ = sshClient.Close()
		return nil, fmt.Errorf("ts3: ssh stdin pipe failed: %w", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		_ = sshClient.Close()
		return nil, fmt.Errorf("ts3: ssh stdout pipe failed: %w", err)
	}

	if err := session.Shell(); err != nil {
		_ = session.Close()
		_ = sshClient.Close()
		return nil, fmt.Errorf("ts3: ssh shell failed: %w", err)
	}

	wrapper := &sshConnWrapper{
		stdin:   stdin,
		stdout:  stdout,
		session: session,
		client:  sshClient,
	}

	cfg.Host = host
	cfg.Port = port
	cfg.Timeout = timeout
	return newClientFromConn(wrapper, cfg, true)
}
