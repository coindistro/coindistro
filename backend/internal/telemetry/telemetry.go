package telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracerProvider wraps OpenTelemetry tracer provider.
type TracerProvider struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	logger   *zap.Logger
}

// Config holds OpenTelemetry configuration.
type Config struct {
	Enabled     bool    `mapstructure:"enabled"`
	ServiceName string  `mapstructure:"service_name"`
	Endpoint    string  `mapstructure:"endpoint"`
	Environment string  `mapstructure:"environment"`
	SampleRate  float64 `mapstructure:"sample_rate"`
}

// NewTracerProvider creates a new OpenTelemetry tracer provider.
func NewTracerProvider(cfg Config, logger *zap.Logger) (*TracerProvider, error) {
	if !cfg.Enabled {
		logger.Info("telemetry disabled")
		return &TracerProvider{
			tracer: otel.Tracer(cfg.ServiceName),
			logger: logger,
		}, nil
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = "localhost:4318"
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "coindistro"
	}
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 0.1
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion("1.0.0"),
		attribute.String("environment", cfg.Environment),
	)

	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer(cfg.ServiceName)

	logger.Info("telemetry initialized",
		zap.String("service", cfg.ServiceName),
		zap.String("endpoint", cfg.Endpoint),
	)

	return &TracerProvider{
		provider: tp,
		tracer:   tracer,
		logger:   logger,
	}, nil
}

// Tracer returns the OpenTelemetry tracer.
func (tp *TracerProvider) Tracer() trace.Tracer {
	return tp.tracer
}

// Shutdown gracefully shuts down the tracer provider.
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider != nil {
		return tp.provider.Shutdown(ctx)
	}
	return nil
}

// StartSpan starts a new span with the given name and options.
func (tp *TracerProvider) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tp.tracer.Start(ctx, name, opts...)
}

// SpanContextFromContext extracts span context from a Go context.
func SpanContextFromContext(ctx context.Context) trace.SpanContext {
	return trace.SpanContextFromContext(ctx)
}

// SpanFromContext returns the current span from a context.
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanEvent adds an event to the current span.
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanError sets an error on the current span.
func SetSpanError(ctx context.Context, err error) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.RecordError(err)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
	}
}

// SetSpanAttributes sets attributes on the current span.
func SetSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if span := trace.SpanFromContext(ctx); span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// Predefined attribute keys for Coindistro spans.
var (
	AttrUserID     = attribute.Key("coindistro.user_id")
	AttrEmail      = attribute.Key("coindistro.email")
	AttrEntityType = attribute.Key("coindistro.entity_type")
	AttrEntityID   = attribute.Key("coindistro.entity_id")
	AttrAction     = attribute.Key("coindistro.action")
	AttrModule     = attribute.Key("coindistro.module")
	AttrDBQuery    = attribute.Key("coindistro.db_query")
	AttrCacheKey   = attribute.Key("coindistro.cache_key")
)
