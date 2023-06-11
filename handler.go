package zaphandler

import (
	"context"
	"fmt"
	"log/slog"

	"go.mrchanchal.com/zaphandler/types"
	"go.uber.org/zap/zapcore"
)

const ErrorKey = "error"

func level(lvl slog.Level) (zapcore.Level, bool) {
	val, found := map[slog.Level]zapcore.Level{
		slog.LevelDebug: zapcore.DebugLevel,
		slog.LevelInfo:  zapcore.InfoLevel,
		slog.LevelWarn:  zapcore.WarnLevel,
		slog.LevelError: zapcore.ErrorLevel,
	}[lvl]

	return val, found
}

type Option func(*ZapHandler)

func AddSource() Option { return func(h *ZapHandler) { h.AddSource = true } }

type ZapHandler struct {
	AddSource bool
	groups    []string
	core      zapcore.Core
	pool      *poolT
}

func NewFromCore(core zapcore.Core, options ...Option) *ZapHandler {
	hand := ZapHandler{core: core, pool: newPool()}

	for _, opt := range options {
		opt(&hand)
	}

	return &hand
}

func New(logger interface{ Core() zapcore.Core }, options ...Option) *ZapHandler {
	return NewFromCore(logger.Core(), options...)
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

	checked := hand.core.Check(zapcore.Entry{
		Level:   lvl,
		Time:    rec.Time,
		Message: rec.Message,
		Caller:  hand.pool.caller(rec.PC, hand.AddSource),
	}, nil)
	if checked == nil {
		return nil
	}

	var errOut Error
	checked.ErrorOutput = &errOut

	hand.write(rec, checked)

	return errOut.Err()
}

func (hand *ZapHandler) write(rec slog.Record, checked *zapcore.CheckedEntry) {
	if len(hand.groups) == 0 {
		hand.pool.withFields(func(field []zapcore.Field) {
			rec.Attrs(func(attr slog.Attr) bool {
				field = hand.appendAttr(field, attr)

				return true
			})

			checked.Write(field...)
		})
	} else {
		hand.pool.withAttrs(func(attrs []slog.Attr) {
			rec.Attrs(func(attr slog.Attr) bool {
				attrs = append(attrs, attr)

				return true
			})

			grp := slog.Attr{
				Key:   hand.groups[len(hand.groups)-1],
				Value: slog.GroupValue(attrs...),
			}

			for i := len(hand.groups) - 1 - 1; i >= 0; i-- {
				grp = slog.Group(hand.groups[i], grp)
			}

			checked.Write(types.NewFieldType(grp.Value).Field(grp.Key))
		})
	}
}

// WithAttrs returns a new Handler whose attributes consist of
// both the receiver's attributes and the arguments.
// The Handler owns the slice: it may retain, modify or discard it.
func (hand *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cloned := hand.clone()

	hand.pool.withFields(func(f []zapcore.Field) {
		for _, attr := range attrs {
			f = hand.appendAttr(f, attr)
		}

		cloned.core = cloned.core.With(f)
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
	cloned.groups = append(cloned.groups, name)

	return &cloned
}

func (hand *ZapHandler) appendAttr(fields []zapcore.Field, attr slog.Attr) []zapcore.Field {
	// If an Attr's key and value are both the zero value, ignore the Attr.
	if attr.Equal(slog.Attr{}) {
		return fields
	}

	// If a group's key is empty, inline the group's Attrs.
	if attr.Value.Kind() == slog.KindGroup && attr.Key == "" {
		for _, a := range attr.Value.Group() {
			fields = hand.appendAttr(fields, a)
		}

		return fields
	}

	return append(fields, types.NewFieldType(attr.Value).Field(attr.Key))
}
