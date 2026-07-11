// Package notify は通知管理 (Supporting BC) のドメインルールを持つ。
// ジャーナルイベントを OTLP トレースのスパン記述へ投影する。
// OTel SDK には依存せず、スパン記述の値 (Span) だけを組み立てる
// (SDK への変換・エクスポートは gateway の責務)。
package notify

import "time"

// SpanStatus はスパンステータス。OTel の status code (Unset | Ok | Error) に対応する。
type SpanStatus string

const (
	SpanStatusUnset SpanStatus = "Unset"
	SpanStatusOK    SpanStatus = "Ok"
	SpanStatusError SpanStatus = "Error"
)

// スパン属性キー。旧 webhook payload が持っていた実行コンテキストを引き継ぐ。
const (
	AttrRunID       = "stfw.run_id"
	AttrNodeType    = "stfw.node.type"
	AttrNodeID      = "stfw.node.id"
	AttrNodeStatus  = "stfw.node.status"
	AttrRunMode     = "stfw.run.mode"
	AttrBizdate     = "stfw.bizdate"
	AttrSeq         = "stfw.seq"
	AttrGroup       = "stfw.group"
	AttrProcessType = "stfw.process.type"
	AttrStepStatus  = "stfw.step.status"
	AttrStepExit    = "stfw.step.exit_code"
)

// Attr はスパン属性 1 件。値は string または int64 のみを許す
// (stdlib だけで表現し、OTel の attribute 型へは gateway で変換する)。
type Attr struct {
	Key   string
	Value any
}

// Span は投影済みのスパン記述 1 件。
// 開始・終了時刻はジャーナルイベントの時刻と一致する。
type Span struct {
	// ID はスパンの同一性キー。階層は NodeID、step は {node_id}+{step名}。
	ID string
	// ParentID は親スパンの ID。ルート (run) は空。
	ParentID string
	// Name はスパン名。run は "stfw run"、他はディレクトリ名
	// (scenario はシナリオ名、step はスクリプト名)。
	Name  string
	Start time.Time
	End   time.Time
	// Status は実行ステータスのマッピング。Success → Ok / Error → Error /
	// Blocked ステップ → Unset (属性 stfw.step.status で表現)。
	// Warn は OTel にスパンステータス相当が無いため Ok + 属性で表現する
	// (階層は stfw.node.status=Warn、step は stfw.step.status=Warn。SPEC-024-03)。
	Status SpanStatus
	// StatusMessage は Status が Error の場合のみ持つ。
	StatusMessage string
	Attrs         []Attr
}
