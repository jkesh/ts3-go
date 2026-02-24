package models

// OnlineClient is one row from "clientlist".
type OnlineClient struct {
	ID                int    `ts3:"clid"`
	ChannelID         int    `ts3:"cid"`
	DatabaseID        int    `ts3:"client_database_id"`
	Nickname          string `ts3:"client_nickname"`
	Type              int    `ts3:"client_type"` // 0=voice client, 1=server query client
	Away              int    `ts3:"client_away"`
	AwayMessage       string `ts3:"client_away_message"`
	InputMuted        int    `ts3:"client_input_muted"`
	OutputMuted       int    `ts3:"client_output_muted"`
	OutputOnlyMuted   int    `ts3:"client_outputonly_muted"`
	InputHardware     int    `ts3:"client_input_hardware"`
	OutputHardware    int    `ts3:"client_output_hardware"`
	TalkPower         int    `ts3:"client_talk_power"`
	IsTalker          int    `ts3:"client_is_talker"`
	IsPrioritySpeaker int    `ts3:"client_is_priority_speaker"`
	IsRecording       int    `ts3:"client_is_recording"`
	UniqueIdentifier  string `ts3:"client_unique_identifier"` // requires -uid in command options
	ServerGroups      []int  `ts3:"client_servergroups"`      // requires -groups in command options
	ChannelGroupID    int    `ts3:"client_channel_group_id"`
}

// ClientInfo is returned by "clientinfo".
type ClientInfo struct {
	ID               int    `ts3:"clid"`
	ChannelID        int    `ts3:"cid"`
	DatabaseID       int    `ts3:"client_database_id"`
	Nickname         string `ts3:"client_nickname"`
	Type             int    `ts3:"client_type"`
	UniqueIdentifier string `ts3:"client_unique_identifier"`
	Created          int64  `ts3:"client_created"`
	LastConnected    int64  `ts3:"client_lastconnected"`
	Connections      int    `ts3:"client_totalconnections"`
	Country          string `ts3:"client_country"`
	IdleTimeMS       int64  `ts3:"client_idle_time"`
	Platform         string `ts3:"client_platform"`
	Version          string `ts3:"client_version"`
	InputMuted       int    `ts3:"client_input_muted"`
	OutputMuted      int    `ts3:"client_output_muted"`
	TalkPower        int    `ts3:"client_talk_power"`
	ServerGroups     []int  `ts3:"client_servergroups"`
	ChannelGroupID   int    `ts3:"client_channel_group_id"`
}

// DBClient is one row from "clientdblist".
type DBClient struct {
	DatabaseID       int    `ts3:"cldbid"`
	UniqueIdentifier string `ts3:"client_unique_identifier"`
	Nickname         string `ts3:"client_nickname"`
	Created          int64  `ts3:"client_created"`       // unix timestamp
	LastConnected    int64  `ts3:"client_lastconnected"` // unix timestamp
	TotalConnections int    `ts3:"client_totalconnections"`
}
