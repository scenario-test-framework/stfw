package run

import "fmt"

// NodeStatus は階層実行ステータス (run / scenario / bizdate / process の各階層)。
// 遷移は Started → Success | Error のみ
// (docs/harvest/latest/05-internal.md の状態モデル「階層実行ステータス」)。
type NodeStatus string

const (
	NodeStarted NodeStatus = "Started"
	NodeSuccess NodeStatus = "Success"
	NodeError   NodeStatus = "Error"
)

// Transition は自状態から to への遷移を検証する。不正遷移は error を返す。
func (s NodeStatus) Transition(to NodeStatus) error {
	if s != NodeStarted {
		return fmt.Errorf("node status can not transition from %s to %s", s, to)
	}
	if to != NodeSuccess && to != NodeError {
		return fmt.Errorf("node status can not transition from %s to %s", s, to)
	}
	return nil
}

// StepStatus はステップ実行ステータス (scripts プロセスのスクリプト単位)。
// 遷移は Pending → Success | Error | Blocked のみ
// (docs/harvest/latest/05-internal.md の状態モデル「ステップ実行ステータス」)。
type StepStatus string

const (
	StepPending StepStatus = "Pending"
	StepSuccess StepStatus = "Success"
	StepError   StepStatus = "Error"
	StepBlocked StepStatus = "Blocked"
)

// Transition は自状態から to への遷移を検証する。不正遷移は error を返す。
func (s StepStatus) Transition(to StepStatus) error {
	if s != StepPending {
		return fmt.Errorf("step status can not transition from %s to %s", s, to)
	}
	if to != StepSuccess && to != StepError && to != StepBlocked {
		return fmt.Errorf("step status can not transition from %s to %s", s, to)
	}
	return nil
}
