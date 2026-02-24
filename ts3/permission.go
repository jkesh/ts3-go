package ts3

import (
	"context"
	"fmt"

	"github.com/jkesh/ts3-go/ts3/models"
)

// --- 频道权限 (Channel Permissions) ---

// ChannelAddPerm 设置或修改频道权限
// cid: 频道 ID
// permName: 权限名称 (建议使用字符串 ID，如 "i_channel_needed_join_power")
// permValue: 权限值
func (c *Client) ChannelAddPerm(ctx context.Context, cid int, permName string, permValue int) error {
	// permsid 表示使用字符串形式的权限名，比数字 ID 更易读且更稳定
	cmd := fmt.Sprintf("channeladdperm cid=%d permsid=%s permvalue=%d", cid, Escape(permName), permValue)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ChannelDelPerm 删除频道权限 (重置为默认)
// cid: 频道 ID
// permName: 权限名称
func (c *Client) ChannelDelPerm(ctx context.Context, cid int, permName string) error {
	cmd := fmt.Sprintf("channeldelperm cid=%d permsid=%s", cid, Escape(permName))
	_, err := c.Exec(ctx, cmd)
	return err
}

// --- 服务器组权限 (Server Group Permissions) ---

// ServerGroupAddPerm 修改服务器组权限
// sgid: 服务器组 ID
// permName: 权限名称
// permValue: 权限值
// negated: 是否忽略 (Negate)，通常设为 false
// skip: 是否跳过频道权限覆盖 (Skip)，通常设为 false
func (c *Client) ServerGroupAddPerm(ctx context.Context, sgid int, permName string, permValue int, negated, skip bool) error {
	n, s := 0, 0
	if negated {
		n = 1
	}
	if skip {
		s = 1
	}

	// TS3 命令格式: servergroupaddperm sgid={sgid} permsid={perm} permvalue={value} permnegated={1|0} permskip={1|0}
	cmd := fmt.Sprintf("servergroupaddperm sgid=%d permsid=%s permvalue=%d permnegated=%d permskip=%d",
		sgid, Escape(permName), permValue, n, s)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ServerGroupDelPerm 删除服务器组权限
// sgid: 服务器组 ID
// permName: 权限名称
func (c *Client) ServerGroupDelPerm(ctx context.Context, sgid int, permName string) error {
	cmd := fmt.Sprintf("servergroupdelperm sgid=%d permsid=%s", sgid, Escape(permName))
	_, err := c.Exec(ctx, cmd)
	return err
}

// --- 频道组权限 (Channel Group Permissions) ---

// ChannelGroupAddPerm 修改频道组权限
// cgid: 频道组 ID
// permName: 权限名称
// permValue: 权限值
func (c *Client) ChannelGroupAddPerm(ctx context.Context, cgid int, permName string, permValue int) error {
	cmd := fmt.Sprintf("channelgroupaddperm cgid=%d permsid=%s permvalue=%d", cgid, Escape(permName), permValue)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ChannelGroupDelPerm 删除频道组权限
// cgid: 频道组 ID
// permName: 权限名称
func (c *Client) ChannelGroupDelPerm(ctx context.Context, cgid int, permName string) error {
	cmd := fmt.Sprintf("channelgroupdelperm cgid=%d permsid=%s", cgid, Escape(permName))
	_, err := c.Exec(ctx, cmd)
	return err
}

// --- 客户端特殊权限 (Client Permissions) ---

// ClientAddPerm 给特定用户(Database ID)添加特殊权限
// cldbid: 用户的数据库 ID (Client Database ID)
// permName: 权限名称
// permValue: 权限值
// skip: 是否跳过 (Skip)
func (c *Client) ClientAddPerm(ctx context.Context, cldbid int, permName string, permValue int, skip bool) error {
	s := 0
	if skip {
		s = 1
	}
	cmd := fmt.Sprintf("clientaddperm cldbid=%d permsid=%s permvalue=%d permskip=%d", cldbid, Escape(permName), permValue, s)
	_, err := c.Exec(ctx, cmd)
	return err
}

// ClientDelPerm 删除特定用户的特殊权限
// cldbid: 用户的数据库 ID
// permName: 权限名称
func (c *Client) ClientDelPerm(ctx context.Context, cldbid int, permName string) error {
	cmd := fmt.Sprintf("clientdelperm cldbid=%d permsid=%s", cldbid, Escape(permName))
	_, err := c.Exec(ctx, cmd)
	return err
}

// --- 权限查询 (Permission Listing) ---

// PermissionList 返回实例上全部权限定义。
func (c *Client) PermissionList(ctx context.Context) ([]models.PermissionEntry, error) {
	resp, err := c.Exec(ctx, "permissionlist")
	if err != nil {
		return nil, err
	}

	var out []models.PermissionEntry
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ServerGroupPermList 返回服务器组权限。
// includePermSID=true 时追加 "-permsid" 选项。
func (c *Client) ServerGroupPermList(ctx context.Context, sgid int, includePermSID bool) ([]models.PermissionEntry, error) {
	cmd := fmt.Sprintf("servergrouppermlist sgid=%d", sgid)
	if includePermSID {
		cmd += " -permsid"
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out []models.PermissionEntry
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelGroupPermList 返回频道组权限。
// cid 可为 0；includePermSID=true 时追加 "-permsid" 选项。
func (c *Client) ChannelGroupPermList(ctx context.Context, cgid int, cid int, includePermSID bool) ([]models.PermissionEntry, error) {
	cmd := fmt.Sprintf("channelgrouppermlist cgid=%d cid=%d", cgid, cid)
	if includePermSID {
		cmd += " -permsid"
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out []models.PermissionEntry
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelPermList 返回频道权限。
func (c *Client) ChannelPermList(ctx context.Context, cid int, includePermSID bool) ([]models.PermissionEntry, error) {
	cmd := fmt.Sprintf("channelpermlist cid=%d", cid)
	if includePermSID {
		cmd += " -permsid"
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out []models.PermissionEntry
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ClientPermList 返回客户端数据库账号的特殊权限。
func (c *Client) ClientPermList(ctx context.Context, cldbid int, includePermSID bool) ([]models.PermissionEntry, error) {
	cmd := fmt.Sprintf("clientpermlist cldbid=%d", cldbid)
	if includePermSID {
		cmd += " -permsid"
	}

	resp, err := c.Exec(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var out []models.PermissionEntry
	if err := NewDecoder().Decode(resp, &out); err != nil {
		return nil, err
	}
	return out, nil
}
