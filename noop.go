package zaphandler

import (
	"context"
	"log/slog"
)

type NoOpHandler struct{}

func (handler NoOpHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (handler NoOpHandler) Handle(context.Context, slog.Record) error { return nil }
func (handler NoOpHandler) WithAttrs([]slog.Attr) slog.Handler        { return handler }
func (handler NoOpHandler) WithGroup(string) slog.Handler             { return handler }
