package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	zapotlp "github.com/SigNoz/zap_otlp"
	zapotlpencoder "github.com/SigNoz/zap_otlp/zap_otlp_encoder"
	zapotlpsync "github.com/SigNoz/zap_otlp/zap_otlp_sync"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var log *zap.Logger
var otlpSyncer *zapotlpsync.OtelSyncer

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
	// Use zap_otlp's SpanCtx method to add trace context to logs
	f.Append(zapotlp.SpanCtx(ctx))
	return f
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

func Setup() {
	// Set up the OpenTelemetry connection
	conn, err := grpc.NewClient("localhost:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		// Fall back to stdout-only logging if OTLP connection fails
		fmt.Printf("Failed to connect to OpenTelemetry collector: %v, logging to stdout only\n", err)

		// Create standard production logger with JSON encoding
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		config.Encoding = "json"
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
		config.Sampling = &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		}
		log, err = config.Build()
		if err != nil {
			panic(fmt.Sprintf("failed to initialize logger: %v", err))
		}
		return
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create JSON encoder for console output
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create OTLP encoder for logs sent to SignOz
	otlpEncoder := zapotlpencoder.NewOTLPEncoder(encoderConfig)

	// Create OTLP syncer with options
	otlpSyncer = zapotlpsync.NewOtlpSyncer(conn, zapotlpsync.Options{
		BatchSize: 100,
	})

	// Create core with both encoders
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		zapcore.NewCore(otlpEncoder, zapcore.AddSync(otlpSyncer), zapcore.InfoLevel),
	)

	// Create logger with recommended options
	log = zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(zap.String("service.name", "hanif-skeleton")),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}),
	)
}

// Cleanup shuts down the logger and flushes any buffered logs
func Cleanup() {
	if log != nil {
		_ = log.Sync()
	}

	if otlpSyncer != nil {
		_ = otlpSyncer.Sync()
	}
}

func Info(msg string, fields ...*Fields) {
	if len(fields) > 0 {
		// Add service.name for consistent correlation with traces
		fields[0].Append(zap.String("service.name", "hanif-skeleton"))
		log.Info(msg, fields[0].fields...)
		return
	}
	log.Info(msg, zap.String("service.name", "hanif-skeleton"))
}

func Error(msg string, fields ...*Fields) {
	if len(fields) > 0 {
		// Add service.name for consistent correlation with traces
		fields[0].Append(zap.String("service.name", "hanif-skeleton"))
		log.Error(msg, fields[0].fields...)
		return
	}
	log.Error(msg, zap.String("service.name", "hanif-skeleton"))
}

func Fatal(msg string, fields ...*Fields) {
	if len(fields) > 0 {
		// Add service.name for consistent correlation with traces
		fields[0].Append(zap.String("service.name", "hanif-skeleton"))
		log.Fatal(msg, fields[0].fields...)
		return
	}
	log.Fatal(msg, zap.String("service.name", "hanif-skeleton"))
}
