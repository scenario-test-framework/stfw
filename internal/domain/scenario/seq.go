package scenario

import "fmt"

// Seq は階層ディレクトリの実行順を決める連番の値オブジェクト。
// 数字のみであることを生成時に保証する (v0.2 の checks.must_be_number と同じ規則)。
// 先頭ゼロ (例: "010") を保持するため内部表現は文字列とする。
type Seq struct {
	value string
}

// NewSeq は連番文字列を検証して Seq を生成する。
func NewSeq(s string) (Seq, error) {
	if !isDigits(s) {
		return Seq{}, fmt.Errorf("%s must be number", s)
	}
	return Seq{value: s}, nil
}

// String は連番の文字列表現を返す。
func (s Seq) String() string { return s.value }
