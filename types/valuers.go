package types

import (
	"log/slog"
	"math"

	"go.uber.org/zap/zapcore"
)

var _ slog.LogValuer = ByteString{}

type ByteString []byte

func (v ByteString) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.ByteStringType, Interface: []byte(v)})
}

var _ slog.LogValuer = Float32(0)

type Float32 float32

func (v Float32) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Float32Type, Integer: int64(math.Float32bits(float32(v)))})
}

var _ slog.LogValuer = Int32(0)

type Int32 int32

func (v Int32) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Int32Type, Integer: int64(v)})
}

var _ slog.LogValuer = Int16(0)

type Int16 int16

func (v Int16) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Int16Type, Integer: int64(v)})
}

var _ slog.LogValuer = Int8(0)

type Int8 int8

func (v Int8) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Int8Type, Integer: int64(v)})
}

var _ slog.LogValuer = Uint32(0)

type Uint32 uint32

func (v Uint32) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Uint32Type, Integer: int64(v)})
}

var _ slog.LogValuer = Uint16(0)

type Uint16 uint16

func (v Uint16) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Uint16Type, Integer: int64(v)})
}

var _ slog.LogValuer = Uint8(0)

type Uint8 uint8

func (v Uint8) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.Uint8Type, Integer: int64(v)})
}

var _ slog.LogValuer = Uintptr(0)

type Uintptr uintptr

func (v Uintptr) LogValue() slog.Value {
	return slog.AnyValue(FieldType{Type: zapcore.UintptrType, Integer: int64(v)})
}

var _ slog.LogValuer = Array{}

type Array struct{ zapcore.ArrayMarshaler }

func (v Array) FieldType() FieldType {
	return FieldType{Type: zapcore.ArrayMarshalerType, Interface: v.ArrayMarshaler}
}

func (v Array) LogValue() slog.Value {
	return slog.AnyValue(v.FieldType())
}

func Skip() slog.Attr {
	return slog.Any("", zapcore.Field{Type: zapcore.SkipType})
}
