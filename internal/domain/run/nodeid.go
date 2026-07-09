package run

import (
	"fmt"
	"strings"
)

// NodeType は実行ツリーの階層種別。
type NodeType string

const (
	NodeTypeRun      NodeType = "run"
	NodeTypeScenario NodeType = "scenario"
	NodeTypeBizdate  NodeType = "bizdate"
	NodeTypeProcess  NodeType = "process"
)

// depth は run_id を除いた NodeID のセグメント数。未知の種別は 0 を返す。
func (t NodeType) depth() int {
	switch t {
	case NodeTypeRun:
		return 1
	case NodeTypeScenario:
		return 2
	case NodeTypeBizdate:
		return 3
	case NodeTypeProcess:
		return 4
	}
	return 0
}

// NodeID は実行ツリーのノード ID。v0.2 の webhook_id 導出規則と同一で、
// run_id にツリーのパスセグメントを `+` で連結する
// (例: _20200101120000_123+run+scenario1+_10_99990101+_10_pre_scripts)。
// v0.2 が余分な `}` を付与していたバグは修正済み。digdag 廃止により
// `^sub` セグメントは含まれない。
type NodeID struct {
	value string
}

// NewRunNodeID は run 階層 (ルート) の NodeID を導出する。
func NewRunNodeID(runID RunID) NodeID {
	return NodeID{value: runID.String() + "+" + string(NodeTypeRun)}
}

// Child は子階層の NodeID を導出する。segment はディレクトリ名等の 1 セグメント。
func (n NodeID) Child(segment string) (NodeID, error) {
	if segment == "" {
		return NodeID{}, fmt.Errorf("node_id segment must not null")
	}
	if strings.ContainsAny(segment, `+^/\`) {
		return NodeID{}, fmt.Errorf("node_id segment %q can not contains %q", segment, `+^/\`)
	}
	return NodeID{value: n.value + "+" + segment}, nil
}

// Parent は親 ID の文字列を返す (v0.2 の webhook parent_id の導出規則:
// `+` を `/` に置換 → dirname → `+` 連結に戻す、と同じ結果)。
// run 階層の親は run_id 自身になる。
func (n NodeID) Parent() string {
	i := strings.LastIndex(n.value, "+")
	if i < 0 {
		return n.value
	}
	return n.value[:i]
}

// String はノード ID の文字列表現を返す。
func (n NodeID) String() string { return n.value }

// ParseNodeID は文字列を NodeID として検証・復元する (リプレイ経路)。
// runID 配下であること・種別と深さの対応を検証する。
func ParseNodeID(runID RunID, s string, nodeType NodeType) (NodeID, error) {
	want := nodeType.depth()
	if want == 0 {
		return NodeID{}, fmt.Errorf("%s is not node_type (run|scenario|bizdate|process)", nodeType)
	}

	prefix := runID.String() + "+"
	if !strings.HasPrefix(s, prefix) {
		return NodeID{}, fmt.Errorf("node_id %s does not belong to run %s", s, runID)
	}
	segments := strings.Split(strings.TrimPrefix(s, prefix), "+")
	for _, seg := range segments {
		if seg == "" {
			return NodeID{}, fmt.Errorf("node_id %s has empty segment", s)
		}
	}
	if segments[0] != string(NodeTypeRun) {
		return NodeID{}, fmt.Errorf("node_id %s must start with %s+run", s, runID)
	}
	if len(segments) != want {
		return NodeID{}, fmt.Errorf("node_id %s depth %d does not match node_type %s", s, len(segments), nodeType)
	}
	return NodeID{value: s}, nil
}
