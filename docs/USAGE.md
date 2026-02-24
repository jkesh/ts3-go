# ts3-go 使用手册与命令示例

本文档给出 `ts3-go` 的实战调用方式，按常见运维任务分组，示例可以直接改参数后使用。

## 1. 连接与会话

### 1.1 TCP 连接（10011）

```go
client, err := ts3.NewClient(ts3.Config{
	Host:            "127.0.0.1",
	Port:            10011,
	Timeout:         10 * time.Second,
	KeepAlivePeriod: 30 * time.Second,
	MaxLineSize:     2 * 1024 * 1024,
})
if err != nil {
	log.Fatal(err)
}
defer client.Close()
```

### 1.2 SSH 连接（10022）

```go
client, err := ts3.NewSSHClient("127.0.0.1", 10022, "serveradmin", "your_password")
if err != nil {
	log.Fatal(err)
}
defer client.Close()
```

### 1.3 登录并选服

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := client.Login(ctx, "serveradmin", "your_password"); err != nil {
	log.Fatal(err)
}

// 方式一：按虚拟服务器 SID
if err := client.Use(ctx, 1); err != nil {
	log.Fatal(err)
}

// 方式二：按语音端口（通常 9987）
if err := client.UseByPort(ctx, 9987); err != nil {
	log.Fatal(err)
}
```

## 2. 基础查询命令

### 2.1 实例与服务器信息

```go
version, _ := client.Version(ctx)
host, _ := client.HostInfo(ctx)
me, _ := client.WhoAmI(ctx)
server, _ := client.ServerInfo(ctx)

log.Printf("version=%s build=%s", version.Version, version.Build)
log.Printf("instance uptime=%d", host.InstanceUptime)
log.Printf("my clid=%d", me.ClientID)
log.Printf("server=%s online=%d", server.Name, server.MaxClients)
```

### 2.2 虚拟服务器列表

```go
servers, err := client.ServerList(ctx, "-uid", "-all")
if err != nil {
	log.Fatal(err)
}
for _, s := range servers {
	log.Printf("sid=%d port=%d name=%s", s.ID, s.Port, s.Name)
}
```

### 2.3 在线客户端

```go
clients, err := client.ClientList(ctx, "-uid", "-groups", "-away", "-times", "-country")
if err != nil {
	log.Fatal(err)
}
for _, c := range clients {
	log.Printf("clid=%d nick=%s uid=%s groups=%v", c.ID, c.Nickname, c.UniqueIdentifier, c.ServerGroups)
}
```

### 2.4 客户端详情与数据库检索

```go
info, err := client.ClientInfo(ctx, 12)
if err != nil {
	log.Fatal(err)
}
log.Printf("clid=%d dbid=%d country=%s", info.ID, info.DatabaseID, info.Country)

rows, err := client.ClientDBFind(ctx, "Alice", "-uid")
if err != nil {
	log.Fatal(err)
}
for _, r := range rows {
	log.Printf("dbid=%d nick=%s uid=%s", r.DatabaseID, r.Nickname, r.UniqueIdentifier)
}
```

### 2.5 用户名 / UID / DBID 互查

```go
dbid, _ := client.ClientGetDBIDFromUID(ctx, "some-uid")
name1, _ := client.ClientGetNameFromUID(ctx, "some-uid")
name2, _ := client.ClientGetNameFromDBID(ctx, dbid)

log.Println(dbid, name1, name2)
```

### 2.6 频道查询

```go
channels, err := client.ChannelList(ctx, "-flags", "-voice", "-limits")
if err != nil {
	log.Fatal(err)
}
for _, ch := range channels {
	log.Printf("cid=%d name=%s clients=%d", ch.ID, ch.Name, ch.TotalClients)
}

ch, err := client.ChannelInfo(ctx, 20)
if err != nil {
	log.Fatal(err)
}
log.Printf("topic=%s", ch.Topic)
```

## 3. 消息与客户端管理

### 3.1 发送消息

```go
_ = client.SendPrivateMessage(ctx, 12, "你好")
_ = client.SendChannelMessage(ctx, 20, "频道公告")
_ = client.Broadcast(ctx, "全服重启通知")
```

### 3.2 Poke / 移动 / 踢出 / 封禁

```go
_ = client.PokeClient(ctx, 12, "请看私信")
_ = client.ClientMove(ctx, 12, 20, "")
_ = client.KickFromChannel(ctx, 12, "请勿刷屏")
_ = client.KickFromServer(ctx, 12, "严重违规")

banID, err := client.BanClient(ctx, 12, 3600, "spam")
if err != nil {
	log.Fatal(err)
}
log.Printf("banID=%d", banID)
```

## 4. 频道管理

### 4.1 创建频道

```go
cid, err := client.ChannelCreate(ctx, ts3.ChannelCreateOptions{
	Name:             "Music",
	Topic:            "24h music",
	Description:      "请文明交流",
	MaxClients:       50,
	MaxFamilyClients: 100,
	CodecQuality:     7,
	IsPermanent:      true,
})
if err != nil {
	log.Fatal(err)
}
log.Printf("new cid=%d", cid)
```

### 4.2 编辑频道

```go
err := client.ChannelEdit(ctx, 20, ts3.ChannelEditOptions{
	Topic:                "夜间安静模式",
	NeededTalkPower:      20,
	NeededSubscribePower: 10,
})
if err != nil {
	log.Fatal(err)
}
```

### 4.3 移动 / 删除频道

```go
_ = client.ChannelMove(ctx, 20, 1, 0)
_ = client.ChannelDelete(ctx, 20, false)
```

### 4.4 频道订阅

```go
_ = client.ChannelSubscribe(ctx, 1, 2, 3)
_ = client.ChannelUnsubscribe(ctx, 2)
_ = client.ChannelSubscribeAll(ctx)
_ = client.ChannelUnsubscribeAll(ctx)
```

## 5. 服务器配置与临时密码

### 5.1 修改服务器参数

```go
err := client.ServerEdit(ctx, ts3.ServerEditOptions{
	Name:            "My TS3",
	WelcomeMessage:  "欢迎来到服务器",
	MaxClients:      128,
	HostMessage:     "请遵守规则",
	HostMessageMode: 2,
})
if err != nil {
	log.Fatal(err)
}
```

### 5.2 临时密码

```go
_ = client.ServerTempPasswordAdd(ctx, ts3.ServerTempPasswordOptions{
	Password:        "temp-123",
	Description:     "活动临时密码",
	DurationSeconds: 3600,
	TargetChannelID: 20,
})

pwList, _ := client.ServerTempPasswordList(ctx)
for _, pw := range pwList {
	log.Printf("pw=%s end=%d", pw.Password, pw.EndTime)
}

_ = client.ServerTempPasswordDelete(ctx, "temp-123")
```

## 6. 组管理

### 6.1 服务器组

```go
sgid, err := client.ServerGroupAdd(ctx, "活动管理员", 1)
if err != nil {
	log.Fatal(err)
}

_ = client.ServerGroupRename(ctx, sgid, "活动管理")
_ = client.ServerGroupAddClient(ctx, sgid, 42)

members, _ := client.ServerGroupClientList(ctx, sgid, "-names")
log.Printf("members=%d", len(members))

copyID, _ := client.ServerGroupCopy(ctx, sgid, "活动管理-备份", 1)
_ = client.ServerGroupDelete(ctx, copyID, true)
```

### 6.2 频道组

```go
cgid, err := client.ChannelGroupAdd(ctx, "临时主持", 1)
if err != nil {
	log.Fatal(err)
}

_ = client.ChannelGroupRename(ctx, cgid, "主持人")
_ = client.SetClientChannelGroup(ctx, cgid, 20, 42)

rows, _ := client.ChannelGroupClientList(ctx, cgid, 20, "-names")
log.Printf("assignments=%d", len(rows))

_ = client.ChannelGroupDelete(ctx, cgid, true)
```

## 7. 权限命令

### 7.1 查询权限

```go
allPerms, _ := client.PermissionList(ctx)
sgPerms, _ := client.ServerGroupPermList(ctx, 6, true)
cgPerms, _ := client.ChannelGroupPermList(ctx, 5, 20, true)
chPerms, _ := client.ChannelPermList(ctx, 20, true)
clPerms, _ := client.ClientPermList(ctx, 42, true)

log.Println(len(allPerms), len(sgPerms), len(cgPerms), len(chPerms), len(clPerms))
```

### 7.2 增删权限

```go
_ = client.ServerGroupAddPerm(ctx, 6, "i_channel_join_power", 75, false, false)
_ = client.ServerGroupDelPerm(ctx, 6, "i_channel_join_power")

_ = client.ChannelAddPerm(ctx, 20, "i_channel_needed_join_power", 50)
_ = client.ChannelDelPerm(ctx, 20, "i_channel_needed_join_power")

_ = client.ChannelGroupAddPerm(ctx, 5, "i_channel_needed_talk_power", 20)
_ = client.ChannelGroupDelPerm(ctx, 5, "i_channel_needed_talk_power")

_ = client.ClientAddPerm(ctx, 42, "i_client_ignore_antiflood", 1, false)
_ = client.ClientDelPerm(ctx, 42, "i_client_ignore_antiflood")
```

## 8. Token 与 Query Login

### 8.1 Token

```go
// tokentype: 0=server group, 1=channel group
token, err := client.TokenAdd(ctx, 0, 6, 0, "活动管理员 token")
if err != nil {
	log.Fatal(err)
}
log.Printf("token=%s", token)

tokens, _ := client.TokenList(ctx)
for _, t := range tokens {
	log.Printf("token=%s desc=%s", t.Token, t.Description)
}

_ = client.TokenDelete(ctx, token)
```

### 8.2 Query Login

```go
cred, err := client.QueryLoginAdd(ctx, 42, 1)
if err != nil {
	log.Fatal(err)
}
log.Printf("query user=%s pass=%s", cred.LoginName, cred.Password)

logins, _ := client.QueryLoginList(ctx)
for _, q := range logins {
	log.Printf("cldbid=%d user=%s", q.ClientDBID, q.LoginName)
}

_ = client.QueryLoginDelete(ctx, 42)
```

## 9. 封禁与举报

### 9.1 封禁管理

```go
bans, _ := client.BanList(ctx)
for _, b := range bans {
	log.Printf("banid=%d reason=%s", b.BanID, b.Reason)
}

_ = client.BanDelete(ctx, 1)
// _ = client.BanDeleteAll(ctx)
```

### 9.2 举报管理

```go
_ = client.ComplainAdd(ctx, 100, 42, "辱骂")
list, _ := client.ComplainList(ctx, 100)
log.Printf("complains=%d", len(list))

_ = client.ComplainDelete(ctx, 100, 42)
// _ = client.ComplainDeleteAll(ctx, 100)
```

## 10. 事件通知

### 10.1 文本消息事件

```go
if err := client.OnTextMessage(ctx, func(payload string) {
	var evt struct {
		Invoker string `ts3:"invokername"`
		Message string `ts3:"msg"`
	}
	_ = ts3.NewDecoder().Decode(payload, &evt)
	log.Printf("text from=%s msg=%s", evt.Invoker, evt.Message)
}); err != nil {
	log.Fatal(err)
}
```

### 10.2 进服事件

```go
if err := client.OnClientEnter(ctx, func(payload string) {
	var evt struct {
		Nickname string `ts3:"client_nickname"`
		ClientID int    `ts3:"clid"`
	}
	_ = ts3.NewDecoder().Decode(payload, &evt)
	log.Printf("join clid=%d nick=%s", evt.ClientID, evt.Nickname)
}); err != nil {
	log.Fatal(err)
}
```

### 10.3 手动注册/取消事件

```go
_ = client.RegisterTextEvents(ctx)
client.Register("notifytextmessage", func(payload string) {
	log.Println(payload)
})

_ = client.UnregisterNotify(ctx)
client.Unregister("notifytextmessage")
```

## 11. 原始命令兜底（Exec）

当库里还没封装某个命令时，直接用 `Exec`：

```go
raw, err := client.Exec(ctx, "channelfind pattern=Music")
if err != nil {
	log.Fatal(err)
}

var rows []struct {
	ID   int    `ts3:"cid"`
	Name string `ts3:"channel_name"`
}
if err := ts3.NewDecoder().Decode(raw, &rows); err != nil {
	log.Fatal(err)
}
```

## 12. 错误处理建议

```go
err := client.KickFromServer(ctx, 12, "test")
if err != nil {
	var qerr *ts3.Error
	if errors.As(err, &qerr) {
		switch {
		case qerr.Is(ts3.ErrPermissions):
			log.Printf("权限不足: %s", qerr.Msg)
		case qerr.Is(ts3.ErrFloodBan):
			log.Printf("触发防洪: %s", qerr.Msg)
		default:
			log.Printf("ts3 error: id=%d msg=%s", qerr.ID, qerr.Msg)
		}
		return
	}
	log.Printf("network/ctx error: %v", err)
}
```

## 13. 最佳实践

- 每次调用都使用 `context.WithTimeout`，避免命令阻塞。
- 同一个 `Client` 内命令是串行语义，避免在同一实例上做高并发管理；并发场景建议多个连接实例。
- 长连接务必开启 `KeepAlivePeriod`，并在退出时 `defer client.Close()`。
- 先查询再变更（如先 `ServerGroupList` 再做 `ServerGroupAddClient`），便于审计和回滚。
