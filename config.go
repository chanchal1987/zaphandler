package zaphandler

import (
	"fmt"
	"io"
	"runtime"
	"sync"

	"go.uber.org/zap/zapcore"
)

const initFieldPoolSize = 32

var _ zapcore.WriteSyncer = (*CloseSyncer)(nil)

type CloseSyncer struct{ io.WriteCloser }

func (c *CloseSyncer) Sync() error {
	if err := c.WriteCloser.Close(); err != nil {
		return fmt.Errorf("unable to close writesyncer: %w", err)
	}

	return nil
}

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
	ErrorOutput    io.Writer
}

func DefaultConfig() Config {
	return Config{
		GroupSeparator: ".",
		AddSource:      false,
		ErrorOutput:    nil,
	}
}

func (c Config) Build(core zapcore.Core) *ZapHandler {
	hand := ZapHandler{
		prefix:     "",
		core:       core,
		config:     c,
		errOutput:  nil,
		stackPool:  &sync.Pool{New: func() any { return &stacktrace{pc: [1]uintptr{}} }},
		fieldsPool: &sync.Pool{New: func() any { return &fields{fields: make([]zapcore.Field, 0, initFieldPoolSize)} }},
	}

	if c.ErrorOutput == io.Writer(nil) {
		c.ErrorOutput = io.Discard
	}

	switch errOut := c.ErrorOutput.(type) {
	case zapcore.WriteSyncer:
		hand.errOutput = errOut
	case io.WriteCloser:
		hand.errOutput = &CloseSyncer{WriteCloser: errOut}
	default:
		hand.errOutput = zapcore.AddSync(errOut)
	}

	return &hand
}
