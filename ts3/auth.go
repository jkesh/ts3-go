package ts3

import "fmt"

// Login 登录 ServerQuery 账号
func (c *Client) Login(username, password string) error {
	cmd := fmt.Sprintf("login %s %s", Escape(username), Escape(password))

	_, err := c.Exec(cmd)
	return err
}

// Use 选择虚拟服务器 (通常是端口 9987 的 server id=1)
func (c *Client) Use(virtualServerID int) error {
	cmd := fmt.Sprintf("use sid=%d", virtualServerID)
	_, err := c.Exec(cmd)
	return err
}
