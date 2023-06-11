package types

import (
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"time"

	"go.uber.org/zap/zapcore"
)

type FieldType struct {
	Type      zapcore.FieldType
	Integer   int64
	String    string
	Interface any
}

func (t FieldType) Field(key string) zapcore.Field {
	return zapcore.Field{
		Key:       key,
		Type:      t.Type,
		Integer:   t.Integer,
		String:    t.String,
		Interface: t.Interface,
	}
}

func handleAny(val any) (FieldType, bool) {
	out := FieldType{Interface: val}

	switch typed := val.(type) {
	case zapcore.Field:
		return FieldType{
			Type:      typed.Type,
			Integer:   typed.Integer,
			String:    typed.String,
			Interface: typed.Interface,
		}, true
	case *time.Time, *time.Duration:
		// Do not match it with fmt.Stringer
	case error:
		out.Type = zapcore.ErrorType
	case fmt.Stringer:
		out.Type = zapcore.StringerType
	case []byte:
		out.Type = zapcore.BinaryType
	case complex128:
		out.Type = zapcore.Complex128Type
	case complex64:
		out.Type = zapcore.Complex64Type
	}

	return out, out.Type != zapcore.UnknownType
}

func isNil(val any) bool {
	rval := reflect.ValueOf(val)

	switch rval.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Pointer,
		reflect.UnsafePointer,
		reflect.Interface,
		reflect.Slice:
		return rval.IsNil()
	default:
		return val == nil
	}
}

func handleReflect(val reflect.Value) (FieldType, bool) {
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return FieldType{Type: zapcore.ReflectType}, true
		}

		switch v := val.Elem().Interface().(type) {
		case slog.LogValuer:
			return NewFieldType(v.LogValue()), true
		default:
			return NewFieldType(slog.AnyValue(v)), true
		}
	}

	return FieldType{}, false
}

func anyType(val any) FieldType {
	if got, ok := handleAny(val); ok {
		return got
	}

	if got, ok := ArrayMarahaler(val); ok {
		return Array{got}.FieldType()
	}

	if got, ok := handleReflect(reflect.ValueOf(val)); ok {
		return got
	}

	return FieldType{Type: zapcore.ReflectType, Interface: val}
}

func logValuerType(val slog.LogValuer) FieldType {
	if isNil(val) {
		return FieldType{Type: zapcore.ReflectType}
	}

	if got, ok := val.LogValue().Any().(FieldType); ok {
		return got
	}

	return NewFieldType(val.LogValue())
}

func timeType(t time.Time) FieldType {
	if t.Before(time.Unix(0, math.MinInt64)) || t.After(time.Unix(0, math.MaxInt64)) {
		return FieldType{Type: zapcore.TimeFullType, Interface: t}
	}

	return FieldType{Type: zapcore.TimeType, Integer: t.UnixNano(), Interface: t.Location()}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}

	return 0
}

func NewFieldType(val slog.Value) FieldType { //nolint:cyclop
	switch val.Kind() {
	case slog.KindLogValuer:
		return logValuerType(val.LogValuer())
	case slog.KindAny:
		return anyType(val.Any())
	case slog.KindBool:
		return FieldType{Type: zapcore.BoolType, Integer: boolToInt(val.Bool())}
	case slog.KindDuration:
		return FieldType{Type: zapcore.DurationType, Integer: int64(val.Duration())}
	case slog.KindFloat64:
		return FieldType{Type: zapcore.Float64Type, Integer: int64(math.Float64bits(val.Float64()))}
	case slog.KindInt64:
		return FieldType{Type: zapcore.Int64Type, Integer: val.Int64()}
	case slog.KindString:
		return FieldType{Type: zapcore.StringType, String: val.String()}
	case slog.KindTime:
		return timeType(val.Time())
	case slog.KindUint64:
		return FieldType{Type: zapcore.Uint64Type, Integer: int64(val.Uint64())}
	case slog.KindGroup:
		return FieldType{Type: zapcore.ObjectMarshalerType, Interface: Group(val.Group())}
	}

	panic("not reachable")
}
