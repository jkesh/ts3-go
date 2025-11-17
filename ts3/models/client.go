package models

// OnlineClient client online
type OnlineClient struct {
	ID                int    `ts3:"clid"`               // client id
	ChannelID         int    `ts3:"cid"`                // channel id
	DatabaseID        int    `ts3:"client_database_id"` // database id
	Nickname          string `ts3:"client_nickname"`    // client nickname
	Type              int    `ts3:"client_type"`        // 0: normal client 1:query client
	Away              int    `ts3:"client_away"`        // 0 = online, 1 = AFK
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
	UniqueIdentifier  string `ts3:"client_unique_identifier"` // need -uid
	ServerGroups      []int  `ts3:"client_servergroups"`      // need -groups
	ChannelGroupID    int    `ts3:"client_channel_group_id"`
}

// DBClient client from database
type DBClient struct {
	DatabaseID       int    `ts3:"cldbid"`
	UniqueIdentifier string `ts3:"client_unique_identifier"`
	Nickname         string `ts3:"client_nickname"`
	Created          int64  `ts3:"client_created"`       // Unix Timestamp
	LastConnected    int64  `ts3:"client_lastconnected"` // Unix Timestamp
	TotalConnections int    `ts3:"client_totalconnections"`
}
