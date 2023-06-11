package zaphandler_test

import (
	"errors"
	"log/slog"
	"testing"
	"time"

	"go.mrchanchal.com/zaphandler"
	"go.mrchanchal.com/zaphandler/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DummyStringer string

func (s DummyStringer) String() string {
	return string(s)
}

//nolint:godox,lll,funlen,gocognit,cyclop,maintidx
func FuzzFields(f *testing.F) {
	match := func(t *testing.T, z func(*zap.Logger), s func(*slog.Logger)) {
		t.Helper()
		Match(t, zapcore.DebugLevel, func(_ *slog.Logger, l *zap.Logger) { z(l) }, func(l *slog.Logger, _ *zap.Logger) { s(l) })
	}

	matchParallel := func(t *testing.T, z func(*zap.Logger), s func(*slog.Logger)) {
		t.Helper()
		t.Parallel()
		match(t, z, s)
	}

	f.Add("", "", false,
		[]byte{}, []byte{},
		false, false,
		0.0, 0.0, float32(0.0), float32(0.0),
		0, 0, int64(0), int64(0), int32(0), int32(0), int16(0), int16(0), int8(0), int8(0),
		uint(0), uint(0), uint64(0), uint64(0), uint32(0), uint32(0), uint16(0), uint16(0), uint8(0), uint8(0),
	)

	f.Add("test log", "test-key", true,
		[]byte{'\x01'}, []byte{'\x02', '\x01'},
		true, false,
		1.22, 1.23, float32(1.24), float32(1.25),
		7, 18, int64(8), int64(9), int32(10), int32(19), int16(11), int16(20), int8(12), int8(21),
		uint(13), uint(22), uint64(14), uint64(23), uint32(15), uint32(24), uint16(16), uint16(25), uint8(17), uint8(26),
	)

	f.Fuzz(func(t *testing.T, msg, key string, nilPtr bool,
		bytesVal1, bytesVal2 []byte,
		boolVal1, boolVal2 bool,
		float64Val1, float64Val2 float64, float32Val1, float32Val2 float32,
		intVal1, intVal2 int, int64Val1, int64Val2 int64, int32Val1, int32Val2 int32, int16Val1, int16Val2 int16, int8Val1, int8Val2 int8,
		uintVal1, uintVal2 uint, uint64Val1, uint64Val2 uint64, uint32Val1, uint32Val2 uint32, uint16Val1, uint16Val2 uint16, uint8Val1, uint8Val2 uint8,
	) {
		t.Parallel()

		complex128Val1, complex128Val2 := complex(float64Val1, float64Val2), complex(float64Val2, float64Val1)
		complex64Val1, complex64Val2 := complex(float32Val1, float32Val2), complex(float32Val2, float32Val1)
		uintptrVal1, uintptrVal2 := uintptr(uint64Val1), uintptr(uint64Val2)
		referenceVal := struct{ String string }{String: msg}
		stringerVal := DummyStringer(msg)
		timeVal1, timeVal2 := time.Unix(int64Val1, int64Val2), time.Unix(int64Val2, int64Val1)
		durationVal1, durationVal2 := time.Duration(int64Val1), time.Duration(int64Val2)
		errVal1, errVal2 := errors.New(msg), errors.New(key) //nolint:goerr113

		t.Run("Message", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg) }, func(l *slog.Logger) { l.Info(msg) })
		})
		t.Run("Binary", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Binary(key, bytesVal1)) }, func(l *slog.Logger) { l.Info(msg, key, bytesVal1) })
		})
		t.Run("Bool", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Bool(key, boolVal1)) }, func(l *slog.Logger) { l.Info(msg, key, boolVal1) })
		})
		t.Run("Boolp", func(t *testing.T) {
			ptr := &boolVal1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Boolp(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("ByteString", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.ByteString(key, bytesVal1)) }, func(l *slog.Logger) { l.Info(msg, key, types.ByteString(bytesVal1)) })
		})
		t.Run("Complex128", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Complex128(key, complex128Val1)) }, func(l *slog.Logger) { l.Info(msg, key, complex128Val1) })
		})
		t.Run("Complex128p", func(t *testing.T) {
			ptr := &complex128Val1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Complex128p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Complex64", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Complex64(key, complex64Val1)) }, func(l *slog.Logger) { l.Info(msg, key, complex64Val1) })
		})
		t.Run("Complex64p", func(t *testing.T) {
			ptr := &complex64Val1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Complex64p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Float64", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Float64(key, float64Val1)) }, func(l *slog.Logger) { l.Info(msg, key, float64Val1) })
		})
		t.Run("Float64p", func(t *testing.T) {
			ptr := &float64Val1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Float64p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Float32", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Float32(key, float32Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Float32(float32Val1)) })
		})
		t.Run("Float32p", func(t *testing.T) {
			ptr := &float32Val1
			sv := types.Float32(float32Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Float32p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Int", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int(key, intVal1)) }, func(l *slog.Logger) { l.Info(msg, key, intVal1) })
		})
		t.Run("Intp", func(t *testing.T) {
			ptr := &intVal1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Intp(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Int64", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int64(key, int64Val1)) }, func(l *slog.Logger) { l.Info(msg, key, int64Val1) })
		})
		t.Run("Int64p", func(t *testing.T) {
			ptr := &int64Val1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int64p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Int32", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int32(key, int32Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Int32(int32Val1)) })
		})
		t.Run("Int32p", func(t *testing.T) {
			ptr := &int32Val1
			sv := types.Int32(int32Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int32p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Int16", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int16(key, int16Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Int16(int16Val1)) })
		})
		t.Run("Int16p", func(t *testing.T) {
			ptr := &int16Val1
			sv := types.Int16(int16Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int16p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Int8", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int8(key, int8Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Int8(int8Val1)) })
		})
		t.Run("Int8p", func(t *testing.T) {
			ptr := &int8Val1
			sv := types.Int8(int8Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int8p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("String", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.String(key, msg)) }, func(l *slog.Logger) { l.Info(msg, key, msg) })
		})
		t.Run("Stringp", func(t *testing.T) {
			ptr := &msg

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Stringp(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Uint", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint(key, uintVal1)) }, func(l *slog.Logger) { l.Info(msg, key, uintVal1) })
		})
		t.Run("Uintp", func(t *testing.T) {
			ptr := &uintVal1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uintp(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Uint64", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint64(key, uint64Val1)) }, func(l *slog.Logger) { l.Info(msg, key, uint64Val1) })
		})
		t.Run("Uint64p", func(t *testing.T) {
			ptr := &uint64Val1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint64p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		t.Run("Uint32", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint32(key, uint32Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Uint32(uint32Val1)) })
		})
		t.Run("Uint32p", func(t *testing.T) {
			ptr := &uint32Val1
			sv := types.Uint32(uint32Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint32p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Uint16", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint16(key, uint16Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Uint16(uint16Val1)) })
		})
		t.Run("Uint16p", func(t *testing.T) {
			ptr := &uint16Val1
			sv := types.Uint16(uint16Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint16p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Uint8", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint8(key, uint8Val1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Uint8(uint8Val1)) })
		})
		t.Run("Uint8p", func(t *testing.T) {
			ptr := &uint8Val1
			sv := types.Uint8(uint8Val1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint8p(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Uintptr", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uintptr(key, uintptrVal1)) }, func(l *slog.Logger) { l.Info(msg, key, types.Uintptr(uintptrVal1)) })
		})
		t.Run("Uintptrp", func(t *testing.T) {
			ptr := &uintptrVal1
			sv := types.Uintptr(uintptrVal1)
			sptr := &sv

			if nilPtr {
				ptr = nil
				sptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uintptrp(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, sptr) })
		})
		t.Run("Reflect", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Reflect(key, referenceVal)) }, func(l *slog.Logger) { l.Info(msg, key, referenceVal) })
			// below will not match as zaphandler will dereference the pointer but zap will not
			// match(t, func(l *zap.Logger) { l.Info(msg, zap.Reflect(key, &ref)) }, func(l *slog.Logger) { l.Info(msg, key, &ref) })
		})
		t.Run("Stringer", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Stringer(key, stringerVal)) }, func(l *slog.Logger) { l.Info(msg, key, stringerVal) })
		})
		t.Run("Time", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Time(key, timeVal1)) }, func(l *slog.Logger) { l.Info(msg, key, timeVal1) })
		})
		t.Run("Timep", func(t *testing.T) {
			ptr := &timeVal1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Timep(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		// TODO: Stack
		// TODO: StackSkip
		t.Run("Duration", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Duration(key, durationVal1)) }, func(l *slog.Logger) { l.Info(msg, key, durationVal1) })
		})
		t.Run("Durationp", func(t *testing.T) {
			ptr := &durationVal1

			if nilPtr {
				ptr = nil
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Durationp(key, ptr)) }, func(l *slog.Logger) { l.Info(msg, key, ptr) })
		})
		// TODO: Object
		// TODO: Objects
		// TODO: ObjectValues
		// TODO: Stringers
		// TODO: Inline
		// TODO: Any
		// TODO: Namespace
		t.Run("Error", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Error(errVal1)) }, func(l *slog.Logger) { l.Info(msg, zaphandler.ErrorKey, errVal1) })
		})
		t.Run("NamedError", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.NamedError(key, errVal1)) }, func(l *slog.Logger) { l.Info(msg, key, errVal1) })
		})
		t.Run("Skip", func(t *testing.T) {
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Skip()) }, func(l *slog.Logger) { l.Info(msg, types.Skip()) })
		})
		t.Run("Array", func(t *testing.T) {
			v, ok := zap.Bools("", []bool{boolVal1, boolVal2}).Interface.(zapcore.ArrayMarshaler)
			if !ok {
				t.Fail()
			}

			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Array(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Bools", func(t *testing.T) {
			v := []bool{boolVal1, boolVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Bools(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("ByteStrings", func(t *testing.T) {
			v := [][]byte{bytesVal1, bytesVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.ByteStrings(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Complex128s", func(t *testing.T) {
			v := []complex128{complex128Val1, complex128Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Complex128s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Complex64s", func(t *testing.T) {
			v := []complex64{complex64Val1, complex64Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Complex64s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Durations", func(t *testing.T) {
			v := []time.Duration{durationVal1, durationVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Durations(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Float64s", func(t *testing.T) {
			v := []float64{float64Val1, float64Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Float64s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Float32s", func(t *testing.T) {
			v := []float32{float32Val1, float32Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Float32s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Ints", func(t *testing.T) {
			v := []int{intVal1, intVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Ints(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Int64s", func(t *testing.T) {
			v := []int64{int64Val1, int64Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int64s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Int32s", func(t *testing.T) {
			v := []int32{int32Val1, int32Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int32s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Int16s", func(t *testing.T) {
			v := []int16{int16Val1, int16Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int16s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Int8s", func(t *testing.T) {
			v := []int8{int8Val1, int8Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Int8s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Strings", func(t *testing.T) {
			v := []string{msg, key}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Strings(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Times", func(t *testing.T) {
			v := []time.Time{timeVal1, timeVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Times(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Uints", func(t *testing.T) {
			v := []uint{uintVal1, uintVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uints(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Uint64s", func(t *testing.T) {
			v := []uint64{uint64Val1, uint64Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint64s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Uint32s", func(t *testing.T) {
			v := []uint32{uint32Val1, uint32Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint32s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Uint16s", func(t *testing.T) {
			v := []uint16{uint16Val1, uint16Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint16s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Uint8s", func(t *testing.T) {
			v := []uint8{uint8Val1, uint8Val2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uint8s(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, types.Uint8s(v)) })
		})
		t.Run("Uintptrs", func(t *testing.T) {
			v := []uintptr{uintptrVal1, uintptrVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Uintptrs(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
		t.Run("Errors", func(t *testing.T) {
			v := []error{errVal1, errVal2}
			matchParallel(t, func(l *zap.Logger) { l.Info(msg, zap.Errors(key, v)) }, func(l *slog.Logger) { l.Info(msg, key, v) })
		})
	})
}
