package ts3

import "fmt"

// Error represents an error returned by the TS3 ServerQuery API.
type Error struct {
	ID  int    `ts3:"id"`
	Msg string `ts3:"msg"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("ts3 error %d: %s", e.ID, e.Msg)
}

// Is reports whether the error code matches the provided TS3 error id.
func (e *Error) Is(id int) bool {
	return e != nil && e.ID == id
}

// Common TS3 error codes.
const (
	ErrOK                  = 0
	ErrOk                  = ErrOK // kept for backward compatibility
	ErrCommandNotFound     = 256
	ErrParameterNotFound   = 257
	ErrDatabaseEmptyResult = 1281
	ErrPermissions         = 2568
	ErrNicknameInUse       = 513
	ErrFloodBan            = 3329
)

// NewError returns nil when id is zero, otherwise it returns *Error.
func NewError(id int, msg string) error {
	if id == ErrOK {
		return nil
	}
	return &Error{ID: id, Msg: msg}
}
