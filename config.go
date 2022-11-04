package zaphandler

import (
	"io"

	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slog"
)

type Config struct {
	StackF         func(slog.Level) string
	GroupSeparator string
	AddSource      bool
	ErrorOutput    io.WriteCloser
}

func DefaultConfig() Config {
	return Config{
		StackF:         nil,
		GroupSeparator: ".",
		AddSource:      false,
		ErrorOutput:    nil,
	}
}

func (c Config) Build(core zapcore.Core) *ZapHandler {
	hand := ZapHandler{
		prefix:    "",
		core:      core,
		config:    c,
		errOutput: nil,
	}

	if c.ErrorOutput != nil {
		hand.errOutput = &closeSyncer{c.ErrorOutput}
	}

	return &hand
}
