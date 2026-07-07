package notify

import (
	"fmt"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// EventKind は webhook イベント種別 (v0.2 の start / end)。
type EventKind string

const (
	EventStart EventKind = "start"
	EventEnd   EventKind = "end"
)

// Notification は投影済みの webhook 通知 1 件。送信可否は Settings.ShouldNotify で判定する。
type Notification struct {
	Event   EventKind
	Status  string
	Payload Payload
}

// nodeState は payload 組み立てに必要なノードの投影状態。
type nodeState struct {
	id       string
	parentID string
	nodeType run.NodeType
	attrs    map[string]string
	startTS  string
	steps    []*stepState
}

// stepState はステップ実行結果の投影状態。未実行は時刻を空文字で持つ。
type stepState struct {
	name    string
	result  string
	startTS string
	endTS   string
}

// Projector はジャーナルイベントを webhook 通知へ投影する。
// ジャーナルが webhook payload の唯一のソース (イベントの投影) であり、
// イベントは Run 集約で検証済みの順序で渡される前提。
type Projector struct {
	ctx   Context
	nodes map[string]*nodeState
	// pendingStart は start 通知を保留中の process ノード。
	// v0.2 の process start payload はステップの Pending 列挙を含むため、
	// node_start 直後の steps_enumerated を待ってから通知を組み立てる。
	pendingStart *nodeState
}

// NewProjector は投影器を生成する。
func NewProjector(ctx Context) *Projector {
	return &Projector{ctx: ctx, nodes: map[string]*nodeState{}}
}

// Project はイベント 1 件を投影し、送信対象の通知 (0..2 件) を返す。
// now は create_time に記録する時刻源。
func (p *Projector) Project(ev run.Event, now time.Time) ([]Notification, error) {
	switch ev.Type {
	case run.EventNodeStart:
		return p.projectNodeStart(ev, now)
	case run.EventStepsEnumerated:
		return p.projectStepsEnumerated(ev, now)
	case run.EventStepEnd:
		return p.projectStepEnd(ev, now)
	case run.EventNodeEnd:
		return p.projectNodeEnd(ev, now)
	default:
		return nil, fmt.Errorf("%s is not journal event type", ev.Type)
	}
}

func (p *Projector) projectNodeStart(ev run.Event, now time.Time) ([]Notification, error) {
	notifs, err := p.flushPending(now)
	if err != nil {
		return nil, err
	}
	attrs := make(map[string]string, len(ev.Attrs))
	for k, v := range ev.Attrs {
		attrs[k] = v
	}
	n := &nodeState{
		id:       ev.NodeID,
		parentID: ev.ParentID,
		nodeType: ev.NodeType,
		attrs:    attrs,
		startTS:  ev.TS,
	}
	p.nodes[ev.NodeID] = n

	// process の start 通知はステップ列挙イベントを待って保留する
	if n.nodeType == run.NodeTypeProcess {
		p.pendingStart = n
		return notifs, nil
	}
	start, err := p.build(n, EventStart, string(run.NodeStarted), "", now)
	if err != nil {
		return nil, err
	}
	return append(notifs, start), nil
}

func (p *Projector) projectStepsEnumerated(ev run.Event, now time.Time) ([]Notification, error) {
	n, ok := p.nodes[ev.NodeID]
	if !ok {
		return nil, fmt.Errorf("node %s is not started", ev.NodeID)
	}
	for _, step := range ev.Steps {
		n.steps = append(n.steps, &stepState{name: step, result: string(run.StepPending)})
	}
	return p.flushPending(now)
}

func (p *Projector) projectStepEnd(ev run.Event, now time.Time) ([]Notification, error) {
	notifs, err := p.flushPending(now)
	if err != nil {
		return nil, err
	}
	n, ok := p.nodes[ev.NodeID]
	if !ok {
		return nil, fmt.Errorf("node %s is not started", ev.NodeID)
	}
	for _, step := range n.steps {
		if step.name != ev.Step {
			continue
		}
		step.result = ev.Status
		step.startTS = ev.StartTS
		step.endTS = ev.EndTS
		return notifs, nil
	}
	return nil, fmt.Errorf("node %s: step %s is not enumerated", ev.NodeID, ev.Step)
}

func (p *Projector) projectNodeEnd(ev run.Event, now time.Time) ([]Notification, error) {
	notifs, err := p.flushPending(now)
	if err != nil {
		return nil, err
	}
	n, ok := p.nodes[ev.NodeID]
	if !ok {
		return nil, fmt.Errorf("node %s is not started", ev.NodeID)
	}
	end, err := p.build(n, EventEnd, ev.Status, ev.TS, now)
	if err != nil {
		return nil, err
	}
	return append(notifs, end), nil
}

// flushPending は保留中の process start 通知を組み立てて返す。
// ステップ列挙後 (またはステップ 0 件のまま次のイベントが来た時点) で確定する。
func (p *Projector) flushPending(now time.Time) ([]Notification, error) {
	if p.pendingStart == nil {
		return nil, nil
	}
	n := p.pendingStart
	p.pendingStart = nil
	start, err := p.build(n, EventStart, string(run.NodeStarted), "", now)
	if err != nil {
		return nil, err
	}
	return []Notification{start}, nil
}

// build はノードの投影状態から通知 1 件を組み立てる。
func (p *Projector) build(n *nodeState, kind EventKind, status, endTS string, now time.Time) (Notification, error) {
	body := Body{
		ID:         n.id,
		ParentID:   yamlScalar(n.parentID),
		Type:       string(n.nodeType),
		Status:     status,
		CreateTime: now.Format(payloadTSFormat),
		StartTime:  payloadTime(n.startTS),
		Stfw: Stfw{
			Host:    p.ctx.Host,
			User:    p.ctx.User,
			Version: p.ctx.Version,
			Project: Project{Home: p.ctx.ProjectHome, Version: p.ctx.ProjectVersion},
		},
	}
	if kind == EventEnd {
		elapsed, err := run.ElapsedString(n.startTS, endTS)
		if err != nil {
			return Notification{}, fmt.Errorf("node %s: %w", n.id, err)
		}
		body.EndTime = payloadTime(endTS)
		body.ProcessingTime = elapsed
	}

	runNode, err := p.buildTree(n)
	if err != nil {
		return Notification{}, err
	}
	body.Run = runNode
	return Notification{Event: kind, Status: status, Payload: Payload{Payload: body}}, nil
}

// buildTree は run > scenario > bizdate > process の階層別属性を
// 親子チェーンから組み立てる (v0.2 の階層別テンプレート合成の再現)。
func (p *Projector) buildTree(n *nodeState) (*RunNode, error) {
	chain := map[run.NodeType]*nodeState{}
	for cur := n; ; {
		chain[cur.nodeType] = cur
		if cur.nodeType == run.NodeTypeRun {
			break
		}
		parent, ok := p.nodes[cur.parentID]
		if !ok {
			return nil, fmt.Errorf("node %s: parent %s is not started", cur.id, cur.parentID)
		}
		cur = parent
	}

	root := chain[run.NodeTypeRun]
	runNode := &RunNode{
		RunID:        yamlScalar(root.attrs["run_id"]),
		WorkspaceDir: yamlScalar(p.ctx.WorkspaceDir),
		Params:       yamlScalar(root.attrs["params"]),
	}
	scenario, ok := chain[run.NodeTypeScenario]
	if !ok {
		return runNode, nil
	}
	runNode.Scenario = &ScenarioNode{Name: yamlScalar(scenario.attrs["name"])}

	bizdate, ok := chain[run.NodeTypeBizdate]
	if !ok {
		return runNode, nil
	}
	runNode.Scenario.Bizdate = &BizdateNode{
		DirName: yamlScalar(bizdate.attrs["dirname"]),
		Seq:     yamlScalar(bizdate.attrs["seq"]),
		Bizdate: yamlScalar(bizdate.attrs["bizdate"]),
	}

	process, ok := chain[run.NodeTypeProcess]
	if !ok {
		return runNode, nil
	}
	runNode.Scenario.Bizdate.Process = &ProcessNode{
		DirName: yamlScalar(process.attrs["dirname"]),
		Seq:     yamlScalar(process.attrs["seq"]),
		Group:   yamlScalar(process.attrs["group"]),
		Plugin:  buildPlugin(process),
	}
	return runNode, nil
}

// buildPlugin は scripts プロセスのステップ詳細 (v0.2 の webhook_detail) を組み立てる。
// ステップ詳細を投影できるのはジャーナルにステップイベントを記録する組込み
// scripts タイプのみ。独自プラグインの webhook 用コンテンツ生成
// (bin/webhook/get_*_content) は v1.0 で廃止したため plugin キー自体を省略する。
func buildPlugin(n *nodeState) *Plugin {
	processType := n.attrs["process_type"]
	if processType != "scripts" {
		return nil
	}
	plugin := &Plugin{Type: processType}
	for _, step := range n.steps {
		target := Target{Result: step.result}
		if step.startTS != "" && step.endTS != "" {
			target.StartTime = payloadTime(step.startTS)
			target.EndTime = payloadTime(step.endTS)
			if elapsed, err := run.ElapsedString(step.startTS, step.endTS); err == nil {
				target.ProcessingTime = elapsed
			}
		}
		plugin.Targets = append(plugin.Targets, map[string]Target{step.name: target})
	}
	return plugin
}
