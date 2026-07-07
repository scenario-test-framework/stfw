package gateway

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

// webhookTimeout は webhook 1 リクエストのタイムアウト。
const webhookTimeout = 30 * time.Second

// WebhookSender は webhook payload を非同期に POST する
// (v0.2 の webhook_gateway.async_execute の置き換え)。
// 送信失敗はログのみでエラーにしない (v0.2 互換)。
// ログは Masker 経由の slog へ出力するためシークレットはマスクされる。
type WebhookSender struct {
	log    *slog.Logger
	client *http.Client
	wg     sync.WaitGroup
}

// NewWebhookSender は送信器を生成する。
func NewWebhookSender(log *slog.Logger) *WebhookSender {
	return &WebhookSender{
		log:    log,
		client: &http.Client{Timeout: webhookTimeout},
	}
}

// AsyncPost は payload 1 件を全 URL へ非同期 (goroutine) で POST する。
// 全送信の完了は Wait で待ち合わせる。
func (s *WebhookSender) AsyncPost(urls []string, body []byte) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for _, url := range urls {
			s.post(url, body)
		}
	}()
}

// Wait は非同期送信の全完了を待つ (run 終了時に呼ぶ)。
func (s *WebhookSender) Wait() {
	s.wg.Wait()
}

// post は payload を 1 URL へ POST する。
// v0.2 と同じく response body が "ok" の場合のみ成功として扱い、
// それ以外は警告ログを出力する (リターンコードへは影響しない)。
func (s *WebhookSender) post(url string, body []byte) {
	resp, err := s.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		s.log.Warn("[webhook] Error", "target", url, "message", err.Error())
		return
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Warn("[webhook] Error", "target", url, "message", err.Error())
		return
	}
	message := strings.TrimRight(string(raw), "\n")
	if message == "ok" {
		s.log.Debug("[webhook] Success", "target", url)
		return
	}
	s.log.Warn("[webhook] Error", "target", url, "message", message)
}
