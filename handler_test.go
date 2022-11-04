package zaphandler_test

import (
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"

	"go.mrchanchal.com/zaphandler"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"golang.org/x/exp/slog"
)

func MatchEntry(t *testing.T, expected, got observer.LoggedEntry) {
	equal := func(name string, expected, got any) {
		if !reflect.DeepEqual(expected, got) {
			t.Errorf("mismatched %s\n%v\n%v", name, expected, got)
		}
	}

	equal("entry", expected.Entry, got.Entry)
	equal("context", expected.Context, got.Context)
}

func TestMain(m *testing.M) {
	backupStderr := os.Stderr
	os.Stderr = os.NewFile(uintptr(syscall.Stdin), os.DevNull)
	ret := m.Run()
	os.Stderr = backupStderr
	os.Exit(ret)
}

func BenchmarkZapHandler(b *testing.B) {
	for _, zapF := range []struct {
		Name string
		F    func(...zap.Option) (*zap.Logger, error)
	}{
		{
			Name: "Development",
			F:    zap.NewDevelopment,
		},
		{
			Name: "Production",
			F:    zap.NewProduction,
		},
	} {
		b.Run(zapF.Name, func(b *testing.B) {
			zapL, _ := zapF.F()
			defer func() {
				b.StopTimer()
				zapL.Sync()
				b.StartTimer()
			}()

			logger := slog.New(zaphandler.New(zapL))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("sample log message", "field1", "value1", "field2", 33)
			}
		})
	}
}

func BenchmarkZap(b *testing.B) {
	for _, zapF := range []struct {
		Name string
		F    func(...zap.Option) (*zap.Logger, error)
	}{
		{
			Name: "Development",
			F:    zap.NewDevelopment,
		},
		{
			Name: "Production",
			F:    zap.NewProduction,
		},
	} {
		b.Run(zapF.Name, func(b *testing.B) {
			zapL, _ := zapF.F()
			defer func() {
				b.StopTimer()
				zapL.Sync()
				b.StartTimer()
			}()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				zapL.Info("sample log message",
					zap.String("field1", "value1"),
					zap.Int("field2", 33),
				)
			}
		})
	}
}

func FuzzZapHandler(f *testing.F) {
	core, obs := observer.New(zap.DebugLevel)
	defer core.Sync()

	logger := slog.New(zaphandler.NewFromCore(core))

	f.Add("sample log", "string field", 9, false, int64(88),
		int64(42), 3.55, uint64(55))
	f.Add("sample log", "string field", 9, true, int64(88),
		int64(42), 3.55, uint64(55))

	f.Fuzz(func(t *testing.T, msg string, stringF string, intF int,
		boolF bool, int64F int64, int64F2 int64, float64F float64,
		uint64F uint64,
	) {
		logger.Info(msg,
			"stringF", stringF,
			"intF", intF,
			"boolF", boolF,
			"int64F", int64F,
			"durationF", time.Duration(int64F),
			"float64F", float64F,
			"timeF", time.Unix(int64F, int64F2),
			"uint64F", uint64F,
		)

		got := obs.AllUntimed()[obs.Len()-1]
		expected := observer.LoggedEntry{
			Entry: zapcore.Entry{
				Level:   zap.InfoLevel,
				Caller:  got.Caller,
				Message: msg,
			},
			Context: []zap.Field{
				zap.String("stringF", stringF),
				zap.Int("intF", intF),
				zap.Bool("boolF", boolF),
				zap.Int64("int64F", int64F),
				zap.Duration("durationF", time.Duration(int64F)),
				zap.Float64("float64F", float64F),
				zap.Time("timeF", time.Unix(int64F, int64F2)),
				zap.Uint64("uint64F", uint64F),
			},
		}

		MatchEntry(t, expected, got)
	})
}

func TestZapHandlerGroup(t *testing.T) {
	core, obs := observer.New(zap.DebugLevel)
	defer core.Sync()

	getObs := func() observer.LoggedEntry {
		return obs.AllUntimed()[obs.Len()-1]
	}

	logger := slog.New(zaphandler.NewFromCore(core))

	logger.WithGroup("s").LogAttrs(slog.InfoLevel, "", slog.Int("a", 1), slog.Int("b", 2))
	got := getObs()

	logger.LogAttrs(slog.InfoLevel, "", slog.Group("s", slog.Int("a", 1), slog.Int("b", 2)))
	MatchEntry(t, getObs(), got)

	logger.WithGroup("s1").With("a", 1).WithGroup("s2").LogAttrs(slog.InfoLevel, "", slog.Int("b", 2))
	got = getObs()

	logger.LogAttrs(slog.InfoLevel, "", slog.Group("s1", slog.Int("a", 1), slog.Group("s2", slog.Int("b", 2))))
	MatchEntry(t, getObs(), got)
}
