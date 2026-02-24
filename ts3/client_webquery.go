package ts3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// WebQueryConfig configures TS3 WebQuery (HTTP/HTTPS) transport.
//
// TeamSpeak 6 server provides query-http/query-https endpoints and expects
// the API key in header "x-api-key".
type WebQueryConfig struct {
	Host            string
	Port            int
	HTTPS           bool
	APIKey          string
	BasePath        string
	Timeout         time.Duration
	KeepAlivePeriod time.Duration
	VirtualServerID int
	HTTPClient      *http.Client
}

type webQueryRuntime struct {
	baseURL    string
	basePath   string
	apiKey     string
	httpClient *http.Client
}

type webQueryArg struct {
	key      string
	value    string
	hasValue bool
}

var webQueryGlobalCommands = map[string]struct{}{
	"help":              {},
	"version":           {},
	"hostinfo":          {},
	"instanceinfo":      {},
	"instanceedit":      {},
	"bindinglist":       {},
	"serverlist":        {},
	"servercreate":      {},
	"serverdelete":      {},
	"serverstart":       {},
	"serverstop":        {},
	"serverprocessstop": {},
	"serveridgetbyport": {},
	"apikeyadd":         {},
	"apikeydel":         {},
	"apikeylist":        {},
	"permissionlist":    {},
	"permidgetbyname":   {},
	"permfind":          {},
	"permget":           {},
}

// NewWebQueryClient creates a TS3 client that talks to WebQuery REST endpoint.
func NewWebQueryClient(cfg WebQueryConfig) (*Client, error) {
	if strings.TrimSpace(cfg.Host) == "" {
		return nil, errors.New("ts3: host is required")
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("ts3: webquery api key is required")
	}

	port := cfg.Port
	if port == 0 {
		if cfg.HTTPS {
			port = defaultWebQueryTLSPort
		} else {
			port = defaultWebQueryPort
		}
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultDialTimeout
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}

	scheme := "http"
	if cfg.HTTPS {
		scheme = "https"
	}

	basePath := normalizeWebQueryBasePath(cfg.BasePath)
	c := &Client{
		web: &webQueryRuntime{
			baseURL:    fmt.Sprintf("%s://%s:%d", scheme, cfg.Host, port),
			basePath:   basePath,
			apiKey:     cfg.APIKey,
			httpClient: httpClient,
		},
		transport:     transportWebQuery,
		selectedSID:   cfg.VirtualServerID,
		notifications: make(map[string][]func(string)),
		quit:          make(chan struct{}),
		logger:        &NopLogger{},
	}

	if cfg.KeepAlivePeriod > 0 {
		go c.keepAliveLoop(cfg.KeepAlivePeriod)
	}

	return c, nil
}

func normalizeWebQueryBasePath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	return "/" + path
}

func (c *Client) execWebQuery(ctx context.Context, rawCmd string) (string, error) {
	cmd, args, err := parseWebQueryCommand(rawCmd)
	if err != nil {
		return "", err
	}

	path := c.web.basePath
	if sid := c.selectedSID; sid > 0 && !isWebQueryGlobalCommand(cmd) {
		path += "/" + strconv.Itoa(sid)
	}
	path += "/" + cmd

	if query := buildWebQueryQuery(args); query != "" {
		path += "?" + query
	}
	endpoint := c.web.baseURL + path
	c.debugf("-> %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("ts3: create webquery request failed: %w", err)
	}
	req.Header.Set("x-api-key", c.web.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.web.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ts3: webquery request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ts3: read webquery response failed: %w", err)
	}

	response, status, err := parseWebQueryResponse(bodyBytes)
	if err != nil {
		if resp.StatusCode >= http.StatusBadRequest {
			return "", fmt.Errorf("ts3: webquery http %d: %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		}
		return "", err
	}

	if resp.StatusCode >= http.StatusBadRequest && status.ID == ErrOK {
		status.ID = resp.StatusCode
		if status.Msg == "" {
			status.Msg = strings.TrimSpace(string(bodyBytes))
		}
	}
	if status.ID != ErrOK {
		return response, &status
	}

	return response, nil
}

func parseWebQueryCommand(raw string) (string, []webQueryArg, error) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		return "", nil, errors.New("ts3: empty command")
	}

	cmd := fields[0]
	args := make([]webQueryArg, 0, len(fields)-1)
	for _, field := range fields[1:] {
		chunks := strings.Split(field, "|")
		for _, chunk := range chunks {
			if chunk == "" {
				continue
			}
			kv := strings.SplitN(chunk, "=", 2)
			if len(kv) == 2 {
				args = append(args, webQueryArg{
					key:      kv[0],
					value:    kv[1],
					hasValue: true,
				})
				continue
			}

			args = append(args, webQueryArg{
				key:      chunk,
				hasValue: false,
			})
		}
	}

	return cmd, args, nil
}

func buildWebQueryQuery(args []webQueryArg) string {
	if len(args) == 0 {
		return ""
	}

	parts := make([]string, 0, len(args))
	for _, arg := range args {
		key := url.QueryEscape(arg.key)
		if !arg.hasValue {
			parts = append(parts, key)
			continue
		}
		parts = append(parts, key+"="+url.QueryEscape(arg.value))
	}
	return strings.Join(parts, "&")
}

func parseWebQueryResponse(payload []byte) (string, Error, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(payload, &root); err != nil {
		return "", Error{}, fmt.Errorf("ts3: decode webquery json failed: %w", err)
	}

	status := Error{ID: ErrOK, Msg: "ok"}
	if rawStatus, ok := root["status"]; ok {
		if m, ok := rawStatus.(map[string]interface{}); ok {
			if code, ok := toInt(m["code"]); ok {
				status.ID = code
			}
			if msg, ok := m["message"].(string); ok {
				status.Msg = msg
			}
		}
	}

	rawBody, exists := root["body"]
	if !exists || rawBody == nil {
		return "", status, nil
	}

	lines, err := normalizeWebQueryBody(rawBody)
	if err != nil {
		return "", status, err
	}
	return strings.Join(lines, "|"), status, nil
}

func normalizeWebQueryBody(rawBody interface{}) ([]string, error) {
	switch body := rawBody.(type) {
	case map[string]interface{}:
		return []string{normalizeWebQueryRow(body)}, nil
	case []interface{}:
		lines := make([]string, 0, len(body))
		for _, item := range body {
			row, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("ts3: unexpected webquery row type %T", item)
			}
			lines = append(lines, normalizeWebQueryRow(row))
		}
		return lines, nil
	default:
		return nil, fmt.Errorf("ts3: unexpected webquery body type %T", rawBody)
	}
}

func normalizeWebQueryRow(row map[string]interface{}) string {
	if len(row) == 0 {
		return ""
	}

	keys := make([]string, 0, len(row))
	for k := range row {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+normalizeWebQueryValue(row[k]))
	}
	return strings.Join(parts, " ")
}

func normalizeWebQueryValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return Escape(Unescape(val))
	case bool:
		if val {
			return "1"
		}
		return "0"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case []interface{}:
		items := make([]string, 0, len(val))
		for _, item := range val {
			items = append(items, webQueryScalarString(item))
		}
		return Escape(Unescape(strings.Join(items, ",")))
	default:
		return Escape(Unescape(fmt.Sprint(v)))
	}
}

func webQueryScalarString(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return Unescape(val)
	case bool:
		if val {
			return "1"
		}
		return "0"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func isWebQueryGlobalCommand(cmd string) bool {
	_, ok := webQueryGlobalCommands[strings.ToLower(cmd)]
	return ok
}

func toInt(v interface{}) (int, bool) {
	switch x := v.(type) {
	case float64:
		return int(x), true
	case float32:
		return int(x), true
	case int:
		return x, true
	case int64:
		return int(x), true
	case json.Number:
		n, err := x.Int64()
		if err != nil {
			return 0, false
		}
		return int(n), true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(x))
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}
