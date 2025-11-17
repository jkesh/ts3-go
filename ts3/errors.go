package ts3

import "fmt"

// Error 代表一个 TS3 服务器端返回的错误
type Error struct {
	ID  int
	Msg string
}

func (e *Error) Error() string {
	return fmt.Sprintf("ts3 error %d: %s", e.ID, e.Msg)
}

// IsToCheck 检查错误 ID 是否为指定的 ID
func (e *Error) Is(id int) bool {
	return e.ID == id
}

// 常见 TS3 错误码常量 (参考官方文档)
const (
	ErrOk                  = 0
	ErrCommandNotFound     = 256
	ErrParameterNotFound   = 257
	ErrDatabaseEmptyResult = 1281
	ErrPermissions         = 2568
	ErrNicknameInUse       = 513
	ErrFloodBan            = 3329
)

// NewError 构造函数
func NewError(id int, msg string) error {
	if id == 0 {
		return nil
	}
	return &Error{ID: id, Msg: msg}
}
