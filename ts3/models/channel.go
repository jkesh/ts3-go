package models

// Channel channel info
type Channel struct {
	ID                   int    `ts3:"cid"`
	ParentID             int    `ts3:"pid"`
	Order                int    `ts3:"channel_order"`
	Name                 string `ts3:"channel_name"`
	Topic                string `ts3:"channel_topic"`
	IsDefault            int    `ts3:"channel_flag_default"`
	Password             int    `ts3:"channel_flag_password"`
	Permanent            int    `ts3:"channel_flag_permanent"`
	SemiPermanent        int    `ts3:"channel_flag_semi_permanent"`
	Codec                int    `ts3:"channel_codec"`
	CodecQuality         int    `ts3:"channel_codec_quality"`
	NeededSubscribePower int    `ts3:"channel_needed_subscribe_power"`
	TotalClients         int    `ts3:"total_clients"`
	MaxClients           int    `ts3:"channel_maxclients"`
	FamilyMaxClients     int    `ts3:"channel_maxfamilyclients"`
}
