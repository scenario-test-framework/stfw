package run

import "fmt"

// NodeStatus は階層実行ステータス (run / scenario / bizdate / process の各階層)。
// 遷移は Started → Success | Warn | Error のみ (AS-BUILT §5.4)。
// Warn は v1.2.0 で追加した一級ステータス (REQ-023): 子の Warn を
// Error > Warn > Success の優先度で集約し、実行は止めない。
type NodeStatus string

const (
	NodeStarted NodeStatus = "Started"
	NodeSuccess NodeStatus = "Success"
	NodeWarn    NodeStatus = "Warn"
	NodeError   NodeStatus = "Error"
)

// Transition は自状態から to への遷移を検証する。不正遷移は error を返す。
func (s NodeStatus) Transition(to NodeStatus) error {
	if s != NodeStarted {
		return fmt.Errorf("node status can not transition from %s to %s", s, to)
	}
	if to != NodeSuccess && to != NodeWarn && to != NodeError {
		return fmt.Errorf("node status can not transition from %s to %s", s, to)
	}
	return nil
}

// StepStatus はステップ実行ステータス (scripts プロセスのスクリプト単位)。
// 遷移は Pending → Success | Warn | Error | Blocked のみ (AS-BUILT §5.4)。
// Warn は v1.2.0 で追加した一級ステータス (REQ-023): exit 3 のステップを
// Warn として記録し、後続ステップは止めない (Error のみ Blocked 伝播)。
type StepStatus string

const (
	StepPending StepStatus = "Pending"
	StepSuccess StepStatus = "Success"
	StepWarn    StepStatus = "Warn"
	StepError   StepStatus = "Error"
	StepBlocked StepStatus = "Blocked"
)

// Transition は自状態から to への遷移を検証する。不正遷移は error を返す。
func (s StepStatus) Transition(to StepStatus) error {
	if s != StepPending {
		return fmt.Errorf("step status can not transition from %s to %s", s, to)
	}
	if to != StepSuccess && to != StepWarn && to != StepError && to != StepBlocked {
		return fmt.Errorf("step status can not transition from %s to %s", s, to)
	}
	return nil
}
