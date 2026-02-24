package ts3

import (
	"context"
	"fmt"
	"strings"

	"github.com/jkesh/ts3-go/ts3/models"
)

const (
	// Message target modes for sendtextmessage.
	TextTargetClient  = 1
	TextTargetChannel = 2
	TextTargetServer  = 3

	// Kick reason ids for clientkick.
	KickReasonChannel = 4
	KickReasonServer  = 5
)

func withOptions(base string, options []string) string {
	if len(options) == 0 {
		return base
	}
	return base + " " + strings.Join(options, " ")
}

// Version returns server query version/build/platform information.
func (c *Client) Version(ctx context.Context) (*models.VersionInfo, error) {
	resp, err := c.Exec(ctx, "version")
	if err != nil {
		return nil, err
	}

	var info models.VersionInfo
	if err := NewDecoder().Decode(resp, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// HostInfo returns instance-level host information.
func (c *Client) HostInfo(ctx context.Context) (*models.HostInfo, error) {
	resp, err := c.Exec(ctx, "hostinfo")
	if err != nil {
		return nil, err
	}

	var info models.HostInfo
	if err := NewDecoder().Decode(resp, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// WhoAmI returns information about the current query session.
func (c *Client) WhoAmI(ctx context.Context) (*models.WhoAmI, error) {
	resp, err := c.Exec(ctx, "whoami")
	if err != nil {
		return nil, err
	}

	var me models.WhoAmI
	if err := NewDecoder().Decode(resp, &me); err != nil {
		return nil, err
	}
	return &me, nil
}

// ServerList returns virtual servers on this instance.
func (c *Client) ServerList(ctx context.Context, options ...string) ([]models.VirtualServer, error) {
	resp, err := c.Exec(ctx, withOptions("serverlist", options))
	if err != nil {
		return nil, err
	}

	var servers []models.VirtualServer
	if err := NewDecoder().Decode(resp, &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

// ServerInfo returns details for the currently selected virtual server.
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

// ServerGroupList returns server groups in the selected virtual server.
func (c *Client) ServerGroupList(ctx context.Context) ([]models.ServerGroup, error) {
	resp, err := c.Exec(ctx, "servergrouplist")
	if err != nil {
		return nil, err
	}

	var groups []models.ServerGroup
	if err := NewDecoder().Decode(resp, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// ClientList returns online clients.
//
// options can include official command switches, such as "-uid", "-away",
// "-voice", "-groups", "-times", "-country" and so on.
func (c *Client) ClientList(ctx context.Context, options ...string) ([]models.OnlineClient, error) {
	resp, err := c.Exec(ctx, withOptions("clientlist", options))
	if err != nil {
		return nil, err
	}

	var clients []models.OnlineClient
	if err := NewDecoder().Decode(resp, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

// ClientInfo returns details for one connected client.
func (c *Client) ClientInfo(ctx context.Context, clientID int) (*models.ClientInfo, error) {
	resp, err := c.Exec(ctx, fmt.Sprintf("clientinfo clid=%d", clientID))
	if err != nil {
		return nil, err
	}

	var info models.ClientInfo
	if err := NewDecoder().Decode(resp, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// ClientDBList returns clients from the database.
//
// start and duration follow ServerQuery defaults:
//   - start: first row offset
//   - duration: max rows to return
func (c *Client) ClientDBList(ctx context.Context, start, duration int, options ...string) ([]models.DBClient, error) {
	cmd := fmt.Sprintf("clientdblist start=%d duration=%d", start, duration)
	cmd = withOptions(cmd, options)

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var clients []models.DBClient
	if err := NewDecoder().Decode(resp, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

// ClientDBFind finds client database entries by nickname pattern.
func (c *Client) ClientDBFind(ctx context.Context, pattern string, options ...string) ([]models.DBClient, error) {
	cmd := withOptions(fmt.Sprintf("clientdbfind pattern=%s", Escape(pattern)), options)
	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var clients []models.DBClient
	if err := NewDecoder().Decode(resp, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}

// ClientGetDBIDFromUID resolves a client database id by unique identifier.
func (c *Client) ClientGetDBIDFromUID(ctx context.Context, uid string) (int, error) {
	resp, err := c.Exec(ctx, fmt.Sprintf("clientgetdbidfromuid cluid=%s", Escape(uid)))
	if err != nil {
		return 0, err
	}

	var out struct {
		DBID int `ts3:"cldbid"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return 0, err
	}
	return out.DBID, nil
}

// ClientGetNameFromDBID resolves a nickname by client database id.
func (c *Client) ClientGetNameFromDBID(ctx context.Context, dbid int) (string, error) {
	resp, err := c.Exec(ctx, fmt.Sprintf("clientgetnamefromdbid cldbid=%d", dbid))
	if err != nil {
		return "", err
	}

	var out struct {
		Name string `ts3:"name"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return "", err
	}
	return out.Name, nil
}

// ClientGetNameFromUID resolves a nickname by unique identifier.
func (c *Client) ClientGetNameFromUID(ctx context.Context, uid string) (string, error) {
	resp, err := c.Exec(ctx, fmt.Sprintf("clientgetnamefromuid cluid=%s", Escape(uid)))
	if err != nil {
		return "", err
	}

	var out struct {
		Name2 string `ts3:"name"`
		Name  string `ts3:"clname"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return "", err
	}
	if out.Name == "" {
		out.Name = out.Name2
	}
	return out.Name, nil
}

// ChannelList returns channels in the selected virtual server.
func (c *Client) ChannelList(ctx context.Context, options ...string) ([]models.Channel, error) {
	resp, err := c.Exec(ctx, withOptions("channellist", options))
	if err != nil {
		return nil, err
	}

	var channels []models.Channel
	if err := NewDecoder().Decode(resp, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

// ChannelInfo returns details for a channel id.
func (c *Client) ChannelInfo(ctx context.Context, channelID int) (*models.Channel, error) {
	resp, err := c.Exec(ctx, fmt.Sprintf("channelinfo cid=%d", channelID))
	if err != nil {
		return nil, err
	}

	var ch models.Channel
	if err := NewDecoder().Decode(resp, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// ChannelCreateOptions defines optional arguments for ChannelCreate.
type ChannelCreateOptions struct {
	Name              string
	Topic             string
	Description       string
	Password          string
	Codec             int
	CodecQuality      int
	MaxClients        int
	MaxFamilyClients  int
	NeededTalkPower   int
	ParentID          int
	Order             int
	IsPermanent       bool
	IsSemiPermanent   bool
	IsDefault         bool
	DeleteDelaySecond int
}

// ChannelCreate creates a channel and returns the new channel id.
func (c *Client) ChannelCreate(ctx context.Context, opt ChannelCreateOptions) (int, error) {
	if strings.TrimSpace(opt.Name) == "" {
		return 0, fmt.Errorf("ts3: channel name is required")
	}

	parts := []string{
		fmt.Sprintf("channel_name=%s", Escape(opt.Name)),
	}
	if opt.Topic != "" {
		parts = append(parts, fmt.Sprintf("channel_topic=%s", Escape(opt.Topic)))
	}
	if opt.Description != "" {
		parts = append(parts, fmt.Sprintf("channel_description=%s", Escape(opt.Description)))
	}
	if opt.Password != "" {
		parts = append(parts, fmt.Sprintf("channel_password=%s", Escape(opt.Password)))
	}
	if opt.Codec != 0 {
		parts = append(parts, fmt.Sprintf("channel_codec=%d", opt.Codec))
	}
	if opt.CodecQuality != 0 {
		parts = append(parts, fmt.Sprintf("channel_codec_quality=%d", opt.CodecQuality))
	}
	if opt.MaxClients != 0 {
		parts = append(parts, fmt.Sprintf("channel_maxclients=%d", opt.MaxClients))
	}
	if opt.MaxFamilyClients != 0 {
		parts = append(parts, fmt.Sprintf("channel_maxfamilyclients=%d", opt.MaxFamilyClients))
	}
	if opt.NeededTalkPower != 0 {
		parts = append(parts, fmt.Sprintf("channel_needed_talk_power=%d", opt.NeededTalkPower))
	}
	if opt.ParentID != 0 {
		parts = append(parts, fmt.Sprintf("cpid=%d", opt.ParentID))
	}
	if opt.Order != 0 {
		parts = append(parts, fmt.Sprintf("channel_order=%d", opt.Order))
	}
	if opt.IsPermanent {
		parts = append(parts, "channel_flag_permanent=1")
	}
	if opt.IsSemiPermanent {
		parts = append(parts, "channel_flag_semi_permanent=1")
	}
	if opt.IsDefault {
		parts = append(parts, "channel_flag_default=1")
	}
	if opt.DeleteDelaySecond > 0 {
		parts = append(parts, fmt.Sprintf("channel_delete_delay=%d", opt.DeleteDelaySecond))
	}

	resp, err := c.Exec(ctx, "channelcreate "+strings.Join(parts, " "))
	if err != nil {
		return 0, err
	}

	var out struct {
		ChannelID int `ts3:"cid"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return 0, err
	}
	return out.ChannelID, nil
}

// ChannelDelete deletes a channel.
func (c *Client) ChannelDelete(ctx context.Context, channelID int, force bool) error {
	forceNum := 0
	if force {
		forceNum = 1
	}

	cmd := fmt.Sprintf("channeldelete cid=%d force=%d", channelID, forceNum)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ChannelMove moves a channel under a parent channel and order.
func (c *Client) ChannelMove(ctx context.Context, channelID, parentID, order int) error {
	cmd := fmt.Sprintf("channelmove cid=%d cpid=%d order=%d", channelID, parentID, order)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ClientMove moves a connected client to another channel.
func (c *Client) ClientMove(ctx context.Context, clientID, channelID int, channelPassword string) error {
	cmd := fmt.Sprintf("clientmove clid=%d cid=%d", clientID, channelID)
	if channelPassword != "" {
		cmd += fmt.Sprintf(" cpw=%s", Escape(channelPassword))
	}
	_, err := c.Exec(ctx, cmd)
	return err
}

// PokeClient sends a poke popup message to a client.
func (c *Client) PokeClient(ctx context.Context, clientID int, msg string) error {
	cmd := fmt.Sprintf("clientpoke clid=%d msg=%s", clientID, Escape(msg))
	_, err := c.Exec(ctx, cmd)
	return err
}

// KickClient kicks a client from channel/server.
func (c *Client) KickClient(ctx context.Context, clientID int, reasonID int, msg string) error {
	cmd := fmt.Sprintf("clientkick clid=%d reasonid=%d reasonmsg=%s", clientID, reasonID, Escape(msg))
	_, err := c.Exec(ctx, cmd)
	return err
}

// SendTextMessage sends a text message.
func (c *Client) SendTextMessage(ctx context.Context, targetMode int, targetID int, msg string) error {
	cmd := fmt.Sprintf("sendtextmessage targetmode=%d target=%d msg=%s", targetMode, targetID, Escape(msg))
	_, err := c.Exec(ctx, cmd)
	return err
}

// SendPrivateMessage sends a private message to one online client.
func (c *Client) SendPrivateMessage(ctx context.Context, clientID int, msg string) error {
	return c.SendTextMessage(ctx, TextTargetClient, clientID, msg)
}

// SendChannelMessage sends a message to a channel.
func (c *Client) SendChannelMessage(ctx context.Context, channelID int, msg string) error {
	return c.SendTextMessage(ctx, TextTargetChannel, channelID, msg)
}

// Broadcast sends a server-wide text message.
func (c *Client) Broadcast(ctx context.Context, msg string) error {
	return c.SendTextMessage(ctx, TextTargetServer, 0, msg)
}

// KickFromChannel kicks a client from the current channel.
func (c *Client) KickFromChannel(ctx context.Context, clientID int, reason string) error {
	return c.KickClient(ctx, clientID, KickReasonChannel, reason)
}

// KickFromServer kicks a client from the virtual server.
func (c *Client) KickFromServer(ctx context.Context, clientID int, reason string) error {
	return c.KickClient(ctx, clientID, KickReasonServer, reason)
}

// BanClient bans a client and returns ban id.
//
// timeInSeconds=0 means permanent ban.
func (c *Client) BanClient(ctx context.Context, clientID int, timeInSeconds int64, reason string) (int, error) {
	cmd := fmt.Sprintf("banclient clid=%d time=%d banreason=%s", clientID, timeInSeconds, Escape(reason))
	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return 0, err
	}

	var out struct {
		BanID int `ts3:"banid"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return 0, err
	}
	return out.BanID, nil
}

// ServerGroupAddClient adds a client database id to a server group.
func (c *Client) ServerGroupAddClient(ctx context.Context, sgid, cldbid int) error {
	cmd := fmt.Sprintf("servergroupaddclient sgid=%d cldbid=%d", sgid, cldbid)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ServerGroupDelClient removes a client database id from a server group.
func (c *Client) ServerGroupDelClient(ctx context.Context, sgid, cldbid int) error {
	cmd := fmt.Sprintf("servergroupdelclient sgid=%d cldbid=%d", sgid, cldbid)
	_, err := c.Exec(ctx, cmd)
	return err
}

// SetClientChannelGroup sets the channel group for a specific client.
func (c *Client) SetClientChannelGroup(ctx context.Context, cgid, cid, cldbid int) error {
	cmd := fmt.Sprintf("setclientchannelgroup cgid=%d cid=%d cldbid=%d", cgid, cid, cldbid)
	_, err := c.Exec(ctx, cmd)
	return err
}

// TokenAdd creates a new privilege key.
//
// tokenType: 0=server group, 1=channel group.
// id1: group id.
// id2: channel id (set 0 for server group token).
func (c *Client) TokenAdd(ctx context.Context, tokenType, id1, id2 int, description string) (string, error) {
	cmd := fmt.Sprintf(
		"tokenadd tokentype=%d tokenid1=%d tokenid2=%d tokendescription=%s",
		tokenType,
		id1,
		id2,
		Escape(description),
	)

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return "", err
	}

	var out struct {
		Token string `ts3:"token"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return "", err
	}
	return out.Token, nil
}

// TokenList returns all currently unused privilege keys.
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

// TokenDelete deletes one privilege key.
func (c *Client) TokenDelete(ctx context.Context, token string) error {
	cmd := fmt.Sprintf("tokendelete token=%s", Escape(token))
	_, err := c.Exec(ctx, cmd)
	return err
}
