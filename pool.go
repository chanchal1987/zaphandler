package zaphandler

import (
	"log/slog"
	"runtime"
	"sync"

	"go.uber.org/zap/zapcore"
)

const (
	invalidType = "invalid pool type"
	poolSize    = 32
)

type (
	poolS[T any] struct{ item []T }
	poolT        struct{ stack, fields, attrs sync.Pool }
)

func initPool[T any](f func() []T) func() any { return func() any { return &poolS[T]{item: f()} } }
func newPool() *poolT {
	return &poolT{
		stack:  sync.Pool{New: initPool(func() []uintptr { return make([]uintptr, 1) })},
		fields: sync.Pool{New: initPool(func() []zapcore.Field { return make([]zapcore.Field, 0, poolSize) })},
		attrs:  sync.Pool{New: initPool(func() []slog.Attr { return make([]slog.Attr, 0, poolSize) })},
	}
}

func (p *poolT) caller(programCounter uintptr, addSource bool) zapcore.EntryCaller {
	if !addSource || programCounter == 0 {
		return zapcore.EntryCaller{}
	}

	stack, ok := p.stack.Get().(*poolS[uintptr])
	if !ok || len(stack.item) != 1 {
		panic(invalidType)
	}

	defer p.stack.Put(stack)

	stack.item[0] = programCounter
	frame, _ := runtime.CallersFrames((stack.item)).Next()

	return zapcore.EntryCaller{
		Defined:  true,
		PC:       frame.PC,
		File:     frame.File,
		Line:     frame.Line,
		Function: frame.Function,
	}
}

func (p *poolT) withFields(fieldsF func([]zapcore.Field)) {
	fields, ok := p.fields.Get().(*poolS[zapcore.Field])
	if !ok {
		panic(invalidType)
	}

	defer p.fields.Put(fields)

	fieldsF(fields.item)
}

func (p *poolT) withAttrs(attrF func([]slog.Attr)) {
	attrs, ok := p.attrs.Get().(*poolS[slog.Attr])
	if !ok {
		panic(invalidType)
	}

	defer p.attrs.Put(attrs)

	attrF(attrs.item)
}
