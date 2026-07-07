package notify

import (
	"reflect"
	"testing"
)

func TestNewSettingsURLOrder(t *testing.T) {
	// URL は添字の昇順。空値 (未定義の環境変数展開等) はスキップする
	flat := map[string]string{
		"stfw_webhooks_urls_10":  "http://example.com/10",
		"stfw_webhooks_urls_0":   "http://example.com/0",
		"stfw_webhooks_urls_1":   "",
		"stfw_webhooks_urls_2":   "http://example.com/2",
		"stfw_webhooks_on_start": "true",
	}
	s := NewSettings(flat)
	want := []string{"http://example.com/0", "http://example.com/2", "http://example.com/10"}
	if !reflect.DeepEqual(s.URLs(), want) {
		t.Errorf("URLs() = %v, want %v", s.URLs(), want)
	}
	if !s.Enabled() {
		t.Error("Enabled() = false, want true")
	}
}

func TestNewSettingsDisabled(t *testing.T) {
	// URL 未設定 (または全て空) なら一切送信しない
	for _, flat := range []map[string]string{
		{"stfw_webhooks_on_start": "true", "stfw_webhooks_on_success": "true", "stfw_webhooks_on_error": "true"},
		{"stfw_webhooks_urls_0": "", "stfw_webhooks_on_start": "true"},
	} {
		s := NewSettings(flat)
		if s.Enabled() {
			t.Errorf("Enabled() = true, want false (flat=%v)", flat)
		}
		if s.ShouldNotify(Notification{Event: EventStart, Status: "Started"}) {
			t.Errorf("ShouldNotify(start) = true, want false (flat=%v)", flat)
		}
	}
}

func TestShouldNotify(t *testing.T) {
	// 送信判定: start は on_start、end は status に応じて on_success / on_error
	// (v0.2 の `-eq` 比較バグを修正した「抑制が機能する」仕様)
	tests := []struct {
		name      string
		onStart   string
		onSuccess string
		onError   string
		event     EventKind
		status    string
		want      bool
	}{
		{name: "start 有効", onStart: "true", event: EventStart, status: "Started", want: true},
		{name: "start 抑制", onStart: "false", event: EventStart, status: "Started", want: false},
		{name: "start 未設定は抑制", event: EventStart, status: "Started", want: false},
		{name: "success 有効", onSuccess: "true", event: EventEnd, status: "Success", want: true},
		{name: "success 抑制", onSuccess: "false", onError: "true", event: EventEnd, status: "Success", want: false},
		{name: "error 有効", onError: "true", event: EventEnd, status: "Error", want: true},
		{name: "error 抑制", onSuccess: "true", onError: "false", event: EventEnd, status: "Error", want: false},
	}
	for _, tt := range tests {
		flat := map[string]string{
			"stfw_webhooks_urls_0":     "http://example.com/hook",
			"stfw_webhooks_on_start":   tt.onStart,
			"stfw_webhooks_on_success": tt.onSuccess,
			"stfw_webhooks_on_error":   tt.onError,
		}
		s := NewSettings(flat)
		got := s.ShouldNotify(Notification{Event: tt.event, Status: tt.status})
		if got != tt.want {
			t.Errorf("%s: ShouldNotify() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
