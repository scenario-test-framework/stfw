package scenario

import (
	"fmt"
	"strings"
)

// Group はプロセスディレクトリ名 `_{seq}_{group}_{type}` の group 部の値オブジェクト。
// `_` を含まないことを生成時に保証する (ディレクトリ名の `_` 区切りパースの保護。
// v0.2 の checks.must_not_contains と同じ規則)。
// ディレクトリ名になるため、パス区切り文字の禁止は v1.0 で追加したガード。
type Group struct {
	value string
}

// NewGroup はグループ名を検証して Group を生成する。
func NewGroup(s string) (Group, error) {
	if s == "" {
		return Group{}, fmt.Errorf("group must not null")
	}
	if strings.Contains(s, "_") {
		return Group{}, fmt.Errorf("%q can not contains %q", s, "_")
	}
	if strings.ContainsAny(s, `/\`) {
		return Group{}, fmt.Errorf("%q can not contains path separator", s)
	}
	return Group{value: s}, nil
}

// String はグループ名の文字列表現を返す。
func (g Group) String() string { return g.value }
