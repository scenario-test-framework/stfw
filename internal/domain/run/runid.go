package run

import (
	"fmt"
	"regexp"
	"time"
)

// RunID は実行 1 回を識別する ID の値オブジェクト。
// 採番規則は v0.2 の run_spec.uniq_id と同一の `_{yyyymmddhhmmss}_{pid}`。
type RunID struct {
	value string
}

var runIDPattern = regexp.MustCompile(`^_\d{14}_\d+$`)

// NewRunID は採番時刻とプロセス ID から RunID を採番する。
// 時刻は usecase から引数で渡す (テスト容易性)。
func NewRunID(t time.Time, pid int) RunID {
	return RunID{value: fmt.Sprintf("_%s_%d", t.Format("20060102150405"), pid)}
}

// ParseRunID は文字列を RunID として検証・復元する (リプレイ経路)。
func ParseRunID(s string) (RunID, error) {
	if !runIDPattern.MatchString(s) {
		return RunID{}, fmt.Errorf("%s is not run_id format (_{yyyymmddhhmmss}_{pid})", s)
	}
	return RunID{value: s}, nil
}

// String は run_id の文字列表現を返す。
func (r RunID) String() string { return r.value }

// Time は run_id に埋め込まれた採番時刻を返す (ハウスキープの保存期間判定用)。
// NewRunID がローカル時刻で採番するため、ローカルタイムゾーンで解釈する。
func (r RunID) Time() (time.Time, error) {
	if !runIDPattern.MatchString(r.value) {
		return time.Time{}, fmt.Errorf("%q is not run_id format (_{yyyymmddhhmmss}_{pid})", r.value)
	}
	return time.ParseInLocation("20060102150405", r.value[1:15], time.Local)
}
