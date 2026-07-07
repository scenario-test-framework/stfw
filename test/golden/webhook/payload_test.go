// Package webhook は webhook payload の v0.2 互換ゴールデンテスト。
// 旧テンプレート (src/config/webhook/*.yml + scripts プラグインの詳細テンプレート)
// から手動展開した期待 JSON と、httptest.Server で受信した payload を比較する。
// 実行時刻・run_id・実行環境に依存する値は正規化して比較する。
package webhook

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/repository"
	"github.com/scenario-test-framework/stfw/internal/usecase/runscenario"
)

// runIDPattern は run_id (`_{yyyymmddhhmmss}_{pid}`) の出現箇所。
var runIDPattern = regexp.MustCompile(`_\d{14}_\d+`)

// receiver は payload を記録する webhook 受信サーバ。
type receiver struct {
	mu       sync.Mutex
	payloads [][]byte
	badReqs  []string
}

// newReceiver は response body "ok" を返す受信サーバを起動する
// (v0.2 の webhook_gateway は "ok" 以外を送信失敗として扱う)。
func newReceiver(t *testing.T) (*receiver, *httptest.Server) {
	t.Helper()
	r := &receiver{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		r.mu.Lock()
		defer r.mu.Unlock()
		if err != nil || req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/json" {
			r.badReqs = append(r.badReqs, fmt.Sprintf("method=%s content-type=%s err=%v", req.Method, req.Header.Get("Content-Type"), err))
		}
		r.payloads = append(r.payloads, body)
		fmt.Fprint(w, "ok")
	}))
	t.Cleanup(srv.Close)
	return r, srv
}

// projectOption は fixture プロジェクトの生成オプション。
type projectOption struct {
	webhooks  string // stfw.yml の webhooks セクション (空なら未設定)
	firstStep string // 100_step1 の内容
}

// writeProject は run > scenario(demo) > bizdate x1 > process(scripts) x1 x steps x2
// の fixture プロジェクトを組み立てる。
func writeProject(t *testing.T, opt projectOption) string {
	t.Helper()
	projDir := t.TempDir()

	stfwYml := "stfw:\n  project_version: 0.1.0\n  loglevel: \"error\"\n  timezone: \"Asia/Tokyo\"\n" + opt.webhooks
	files := map[string]string{
		"stfw.yml":                                stfwYml,
		"scenario/demo/metadata.yml":              "description:\n",
		"scenario/demo/_10_99990101/metadata.yml": "description:\n",
		"scenario/demo/_10_99990101/_10_pre_scripts/metadata.yml":      "description:\n",
		"scenario/demo/_10_99990101/_10_pre_scripts/config/config.yml": "stfw:\n  process:\n    scripts:\n      some_key: value\n",
		"scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_step1": opt.firstStep,
		"scenario/demo/_10_99990101/_10_pre_scripts/scripts/200_step2": "#!/bin/bash\nexit 0\n",
	}
	for path, content := range files {
		full := filepath.Join(projDir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(0o644)
		if strings.Contains(path, "/scripts/") {
			mode = 0o755
		}
		if err := os.WriteFile(full, []byte(content), mode); err != nil {
			t.Fatal(err)
		}
	}
	return projDir
}

// webhooksSection は stfw.yml の webhooks 設定を組み立てる。
func webhooksSection(url string, onStart, onSuccess, onError bool) string {
	return fmt.Sprintf("  webhooks:\n    urls:\n      - %s\n    on_start: %t\n    on_success: %t\n    on_error: %t\n",
		url, onStart, onSuccess, onError)
}

// runScenario は fixture プロジェクトで stfw run demo 相当を実行する。
// wantErr は実行結果 (Error 終了) の期待値。
func runScenario(t *testing.T, projDir string, wantErr bool) {
	t.Helper()
	cfg, _, err := repository.LoadConfig(projDir)
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	err = runscenario.Run(log, io.Discard, io.Discard, projDir, cfg, "1.0.0-test", []string{"demo"}, false, time.Now)
	if wantErr && err == nil {
		t.Fatal("run should finish with error")
	}
	if !wantErr && err != nil {
		t.Fatal(err)
	}
}

// classify は受信 payload を「{type}_{start|end}」で分類する。
func classify(t *testing.T, payloads [][]byte) map[string]map[string]any {
	t.Helper()
	classified := map[string]map[string]any{}
	for _, raw := range payloads {
		var payload map[string]any
		if err := json.Unmarshal(raw, &payload); err != nil {
			t.Fatalf("payload %s: %v", raw, err)
		}
		body, ok := payload["payload"].(map[string]any)
		if !ok {
			t.Fatalf("payload key not found: %s", raw)
		}
		kind := "end"
		if body["status"] == "Started" {
			kind = "start"
		}
		key := fmt.Sprintf("%s_%s", body["type"], kind)
		if _, exists := classified[key]; exists {
			t.Fatalf("payload %s is duplicated", key)
		}
		classified[key] = payload
	}
	return classified
}

// normalize は実行時刻・run_id・実行環境依存の値を正規化する。
func normalize(t *testing.T, payload map[string]any, projDir string) map[string]any {
	t.Helper()
	stfw, ok := payload["payload"].(map[string]any)["stfw"].(map[string]any)
	if !ok {
		t.Fatalf("stfw key not found: %v", payload)
	}
	stfw["host"] = "HOST"
	stfw["user"] = "USER"
	stfw["version"] = "VERSION"
	return normalizeValue(payload, projDir).(map[string]any)
}

func normalizeValue(v any, projDir string) any {
	switch val := v.(type) {
	case map[string]any:
		for k, child := range val {
			switch k {
			case "create_time", "start_time", "end_time":
				if s, ok := child.(string); ok && s != "" {
					val[k] = "TS"
					continue
				}
			case "processing_time":
				if s, ok := child.(string); ok && s != "" {
					val[k] = "PT"
					continue
				}
			}
			val[k] = normalizeValue(child, projDir)
		}
		return val
	case []any:
		for i, child := range val {
			val[i] = normalizeValue(child, projDir)
		}
		return val
	case string:
		s := strings.ReplaceAll(val, projDir, "PROJ_DIR")
		return runIDPattern.ReplaceAllString(s, "RUN_ID")
	default:
		return v
	}
}

// assertGolden は正規化済み payload をゴールデン JSON と比較する。
func assertGolden(t *testing.T, goldenName string, payload map[string]any) {
	t.Helper()
	raw, err := os.ReadFile(goldenName + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var golden map[string]any
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("%s.json: %v", goldenName, err)
	}
	want, err := json.MarshalIndent(golden, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("payload %s mismatch\n--- want\n%s\n--- got\n%s", goldenName, want, got)
	}
}

func TestPayloadGoldenSuccess(t *testing.T) {
	// 正常系: run / scenario / bizdate / process の各階層 x start / end の 8 通知
	recv, srv := newReceiver(t)
	projDir := writeProject(t, projectOption{
		webhooks:  webhooksSection(srv.URL, true, true, true),
		firstStep: "#!/bin/bash\nexit 0\n",
	})
	runScenario(t, projDir, false)

	if len(recv.badReqs) > 0 {
		t.Fatalf("bad requests: %v", recv.badReqs)
	}
	if len(recv.payloads) != 8 {
		t.Fatalf("received = %d, want 8", len(recv.payloads))
	}
	classified := classify(t, recv.payloads)
	for _, key := range []string{
		"run_start", "run_end", "scenario_start", "scenario_end",
		"bizdate_start", "bizdate_end", "process_start", "process_end",
	} {
		payload, ok := classified[key]
		if !ok {
			t.Errorf("payload %s is not received", key)
			continue
		}
		assertGolden(t, key, normalize(t, payload, projDir))
	}
}

func TestPayloadGoldenError(t *testing.T) {
	// 異常系: step1 がエラー終了 → step2 は Blocked、全階層の end が Error
	recv, srv := newReceiver(t)
	projDir := writeProject(t, projectOption{
		webhooks:  webhooksSection(srv.URL, true, true, true),
		firstStep: "#!/bin/bash\nexit 6\n",
	})
	runScenario(t, projDir, true)

	if len(recv.payloads) != 8 {
		t.Fatalf("received = %d, want 8", len(recv.payloads))
	}
	classified := classify(t, recv.payloads)
	// start payload は正常系と同一
	for _, key := range []string{"run_start", "scenario_start", "bizdate_start", "process_start"} {
		payload, ok := classified[key]
		if !ok {
			t.Errorf("payload %s is not received", key)
			continue
		}
		assertGolden(t, key, normalize(t, payload, projDir))
	}
	for _, key := range []string{"run_end", "scenario_end", "bizdate_end", "process_end"} {
		payload, ok := classified[key]
		if !ok {
			t.Errorf("payload %s is not received", key)
			continue
		}
		assertGolden(t, key+"_error", normalize(t, payload, projDir))
	}
}

func TestSuppression(t *testing.T) {
	// 送信抑制 (v0.2 の -eq 比較バグを修正した「抑制が機能する」仕様)
	tests := []struct {
		name      string
		webhooks  func(url string) string
		failStep  bool
		wantCount int
		wantKinds []string // 受信 payload の期待キー (nil なら件数のみ検証)
	}{
		{
			name:      "URL 未設定なら一切送信しない",
			webhooks:  func(url string) string { return "" },
			wantCount: 0,
		},
		{
			name:      "on_start=false で start が送信されない",
			webhooks:  func(url string) string { return webhooksSection(url, false, true, true) },
			wantCount: 4,
			wantKinds: []string{"bizdate_end", "process_end", "run_end", "scenario_end"},
		},
		{
			name:      "on_success=false で Success の end が送信されない",
			webhooks:  func(url string) string { return webhooksSection(url, true, false, true) },
			wantCount: 4,
			wantKinds: []string{"bizdate_start", "process_start", "run_start", "scenario_start"},
		},
		{
			name:      "on_error=false で Error の end が送信されない",
			webhooks:  func(url string) string { return webhooksSection(url, false, true, false) },
			failStep:  true,
			wantCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recv, srv := newReceiver(t)
			firstStep := "#!/bin/bash\nexit 0\n"
			if tt.failStep {
				firstStep = "#!/bin/bash\nexit 6\n"
			}
			projDir := writeProject(t, projectOption{webhooks: tt.webhooks(srv.URL), firstStep: firstStep})
			runScenario(t, projDir, tt.failStep)

			if len(recv.payloads) != tt.wantCount {
				t.Fatalf("received = %d, want %d", len(recv.payloads), tt.wantCount)
			}
			if tt.wantKinds == nil {
				return
			}
			classified := classify(t, recv.payloads)
			for _, key := range tt.wantKinds {
				if _, ok := classified[key]; !ok {
					t.Errorf("payload %s is not received (got %v)", key, keysOf(classified))
				}
			}
		})
	}
}

func keysOf(m map[string]map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
