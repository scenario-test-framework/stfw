package run

import (
	"cmp"
	"fmt"
	"regexp"
	"strings"
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

// maxIssueLead は、最新 run の採番時刻が現在時刻より先行していても「最新 run の 1 秒後」で
// 採番する上限。1 秒に複数回実行すると繰り上げで採番時刻が実時刻より先行するため、
// 実運用の連続実行では届かない長さにする (届くと現在時刻に戻り、新旧が逆転する)。
// これを超える先行は時計の異常とみなし、現在時刻で採番する
// (1 つの未来時刻の run に以後の run_id が引きずられ続けることを防ぐ)。
const maxIssueLead = 24 * time.Hour

// NextIssueTime は新しい run_id の採番時刻を決める (AS-BUILT §5.2)。
// latest は既存の最新 run の採番時刻 (run が無ければゼロ値)。
// 同一秒に続けて実行しても採番時刻で新旧が決まるよう、最新 run 以上の時刻なら 1 秒後にする
// (pid は時刻順にならないため、pid では同一秒の新旧を決められない)。
func NextIssueTime(now, latest time.Time) time.Time {
	t := now.Truncate(time.Second)
	if latest.IsZero() || latest.Before(t) || latest.Sub(t) > maxIssueLead {
		return t
	}
	return latest.Add(time.Second)
}

// CompareRunID は run_id の新旧を比較する (a が古ければ負、新しければ正、同一なら 0)。
// 採番時刻で比較し、同一の採番時刻どうしは pid を数値として比較する
// (文字列比較では桁数の異なる pid で順序が逆転する)。
func CompareRunID(a, b RunID) int {
	if c := strings.Compare(a.value[1:15], b.value[1:15]); c != 0 {
		return c
	}
	// pid は数字のみ (runIDPattern) なので、先頭ゼロを除いた桁数 → 辞書順で数値比較になる
	// (桁あふれしない)
	pa := strings.TrimLeft(a.value[16:], "0")
	pb := strings.TrimLeft(b.value[16:], "0")
	if len(pa) != len(pb) {
		return cmp.Compare(len(pa), len(pb))
	}
	if c := strings.Compare(pa, pb); c != 0 {
		return c
	}
	// 数値として同じ (先頭ゼロだけが違う) 場合も全順序にするため文字列で決める
	return strings.Compare(a.value, b.value)
}

// Time は run_id に埋め込まれた採番時刻を返す (ハウスキープの保存期間判定用)。
// NewRunID がローカル時刻で採番するため、ローカルタイムゾーンで解釈する。
func (r RunID) Time() (time.Time, error) {
	if !runIDPattern.MatchString(r.value) {
		return time.Time{}, fmt.Errorf("%q is not run_id format (_{yyyymmddhhmmss}_{pid})", r.value)
	}
	return time.ParseInLocation("20060102150405", r.value[1:15], time.Local)
}
