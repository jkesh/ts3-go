package ts3

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// Login authenticates with ServerQuery credentials.
func (c *Client) Login(ctx context.Context, username, password string) error {
	if c.isWebQuery() {
		if c.web == nil || strings.TrimSpace(c.web.apiKey) == "" {
			return errors.New("ts3: webquery api key is required")
		}
		return nil
	}

	cmd := fmt.Sprintf(
		"login client_login_name=%s client_login_password=%s",
		Escape(username),
		Escape(password),
	)
	_, err := c.Exec(ctx, cmd)
	return err
}

// Use selects the target virtual server by server id.
func (c *Client) Use(ctx context.Context, virtualServerID int) error {
	if c.isWebQuery() {
		c.setSelectedSID(virtualServerID)
		return nil
	}

	cmd := fmt.Sprintf("use sid=%d", virtualServerID)
	_, err := c.Exec(ctx, cmd)
	return err
}

// UseByPort selects the target virtual server by voice port (e.g. 9987).
func (c *Client) UseByPort(ctx context.Context, port int) error {
	if c.isWebQuery() {
		resp, err := c.Exec(ctx, fmt.Sprintf("serveridgetbyport virtualserver_port=%d", port))
		if err != nil {
			return err
		}

		var out struct {
			ServerID  int `ts3:"server_id"`
			SID       int `ts3:"sid"`
			ServerID2 int `ts3:"virtualserver_id"`
		}
		if err := NewDecoder().Decode(resp, &out); err != nil {
			return err
		}

		sid := out.ServerID
		if sid == 0 {
			sid = out.SID
		}
		if sid == 0 {
			sid = out.ServerID2
		}
		if sid <= 0 {
			return errors.New("ts3: serveridgetbyport returned empty sid")
		}

		c.setSelectedSID(sid)
		return nil
	}

	cmd := fmt.Sprintf("use port=%d", port)
	_, err := c.Exec(ctx, cmd)
	return err
}

// Logout logs out the current ServerQuery session.
func (c *Client) Logout(ctx context.Context) error {
	if c.isWebQuery() {
		return nil
	}

	_, err := c.Exec(ctx, "logout")
	return err
}
