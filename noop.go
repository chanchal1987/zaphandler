package zaphandler

import (
	"context"

	"golang.org/x/exp/slog"
)

func NewNoOp() NoOpHandler {
	return NoOpHandler{}
}

type NoOpHandler struct{}

func (handler NoOpHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (handler NoOpHandler) Handle(context.Context, slog.Record) error { return nil }
func (handler NoOpHandler) WithAttrs([]slog.Attr) slog.Handler        { return handler } //nolint:ireturn,nolintlint
func (handler NoOpHandler) WithGroup(string) slog.Handler             { return handler } //nolint:ireturn,nolintlint
