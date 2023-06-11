package zaphandler_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"testing"
	"time"

	"go.mrchanchal.com/zaphandler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func MatchEntry(tb testing.TB, expected, got []observer.LoggedEntry) {
	tb.Helper()

	if len(expected) != len(got) {
		tb.Errorf("length of expected(%d) is not matching with got(%d)", len(expected), len(got))
	}

	for i := 0; i < len(expected); i++ {
		if !reflect.DeepEqual(expected, got) {
			tb.Errorf("mismatched entry\nExpected: %+v\nGot:      %+v", expected, got)
		}
	}
}

func Take(o *observer.ObservedLogs) []observer.LoggedEntry {
	ret := o.TakeAll()
	for i := range ret {
		ret[i].Time = time.Time{}
		ret[i].Caller.PC = 1
		ret[i].Caller.Line = 1
		ret[i].Caller.Function = ""
	}

	return ret
}

func Match(tb testing.TB, lvl zapcore.LevelEnabler, expected, got func(*slog.Logger, *zap.Logger)) {
	tb.Helper()

	core, obs := observer.New(lvl)
	zLogger := zap.New(core, zap.AddCaller())
	sLogger := slog.New(zaphandler.NewFromCore(core, zaphandler.AddSource()))

	expected(sLogger, zLogger)

	e := Take(obs)

	got(sLogger, zLogger)
	MatchEntry(tb, e, Take(obs))
}

func ObsLogger(lvl zapcore.LevelEnabler) (*slog.Logger, *zap.Logger, *observer.ObservedLogs) {
	core, obs := observer.New(lvl)

	return slog.New(zaphandler.NewFromCore(core)), zap.New(core), obs
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
				logger.Info("sample log message", "field1", "value1", "field2", 33, "field3", []int{32, 33})
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

			hand := zaphandler.New(zapL)
			hand.AddSource = true
			logger := slog.New(hand)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("sample log message", "field1", "value1", "field2", 33, "field3", []int{32, 33})
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
					zap.Ints("field3", []int{32, 33}),
				)
			}

			b.StopTimer()
		})
	}
}

func TestWith(t *testing.T) {
	t.Parallel()

	Match(t, zap.DebugLevel, func(l *slog.Logger, _ *zap.Logger) {
		l.Info("test", "test-field", "test-value")
	}, func(l *slog.Logger, _ *zap.Logger) {
		l.With("test-field", "test-value").Info("test")
	})
}

func TestZapHandlerGroup(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	Match(t, zap.DebugLevel, func(l *slog.Logger, _ *zap.Logger) {
		l.WithGroup("s").LogAttrs(ctx, slog.LevelInfo, "", slog.Int("a", 1), slog.Int("b", 2))
	}, func(l *slog.Logger, _ *zap.Logger) {
		l.LogAttrs(ctx, slog.LevelInfo, "", slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
	})
}
