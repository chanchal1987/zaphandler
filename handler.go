package zaphandler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sync"

	"go.uber.org/zap/zapcore"
	"log/slog"
)

const initFieldPoolSize = 32

type stacktrace struct{ pc [1]uintptr }

func (ss *stacktrace) Frame(pc uintptr) runtime.Frame {
	ss.pc[0] = pc
	f, _ := runtime.CallersFrames(ss.pc[:]).Next()

	return f
}

type fields struct{ fields []zapcore.Field }

func level(lvl slog.Level) (zapcore.Level, bool) {
	val, found := map[slog.Level]zapcore.Level{
		slog.LevelDebug: zapcore.DebugLevel,
		slog.LevelInfo:  zapcore.InfoLevel,
		slog.LevelWarn:  zapcore.WarnLevel,
		slog.LevelError: zapcore.ErrorLevel,
	}[lvl]

	return val, found
}

func fieldType(kind slog.Kind) (zapcore.FieldType, bool) {
	val, found := map[slog.Kind]zapcore.FieldType{
		slog.KindBool:     zapcore.BoolType,
		slog.KindDuration: zapcore.DurationType,
		slog.KindFloat64:  zapcore.Float64Type,
		slog.KindInt64:    zapcore.Int64Type,
		slog.KindString:   zapcore.StringType,
		slog.KindTime:     zapcore.TimeType,
		slog.KindUint64:   zapcore.Uint64Type,
	}[kind]

	return val, found
}

var ErrInternal = errors.New("internal error")

type ZapHandler struct {
	GroupSeparator string
	AddSource      bool
	prefix         string
	core           zapcore.Core
	stackPool      *sync.Pool
	fieldsPool     *sync.Pool
}

func NewFromCore(core zapcore.Core) *ZapHandler {
	return &ZapHandler{
		GroupSeparator: ".",
		AddSource:      false,
		prefix:         "",
		core:           core,
		stackPool:      &sync.Pool{New: func() any { return &stacktrace{pc: [1]uintptr{}} }},
		fieldsPool:     &sync.Pool{New: func() any { return &fields{fields: make([]zapcore.Field, 0, initFieldPoolSize)} }},
	}
}

func New(logger interface{ Core() zapcore.Core }) *ZapHandler {
	return NewFromCore(logger.Core())
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// Enabled is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
func (hand *ZapHandler) Enabled(_ context.Context, l slog.Level) bool {
	if v, ok := level(l); ok {
		return hand.core.Enabled(v)
	}

	return false
}

func (hand *ZapHandler) frame(caller uintptr) (runtime.Frame, error) {
	stack, ok := hand.stackPool.Get().(*stacktrace)
	if !ok {
		return runtime.Frame{}, fmt.Errorf("%w: invalid stack from pool", ErrInternal)
	}

	defer hand.stackPool.Put(stack)

	return stack.Frame(caller), nil
}

func (hand *ZapHandler) withFields(fieldFunc func([]zapcore.Field) error) error {
	fields, ok := hand.fieldsPool.Get().(*fields)
	if !ok {
		return fmt.Errorf("%w: invalid field from pool", ErrInternal)
	}

	defer hand.fieldsPool.Put(fields)

	return fieldFunc(fields.fields[:0])
}

func (hand *ZapHandler) entryCaller(caller uintptr) (zapcore.EntryCaller, error) {
	if !hand.AddSource || caller == 0 {
		//nolint:exhaustivestruct,exhaustruct
		return zapcore.EntryCaller{}, nil
	}

	frame, err := hand.frame(caller)
	if err != nil || frame.PC == 0 {
		return zapcore.EntryCaller{}, err
	}

	return zapcore.EntryCaller{
		Defined:  true,
		PC:       frame.PC,
		File:     frame.File,
		Line:     frame.Line,
		Function: frame.Function,
	}, nil
}

// Handle handles the Record.
// It will only be called if Enabled returns true.
func (hand *ZapHandler) Handle(ctx context.Context, rec slog.Record) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("error from context: %w", err)
	}

	lvl := zapcore.Level(rec.Level)
	if l, ok := level(rec.Level); ok {
		lvl = l
	}

	caller, err := hand.entryCaller(rec.PC)
	if err != nil {
		return err
	}

	checked := hand.core.Check(zapcore.Entry{
		Level:      lvl,
		Time:       rec.Time,
		LoggerName: "",
		Message:    rec.Message,
		Caller:     caller,
		Stack:      "",
	}, nil)
	if checked == nil {
		return nil
	}

	var errOut Error
	checked.ErrorOutput = &errOut

	return hand.withFields(func(field []zapcore.Field) error {
		rec.Attrs(func(attr slog.Attr) bool {
			field = hand.appendAttr(field, attr, "")

			return true
		})

		checked.Write(field...)

		return errOut.Err()
	})
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
func (hand *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler { //nolint:ireturn
	var fields []zapcore.Field

	_ = hand.withFields(func(f []zapcore.Field) error {
		for _, attr := range attrs {
			f = hand.appendAttr(f, attr, "")
		}

		copy(fields, f)

		return nil
	})

	return &ZapHandler{
		GroupSeparator: hand.GroupSeparator,
		AddSource:      hand.AddSource,
		prefix:         hand.prefix,
		core:           hand.core.With(fields),
		stackPool:      hand.stackPool,
		fieldsPool:     hand.fieldsPool,
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
		GroupSeparator: hand.GroupSeparator,
		AddSource:      hand.AddSource,
		prefix:         hand.prefix + name + hand.GroupSeparator,
		core:           hand.core,
		stackPool:      hand.stackPool,
		fieldsPool:     hand.fieldsPool,
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
	if t, found := fieldType(Kind); found {
		Type = t
	}

	switch Kind {
	case slog.KindAny:
		Interface = val.Any()

		switch Interface.(type) {
		case fmt.Stringer:
			Type = zapcore.StringerType
		case error:
			Type = zapcore.ErrorType
		}
	case slog.KindBool:
		if val.Bool() {
			Integer = 1
		}
	case slog.KindDuration:
		Integer = int64(val.Duration())
	case slog.KindFloat64:
		Integer = int64(math.Float64bits(val.Float64()))
	case slog.KindInt64:
		Integer = val.Int64()
	case slog.KindString:
		String = val.String()
	case slog.KindTime:
		t := val.Time()
		Integer = t.UnixNano()
		Interface = t.Location()
	case slog.KindUint64:
		Type = zapcore.Uint64Type
		Integer = int64(val.Uint64())
	case slog.KindGroup:
	case slog.KindLogValuer:
	}

	return Type, Integer, String, Interface
}

func (hand *ZapHandler) appendAttr(fields []zapcore.Field, attr slog.Attr, prefix string) []zapcore.Field {
	if attr.Value.Kind() != slog.KindGroup {
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
		fields = hand.appendAttr(fields, gAttr, prefix+attr.Key+hand.GroupSeparator)
	}

	return fields
}
