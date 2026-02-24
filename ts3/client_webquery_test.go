package ts3

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

type webQueryTestResponse struct {
	Status map[string]interface{}   `json:"status"`
	Body   []map[string]interface{} `json:"body,omitempty"`
}

func newWebQueryTestClient(t *testing.T, handler http.HandlerFunc, sid int) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url failed: %v", err)
	}
	host, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatalf("split host port failed: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("atoi port failed: %v", err)
	}

	client, err := NewWebQueryClient(WebQueryConfig{
		Host:            host,
		Port:            port,
		APIKey:          "test-api-key",
		VirtualServerID: sid,
		Timeout:         5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewWebQueryClient failed: %v", err)
	}
	return client, srv
}

func writeWebQueryOK(t *testing.T, w http.ResponseWriter, body []map[string]interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	out := webQueryTestResponse{
		Status: map[string]interface{}{
			"code":    0,
			"message": "ok",
		},
		Body: body,
	}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		t.Fatalf("encode response failed: %v", err)
	}
}

func TestWebQueryClientClientList(t *testing.T) {
	client, srv := newWebQueryTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "test-api-key" {
			t.Fatalf("unexpected api key: %q", got)
		}
		if r.URL.Path != "/1/clientlist" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if _, ok := r.URL.Query()["-uid"]; !ok {
			t.Fatalf("missing -uid option in query: %v", r.URL.Query())
		}
		if _, ok := r.URL.Query()["-groups"]; !ok {
			t.Fatalf("missing -groups option in query: %v", r.URL.Query())
		}

		writeWebQueryOK(t, w, []map[string]interface{}{
			{
				"clid":                "1",
				"client_type":         "0",
				"client_nickname":     "Alice Bob",
				"client_servergroups": "6,7",
			},
		})
	}, 1)
	defer srv.Close()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	clients, err := client.ClientList(ctx, "-uid", "-groups")
	if err != nil {
		t.Fatalf("ClientList failed: %v", err)
	}
	if len(clients) != 1 {
		t.Fatalf("unexpected clients len: %d", len(clients))
	}
	if clients[0].Nickname != "Alice Bob" {
		t.Fatalf("unexpected nickname: %q", clients[0].Nickname)
	}
	if len(clients[0].ServerGroups) != 2 || clients[0].ServerGroups[0] != 6 {
		t.Fatalf("unexpected server groups: %+v", clients[0].ServerGroups)
	}
}

func TestWebQueryClientUseByPortAndServerInfo(t *testing.T) {
	client, srv := newWebQueryTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/serveridgetbyport":
			if got := r.URL.Query().Get("virtualserver_port"); got != "9987" {
				t.Fatalf("unexpected virtualserver_port: %q", got)
			}
			writeWebQueryOK(t, w, []map[string]interface{}{
				{"server_id": "2"},
			})
		case "/2/serverinfo":
			writeWebQueryOK(t, w, []map[string]interface{}{
				{
					"virtualserver_id":   "2",
					"virtualserver_name": "WebQuery Test",
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}, 0)
	defer srv.Close()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.UseByPort(ctx, 9987); err != nil {
		t.Fatalf("UseByPort failed: %v", err)
	}

	info, err := client.ServerInfo(ctx)
	if err != nil {
		t.Fatalf("ServerInfo failed: %v", err)
	}
	if info.ID != 2 || info.Name != "WebQuery Test" {
		t.Fatalf("unexpected server info: %+v", info)
	}
}

func TestWebQueryClientChannelSubscribe(t *testing.T) {
	client, srv := newWebQueryTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/3/channelsubscribe" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		cids := r.URL.Query()["cid"]
		if len(cids) != 3 || cids[0] != "1" || cids[2] != "3" {
			t.Fatalf("unexpected cid query: %v", cids)
		}
		writeWebQueryOK(t, w, nil)
	}, 3)
	defer srv.Close()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.ChannelSubscribe(ctx, 1, 2, 3); err != nil {
		t.Fatalf("ChannelSubscribe failed: %v", err)
	}
}

func TestWebQueryClientEventsUnsupported(t *testing.T) {
	client, srv := newWebQueryTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeWebQueryOK(t, w, nil)
	}, 1)
	defer srv.Close()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.RegisterTextEvents(ctx); err == nil {
		t.Fatalf("expected RegisterTextEvents to fail in webquery mode")
	}
	if err := client.OnTextMessage(ctx, func(string) {}); err == nil {
		t.Fatalf("expected OnTextMessage to fail in webquery mode")
	}
}
