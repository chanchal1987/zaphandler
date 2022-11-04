package zaphandler

import (
	"io"

	"go.uber.org/zap/zapcore"
)

var _ zapcore.WriteSyncer = (*closeSyncer)(nil)

type closeSyncer struct{ io.WriteCloser }

func (c *closeSyncer) Sync() error { return c.Close() }
