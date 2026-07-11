// Package run は実行管理 (Core BC) のドメインルールを持つ。
// 実行ツリーの状態遷移・ジャーナルイベント・RunID/NodeID 導出を型と関数で表現する。
package run

import (
	"fmt"
	"strings"
)

// Run は実行 1 回の集約。ジャーナルイベントを Apply して状態を管理する。
// 生成 (実行) 経路とジャーナルからの復元 (リプレイ) 経路の両方が同じ Apply を
// 通るため、外部から編集された不正なジャーナルはリプレイ時に検出される。
type Run struct {
	runID RunID
	nodes map[string]*node
	order []string
}

type node struct {
	id       string
	parentID string
	nodeType NodeType
	status   NodeStatus
	attrs    map[string]string
	steps    *Steps
}

// NewRun は空の Run 集約を生成する。
func NewRun(runID RunID) *Run {
	return &Run{runID: runID, nodes: map[string]*node{}}
}

// Replay はジャーナルイベント列から Run を復元する。
// 生成時と同じ状態遷移検証を通し、不正なイベント列は error を返す。
func Replay(runID RunID, events []Event) (*Run, error) {
	r := NewRun(runID)
	for i, ev := range events {
		if err := r.Apply(ev); err != nil {
			return nil, fmt.Errorf("journal line %d: %w", i+1, err)
		}
	}
	return r, nil
}

// RunID は集約の run_id を返す。
func (r *Run) RunID() RunID { return r.runID }

// Apply はイベントを検証して状態へ反映する。不正な遷移は error を返す。
func (r *Run) Apply(ev Event) error {
	switch ev.Type {
	case EventNodeStart:
		return r.applyNodeStart(ev)
	case EventStepsEnumerated:
		return r.applyStepsEnumerated(ev)
	case EventStepEnd:
		return r.applyStepEnd(ev)
	case EventNodeEnd:
		return r.applyNodeEnd(ev)
	default:
		return fmt.Errorf("%s is not journal event type", ev.Type)
	}
}

// applyNodeStart は階層の実行開始を検証する: 未開始のノードであること・
// 親が Started であること (状態 (初期) → Started)。
func (r *Run) applyNodeStart(ev Event) error {
	id, err := ParseNodeID(r.runID, ev.NodeID, ev.NodeType)
	if err != nil {
		return err
	}
	if _, ok := r.nodes[ev.NodeID]; ok {
		return fmt.Errorf("node %s is already started", ev.NodeID)
	}
	parentID := id.Parent()
	if ev.ParentID != parentID {
		return fmt.Errorf("node %s: parent_id %s does not match derived %s", ev.NodeID, ev.ParentID, parentID)
	}
	if ev.NodeType != NodeTypeRun {
		parent, ok := r.nodes[parentID]
		if !ok {
			return fmt.Errorf("node %s: parent %s is not started", ev.NodeID, parentID)
		}
		if parent.status != NodeStarted {
			return fmt.Errorf("node %s: parent %s is %s (not %s)", ev.NodeID, parentID, parent.status, NodeStarted)
		}
	}

	attrs := make(map[string]string, len(ev.Attrs))
	for k, v := range ev.Attrs {
		attrs[k] = v
	}
	r.nodes[ev.NodeID] = &node{
		id:       ev.NodeID,
		parentID: parentID,
		nodeType: ev.NodeType,
		status:   NodeStarted,
		attrs:    attrs,
	}
	r.order = append(r.order, ev.NodeID)
	return nil
}

// applyStepsEnumerated はステップ列挙 (全件 Pending 登録) を検証する:
// Started な process ノードに対して 1 度だけ列挙できる。
func (r *Run) applyStepsEnumerated(ev Event) error {
	n, err := r.startedNode(ev.NodeID)
	if err != nil {
		return err
	}
	if n.nodeType != NodeTypeProcess {
		return fmt.Errorf("node %s: steps can not be enumerated on %s node", ev.NodeID, n.nodeType)
	}
	if n.steps != nil {
		return fmt.Errorf("node %s: steps are already enumerated", ev.NodeID)
	}
	steps, err := NewSteps(ev.Steps)
	if err != nil {
		return fmt.Errorf("node %s: %w", ev.NodeID, err)
	}
	n.steps = steps
	return nil
}

// applyStepEnd はステップの終了 (Pending → Success/Warn/Error/Blocked) を検証する。
func (r *Run) applyStepEnd(ev Event) error {
	n, err := r.startedNode(ev.NodeID)
	if err != nil {
		return err
	}
	if n.steps == nil {
		return fmt.Errorf("node %s: steps are not enumerated", ev.NodeID)
	}
	if err := n.steps.MarkEnd(ev.Step, StepStatus(ev.Status)); err != nil {
		return fmt.Errorf("node %s: %w", ev.NodeID, err)
	}
	return nil
}

// applyNodeEnd は階層の終了 (Started → Success/Warn/Error) を検証する。
func (r *Run) applyNodeEnd(ev Event) error {
	n, ok := r.nodes[ev.NodeID]
	if !ok {
		return fmt.Errorf("node %s is not started", ev.NodeID)
	}
	to := NodeStatus(ev.Status)
	if err := n.status.Transition(to); err != nil {
		return fmt.Errorf("node %s: %w", ev.NodeID, err)
	}
	n.status = to
	return nil
}

// startedNode は Started 状態のノードを取得する。
func (r *Run) startedNode(nodeID string) (*node, error) {
	n, ok := r.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node %s is not started", nodeID)
	}
	if n.status != NodeStarted {
		return nil, fmt.Errorf("node %s is %s (not %s)", nodeID, n.status, NodeStarted)
	}
	return n, nil
}

// NodeView はノードの表示用スナップショット。
type NodeView struct {
	ID       string
	ParentID string
	Name     string // NodeID の最終セグメント
	Type     NodeType
	Status   NodeStatus
	Attrs    map[string]string
	Steps    []StepView
	Depth    int // run 階層 = 0
}

// NodeViews はイベント適用順 (実行順) のノードビューを返す。
func (r *Run) NodeViews() []NodeView {
	views := make([]NodeView, 0, len(r.order))
	for _, id := range r.order {
		n := r.nodes[id]
		segments := strings.Split(strings.TrimPrefix(id, r.runID.String()+"+"), "+")
		view := NodeView{
			ID:       n.id,
			ParentID: n.parentID,
			Name:     segments[len(segments)-1],
			Type:     n.nodeType,
			Status:   n.status,
			Attrs:    n.attrs,
			Depth:    len(segments) - 1,
		}
		if n.steps != nil {
			view.Steps = n.steps.Views()
		}
		views = append(views, view)
	}
	return views
}
