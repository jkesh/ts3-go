package ts3

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"
)

func newMockServerConn(t *testing.T, handler func(cmd string) []string) net.Conn {
	t.Helper()

	serverConn, clientConn := net.Pipe()
	go func() {
		defer serverConn.Close()

		writer := bufio.NewWriter(serverConn)
		_, _ = writer.WriteString("TS3\n")
		_, _ = writer.WriteString("Welcome to TeamSpeak 3 ServerQuery\n")
		_ = writer.Flush()

		scanner := bufio.NewScanner(serverConn)
		for scanner.Scan() {
			cmd := strings.TrimSpace(scanner.Text())
			lines := handler(cmd)
			for _, line := range lines {
				_, _ = writer.WriteString(line + "\n")
			}
			_ = writer.Flush()
		}
	}()

	return clientConn
}

func TestClientExecAndErrorDecode(t *testing.T) {
	conn := newMockServerConn(t, func(cmd string) []string {
		switch cmd {
		case "whoami":
			return []string{
				"virtualserver_id=1 client_id=5 client_channel_id=2 client_nickname=Bot",
				"error id=0 msg=ok",
			}
		default:
			return []string{
				"error id=256 msg=command\\snot\\sfound",
			}
		}
	})

	client, err := NewClientFromConn(conn, Config{})
	if err != nil {
		t.Fatalf("NewClientFromConn failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Exec(ctx, "whoami")
	if err != nil {
		t.Fatalf("Exec(whoami) failed: %v", err)
	}
	if !strings.Contains(resp, "client_id=5") {
		t.Fatalf("unexpected whoami response: %q", resp)
	}

	_, err = client.Exec(ctx, "unknowncmd")
	if err == nil {
		t.Fatalf("expected command error")
	}

	var ts3Err *Error
	if !errors.As(err, &ts3Err) {
		t.Fatalf("expected ts3.Error, got: %T (%v)", err, err)
	}
	if ts3Err.ID != ErrCommandNotFound {
		t.Fatalf("unexpected error id: %d", ts3Err.ID)
	}
}

func TestWhoAmIMethod(t *testing.T) {
	conn := newMockServerConn(t, func(cmd string) []string {
		if cmd == "whoami" {
			return []string{
				"virtualserver_id=1 client_id=5 client_channel_id=2 client_nickname=QueryBot client_database_id=10",
				"error id=0 msg=ok",
			}
		}
		return []string{"error id=0 msg=ok"}
	})

	client, err := NewClientFromConn(conn, Config{})
	if err != nil {
		t.Fatalf("NewClientFromConn failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	me, err := client.WhoAmI(ctx)
	if err != nil {
		t.Fatalf("WhoAmI failed: %v", err)
	}
	if me.ClientID != 5 || me.VirtualServerID != 1 {
		t.Fatalf("unexpected whoami data: %+v", me)
	}
}

func TestOnTextMessageDispatch(t *testing.T) {
	notifyReady := make(chan struct{}, 1)
	conn := newMockServerConn(t, func(cmd string) []string {
		if strings.HasPrefix(cmd, "servernotifyregister event=text") {
			return []string{"error id=0 msg=ok"}
		}
		return []string{"error id=0 msg=ok"}
	})

	client, err := NewClientFromConn(conn, Config{})
	if err != nil {
		t.Fatalf("NewClientFromConn failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.OnTextMessage(ctx, func(payload string) {
		if strings.Contains(payload, "hello") {
			notifyReady <- struct{}{}
		}
	}); err != nil {
		t.Fatalf("OnTextMessage failed: %v", err)
	}

	// Inject a notify line directly to the dispatcher path.
	client.dispatchNotify("notifytextmessage invokername=Alice msg=hello")

	select {
	case <-notifyReady:
	case <-time.After(2 * time.Second):
		t.Fatalf("text notification handler was not called")
	}
}
