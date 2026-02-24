package ts3

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestChannelSubscribeBuildsMultiCIDCommand(t *testing.T) {
	cmdCh := make(chan string, 1)
	conn := newMockServerConn(t, func(cmd string) []string {
		cmdCh <- cmd
		return []string{"error id=0 msg=ok"}
	})

	client, err := NewClientFromConn(conn, Config{})
	if err != nil {
		t.Fatalf("NewClientFromConn failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.ChannelSubscribe(ctx, 1, 2, 3); err != nil {
		t.Fatalf("ChannelSubscribe failed: %v", err)
	}

	got := <-cmdCh
	want := "channelsubscribe cid=1|cid=2|cid=3"
	if got != want {
		t.Fatalf("unexpected command: got=%q want=%q", got, want)
	}
}

func TestServerEditBuildsExpectedCommand(t *testing.T) {
	cmdCh := make(chan string, 1)
	conn := newMockServerConn(t, func(cmd string) []string {
		cmdCh <- cmd
		return []string{"error id=0 msg=ok"}
	})

	client, err := NewClientFromConn(conn, Config{})
	if err != nil {
		t.Fatalf("NewClientFromConn failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.ServerEdit(ctx, ServerEditOptions{
		Name:           "My Server",
		WelcomeMessage: "hello world",
		MaxClients:     64,
	})
	if err != nil {
		t.Fatalf("ServerEdit failed: %v", err)
	}

	got := <-cmdCh
	if !strings.HasPrefix(got, "serveredit ") {
		t.Fatalf("unexpected command prefix: %q", got)
	}
	if !strings.Contains(got, "virtualserver_name=My\\sServer") {
		t.Fatalf("missing escaped server name: %q", got)
	}
	if !strings.Contains(got, "virtualserver_welcomemessage=hello\\sworld") {
		t.Fatalf("missing escaped welcome message: %q", got)
	}
	if !strings.Contains(got, "virtualserver_maxclients=64") {
		t.Fatalf("missing maxclients: %q", got)
	}
}

func TestQueryLoginAddDecode(t *testing.T) {
	conn := newMockServerConn(t, func(cmd string) []string {
		if cmd == "queryloginadd cldbid=10 sid=1" {
			return []string{
				"cldbid=10 sid=1 client_login_name=q-user client_login_password=q-pass",
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

	creds, err := client.QueryLoginAdd(ctx, 10, 1)
	if err != nil {
		t.Fatalf("QueryLoginAdd failed: %v", err)
	}
	if creds.LoginName != "q-user" || creds.Password != "q-pass" {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
}

func TestBanListDecode(t *testing.T) {
	conn := newMockServerConn(t, func(cmd string) []string {
		if cmd == "banlist" {
			return []string{
				"banid=1 reason=test\\sreason invokername=admin",
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

	out, err := client.BanList(ctx)
	if err != nil {
		t.Fatalf("BanList failed: %v", err)
	}
	if len(out) != 1 || out[0].BanID != 1 || out[0].Reason != "test reason" {
		t.Fatalf("unexpected ban list decode: %+v", out)
	}
}

func TestChannelEditBuildsExpectedCommand(t *testing.T) {
	cmdCh := make(chan string, 1)
	conn := newMockServerConn(t, func(cmd string) []string {
		cmdCh <- cmd
		return []string{"error id=0 msg=ok"}
	})

	client, err := NewClientFromConn(conn, Config{})
	if err != nil {
		t.Fatalf("NewClientFromConn failed: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.ChannelEdit(ctx, 12, ChannelEditOptions{
		Name:         "Music Room",
		Topic:        "all day",
		MaxClients:   20,
		IsPermanent:  true,
		CodecQuality: 7,
	})
	if err != nil {
		t.Fatalf("ChannelEdit failed: %v", err)
	}

	got := <-cmdCh
	if !strings.HasPrefix(got, "channeledit cid=12 ") {
		t.Fatalf("unexpected command prefix: %q", got)
	}
	if !strings.Contains(got, "channel_name=Music\\sRoom") {
		t.Fatalf("missing channel name: %q", got)
	}
	if !strings.Contains(got, "channel_topic=all\\sday") {
		t.Fatalf("missing channel topic: %q", got)
	}
	if !strings.Contains(got, "channel_maxclients=20") {
		t.Fatalf("missing maxclients: %q", got)
	}
	if !strings.Contains(got, "channel_flag_permanent=1") {
		t.Fatalf("missing permanent flag: %q", got)
	}
}
