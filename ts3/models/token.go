package models

type Token struct {
	Token       string `ts3:"token"`
	Type        int    `ts3:"tokentype"`
	ID1         int    `ts3:"tokenid1"` // GroupID
	ID2         int    `ts3:"tokenid2"` // ChannelID
	Created     int64  `ts3:"token_created"`
	Description string `ts3:"tokendescription"`
}
