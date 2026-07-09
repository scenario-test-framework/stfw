package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/scenario-test-framework/stfw/internal/domain/notify"
)

// otelShutdownTimeout は run 終了時の ForceFlush / Shutdown のタイムアウト。
// 送信先が到達不能でも実行をこれ以上待たせない (SPEC-009-04)。
const otelShutdownTimeout = 10 * time.Second

// TraceExporter はスパン記述 (notify.Span) を OTel スパンへ変換して
// OTLP エクスポートする。送信失敗はログのみでエラーにしない
// (v0.2 由来の webhook と同じ「通知の失敗は実行を失敗させない」方針)。
type TraceExporter struct {
	log *slog.Logger
	tp  *sdktrace.TracerProvider
}

// NewOTLPTraceExporter は OTLP/HTTP エクスポーターを組み立てる。
// endpointURL が空の場合は OTel 標準環境変数 (OTEL_EXPORTER_OTLP_ENDPOINT /
// OTEL_EXPORTER_OTLP_TRACES_ENDPOINT) の設定に従う (otlptracehttp が自動で尊重する)。
func NewOTLPTraceExporter(log *slog.Logger, endpointURL, version string) (*TraceExporter, error) {
	var opts []otlptracehttp.Option
	if endpointURL != "" {
		// パス無し URL は /v1/traces が補完される (OTEL_EXPORTER_OTLP_ENDPOINT と同等)
		opts = append(opts, otlptracehttp.WithEndpointURL(endpointURL))
	}
	exp, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return NewTraceExporter(log, exp, version), nil
}

// NewTraceExporter は任意の SpanExporter で TracerProvider を組み立てる
// (tracetest.InMemoryExporter を注入するテスト経路と本番経路の共通部)。
func NewTraceExporter(log *slog.Logger, exp sdktrace.SpanExporter, version string) *TraceExporter {
	res := resource.NewWithAttributes(semconv.SchemaURL,
		semconv.ServiceName("stfw"),
		semconv.ServiceVersion(version),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	return &TraceExporter{log: log, tp: tp}
}

// Export はスパン記述 (親が先の順) を OTel スパンへ変換して記録する。
// スパンの開始・終了時刻は記述の時刻 (= ジャーナルイベントの時刻) をそのまま使う。
func (e *TraceExporter) Export(spans []notify.Span) {
	tracer := e.tp.Tracer("github.com/scenario-test-framework/stfw")
	ctxs := map[string]context.Context{}
	for _, s := range spans {
		parent := context.Background()
		if s.ParentID != "" {
			p, ok := ctxs[s.ParentID]
			if !ok {
				e.log.Warn("[otel] parent span is not exported", "span_id", s.ID, "parent_id", s.ParentID)
				continue
			}
			parent = p
		}
		ctx, span := tracer.Start(parent, s.Name,
			trace.WithTimestamp(s.Start),
			trace.WithAttributes(otelAttrs(s.Attrs)...),
		)
		switch s.Status {
		case notify.SpanStatusOK:
			span.SetStatus(codes.Ok, "")
		case notify.SpanStatusError:
			span.SetStatus(codes.Error, s.StatusMessage)
		}
		span.End(trace.WithTimestamp(s.End))
		ctxs[s.ID] = ctx
	}
}

// Shutdown は未送信スパンを ForceFlush してから TracerProvider を停止する。
// エラーは警告ログのみで実行結果へは影響しない (SPEC-009-04)。
func (e *TraceExporter) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), otelShutdownTimeout)
	defer cancel()
	if err := e.tp.ForceFlush(ctx); err != nil {
		e.log.Warn("[otel] trace export failed", "message", err.Error())
	}
	if err := e.tp.Shutdown(ctx); err != nil {
		e.log.Warn("[otel] tracer provider shutdown failed", "message", err.Error())
	}
}

// otelAttrs はスパン記述の属性 (string | int64) を OTel の属性へ変換する。
func otelAttrs(attrs []notify.Attr) []attribute.KeyValue {
	kvs := make([]attribute.KeyValue, 0, len(attrs))
	for _, a := range attrs {
		switch v := a.Value.(type) {
		case int64:
			kvs = append(kvs, attribute.Int64(a.Key, v))
		case string:
			kvs = append(kvs, attribute.String(a.Key, v))
		default:
			kvs = append(kvs, attribute.String(a.Key, fmt.Sprintf("%v", a.Value)))
		}
	}
	return kvs
}
