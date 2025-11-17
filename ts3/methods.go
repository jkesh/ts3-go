package ts3

import (
	"context"
	"fmt"
	"github.com/jkesh/ts3-go/ts3/models"
)

// ServerInfo 获取当前虚拟服务器详情
func (c *Client) ServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	resp, err := c.Exec(ctx, "serverinfo")
	if err != nil {
		return nil, err
	}

	var info models.ServerInfo
	if err := NewDecoder().Decode(resp, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ClientList 获取在线用户列表
// options: 可选参数，如 "-uid", "-away", "-voice", "-groups"
func (c *Client) ClientList(ctx context.Context, options ...string) ([]models.OnlineClient, error) {
	cmd := "clientlist"
	for _, opt := range options {
		cmd += " " + opt
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var clients []models.OnlineClient
	if err := NewDecoder().Decode(resp, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

// ChannelList 获取频道列表
func (c *Client) ChannelList(ctx context.Context, options ...string) ([]models.Channel, error) {
	cmd := "channellist"
	for _, opt := range options {
		cmd += " " + opt
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var channels []models.Channel
	if err := NewDecoder().Decode(resp, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

// PokeClient 戳一下用户 (发送弹窗消息)
func (c *Client) PokeClient(ctx context.Context, clientID int, msg string) error {
	cmd := fmt.Sprintf("clientpoke clid=%d msg=%s", clientID, Escape(msg))
	_, err := c.Exec(ctx, cmd)
	return err
}

// KickClient 踢出用户
// reasonID: 4=kick from channel, 5=kick from server
func (c *Client) KickClient(ctx context.Context, clientID int, reasonID int, msg string) error {
	cmd := fmt.Sprintf("clientkick clid=%d reasonid=%d reasonmsg=%s", clientID, reasonID, Escape(msg))
	_, err := c.Exec(ctx, cmd)
	return err
}

// SendTextMessage 发送文本消息
// targetMode: 1=Private, 2=Channel, 3=Server
func (c *Client) SendTextMessage(ctx context.Context, targetMode int, targetID int, msg string) error {
	cmd := fmt.Sprintf("sendtextmessage targetmode=%d target=%d msg=%s", targetMode, targetID, Escape(msg))
	_, err := c.Exec(ctx, cmd)
	return err
}

// ServerGroupAddClient 将用户添加到服务器组 (例如添加 Admin)
// sgid: Server Group ID
// cldbid: Client Database ID
func (c *Client) ServerGroupAddClient(ctx context.Context, sgid, cldbid int) error {
	cmd := fmt.Sprintf("servergroupaddclient sgid=%d cldbid=%d", sgid, cldbid)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ServerGroupDelClient 将用户从服务器组移除
func (c *Client) ServerGroupDelClient(ctx context.Context, sgid, cldbid int) error {
	cmd := fmt.Sprintf("servergroupdelclient sgid=%d cldbid=%d", sgid, cldbid)
	_, err := c.Exec(ctx, cmd)
	return err
}

// SetClientChannelGroup 设置用户的频道组 (例如设置某人为频道管理员)
// cgid: Channel Group ID
// cid: Channel ID
// cldbid: Client Database ID
func (c *Client) SetClientChannelGroup(ctx context.Context, cgid, cid, cldbid int) error {
	cmd := fmt.Sprintf("setclientchannelgroup cgid=%d cid=%d cldbid=%d", cgid, cid, cldbid)
	_, err := c.Exec(ctx, cmd)
	return err
}

// Broadcast 发送全服通告 (Server Message)
func (c *Client) Broadcast(ctx context.Context, msg string) error {
	// targetmode=3 (Server)
	return c.SendTextMessage(ctx, 3, 0, msg)
}

// KickFromChannel 将用户踢出频道
func (c *Client) KickFromChannel(ctx context.Context, clid int, reason string) error {
	return c.KickClient(ctx, clid, 4, reason)
}

// KickFromServer 将用户踢出服务器
func (c *Client) KickFromServer(ctx context.Context, clid int, reason string) error {
	return c.KickClient(ctx, clid, 5, reason)
}

// BanClient 封禁用户
// timeInSeconds: 封禁时长 (0 为永久)
func (c *Client) BanClient(ctx context.Context, clid int, timeInSeconds int64, reason string) (int, error) {
	cmd := fmt.Sprintf("banclient clid=%d time=%d banreason=%s", clid, timeInSeconds, Escape(reason))
	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return 0, err
	}

	// 返回 banid
	var res struct {
		BanID int `ts3:"banid"`
	}
	if err := NewDecoder().Decode(resp, &res); err != nil {
		return 0, err
	}
	return res.BanID, nil
}

// --- Token / Privilege Key Management ---

// TokenAdd 创建一个新的权限密钥 (Privilege Key)
// tokenType: 0 = Server Group, 1 = Channel Group
// id1: Group ID
// id2: Channel ID (如果是 Server Group 则填 0)
// description: 备注信息
func (c *Client) TokenAdd(ctx context.Context, tokenType, id1, id2 int, description string) (string, error) {
	cmd := fmt.Sprintf("tokenadd tokentype=%d tokenid1=%d tokenid2=%d tokendescription=%s",
		tokenType, id1, id2, Escape(description))

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return "", err
	}

	// 解析返回的 token 字符串
	var res struct {
		Token string `ts3:"token"`
	}
	if err := NewDecoder().Decode(resp, &res); err != nil {
		return "", err
	}
	return res.Token, nil
}

// TokenList 列出所有可用的(未使用的)权限密钥
func (c *Client) TokenList(ctx context.Context) ([]models.Token, error) {
	resp, err := c.Exec(ctx, "tokenlist")
	if err != nil {
		return nil, err
	}

	var tokens []models.Token
	if err := NewDecoder().Decode(resp, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// TokenDelete 删除一个权限密钥
func (c *Client) TokenDelete(ctx context.Context, token string) error {
	cmd := fmt.Sprintf("tokendelete token=%s", Escape(token))
	_, err := c.Exec(ctx, cmd)
	return err
}
