package ts3

import "strings"

// ts3Escaper 用于发送命令时的转义 (Go string -> TS3 Server)
var ts3Escaper = strings.NewReplacer(
	"\\", "\\\\",
	"/", "\\/",
	" ", "\\s",
	"|", "\\p",
	"\a", "\\a",
	"\b", "\\b",
	"\f", "\\f",
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t",
	"\v", "\\v",
)

// ts3Unescaper 用于接收响应时的反转义 (TS3 Server -> Go string)
var ts3Unescaper = strings.NewReplacer(
	"\\\\", "\\",
	"\\/", "/",
	"\\s", " ",
	"\\p", "|",
	"\\a", "\a",
	"\\b", "\b",
	"\\f", "\f",
	"\\n", "\n",
	"\\r", "\r",
	"\\t", "\t",
	"\\v", "\v",
)

// Escape 将普通字符串转义为 TS3 协议格式
func Escape(s string) string {
	return ts3Escaper.Replace(s)
}

// Unescape 将 TS3 协议格式字符串还原为普通字符串
func Unescape(s string) string {
	return ts3Unescaper.Replace(s)
}
