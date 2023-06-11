package types

import (
	"log/slog"

	"go.uber.org/zap/zapcore"
)

var _ zapcore.ObjectMarshaler = (Group)(nil)

type Group []slog.Attr

func (g Group) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range g {
		NewFieldType(attr.Value).Field(attr.Key).AddTo(enc)
	}

	return nil
}
