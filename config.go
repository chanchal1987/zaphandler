package zaphandler

import (
	"runtime"
	"sync"

	"go.uber.org/zap/zapcore"
)

const initFieldPoolSize = 32

type stacktrace struct{ pc [1]uintptr }

func (ss *stacktrace) Frame(pc uintptr) runtime.Frame {
	ss.pc[0] = pc
	f, _ := runtime.CallersFrames(ss.pc[:]).Next()

	return f
}

type fields struct{ fields []zapcore.Field }

type Config struct {
	GroupSeparator string
	AddSource      bool
}

func DefaultConfig() Config {
	return Config{
		GroupSeparator: ".",
		AddSource:      false,
	}
}

func (c Config) Build(core zapcore.Core) *ZapHandler {
	hand := ZapHandler{
		prefix:     "",
		core:       core,
		config:     c,
		stackPool:  &sync.Pool{New: func() any { return &stacktrace{pc: [1]uintptr{}} }},
		fieldsPool: &sync.Pool{New: func() any { return &fields{fields: make([]zapcore.Field, 0, initFieldPoolSize)} }},
	}

	return &hand
}
