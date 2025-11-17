package ts3

import (
	"context"
	"fmt"
)

// Login 登录 ServerQuery 账号
func (c *Client) Login(ctx context.Context, username, password string) error {
	cmd := fmt.Sprintf("login %s %s", Escape(username), Escape(password))
	_, err := c.Exec(ctx, cmd) // 传入 ctx
	return err
}

// Use 选择虚拟服务器 (通常是端口 9987 的 server id=1)
func (c *Client) Use(ctx context.Context, virtualServerID int) error {
	cmd := fmt.Sprintf("use sid=%d", virtualServerID)
	_, err := c.Exec(ctx, cmd) // 传入 ctx
	return err
}

// Logout 登出 (可选)
func (c *Client) Logout(ctx context.Context) error {
	_, err := c.Exec(ctx, "logout")
	return err
}
