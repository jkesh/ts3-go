package ts3

import (
	"context"
	"fmt"
)

// Login authenticates with ServerQuery credentials.
func (c *Client) Login(ctx context.Context, username, password string) error {
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
	cmd := fmt.Sprintf("use sid=%d", virtualServerID)
	_, err := c.Exec(ctx, cmd)
	return err
}

// UseByPort selects the target virtual server by voice port (e.g. 9987).
func (c *Client) UseByPort(ctx context.Context, port int) error {
	cmd := fmt.Sprintf("use port=%d", port)
	_, err := c.Exec(ctx, cmd)
	return err
}

// Logout logs out the current ServerQuery session.
func (c *Client) Logout(ctx context.Context) error {
	_, err := c.Exec(ctx, "logout")
	return err
}
