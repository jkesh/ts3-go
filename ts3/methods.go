package ts3

import (
	"fmt"
	"ts3-go/ts3/models"
)

// ServerInfo 获取当前虚拟服务器详情
func (c *Client) ServerInfo() (*models.ServerInfo, error) {
	resp, err := c.Exec("serverinfo")
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
func (c *Client) ClientList(options ...string) ([]models.OnlineClient, error) {
	cmd := "clientlist"
	for _, opt := range options {
		cmd += " " + opt
	}

	resp, err := c.Exec(cmd)
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
func (c *Client) ChannelList(options ...string) ([]models.Channel, error) {
	cmd := "channellist"
	for _, opt := range options {
		cmd += " " + opt
	}

	resp, err := c.Exec(cmd)
	if err != nil {
		return nil, err
	}

	var channels []models.Channel
	if err := NewDecoder().Decode(resp, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

// PokeClient 戳一下用户
func (c *Client) PokeClient(clientID int, msg string) error {
	cmd := fmt.Sprintf("clientpoke clid=%d msg=%s", clientID, Escape(msg))
	_, err := c.Exec(cmd)
	return err
}

// KickClient 踢出用户
// reasonID: 4=kick form channel, 5=kick from server
func (c *Client) KickClient(clientID int, reasonID int, msg string) error {
	cmd := fmt.Sprintf("clientkick clid=%d reasonid=%d reasonmsg=%s", clientID, reasonID, Escape(msg))
	_, err := c.Exec(cmd)
	return err
}

// SendTextMessage 发送消息
// targetMode: 1=Private, 2=Channel, 3=Server
func (c *Client) SendTextMessage(targetMode int, targetID int, msg string) error {
	cmd := fmt.Sprintf("sendtextmessage targetmode=%d target=%d msg=%s", targetMode, targetID, Escape(msg))
	_, err := c.Exec(cmd)
	return err
}

func (c *Client) ServerGroupAddClient(sgid, cldbid int) error {
	cmd := fmt.Sprintf("servergroupaddclient sgid=%d cldbid=%d", sgid, cldbid)
	_, err := c.Exec(cmd)
	return err
}

// ServerGroupDelClient 将用户从服务器组移除
func (c *Client) ServerGroupDelClient(sgid, cldbid int) error {
	cmd := fmt.Sprintf("servergroupdelclient sgid=%d cldbid=%d", sgid, cldbid)
	_, err := c.Exec(cmd)
	return err
}

// SetClientChannelGroup 设置用户的频道组 (例如设置某人为频道管理员)
// cgid: 频道组 ID
// cid: 频道 ID
// cldbid: 客户端数据库 ID
func (c *Client) SetClientChannelGroup(cgid, cid, cldbid int) error {
	cmd := fmt.Sprintf("setclientchannelgroup cgid=%d cid=%d cldbid=%d", cgid, cid, cldbid)
	_, err := c.Exec(cmd)
	return err
}

// Broadcast 发送全服通告 (Server Message)
func (c *Client) Broadcast(msg string) error {
	// targetmode=3 (Server)
	return c.SendTextMessage(3, 0, msg)
}

// KickFromChannel 将用户踢出频道
func (c *Client) KickFromChannel(clid int, reason string) error {
	return c.KickClient(clid, 4, reason)
}

// KickFromServer 将用户踢出服务器
func (c *Client) KickFromServer(clid int, reason string) error {
	return c.KickClient(clid, 5, reason)
}

// BanClient 封禁用户
// timeInSeconds: 封禁时长 (0 为永久)
func (c *Client) BanClient(clid int, timeInSeconds int64, reason string) (int, error) {
	cmd := fmt.Sprintf("banclient clid=%d time=%d banreason=%s", clid, timeInSeconds, Escape(reason))
	resp, err := c.Exec(cmd)
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
