package zaphandler

import "golang.org/x/exp/slog"

//nolint:gochecknoglobals
var NoOp = slog.New(NoOpHandler{})

var _ slog.Handler = (*NoOpHandler)(nil)

type NoOpHandler struct{}

func (handler NoOpHandler) Enabled(slog.Level) bool {
	return false
}

func (handler NoOpHandler) Handle(slog.Record) error {
	return nil
}

//nolint:ireturn,nolintlint
func (handler NoOpHandler) WithAttrs([]slog.Attr) slog.Handler {
	return handler
}

//nolint:ireturn,nolintlint
func (handler NoOpHandler) WithGroup(string) slog.Handler {
	return handler
}
