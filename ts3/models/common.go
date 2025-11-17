package models

// Response TS3 Server resp status
type Response struct {
	ID      int    `ts3:"id"`
	Message string `ts3:"msg"`
}

// IsSuccess check successful status
func (r *Response) IsSuccess() bool {
	return r.ID == 0
}

// VersionInfo command version
type VersionInfo struct {
	Version  string `ts3:"version"`
	Build    string `ts3:"build"`
	Platform string `ts3:"platform"`
}

// WhoAmI command whoami
type WhoAmI struct {
	VirtualServerID  int    `ts3:"virtualserver_id"`
	ClientID         int    `ts3:"client_id"`
	ChannelID        int    `ts3:"client_channel_id"`
	Nickname         string `ts3:"client_nickname"`
	DatabaseID       int    `ts3:"client_database_id"`
	LoginName        string `ts3:"client_login_name"`
	UniqueIdentifier string `ts3:"client_unique_identifier"`
	OriginServerID   int    `ts3:"client_origin_server_id"`
}
