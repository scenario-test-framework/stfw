package runscenario

import (
	"log/slog"
	"os"

	"github.com/scenario-test-framework/stfw/internal/domain/notify"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// otelNotifier はジャーナルイベントをスパン記述へ投影し、run 終了時に
// OTLP トレースとしてエクスポートする (1 run = 1 トレース)。
// 投影・送信の失敗はログのみで実行結果へは影響しない (SPEC-009-04)。
type otelNotifier struct {
	log       *slog.Logger
	projector *notify.Projector
	// exporter は送信先未設定時は nil (TracerProvider を組み立てず一切送信しない)。
	exporter *gateway.TraceExporter
}

// newOTelNotifier は送信先設定を解決してエクスポーターを組み立てる。
// OTel 標準環境変数 (OTEL_EXPORTER_OTLP_ENDPOINT / OTEL_EXPORTER_OTLP_TRACES_ENDPOINT)
// を優先し、未設定時は stfw.yml の stfw.otel.endpoint を使う。
// どちらも未設定なら無効 (SPEC-009-03)。
func newOTelNotifier(log *slog.Logger, cfg *repository.Config, version string) *otelNotifier {
	endpoint := cfg.Get("stfw_otel_endpoint")
	if otelEnvConfigured() {
		// 環境変数は otlptracehttp が自動で尊重するためオプション指定しない
		endpoint = ""
	} else if endpoint == "" {
		return &otelNotifier{log: log}
	}
	exporter, err := gateway.NewOTLPTraceExporter(log, endpoint, version)
	if err != nil {
		log.Warn("[otel] trace exporter setup failed", "message", err.Error())
		return &otelNotifier{log: log}
	}
	return &otelNotifier{log: log, projector: notify.NewProjector(), exporter: exporter}
}

// otelEnvConfigured は OTel 標準環境変数で送信先が指定されているかを返す。
func otelEnvConfigured() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
}

// enabled はトレースエクスポートが有効 (送信先設定済み) かを返す。
func (n *otelNotifier) enabled() bool { return n.exporter != nil }

// onEvent はイベント 1 件を投影し、ルート (run) の終了で確定した
// スパンツリーをエクスポートする。
func (n *otelNotifier) onEvent(ev run.Event) {
	if !n.enabled() {
		return
	}
	spans, err := n.projector.Apply(ev)
	if err != nil {
		n.log.Warn("[otel] projection failed", "node_id", ev.NodeID, "message", err.Error())
		return
	}
	if len(spans) > 0 {
		n.exporter.Export(spans)
	}
}

// close は未送信スパンを flush して TracerProvider を停止する (run 終了時に呼ぶ)。
func (n *otelNotifier) close() {
	if !n.enabled() {
		return
	}
	n.exporter.Shutdown()
}
