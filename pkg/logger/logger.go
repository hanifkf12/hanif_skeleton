package logger

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

type Fields struct {
	fields []zap.Field
}

func NewFields(eventName string) *Fields {
	return &Fields{
		fields: []zap.Field{
			zap.String("event", eventName),
		},
	}
}

func (f *Fields) Append(fields ...zap.Field) {
	f.fields = append(f.fields, fields...)
}

func (f *Fields) WithTrace(ctx context.Context) *Fields {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		f.Append(
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return f
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Setup() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	log, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
}

func Info(msg string, fields ...*Fields) {
	if len(fields) > 0 {
		log.Info(msg, fields[0].fields...)
		return
	}
	log.Info(msg)
}

func Error(msg string, fields ...*Fields) {
	if len(fields) > 0 {
		log.Error(msg, fields[0].fields...)
		return
	}
	log.Error(msg)
}

func Fatal(msg string, fields ...*Fields) {
	if len(fields) > 0 {
		log.Fatal(msg, fields[0].fields...)
		return
	}
	log.Fatal(msg)
}
