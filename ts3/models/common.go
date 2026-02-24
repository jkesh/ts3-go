package models

// Response represents the trailing "error id=... msg=..." line fields.
type Response struct {
	ID      int    `ts3:"id"`
	Message string `ts3:"msg"`
}

// IsSuccess returns true when the command succeeded.
func (r *Response) IsSuccess() bool {
	return r.ID == 0
}

// VersionInfo is returned by the "version" command.
type VersionInfo struct {
	Version  string `ts3:"version"`
	Build    string `ts3:"build"`
	Platform string `ts3:"platform"`
}

// HostInfo is returned by the "hostinfo" command.
type HostInfo struct {
	InstanceUptime int64 `ts3:"instance_uptime"`
	HostTimestamp  int64 `ts3:"host_timestamp_utc"`
	VirtualServers int   `ts3:"virtualservers_running_total"`
	ChannelsOnline int   `ts3:"virtualservers_total_channels_online"`
	ClientsOnline  int   `ts3:"virtualservers_total_clients_online"`
	QueriesOnline  int   `ts3:"virtualservers_total_query_clients_online"`
}

// WhoAmI is returned by the "whoami" command.
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
