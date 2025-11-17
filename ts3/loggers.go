package ts3

// Logger 接口，允许用户使用 logrus, zap 或标准 log
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Debug(v ...interface{}) // 用于记录原始的 TS3 通信
	Debugf(format string, v ...interface{})
}

// NopLogger 默认不输出任何日志
type NopLogger struct{}

func (l *NopLogger) Print(v ...interface{})                 {}
func (l *NopLogger) Printf(format string, v ...interface{}) {}
func (l *NopLogger) Debug(v ...interface{})                 {}
func (l *NopLogger) Debugf(format string, v ...interface{}) {}
