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

var ErrInternal = errors.New("internal error, please report a bug")

type stacktrace struct{ pc []uintptr }

func (ss *stacktrace) Frame(pc uintptr) (runtime.Frame, error) {
	if len(ss.pc) != 1 {
		return runtime.Frame{}, fmt.Errorf("%w: invalid stack length from pool", ErrInternal)
	}

	ss.pc[0] = pc
	f, _ := runtime.CallersFrames(ss.pc).Next()

	return f, nil
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
		slog.KindGroup:    zapcore.NamespaceType,
	}[kind]

	return val, found
}

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
		stackPool:      &sync.Pool{New: func() any { return &stacktrace{pc: make([]uintptr, 1)} }},
		fieldsPool:     &sync.Pool{New: func() any { return &fields{fields: make([]zapcore.Field, 0, initFieldPoolSize)} }},
	}
}

func New(logger interface{ Core() zapcore.Core }) *ZapHandler {
	return NewFromCore(logger.Core())
}

func (hand *ZapHandler) clone() ZapHandler {
	return *hand
}

// Enabled reports whether the handler handles records at the given level.
// The handler ignores records whose level is lower.
// It is called early, before any arguments are processed,
// to save effort if the log event should be discarded.
// If called from a Logger method, the first argument is the context
// passed to that method, or context.Background() if nil was passed
// or the method does not take a context.
// The context is passed so Enabled can use its values
// to make a decision.
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

	return stack.Frame(caller)
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
	// If caller is zero, ignore it.
	if !hand.AddSource || caller == 0 {
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
// It will only be called when Enabled returns true.
// The Context argument is as for Enabled.
// It is present solely to provide Handlers access to the context's values.
// Canceling the context should not affect record processing.
// (Among other things, log messages may be necessary to debug a
// cancellation-related problem.)
//
// Handle methods that produce output should observe the following rules:
//   - If r.Time is the zero time, ignore the time.
//   - If r.PC is zero, ignore it.
//   - Attr's values should be resolved.
//   - If an Attr's key and value are both the zero value, ignore the Attr.
//     This can be tested with attr.Equal(Attr{}).
//   - If a group's key is empty, inline the group's Attrs.
//   - If a group has no Attrs (even if it has a non-empty key),
//     ignore it.
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
// The Handler owns the slice: it may retain, modify or discard it.
func (hand *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cloned := hand.clone()

	_ = hand.withFields(func(f []zapcore.Field) error {
		for _, attr := range attrs {
			f = hand.appendAttr(f, attr, "")
		}

		cloned.core = cloned.core.With(f)

		return nil
	})

	return &cloned
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
// A Handler should treat WithGroup as starting a Group of Attrs that ends
// at the end of the log event. That is,
//
//	logger.WithGroup("s").LogAttrs(level, msg, slog.Int("a", 1), slog.Int("b", 2))
//
// should behave like
//
//	logger.LogAttrs(level, msg, slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
//
// If the name is empty, WithGroup returns the receiver.
func (hand *ZapHandler) WithGroup(name string) slog.Handler {
	// If the name is empty, WithGroup returns the receiver.
	if name == "" {
		return hand
	}

	cloned := hand.clone()
	cloned.prefix = hand.prefix + name + hand.GroupSeparator

	return &cloned
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
	// If an Attr's key and value are both the zero value, ignore the Attr. This can be tested with attr.Equal(Attr{}).
	if attr.Equal(slog.Attr{}) {
		return fields
	}

	if attr.Value.Kind() == slog.KindGroup {
		var addPrefix string

		// If a group's key is empty, inline the group's Attrs.
		if attr.Key != "" {
			addPrefix = attr.Key + hand.GroupSeparator
		}

		// If a group has no Attrs (even if it has a non-empty key), ignore it.
		for _, gAttr := range attr.Value.Group() {
			fields = hand.appendAttr(fields, gAttr, prefix+addPrefix)
		}

		return fields
	}

	Type, Integer, String, Interface := getType(attr.Value)

	return append(fields, zapcore.Field{
		Key:       hand.prefix + prefix + attr.Key,
		Type:      Type,
		Integer:   Integer,
		String:    String,
		Interface: Interface,
	})
}
