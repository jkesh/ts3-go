package models

type ServerGroup struct {
	ID                int    `ts3:"sgid"`
	Name              string `ts3:"name"`
	Type              int    `ts3:"type"` // 0=Template, 1=Regular, 2=Query
	IconID            int    `ts3:"iconid"`
	SavedB            int    `ts3:"savedb"`
	SortID            int    `ts3:"sortid"`
	NameMode          int    `ts3:"namemode"`
	ModifyPower       int    `ts3:"n_modifyp"`
	MemberAddPower    int    `ts3:"n_member_addp"`
	MemberRemovePower int    `ts3:"n_member_removep"`
}
