package run

import "time"

// EventType はジャーナルイベント種別。
type EventType string

const (
	EventNodeStart       EventType = "node_start"
	EventStepsEnumerated EventType = "steps_enumerated"
	EventStepEnd         EventType = "step_end"
	EventNodeEnd         EventType = "node_end"
)

// TSFormat はジャーナルのタイムスタンプ形式 (ISO 8601 / ローカルタイムゾーン)。
// v0.2 の timestamp_to_iso と同じ形式。
const TSFormat = "2006-01-02T15:04:05Z07:00"

// Event は journal.jsonl の 1 行に対応する追記専用イベント。
// OTLP トレースと HTML レポートの材料になる属性を欠落なく保持する。
type Event struct {
	Type     EventType         `json:"event"`
	TS       string            `json:"ts"`
	NodeID   string            `json:"node_id"`
	ParentID string            `json:"parent_id,omitempty"`
	NodeType NodeType          `json:"node_type,omitempty"`
	Attrs    map[string]string `json:"attrs,omitempty"`
	Steps    []string          `json:"steps,omitempty"`
	Step     string            `json:"step,omitempty"`
	Status   string            `json:"status,omitempty"`
	ExitCode *int              `json:"exit_code,omitempty"`
	StartTS  string            `json:"start_ts,omitempty"`
	EndTS    string            `json:"end_ts,omitempty"`
}

// NewNodeStartEvent は階層の実行開始 (Started) イベントを組み立てる。
// attrs には投影 (OTLP トレース / HTML レポート) の階層別属性 (run: run_id/run_mode/params,
// scenario: name, bizdate: dirname/seq/bizdate, process: dirname/seq/group/process_type)
// を記録する。
func NewNodeStartEvent(ts time.Time, id NodeID, nodeType NodeType, attrs map[string]string) Event {
	return Event{
		Type:     EventNodeStart,
		TS:       ts.Format(TSFormat),
		NodeID:   id.String(),
		ParentID: id.Parent(),
		NodeType: nodeType,
		Attrs:    attrs,
	}
}

// NewStepsEnumeratedEvent はステップの列挙 (全件 Pending 登録) イベントを組み立てる。
// v0.2 の scripts プラグインが start webhook で全スクリプトを Pending 列挙していた記録に対応する。
func NewStepsEnumeratedEvent(ts time.Time, id NodeID, steps []string) Event {
	return Event{
		Type:   EventStepsEnumerated,
		TS:     ts.Format(TSFormat),
		NodeID: id.String(),
		Steps:  steps,
	}
}

// NewStepEndEvent はステップの実行終了 (Success | Error) イベントを組み立てる。
func NewStepEndEvent(ts time.Time, id NodeID, step string, status StepStatus, exitCode int, start, end time.Time) Event {
	code := exitCode
	return Event{
		Type:     EventStepEnd,
		TS:       ts.Format(TSFormat),
		NodeID:   id.String(),
		Step:     step,
		Status:   string(status),
		ExitCode: &code,
		StartTS:  start.Format(TSFormat),
		EndTS:    end.Format(TSFormat),
	}
}

// NewStepBlockedEvent は先行エラーによるステップの Blocked イベントを組み立てる。
// 実行されていないため exit_code / start_ts / end_ts は持たない (v0.2 の skip 記録と同じ)。
func NewStepBlockedEvent(ts time.Time, id NodeID, step string) Event {
	return Event{
		Type:   EventStepEnd,
		TS:     ts.Format(TSFormat),
		NodeID: id.String(),
		Step:   step,
		Status: string(StepBlocked),
	}
}

// NewNodeEndEvent は階層の実行終了 (Success | Error) イベントを組み立てる。
func NewNodeEndEvent(ts time.Time, id NodeID, status NodeStatus) Event {
	return Event{
		Type:   EventNodeEnd,
		TS:     ts.Format(TSFormat),
		NodeID: id.String(),
		Status: string(status),
	}
}
