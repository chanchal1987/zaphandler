package zaphandler_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"go.mrchanchal.com/zaphandler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/slog"
)

func MatchEntry(log interface{ Errorf(string, ...any) }, expected, got []observer.LoggedEntry) {
	if len(expected) != len(got) {
		log.Errorf("length of expected(%d) is not matching with got(%d)", len(expected), len(got))
	}

	equal := func(name string, expected, got any) {
		if !reflect.DeepEqual(expected, got) {
			log.Errorf("mismatched %s\nExpected: %+v\nGot:      %+v", name, expected, got)
		}
	}

	emptyCaller := zapcore.EntryCaller{
		Defined:  false,
		PC:       0,
		File:     "",
		Line:     0,
		Function: "",
	}

	for i := 0; i < len(expected); i++ {
		equal("caller", expected[i].Entry.Caller.File, got[i].Entry.Caller.File)
		expected[i].Entry.Caller = emptyCaller
		got[i].Entry.Caller = emptyCaller
		equal("entry", expected[i].Entry, got[i].Entry)
		equal("context", expected[i].Context, got[i].Context)
	}
}

func TakeUntimed(o *observer.ObservedLogs) []observer.LoggedEntry {
	ret := o.TakeAll()
	for i := range ret {
		ret[i].Time = time.Time{}
	}

	return ret
}

func BenchData() []struct {
	Name string
	F    func(...zap.Option) (*zap.Logger, error)
} {
	return []struct {
		Name string
		F    func(...zap.Option) (*zap.Logger, error)
	}{
		{
			Name: "Development",
			F: func(o ...zap.Option) (*zap.Logger, error) {
				c := zap.NewDevelopmentConfig()
				c.OutputPaths = []string{os.DevNull}

				l, err := c.Build(o...)
				if err != nil {
					return nil, fmt.Errorf("test error: %w", err)
				}

				return l, nil
			},
		},
		{
			Name: "Production",
			F: func(o ...zap.Option) (*zap.Logger, error) {
				c := zap.NewProductionConfig()
				c.OutputPaths = []string{os.DevNull}

				l, err := c.Build(o...)
				if err != nil {
					return nil, fmt.Errorf("test error: %w", err)
				}

				return l, nil
			},
		},
	}
}

func HandleNullSyncErr(err error) error {
	if err == nil {
		return nil
	}

	var pe *os.PathError
	if errors.As(err, &pe) && pe.Op == "sync" && pe.Path == os.DevNull {
		return nil
	}

	return err
}

func BenchmarkZapHandler(b *testing.B) {
	for _, zapF := range BenchData() {
		zapF := zapF

		b.Run(zapF.Name, func(b *testing.B) {
			zapL, err := zapF.F()
			if err != nil {
				b.Error(err)
			}

			defer func() {
				if err := HandleNullSyncErr(zapL.Sync()); err != nil {
					b.Error(err)
				}
			}()

			logger := slog.New(zaphandler.New(zapL))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("sample log message", "field1", "value1", "field2", 33)
			}

			b.StopTimer()
		})
	}
}

func BenchmarkZapHandlerWithCaller(b *testing.B) {
	for _, zapF := range BenchData() {
		zapF := zapF

		b.Run(zapF.Name, func(b *testing.B) {
			zapL, err := zapF.F()
			if err != nil {
				b.Error(err)
			}

			defer func() {
				if err := HandleNullSyncErr(zapL.Sync()); err != nil {
					b.Error(err)
				}
			}()

			logger := slog.New(zaphandler.Config{
				GroupSeparator: ".",
				AddSource:      true,
				ErrorOutput:    nil,
			}.Build(zapL.Core()))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("sample log message", "field1", "value1", "field2", 33)
			}

			b.StopTimer()
		})
	}
}

func BenchmarkZap(b *testing.B) {
	for _, zapF := range BenchData() {
		zapF := zapF

		b.Run(zapF.Name, func(b *testing.B) {
			zapL, err := zapF.F()
			if err != nil {
				b.Error(err)
			}

			defer func() {
				if err := HandleNullSyncErr(zapL.Sync()); err != nil {
					b.Error(err)
				}
			}()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				zapL.Info("sample log message",
					zap.String("field1", "value1"),
					zap.Int("field2", 33),
				)
			}

			b.StopTimer()
		})
	}
}

type dummyStringer string

func (s dummyStringer) String() string {
	return string(s)
}

var errTest = errors.New("test error")

func FuzzZapHandler(f *testing.F) {
	core, obs := observer.New(zap.DebugLevel)
	defer func() {
		if err := core.Sync(); err != nil {
			f.Error(err)
		}
	}()

	logger := slog.New(zaphandler.Config{
		GroupSeparator: ".",
		AddSource:      true,
		ErrorOutput:    os.Stderr,
	}.Build(core))

	zapL := zap.New(core, zap.AddCaller())

	f.Add("sample log", "string field", 9, false, int64(88),
		int64(42), 3.55, uint64(55))
	f.Add("sample log", "string field", 9, true, int64(88),
		int64(42), 3.55, uint64(55))

	f.Fuzz(func(t *testing.T, msg string, stringF string, intF int,
		boolF bool, int64F int64, int64F2 int64, float64F float64,
		uint64F uint64,
	) {
		err := fmt.Errorf("%w: %s", errTest, msg)
		stringer := dummyStringer(msg)

		zapL.Info(msg,
			zap.String("stringF", stringF),
			zap.Int("intF", intF),
			zap.Bool("boolF", boolF),
			zap.Int64("int64F", int64F),
			zap.Duration("durationF", time.Duration(int64F)),
			zap.Float64("float64F", float64F),
			zap.Time("timeF", time.Unix(int64F, int64F2)),
			zap.Uint64("uint64F", uint64F),
			zap.Error(err),
			zap.Stringer("stringerF", stringer),
		)

		expected := TakeUntimed(obs)

		logger.Info(msg,
			"stringF", stringF,
			"intF", intF,
			"boolF", boolF,
			"int64F", int64F,
			"durationF", time.Duration(int64F),
			"float64F", float64F,
			"timeF", time.Unix(int64F, int64F2),
			"uint64F", uint64F,
			"error", err,
			"stringerF", stringer,
		)

		MatchEntry(t, expected, TakeUntimed(obs))
	})
}

func TestZapHandlerGroup(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	core, obs := observer.New(zap.DebugLevel)
	defer func() {
		if err := core.Sync(); err != nil {
			t.Error(err)
		}
	}()

	logger := slog.New(zaphandler.NewFromCore(core))

	logger.WithGroup("s").LogAttrs(ctx, slog.LevelInfo, "", slog.Int("a", 1), slog.Int("b", 2))

	got := TakeUntimed(obs)

	logger.LogAttrs(ctx, slog.LevelInfo, "", slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
	MatchEntry(t, TakeUntimed(obs), got)
}
