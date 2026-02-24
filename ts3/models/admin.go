package models

// ServerGroupClient is one row from "servergroupclientlist".
type ServerGroupClient struct {
	ServerGroupID    int    `ts3:"sgid"`
	ClientDBID       int    `ts3:"cldbid"`
	Nickname         string `ts3:"name"`
	UniqueIdentifier string `ts3:"cluid"`
}

// ChannelGroup describes one channel group.
type ChannelGroup struct {
	ID                int    `ts3:"cgid"`
	Name              string `ts3:"name"`
	Type              int    `ts3:"type"`
	IconID            int    `ts3:"iconid"`
	SavedB            int    `ts3:"savedb"`
	SortID            int    `ts3:"sortid"`
	NameMode          int    `ts3:"namemode"`
	ModifyPower       int    `ts3:"n_modifyp"`
	MemberAddPower    int    `ts3:"n_member_addp"`
	MemberRemovePower int    `ts3:"n_member_removep"`
}

// ChannelGroupClient is one row from "channelgroupclientlist".
type ChannelGroupClient struct {
	ChannelGroupID    int    `ts3:"cgid"`
	ChannelID         int    `ts3:"cid"`
	ClientDBID        int    `ts3:"cldbid"`
	Nickname          string `ts3:"name"`
	UniqueIdentifier  string `ts3:"cluid"`
	LastNickname      string `ts3:"client_nickname"`
	LastConnectedTime int64  `ts3:"client_lastconnected"`
}

// BanEntry is one row from "banlist".
type BanEntry struct {
	BanID         int    `ts3:"banid"`
	IP            string `ts3:"ip"`
	Name          string `ts3:"name"`
	UID           string `ts3:"uid"`
	Created       int64  `ts3:"created"`
	Duration      int64  `ts3:"duration"`
	InvokerName   string `ts3:"invokername"`
	InvokerUID    string `ts3:"invokeruid"`
	LastNickname  string `ts3:"lastnickname"`
	Reason        string `ts3:"reason"`
	Enforcements  int    `ts3:"enforcements"`
	TargetMode    int    `ts3:"targetmode"`
	Target        string `ts3:"target"`
	TargetNick    string `ts3:"targetnick"`
	TargetUID     string `ts3:"targetuid"`
	TargetIP      string `ts3:"targetip"`
	TargetName    string `ts3:"targetname"`
	Expires       int64  `ts3:"expires"`
	EnforcedTimes int    `ts3:"count"`
}

// ComplainEntry is one row from "complainlist".
type ComplainEntry struct {
	TargetClientDBID int    `ts3:"tcldbid"`
	FromClientDBID   int    `ts3:"fcldbid"`
	TargetName       string `ts3:"tname"`
	FromName         string `ts3:"fname"`
	Message          string `ts3:"message"`
	Timestamp        int64  `ts3:"timestamp"`
}

// ServerTempPassword is one row from "servertemppasswordlist".
type ServerTempPassword struct {
	Password      string `ts3:"pw"`
	Description   string `ts3:"desc"`
	StartTime     int64  `ts3:"start"`
	EndTime       int64  `ts3:"end"`
	TargetChannel int    `ts3:"tcid"`
}

// QueryLoginCredentials is returned by "queryloginadd".
type QueryLoginCredentials struct {
	ClientDBID int    `ts3:"cldbid"`
	ServerID   int    `ts3:"sid"`
	LoginName  string `ts3:"client_login_name"`
	Password   string `ts3:"client_login_password"`
}

// QueryLogin is one row from "queryloginlist".
type QueryLogin struct {
	ClientDBID int    `ts3:"cldbid"`
	ServerID   int    `ts3:"sid"`
	LoginName  string `ts3:"client_login_name"`
	CreatedAt  int64  `ts3:"created_at"`
}

// PermissionEntry is one row from permlist or *permlist commands.
type PermissionEntry struct {
	PermID      int    `ts3:"permid"`
	PermSID     string `ts3:"permsid"`
	PermValue   int    `ts3:"permvalue"`
	PermNegated int    `ts3:"permnegated"`
	PermSkip    int    `ts3:"permskip"`
}
