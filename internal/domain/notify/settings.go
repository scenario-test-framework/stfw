// Package notify は通知管理 (Supporting BC) のドメインルールを持つ。
// ジャーナルイベントを webhook payload へ投影し、送信可否を判定する。
// payload の JSON 構造は v0.2 の src/config/webhook/*.yml テンプレートと互換。
package notify

import (
	"sort"
	"strconv"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// urlKeyPrefix はフラット化済み設定の webhook URL キー接頭辞
// (stfw.webhooks.urls のリスト添字展開)。
const urlKeyPrefix = "stfw_webhooks_urls_"

// Settings は webhook 通知設定 (stfw.webhooks.*)。
// on_start / on_success / on_error は "true" のときのみ有効
// (v0.2 の webhook_spec は `-eq` の文字列比較バグで抑制が機能していなかったが、
// 「設定で抑制できる」仕様側に修正済み)。
type Settings struct {
	urls      []string
	onStart   bool
	onSuccess bool
	onError   bool
}

// NewSettings はフラット化済み設定から通知設定を組み立てる。
// URL は添字の昇順で並べ、空値はスキップする (v0.2 の webhook_gateway と同じ)。
func NewSettings(flat map[string]string) Settings {
	type indexed struct {
		key string
		idx int
	}
	var keys []indexed
	for k := range flat {
		if !strings.HasPrefix(k, urlKeyPrefix) {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimPrefix(k, urlKeyPrefix))
		if err != nil {
			continue
		}
		keys = append(keys, indexed{key: k, idx: idx})
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].idx < keys[j].idx })

	s := Settings{
		onStart:   flat["stfw_webhooks_on_start"] == "true",
		onSuccess: flat["stfw_webhooks_on_success"] == "true",
		onError:   flat["stfw_webhooks_on_error"] == "true",
	}
	for _, k := range keys {
		if flat[k.key] == "" {
			continue
		}
		s.urls = append(s.urls, flat[k.key])
	}
	return s
}

// URLs は送信先 URL のリストを返す。
func (s Settings) URLs() []string { return s.urls }

// Enabled は webhook 通知が有効 (URL が 1 件以上設定済み) かを返す。
func (s Settings) Enabled() bool { return len(s.urls) > 0 }

// ShouldNotify は通知 1 件の送信可否を判定する。
// URL 未設定なら送信しない。start は on_start=true のとき、
// end は status=Success なら on_success=true / status=Error なら on_error=true のとき送信する。
func (s Settings) ShouldNotify(n Notification) bool {
	if !s.Enabled() {
		return false
	}
	switch n.Event {
	case EventStart:
		return s.onStart
	case EventEnd:
		if n.Status == string(run.NodeSuccess) {
			return s.onSuccess
		}
		return s.onError
	}
	return false
}
