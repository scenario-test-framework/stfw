package runscenario

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/notify"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// webhookNotifier はジャーナルイベントを webhook 通知へ投影して送信する。
// 投影・送信の失敗はログのみで実行結果へは影響しない (v0.2 の webhook_gateway 互換)。
type webhookNotifier struct {
	log       *slog.Logger
	settings  notify.Settings
	projector *notify.Projector
	sender    *gateway.WebhookSender
	now       func() time.Time
}

// newWebhookNotifier は通知設定と payload の実行環境情報を組み立てる。
func newWebhookNotifier(log *slog.Logger, cfg *repository.Config, projDir, version string, runID run.RunID, now func() time.Time) *webhookNotifier {
	ctx := notify.Context{
		Host:           gateway.LocalIP(),
		User:           currentUser(),
		Version:        version,
		ProjectVersion: cfg.Get("stfw_project_version"),
		ProjectHome:    projDir,
		WorkspaceDir:   filepath.Dir(repository.JournalPath(projDir, runID.String())),
	}
	return &webhookNotifier{
		log:       log,
		settings:  notify.NewSettings(cfg.Flat()),
		projector: notify.NewProjector(ctx),
		sender:    gateway.NewWebhookSender(log),
		now:       now,
	}
}

// onEvent はイベント 1 件を投影し、送信判定を通過した通知を非同期 POST する。
func (n *webhookNotifier) onEvent(ev run.Event) {
	if !n.settings.Enabled() {
		return
	}
	notifs, err := n.projector.Project(ev, n.now())
	if err != nil {
		n.log.Warn("[webhook] projection failed", "node_id", ev.NodeID, "message", err.Error())
		return
	}
	for _, notif := range notifs {
		if !n.settings.ShouldNotify(notif) {
			continue
		}
		body, err := json.Marshal(notif.Payload)
		if err != nil {
			n.log.Warn("[webhook] payload marshal failed", "node_id", ev.NodeID, "message", err.Error())
			continue
		}
		n.sender.AsyncPost(n.settings.URLs(), body)
	}
}

// wait は非同期送信の全完了を待つ (run 終了時に呼ぶ)。
func (n *webhookNotifier) wait() {
	n.sender.Wait()
}

// currentUser は実行ユーザー名を返す (v0.2 の whoami 相当)。
func currentUser() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return os.Getenv("USER")
}
