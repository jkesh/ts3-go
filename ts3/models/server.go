package models

// VirtualServer resp for virtual server
type VirtualServer struct {
	ID            int    `ts3:"virtualserver_id"`
	Port          int    `ts3:"virtualserver_port"`
	Status        string `ts3:"virtualserver_status"` // online, offline ...
	ClientsOnline int    `ts3:"virtualserver_clientsonline"`
	QueryClients  int    `ts3:"virtualserver_queryclientsonline"`
	MaxClients    int    `ts3:"virtualserver_maxclients"`
	Uptime        int64  `ts3:"virtualserver_uptime"`
	Name          string `ts3:"virtualserver_name"`
	AutoStart     int    `ts3:"virtualserver_autostart"`
	MachineID     string `ts3:"virtualserver_machine_id"`
}

// ServerInfo command serverInfo
type ServerInfo struct {
	ID                  int    `ts3:"virtualserver_id"`
	Name                string `ts3:"virtualserver_name"`
	WelcomeMessage      string `ts3:"virtualserver_welcomemessage"`
	MaxClients          int    `ts3:"virtualserver_maxclients"`
	Platform            string `ts3:"virtualserver_platform"`
	Version             string `ts3:"virtualserver_version"`
	Password            string `ts3:"virtualserver_password"`
	Created             int64  `ts3:"virtualserver_created"`
	Uptime              int64  `ts3:"virtualserver_uptime"`
	Hostmessage         string `ts3:"virtualserver_hostmessage"`
	HostmessageMode     int    `ts3:"virtualserver_hostmessage_mode"`
	FileBase            string `ts3:"virtualserver_filebase"`
	DefaultServerGroup  int    `ts3:"virtualserver_default_server_group"`
	DefaultChannelGroup int    `ts3:"virtualserver_default_channel_group"`
	DownloadQuota       int64  `ts3:"virtualserver_download_quota"`
	UploadQuota         int64  `ts3:"virtualserver_upload_quota"`
}
