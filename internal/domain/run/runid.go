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
