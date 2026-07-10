package runscenario

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/scenario-test-framework/stfw/internal/repository"
)

// testConfig は stfw.yml の内容から Config を組み立てる (空なら stfw.yml 無し)。
func testConfig(t *testing.T, stfwYml string) *repository.Config {
	t.Helper()
	projDir := t.TempDir()
	if stfwYml != "" {
		if err := os.WriteFile(filepath.Join(projDir, "stfw.yml"), []byte(stfwYml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg, _, err := repository.LoadConfig(projDir)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestNewOTelNotifierEndpointResolution(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	otelYml := "stfw:\n  otel:\n    endpoint: http://127.0.0.1:14318\n"

	tests := []struct {
		name        string
		stfwYml     string
		env         map[string]string
		wantEnabled bool
	}{
		{
			name:        "newOTelNotifier_環境変数もstfwymlも未設定の場合_無効であること",
			wantEnabled: false,
		},
		{
			name:        "newOTelNotifier_OTEL_EXPORTER_OTLP_ENDPOINT設定の場合_有効であること",
			env:         map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:14318"},
			wantEnabled: true,
		},
		{
			name:        "newOTelNotifier_OTEL_EXPORTER_OTLP_TRACES_ENDPOINT設定の場合_有効であること",
			env:         map[string]string{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://127.0.0.1:14318/v1/traces"},
			wantEnabled: true,
		},
		{
			name:        "newOTelNotifier_stfwymlのotelendpoint設定の場合_有効であること",
			stfwYml:     otelYml,
			wantEnabled: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			// テスト実行環境の OTel 変数の影響を除く
			t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
			t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			// Act
			n := newOTelNotifier(log, testConfig(t, tt.stfwYml), "1.0.0-test")
			defer n.close()
			// Assert
			if n.enabled() != tt.wantEnabled {
				t.Errorf("enabled = %t, want %t", n.enabled(), tt.wantEnabled)
			}
		})
	}
}
