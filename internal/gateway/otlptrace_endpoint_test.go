package gateway

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/notify"
)

// TestNewOTLPTraceExporterEndpointPath は stfw.otel.endpoint 由来の URL について、
// 実際に送信される HTTP リクエストのパスを実サーバーで検証する。
// otel SDK の WithEndpointURL はバージョンによって既定パスの扱いが変わるため (v1.45.0 以降は補完しない)、
// 依存更新で補完・保持の挙動が退行しないことを固定する。
func TestNewOTLPTraceExporterEndpointPath(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		wantPath string
	}{
		{name: "NewOTLPTraceExporter_パス無しURLの場合_v1tracesへ送信すること", path: "", wantPath: "/v1/traces"},
		{name: "NewOTLPTraceExporter_ルートパスのみの場合_v1tracesへ送信すること", path: "/", wantPath: "/v1/traces"},
		{name: "NewOTLPTraceExporter_独自パスを指定した場合_そのパスへ送信すること", path: "/custom/v1/traces", wantPath: "/custom/v1/traces"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			// 標準環境変数が設定されていると SDK がそちらを優先し得るため、空にして影響を断つ
			for _, k := range []string{
				"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
				"OTEL_EXPORTER_OTLP_PROTOCOL", "OTEL_EXPORTER_OTLP_TRACES_PROTOCOL",
			} {
				t.Setenv(k, "")
			}
			var mu sync.Mutex
			var gotPaths []string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				gotPaths = append(gotPaths, r.URL.Path)
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			exporter, err := NewOTLPTraceExporter(log, srv.URL+tc.path, "1.0.0-test")
			if err != nil {
				t.Fatalf("NewOTLPTraceExporter: %v", err)
			}
			start := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)
			spans := []notify.Span{{ID: "root", Name: "stfw run", Start: start, End: start.Add(time.Second)}}

			// Act
			exporter.Export(spans)
			exporter.Shutdown()

			// Assert
			mu.Lock()
			defer mu.Unlock()
			if len(gotPaths) == 0 {
				t.Fatal("no request reached the OTLP server")
			}
			for _, p := range gotPaths {
				if p != tc.wantPath {
					t.Errorf("request path = %q, want %q", p, tc.wantPath)
				}
			}
		})
	}
}
