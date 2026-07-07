package notify

import (
	"regexp"
	"strconv"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// payloadTSFormat は payload の時刻形式。v0.2 の timestamp_to_iso
// (`date '+%z'` によるコロン無しタイムゾーン) と同じ。
const payloadTSFormat = "2006-01-02T15:04:05-0700"

// Payload は webhook の request body。フィールド順は v0.2 の
// src/config/webhook/payload.yml + run/scenario/bizdate/process.yml の
// テンプレート連結順と同一 (yaml2json がキー順を保持していたことに対応)。
type Payload struct {
	Payload Body `json:"payload"`
}

// Body は payload 直下の共通項目 + 階層別属性ツリー。
type Body struct {
	ID             string   `json:"id"`
	ParentID       any      `json:"parent_id"`
	Type           string   `json:"type"`
	Status         string   `json:"status"`
	CreateTime     string   `json:"create_time"`
	StartTime      string   `json:"start_time"`
	EndTime        string   `json:"end_time"`
	ProcessingTime string   `json:"processing_time"`
	Stfw           Stfw     `json:"stfw"`
	Run            *RunNode `json:"run"`
}

// Stfw は実行環境情報。digdag / home は v1.0 で廃止したためキーのみ維持し
// 空文字を設定する (payload スキーマ互換)。
type Stfw struct {
	Host    string  `json:"host"`
	User    string  `json:"user"`
	Home    string  `json:"home"`
	Version string  `json:"version"`
	Digdag  Digdag  `json:"digdag"`
	Project Project `json:"project"`
}

// Digdag は v0.2 互換のためキーのみ残した digdag 情報 (常に空)。
type Digdag struct {
	URL     string `json:"url"`
	Version string `json:"version"`
}

// Project はプロジェクト情報。
type Project struct {
	Home    string `json:"home"`
	Version string `json:"version"`
}

// RunNode / ScenarioNode / BizdateNode / ProcessNode は
// run > scenario > bizdate > process の階層別属性 (webhook/*.yml の 1:1 再現)。
type RunNode struct {
	RunID        any           `json:"run_id"`
	WorkspaceDir any           `json:"workspace_dir"`
	Params       any           `json:"params"`
	Scenario     *ScenarioNode `json:"scenario,omitempty"`
}

type ScenarioNode struct {
	Name    any          `json:"name"`
	Bizdate *BizdateNode `json:"bizdate,omitempty"`
}

type BizdateNode struct {
	DirName any          `json:"dirname"`
	Seq     any          `json:"seq"`
	Bizdate any          `json:"bizdate"`
	Process *ProcessNode `json:"process,omitempty"`
}

type ProcessNode struct {
	DirName any     `json:"dirname"`
	Seq     any     `json:"seq"`
	Group   any     `json:"group"`
	Plugin  *Plugin `json:"plugin,omitempty"`
}

// Plugin は process payload のプラグイン詳細
// (scripts プラグインの bin/webhook/template.yml + template_detail.yml の再現)。
// targets はステップ実行結果のリスト。ステップが無い場合は null。
type Plugin struct {
	Type    string              `json:"type"`
	Targets []map[string]Target `json:"targets"`
}

// Target はステップ 1 件の実行結果。未実行 (Pending / Blocked) は時刻を空文字で持つ。
type Target struct {
	Result         string `json:"result"`
	StartTime      string `json:"start_time"`
	EndTime        string `json:"end_time"`
	ProcessingTime string `json:"processing_time"`
}

// Context は payload の実行環境情報 (イベントに含まれない値) を保持する。
// I/O を伴う値 (host / user) は usecase 側で解決して渡す。
type Context struct {
	Host           string // 自ホスト IP アドレス
	User           string // 実行ユーザー名
	Version        string // stfw バージョン
	ProjectVersion string // stfw.project_version 設定値
	ProjectHome    string // プロジェクトディレクトリ
	WorkspaceDir   string // 実行データディレクトリ (.stfw/runs/{run_id})
}

var (
	yamlDecimalPattern = regexp.MustCompile(`^[-+]?(0|[1-9][0-9]*)$`)
	yamlOctalPattern   = regexp.MustCompile(`^[-+]?0[0-7]+$`)
)

// yamlScalar は v0.2 テンプレートで unquoted だった値の JSON 型を再現する
// (yaml2json = PyYAML のスカラー解決)。空文字は null、整数表現は数値
// (先頭 0 は 8 進)、それ以外は文字列として扱う。
func yamlScalar(s string) any {
	if s == "" {
		return nil
	}
	if yamlDecimalPattern.MatchString(s) {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			return n
		}
	}
	if yamlOctalPattern.MatchString(s) {
		sign := int64(1)
		digits := s
		switch s[0] {
		case '-':
			sign = -1
			digits = s[1:]
		case '+':
			digits = s[1:]
		}
		if n, err := strconv.ParseInt(digits, 8, 64); err == nil {
			return sign * n
		}
	}
	return s
}

// payloadTime はジャーナルのタイムスタンプ (run.TSFormat) を payload の
// 時刻形式へ変換する。パースできない値はそのまま返す (payload 生成を止めない)。
func payloadTime(ts string) string {
	t, err := time.Parse(run.TSFormat, ts)
	if err != nil {
		return ts
	}
	return t.Format(payloadTSFormat)
}
