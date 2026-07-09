package scenario

import (
	"fmt"
	"strings"
)

// ScenarioName はシナリオ名の値オブジェクト。シナリオディレクトリ名になる。
// v0.2 の scenario_spec は非 null のみを検査していた。
// パス区切り文字と "." / ".." の禁止は v1.0 で追加したガード。
type ScenarioName struct {
	value string
}

// NewScenarioName はシナリオ名を検証して ScenarioName を生成する。
func NewScenarioName(s string) (ScenarioName, error) {
	if s == "" {
		return ScenarioName{}, fmt.Errorf("scenario_name must not null")
	}
	if s == "." || s == ".." {
		return ScenarioName{}, fmt.Errorf("%q can not be used as scenario_name", s)
	}
	if strings.ContainsAny(s, `/\`) {
		return ScenarioName{}, fmt.Errorf("%q can not contains path separator", s)
	}
	return ScenarioName{value: s}, nil
}

// String はシナリオ名の文字列表現を返す。
func (n ScenarioName) String() string { return n.value }
