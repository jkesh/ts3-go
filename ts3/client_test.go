package ts3

import (
	"bufio"
	"net"
	"testing"
)

// mockServer 模拟一个简单的 TS3 服务器行为
func mockServer(conn net.Conn) {
	defer conn.Close()
	writer := bufio.NewWriter(conn)
	scanner := bufio.NewScanner(conn)

	// 1. 发送握手
	writer.WriteString("TS3\n")
	writer.WriteString("Welcome to TeamSpeak 3 Server\n")
	writer.Flush()

	// 2. 简单的命令响应循环
	for scanner.Scan() {
		cmd := scanner.Text()
		switch cmd {
		case "whoami":
			// 模拟 whoami 的响应
			writer.WriteString("virtualserver_status=online virtualserver_id=1 client_id=5\n")
			writer.WriteString("error id=0 msg=ok\n")
		case "clientlist":
			// 模拟 clientlist 响应
			writer.WriteString("clid=1 client_nickname=Alice|clid=2 client_nickname=Bob\n")
			writer.WriteString("error id=0 msg=ok\n")
		default:
			writer.WriteString("error id=256 msg=command\\snot\\sfound\n")
		}
		writer.Flush()
	}
}

func TestClient_WhoAmI(t *testing.T) {
	// 创建管道：serverEnd 给 mockServer 用，clientEnd 给 Client 用
	serverEnd, clientEnd := net.Pipe()

	// 启动模拟服务器
	go mockServer(serverEnd)

	// 初始化 Client，手动注入 connection
	// 注意：我们需要稍微修改 NewClient 逻辑或手动构造 Client 结构体来支持注入 conn
	// 为了测试，我们可以手动构造：
	c := &Client{
		conn:       clientEnd,
		scanner:    bufio.NewScanner(clientEnd),
		cmdResChan: make(chan string),
		errorChan:  make(chan error, 1),
		quit:       make(chan struct{}),
	}

	// 启动 readLoop
	go c.readLoop()
	// 消耗握手信息 (mockServer 发送了 TS3 和 Welcome)
	// 注意：由于 readLoop 可能会抢在 handshake 逻辑之前读取，
	// 在真实测试中，通过 net.Pipe 可能需要更精细的控制。
	// 简单起见，假设 readLoop 会丢弃非 cmd 非 notify 的行，或者我们在这里不处理握手

	// 测试 Exec
	resp, err := c.Exec("whoami")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}

	// 验证原始响应
	expected := "virtualserver_status=online virtualserver_id=1 client_id=5"
	if resp != expected {
		t.Errorf("Expected %s, got %s", expected, resp)
	}
}
