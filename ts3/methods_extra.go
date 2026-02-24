package ts3

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jkesh/ts3-go/ts3/models"
)

// ServerEditOptions contains optional fields for "serveredit".
type ServerEditOptions struct {
	Name                        string
	WelcomeMessage              string
	Password                    string
	MaxClients                  int
	HostMessage                 string
	HostMessageMode             int
	DefaultServerGroup          int
	DefaultChannelGroup         int
	NeededIdentitySecurityLevel int
	MinClientVersion            int64
}

// ServerEdit updates settings of the currently selected virtual server.
func (c *Client) ServerEdit(ctx context.Context, opt ServerEditOptions) error {
	parts := make([]string, 0, 10)
	if opt.Name != "" {
		parts = append(parts, "virtualserver_name="+Escape(opt.Name))
	}
	if opt.WelcomeMessage != "" {
		parts = append(parts, "virtualserver_welcomemessage="+Escape(opt.WelcomeMessage))
	}
	if opt.Password != "" {
		parts = append(parts, "virtualserver_password="+Escape(opt.Password))
	}
	if opt.MaxClients > 0 {
		parts = append(parts, "virtualserver_maxclients="+strconv.Itoa(opt.MaxClients))
	}
	if opt.HostMessage != "" {
		parts = append(parts, "virtualserver_hostmessage="+Escape(opt.HostMessage))
	}
	if opt.HostMessageMode > 0 {
		parts = append(parts, "virtualserver_hostmessage_mode="+strconv.Itoa(opt.HostMessageMode))
	}
	if opt.DefaultServerGroup > 0 {
		parts = append(parts, "virtualserver_default_server_group="+strconv.Itoa(opt.DefaultServerGroup))
	}
	if opt.DefaultChannelGroup > 0 {
		parts = append(parts, "virtualserver_default_channel_group="+strconv.Itoa(opt.DefaultChannelGroup))
	}
	if opt.NeededIdentitySecurityLevel > 0 {
		parts = append(parts, "virtualserver_needed_identity_security_level="+strconv.Itoa(opt.NeededIdentitySecurityLevel))
	}
	if opt.MinClientVersion > 0 {
		parts = append(parts, "virtualserver_min_client_version="+strconv.FormatInt(opt.MinClientVersion, 10))
	}
	if len(parts) == 0 {
		return nil
	}

	_, err := c.Exec(ctx, "serveredit "+strings.Join(parts, " "))
	return err
}

// ChannelEditOptions contains optional fields for "channeledit".
type ChannelEditOptions struct {
	Name                 string
	Topic                string
	Description          string
	Password             string
	Codec                int
	CodecQuality         int
	MaxClients           int
	MaxFamilyClients     int
	NeededTalkPower      int
	NeededSubscribePower int
	IsPermanent          bool
	IsSemiPermanent      bool
	IsDefault            bool
	DeleteDelaySeconds   int
}

// ChannelEdit updates channel properties by channel id.
func (c *Client) ChannelEdit(ctx context.Context, channelID int, opt ChannelEditOptions) error {
	parts := make([]string, 0, 16)
	parts = append(parts, "cid="+strconv.Itoa(channelID))

	if opt.Name != "" {
		parts = append(parts, "channel_name="+Escape(opt.Name))
	}
	if opt.Topic != "" {
		parts = append(parts, "channel_topic="+Escape(opt.Topic))
	}
	if opt.Description != "" {
		parts = append(parts, "channel_description="+Escape(opt.Description))
	}
	if opt.Password != "" {
		parts = append(parts, "channel_password="+Escape(opt.Password))
	}
	if opt.Codec != 0 {
		parts = append(parts, "channel_codec="+strconv.Itoa(opt.Codec))
	}
	if opt.CodecQuality != 0 {
		parts = append(parts, "channel_codec_quality="+strconv.Itoa(opt.CodecQuality))
	}
	if opt.MaxClients > 0 {
		parts = append(parts, "channel_maxclients="+strconv.Itoa(opt.MaxClients))
	}
	if opt.MaxFamilyClients > 0 {
		parts = append(parts, "channel_maxfamilyclients="+strconv.Itoa(opt.MaxFamilyClients))
	}
	if opt.NeededTalkPower > 0 {
		parts = append(parts, "channel_needed_talk_power="+strconv.Itoa(opt.NeededTalkPower))
	}
	if opt.NeededSubscribePower > 0 {
		parts = append(parts, "channel_needed_subscribe_power="+strconv.Itoa(opt.NeededSubscribePower))
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
	if opt.DeleteDelaySeconds > 0 {
		parts = append(parts, "channel_delete_delay="+strconv.Itoa(opt.DeleteDelaySeconds))
	}

	_, err := c.Exec(ctx, "channeledit "+strings.Join(parts, " "))
	return err
}

// ServerTempPasswordOptions holds options for "servertemppasswordadd".
type ServerTempPasswordOptions struct {
	Password              string
	Description           string
	DurationSeconds       int64
	TargetChannelID       int
	TargetChannelPassword string
}

// ServerTempPasswordAdd creates a temporary server password.
func (c *Client) ServerTempPasswordAdd(ctx context.Context, opt ServerTempPasswordOptions) error {
	if strings.TrimSpace(opt.Password) == "" {
		return fmt.Errorf("ts3: temp password is required")
	}
	if opt.DurationSeconds <= 0 {
		return fmt.Errorf("ts3: duration must be > 0")
	}

	parts := []string{
		"pw=" + Escape(opt.Password),
		"desc=" + Escape(opt.Description),
		"duration=" + strconv.FormatInt(opt.DurationSeconds, 10),
	}
	if opt.TargetChannelID > 0 {
		parts = append(parts, "tcid="+strconv.Itoa(opt.TargetChannelID))
	}
	if opt.TargetChannelPassword != "" {
		parts = append(parts, "tcpw="+Escape(opt.TargetChannelPassword))
	}

	_, err := c.Exec(ctx, "servertemppasswordadd "+strings.Join(parts, " "))
	return err
}

// ServerTempPasswordList returns active temporary server passwords.
func (c *Client) ServerTempPasswordList(ctx context.Context) ([]models.ServerTempPassword, error) {
	resp, err := c.Exec(ctx, "servertemppasswordlist")
	if err != nil {
		return nil, err
	}

	var entries []models.ServerTempPassword
	if err := NewDecoder().Decode(resp, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// ServerTempPasswordDelete deletes a temporary server password by value.
func (c *Client) ServerTempPasswordDelete(ctx context.Context, password string) error {
	_, err := c.Exec(ctx, "servertemppassworddel pw="+Escape(password))
	return err
}

// QueryLoginAdd creates a server query login from an existing client database id.
func (c *Client) QueryLoginAdd(ctx context.Context, clientDBID int, serverID int) (*models.QueryLoginCredentials, error) {
	cmd := fmt.Sprintf("queryloginadd cldbid=%d", clientDBID)
	if serverID > 0 {
		cmd += fmt.Sprintf(" sid=%d", serverID)
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out models.QueryLoginCredentials
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QueryLoginDelete deletes a query login by client database id.
func (c *Client) QueryLoginDelete(ctx context.Context, clientDBID int) error {
	_, err := c.Exec(ctx, fmt.Sprintf("querylogindel cldbid=%d", clientDBID))
	return err
}

// QueryLoginList returns all query logins for the current scope.
func (c *Client) QueryLoginList(ctx context.Context) ([]models.QueryLogin, error) {
	resp, err := c.Exec(ctx, "queryloginlist")
	if err != nil {
		return nil, err
	}

	var out []models.QueryLogin
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BanList returns current ban list.
func (c *Client) BanList(ctx context.Context) ([]models.BanEntry, error) {
	resp, err := c.Exec(ctx, "banlist")
	if err != nil {
		return nil, err
	}

	var bans []models.BanEntry
	if err := NewDecoder().Decode(resp, &bans); err != nil {
		return nil, err
	}
	return bans, nil
}

// BanDelete removes a ban by ban id.
func (c *Client) BanDelete(ctx context.Context, banID int) error {
	_, err := c.Exec(ctx, fmt.Sprintf("bandel banid=%d", banID))
	return err
}

// BanDeleteAll removes all bans.
func (c *Client) BanDeleteAll(ctx context.Context) error {
	_, err := c.Exec(ctx, "bandelall")
	return err
}

// ComplainList returns complaints for a specific client DBID.
func (c *Client) ComplainList(ctx context.Context, targetClientDBID int) ([]models.ComplainEntry, error) {
	resp, err := c.Exec(ctx, fmt.Sprintf("complainlist tcldbid=%d", targetClientDBID))
	if err != nil {
		return nil, err
	}

	var entries []models.ComplainEntry
	if err := NewDecoder().Decode(resp, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// ComplainAdd adds a complaint against a client.
func (c *Client) ComplainAdd(ctx context.Context, targetClientDBID int, fromClientDBID int, message string) error {
	cmd := fmt.Sprintf(
		"complainadd tcldbid=%d fcldbid=%d message=%s",
		targetClientDBID,
		fromClientDBID,
		Escape(message),
	)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ComplainDelete deletes a complaint between two client DBIDs.
func (c *Client) ComplainDelete(ctx context.Context, targetClientDBID int, fromClientDBID int) error {
	cmd := fmt.Sprintf("complaindel tcldbid=%d fcldbid=%d", targetClientDBID, fromClientDBID)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ComplainDeleteAll removes all complaints for a target client DBID.
func (c *Client) ComplainDeleteAll(ctx context.Context, targetClientDBID int) error {
	_, err := c.Exec(ctx, fmt.Sprintf("complaindelall tcldbid=%d", targetClientDBID))
	return err
}

// ChannelSubscribe subscribes current query client to one or more channels.
func (c *Client) ChannelSubscribe(ctx context.Context, channelIDs ...int) error {
	if len(channelIDs) == 0 {
		return nil
	}
	parts := make([]string, 0, len(channelIDs))
	for _, cid := range channelIDs {
		if cid <= 0 {
			continue
		}
		parts = append(parts, "cid="+strconv.Itoa(cid))
	}
	if len(parts) == 0 {
		return nil
	}

	_, err := c.Exec(ctx, "channelsubscribe "+strings.Join(parts, "|"))
	return err
}

// ChannelUnsubscribe unsubscribes current query client from channels.
func (c *Client) ChannelUnsubscribe(ctx context.Context, channelIDs ...int) error {
	if len(channelIDs) == 0 {
		return nil
	}
	parts := make([]string, 0, len(channelIDs))
	for _, cid := range channelIDs {
		if cid <= 0 {
			continue
		}
		parts = append(parts, "cid="+strconv.Itoa(cid))
	}
	if len(parts) == 0 {
		return nil
	}

	_, err := c.Exec(ctx, "channelunsubscribe "+strings.Join(parts, "|"))
	return err
}

// ChannelSubscribeAll subscribes query client to all channels.
func (c *Client) ChannelSubscribeAll(ctx context.Context) error {
	_, err := c.Exec(ctx, "channelsubscribeall")
	return err
}

// ChannelUnsubscribeAll unsubscribes query client from all channels.
func (c *Client) ChannelUnsubscribeAll(ctx context.Context) error {
	_, err := c.Exec(ctx, "channelunsubscribeall")
	return err
}

// ServerGroupAdd creates a new server group and returns new group id.
func (c *Client) ServerGroupAdd(ctx context.Context, name string, groupType int) (int, error) {
	if strings.TrimSpace(name) == "" {
		return 0, fmt.Errorf("ts3: server group name is required")
	}

	cmd := "servergroupadd name=" + Escape(name)
	if groupType >= 0 {
		cmd += " type=" + strconv.Itoa(groupType)
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return 0, err
	}

	var out struct {
		GroupID int `ts3:"sgid"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return 0, err
	}
	return out.GroupID, nil
}

// ServerGroupDelete deletes a server group.
func (c *Client) ServerGroupDelete(ctx context.Context, sgid int, force bool) error {
	forceInt := 0
	if force {
		forceInt = 1
	}
	_, err := c.Exec(ctx, fmt.Sprintf("servergroupdel sgid=%d force=%d", sgid, forceInt))
	return err
}

// ServerGroupRename renames a server group.
func (c *Client) ServerGroupRename(ctx context.Context, sgid int, newName string) error {
	_, err := c.Exec(ctx, fmt.Sprintf("servergrouprename sgid=%d name=%s", sgid, Escape(newName)))
	return err
}

// ServerGroupCopy copies an existing server group and returns new group id.
func (c *Client) ServerGroupCopy(ctx context.Context, sourceGroupID int, newName string, groupType int) (int, error) {
	cmd := fmt.Sprintf("servergroupcopy ssgid=%d tsgid=0 name=%s", sourceGroupID, Escape(newName))
	if groupType >= 0 {
		cmd += fmt.Sprintf(" type=%d", groupType)
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return 0, err
	}

	var out struct {
		GroupID int `ts3:"sgid"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return 0, err
	}
	return out.GroupID, nil
}

// ServerGroupClientList returns clients in a server group.
//
// Use options like "-names" to include resolved names/uids.
func (c *Client) ServerGroupClientList(ctx context.Context, sgid int, options ...string) ([]models.ServerGroupClient, error) {
	cmd := withOptions(fmt.Sprintf("servergroupclientlist sgid=%d", sgid), options)
	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out []models.ServerGroupClient
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelGroupList returns channel groups.
func (c *Client) ChannelGroupList(ctx context.Context) ([]models.ChannelGroup, error) {
	resp, err := c.Exec(ctx, "channelgrouplist")
	if err != nil {
		return nil, err
	}

	var out []models.ChannelGroup
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelGroupAdd creates a channel group and returns new id.
func (c *Client) ChannelGroupAdd(ctx context.Context, name string, groupType int) (int, error) {
	if strings.TrimSpace(name) == "" {
		return 0, fmt.Errorf("ts3: channel group name is required")
	}

	cmd := "channelgroupadd name=" + Escape(name)
	if groupType >= 0 {
		cmd += " type=" + strconv.Itoa(groupType)
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return 0, err
	}

	var out struct {
		GroupID int `ts3:"cgid"`
	}
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return 0, err
	}
	return out.GroupID, nil
}

// ChannelGroupDelete deletes a channel group.
func (c *Client) ChannelGroupDelete(ctx context.Context, cgid int, force bool) error {
	forceInt := 0
	if force {
		forceInt = 1
	}
	_, err := c.Exec(ctx, fmt.Sprintf("channelgroupdel cgid=%d force=%d", cgid, forceInt))
	return err
}

// ChannelGroupRename renames a channel group.
func (c *Client) ChannelGroupRename(ctx context.Context, cgid int, newName string) error {
	_, err := c.Exec(ctx, fmt.Sprintf("channelgrouprename cgid=%d name=%s", cgid, Escape(newName)))
	return err
}

// ChannelGroupClientList returns assignments for a channel group in one channel.
func (c *Client) ChannelGroupClientList(ctx context.Context, cgid int, channelID int, options ...string) ([]models.ChannelGroupClient, error) {
	cmd := withOptions(fmt.Sprintf("channelgroupclientlist cgid=%d cid=%d", cgid, channelID), options)
	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out []models.ChannelGroupClient
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}
