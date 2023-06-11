package types

import (
	"log/slog"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func arrayMarshaler[T any](f func(string, []T) zap.Field, v []T) (zapcore.ArrayMarshaler, bool) {
	got, ok := f("", v).Interface.(zapcore.ArrayMarshaler)

	return got, ok
}

func ints(a any) (zapcore.ArrayMarshaler, bool) {
	switch arr := a.(type) {
	case []int:
		return arrayMarshaler(zap.Ints, arr)
	case []int64:
		return arrayMarshaler(zap.Int64s, arr)
	case []int32:
		return arrayMarshaler(zap.Int32s, arr)
	case []int16:
		return arrayMarshaler(zap.Int16s, arr)
	case []int8:
		return arrayMarshaler(zap.Int8s, arr)
	case []uint, []uint64, []uint32, []uint16, []uintptr:
		return uints(arr)
	}

	return nil, false
}

func uints(a any) (zapcore.ArrayMarshaler, bool) {
	switch arr := a.(type) {
	case []uint:
		return arrayMarshaler(zap.Uints, arr)
	case []uint64:
		return arrayMarshaler(zap.Uint64s, arr)
	case []uint32:
		return arrayMarshaler(zap.Uint32s, arr)
	case []uint16:
		return arrayMarshaler(zap.Uint16s, arr)
	case []uintptr:
		return arrayMarshaler(zap.Uintptrs, arr)
	}

	// []uint8 is handled separately
	return nil, false
}

func complexes(a any) (zapcore.ArrayMarshaler, bool) {
	switch arr := a.(type) {
	case []complex128:
		return arrayMarshaler(zap.Complex128s, arr)
	case []complex64:
		return arrayMarshaler(zap.Complex64s, arr)
	}

	return nil, false
}

func floats(a any) (zapcore.ArrayMarshaler, bool) {
	switch arr := a.(type) {
	case []float64:
		return arrayMarshaler(zap.Float64s, arr)
	case []float32:
		return arrayMarshaler(zap.Float32s, arr)
	}

	return nil, false
}

func times(a any) (zapcore.ArrayMarshaler, bool) {
	switch arr := a.(type) {
	case []time.Duration:
		return arrayMarshaler(zap.Durations, arr)
	case []time.Time:
		return arrayMarshaler(zap.Times, arr)
	}

	return nil, false
}

func ArrayMarahaler(val any) (zapcore.ArrayMarshaler, bool) {
	switch arr := val.(type) {
	case zapcore.ArrayMarshaler:
		return arr, true
	case []bool:
		return arrayMarshaler(zap.Bools, arr)
	case [][]byte:
		return arrayMarshaler(zap.ByteStrings, arr)
	case []complex128, []complex64:
		return complexes(arr)
	case []time.Duration, []time.Time:
		return times(val)
	case []float64, []float32:
		return floats(arr)
	case []string:
		return arrayMarshaler(zap.Strings, arr)
	case []int, []int64, []int32, []int16, []int8, []uint, []uint64, []uint32, []uint16, []uintptr:
		return ints(arr)
	case []error:
		return arrayMarshaler(zap.Errors, arr)
	}

	return nil, false
}

var _ slog.LogValuer = Uint8s{}

type Uint8s []uint8

func (v Uint8s) LogValue() slog.Value {
	if got, ok := arrayMarshaler(zap.Uint8s, v); ok {
		return slog.AnyValue(got)
	}

	return slog.Value{}
}
