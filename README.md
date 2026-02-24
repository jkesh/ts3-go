# ts3-go

`ts3-go` 是一个 TeamSpeak ServerQuery 客户端库，支持：

- 传统 ServerQuery（TCP `10011` / SSH `10022`）
- WebQuery REST（HTTP/HTTPS + `x-api-key`）
- 查询/管理命令封装、统一错误处理

## 安装

```bash
go get github.com/jkesh/ts3-go
```

## 先决条件

- TeamSpeak 6 Server（或兼容 ServerQuery/WebQuery 的版本）已开启查询接口
- 传统 Query 默认端口：TCP `10011`，SSH `10022`
- WebQuery 默认端口：HTTP `10080`，HTTPS `10443`
- 已有可登录的 Query 账号（通常是 `serveradmin`）
- 账号具备目标操作权限（例如封禁、改组、改权限）

## 快速开始（TCP ServerQuery）

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/jkesh/ts3-go/ts3"
)

func main() {
	client, err := ts3.NewClient(ts3.Config{
		Host:            "127.0.0.1",
		Port:            10011,
		Timeout:         10 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Login(ctx, "serveradmin", "your_password"); err != nil {
		log.Fatal(err)
	}
	if err := client.UseByPort(ctx, 9987); err != nil {
		log.Fatal(err)
	}

	info, err := client.ServerInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server=%s max=%d", info.Name, info.MaxClients)
}
```

## SSH 连接示例

```go
client, err := ts3.NewSSHClientWithConfig(
	"127.0.0.1",
	10022,
	"serveradmin",
	"your_password",
	ts3.Config{Timeout: 8 * time.Second, KeepAlivePeriod: 30 * time.Second},
)
if err != nil {
	log.Fatal(err)
}
defer client.Close()
```

## TeamSpeak 6 WebQuery REST 模式

```go
client, err := ts3.NewWebQueryClient(ts3.WebQueryConfig{
	Host:            "127.0.0.1",
	Port:            10080, // HTTPS 默认端口是 10443
	APIKey:          "your_api_key",
	VirtualServerID: 1, // 可选，等价于先 Use(sid)
	Timeout:         8 * time.Second,
})
if err != nil {
	log.Fatal(err)
}
defer client.Close()

// REST 模式不需要 Login；可直接调用业务命令
info, err := client.ServerInfo(ctx)
if err != nil {
	log.Fatal(err)
}
log.Println(info.Name)
```

注意：

- WebQuery 模式下 `Login/Logout` 为兼容保留（无实际登录动作，认证靠 API Key）。
- `Use/UseByPort` 会切换 REST 请求路径中的 `sid`。
- 事件推送（`OnTextMessage` 等）在 WebQuery 模式不支持流式通知。

## 常用命令示例

### 查询在线客户端

```go
clients, err := client.ClientList(ctx, "-uid", "-groups", "-away")
if err != nil {
	log.Fatal(err)
}
for _, c := range clients {
	log.Printf("clid=%d nick=%s uid=%s groups=%v", c.ID, c.Nickname, c.UniqueIdentifier, c.ServerGroups)
}
```

### 发送消息 / 踢人 / 封禁

```go
if err := client.SendPrivateMessage(ctx, 12, "hello"); err != nil {
	log.Fatal(err)
}
if err := client.KickFromChannel(ctx, 12, "请遵守频道规则"); err != nil {
	log.Fatal(err)
}
banID, err := client.BanClient(ctx, 12, 3600, "spam")
if err != nil {
	log.Fatal(err)
}
log.Printf("ban id=%d", banID)
```

### 创建并修改频道

```go
cid, err := client.ChannelCreate(ctx, ts3.ChannelCreateOptions{
	Name:        "Music Room",
	Topic:       "全天音乐",
	MaxClients:  20,
	IsPermanent: true,
})
if err != nil {
	log.Fatal(err)
}

if err := client.ChannelEdit(ctx, cid, ts3.ChannelEditOptions{
	Topic:        "夜间轻音乐",
	CodecQuality: 7,
}); err != nil {
	log.Fatal(err)
}
```

### 服务器组管理

```go
sgid, err := client.ServerGroupAdd(ctx, "活动管理员", 1)
if err != nil {
	log.Fatal(err)
}
if err := client.ServerGroupAddClient(ctx, sgid, 42); err != nil {
	log.Fatal(err)
}
```

### 权限修改

```go
if err := client.ServerGroupAddPerm(ctx, 6, "i_channel_subscribe_power", 75, false, false); err != nil {
	log.Fatal(err)
}
perms, err := client.ServerGroupPermList(ctx, 6, true)
if err != nil {
	log.Fatal(err)
}
log.Printf("perms=%d", len(perms))
```

### 事件订阅

```go
if err := client.OnTextMessage(ctx, func(payload string) {
	log.Printf("notifytextmessage: %s", payload)
}); err != nil {
	log.Fatal(err)
}

if err := client.OnClientEnter(ctx, func(payload string) {
	log.Printf("notifycliententerview: %s", payload)
}); err != nil {
	log.Fatal(err)
}
```

## 错误处理

```go
if err != nil {
	var qerr *ts3.Error
	if errors.As(err, &qerr) {
		if qerr.Is(ts3.ErrPermissions) {
			log.Printf("权限不足: %s", qerr.Msg)
		}
		log.Printf("ts3 code=%d msg=%s", qerr.ID, qerr.Msg)
	}
}
```

## 详细手册

完整命令使用示例见：

- [docs/USAGE.md](docs/USAGE.md)

## 连接建议

- 单次管理操作建议 `context.WithTimeout(3~8s)`
- 长连接建议设置 `KeepAlivePeriod=20~60s`
- 大服可调大 `Config.MaxLineSize`（默认 1MB）

## 测试

```bash
go test ./...
go test -race ./ts3
```
