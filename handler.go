package zaphandler

import (
	"fmt"
	"math"

	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slog"
)

//nolint:gochecknoglobals
var labelMap = map[slog.Level]zapcore.Level{
	slog.DebugLevel: zapcore.DebugLevel,
	slog.InfoLevel:  zapcore.InfoLevel,
	slog.WarnLevel:  zapcore.WarnLevel,
	slog.ErrorLevel: zapcore.ErrorLevel,
}

//nolint:gochecknoglobals
var typeMap = map[slog.Kind]zapcore.FieldType{
	slog.BoolKind:     zapcore.BoolType,
	slog.DurationKind: zapcore.DurationType,
	slog.Float64Kind:  zapcore.Float64Type,
	slog.Int64Kind:    zapcore.Int64Type,
	slog.StringKind:   zapcore.StringType,
	slog.TimeKind:     zapcore.TimeType,
	slog.Uint64Kind:   zapcore.Uint64Type,
}

var _ slog.Handler = (*ZapHandler)(nil)

type ZapHandler struct {
	prefix    string
	core      zapcore.Core
	errOutput zapcore.WriteSyncer
	config    Config
}

func NewFromCore(core zapcore.Core) *ZapHandler {
	return DefaultConfig().Build(core)
}

func New(logger interface{ Core() zapcore.Core }) *ZapHandler {
	return NewFromCore(logger.Core())
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// Enabled is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
func (hand *ZapHandler) Enabled(level slog.Level) bool {
	return hand.core.Enabled(labelMap[level])
}

// Handle handles the Record.
// It will only be called if Enabled returns true.
func (hand *ZapHandler) Handle(rec slog.Record) error {
	var stack string
	if hand.config.StackF != nil {
		stack = hand.config.StackF(rec.Level)
	}

	fields := make([]zapcore.Field, 0, rec.NumAttrs())
	rec.Attrs(func(attr slog.Attr) {
		fields = hand.appendAttr(fields, attr, "")
	})

	var Caller zapcore.EntryCaller
	if hand.config.AddSource {
		Caller.Defined = true
		Caller.File, Caller.Line = rec.SourceLine()
	}

	checkedCore := hand.core.Check(zapcore.Entry{
		Level:      labelMap[rec.Level],
		Time:       rec.Time,
		LoggerName: "",
		Message:    rec.Message,
		Caller:     Caller,
		Stack:      stack,
	}, nil)

	if hand.errOutput != nil {
		checkedCore.ErrorOutput = hand.errOutput
	}

	checkedCore.Write(fields...)

	return nil
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (hand *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler { //nolint:ireturn
	fields := make([]zapcore.Field, 0, flatLen(attrs))

	for _, attr := range attrs {
		fields = hand.appendAttr(fields, attr, "")
	}

	return &ZapHandler{
		prefix:    hand.prefix,
		core:      hand.core.With(fields),
		errOutput: hand.errOutput,
		config:    hand.config,
	}
}

// WithGroup returns a new Handler with the given group appended to
// the receiver's existing groups.
// The keys of all subsequent attributes, whether added by With or in a
// Record, should be qualified by the sequence of group names.
//
// How this qualification happens is up to the Handler, so long as
// this Handler's attribute keys differ from those of another Handler
// with a different sequence of group names.
//
//	logger.WithGroup("s").LogAttrs(slog.Int("a", 1), slog.Int("b", 2))
//
// will behave like
//
//	logger.LogAttrs(slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
func (hand *ZapHandler) WithGroup(name string) slog.Handler { //nolint:ireturn
	return &ZapHandler{
		prefix:    hand.prefix + name + hand.config.GroupSeparator,
		core:      hand.core,
		errOutput: hand.errOutput,
		config:    hand.config,
	}
}

//nolint:cyclop
func getType(val slog.Value) (zapcore.FieldType, int64, string, any) {
	var (
		Type      = zapcore.ReflectType
		Integer   = int64(0)
		String    = ""
		Interface = any(nil)
	)

	Kind := val.Kind()
	if gotType, found := typeMap[Kind]; found {
		Type = gotType
	}

	switch Kind {
	case slog.AnyKind:
		Interface = val.Any()

		switch Interface.(type) {
		case fmt.Stringer:
			Type = zapcore.StringerType
		case error:
			Type = zapcore.ErrorType
		}
	case slog.BoolKind:
		if val.Bool() {
			Integer = 1
		}
	case slog.DurationKind:
		Integer = int64(val.Duration())
	case slog.Float64Kind:
		Integer = int64(math.Float64bits(val.Float64()))
	case slog.Int64Kind:
		Integer = val.Int64()
	case slog.StringKind:
		String = val.String()
	case slog.TimeKind:
		t := val.Time()
		Integer = t.UnixNano()
		Interface = t.Location()
	case slog.Uint64Kind:
		Type = zapcore.Uint64Type
		Integer = int64(val.Uint64())
	case slog.GroupKind:
	case slog.LogValuerKind:
	}

	return Type, Integer, String, Interface
}

func flatLen(attrs []slog.Attr) int {
	length := len(attrs)

	for _, attr := range attrs {
		if attr.Value.Kind() == slog.GroupKind {
			length += flatLen(attr.Value.Group()) - 1
		}
	}

	return length
}

func (hand *ZapHandler) appendAttr(fields []zapcore.Field, attr slog.Attr, prefix string) []zapcore.Field {
	if attr.Value.Kind() != slog.GroupKind {
		Type, Integer, String, Interface := getType(attr.Value)

		return append(fields, zapcore.Field{
			Key:       hand.prefix + prefix + attr.Key,
			Type:      Type,
			Integer:   Integer,
			String:    String,
			Interface: Interface,
		})
	}

	for _, gAttr := range attr.Value.Group() {
		fields = hand.appendAttr(fields, gAttr, prefix+attr.Key+hand.config.GroupSeparator)
	}

	return fields
}
