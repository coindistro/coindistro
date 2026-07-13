package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide structured logging.
type Logger struct {
	*zap.Logger
}

// New creates a new Logger based on the provided configuration.
func New(level string, encoding string, outputPaths []string, errorOutputPaths []string) (*Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if encoding == "" {
		encoding = "json"
	}

	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}
	if len(errorOutputPaths) == 0 {
		errorOutputPaths = []string{"stderr"}
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: errorOutputPaths,
	}

	zapLogger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{Logger: zapLogger}, nil
}

// NewDefault creates a default development logger.
func NewDefault() *Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02T15:04:05.000Z0700"))
	}

	zapLogger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		// Fallback to basic logger
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			zap.InfoLevel,
		)
		return &Logger{Logger: zap.New(core, zap.AddCallerSkip(1))}
	}

	return &Logger{Logger: zapLogger}
}

// WithRequestID adds a request ID field to the logger.
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.With(zap.String("request_id", requestID))}
}

// WithComponent adds a component field to the logger.
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{Logger: l.With(zap.String("component", component))}
}

// WithFields adds multiple fields to the logger.
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.With(fields...)}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
