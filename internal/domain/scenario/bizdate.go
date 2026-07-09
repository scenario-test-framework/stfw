// Package scenario はシナリオ構造管理 (Core BC) のドメインルールを持つ。
// ディレクトリ命名規約・階層判定・走査規則を型と関数で表現する。
package scenario

import (
	"fmt"
	"time"
)

// Bizdate は業務日付の値オブジェクト。
// YYYYMMDD の 8 桁数字かつ実在する日付であることを生成時に保証する
// (v0.2 の checks.must_be_date_format は桁数と数字のみの検査だったが、
// v1.0 では実在日付の検証を追加している)。
type Bizdate struct {
	value string
}

// NewBizdate は業務日付文字列を検証して Bizdate を生成する。
func NewBizdate(s string) (Bizdate, error) {
	if len(s) != 8 || !isDigits(s) {
		return Bizdate{}, fmt.Errorf("%s must be YYYYMMDD format", s)
	}
	if _, err := time.Parse("20060102", s); err != nil {
		return Bizdate{}, fmt.Errorf("%s is not a valid date", s)
	}
	return Bizdate{value: s}, nil
}

// String は YYYYMMDD 形式の文字列を返す。
func (b Bizdate) String() string { return b.value }

// isDigits は s が 1 文字以上の半角数字のみで構成されるかを返す。
func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
