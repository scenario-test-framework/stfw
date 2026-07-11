package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// nodeState はスパン記述の組み立てに必要なノードの投影状態。
type nodeState struct {
	id       string
	parentID string
	nodeType run.NodeType
	name     string
	attrs    map[string]string
	start    time.Time
	end      time.Time
	status   SpanStatus
	message  string
	// warn は階層ステータスが Warn だったことを表す (スパンステータスは Ok の
	// まま属性 stfw.node.status=Warn で表現する。SPEC-024-03)。
	warn  bool
	ended bool
	steps []Span
}

// Projector はジャーナルイベントをスパン記述へ投影する。
// ジャーナルがトレースの唯一のソース (イベントの投影) であり、
// イベントは Run 集約で検証済みの順序で渡される前提。
// 1 run = 1 トレース: ルート (run) の node_end でスパンツリー全体を確定して返す。
type Projector struct {
	runID string
	nodes map[string]*nodeState
	// order はノードの開始順 (親が先)。スパンの返却順に使う。
	order []string
}

// NewProjector は投影器を生成する。
func NewProjector() *Projector {
	return &Projector{nodes: map[string]*nodeState{}}
}

// Apply はイベント 1 件を投影する。ルート (run) の node_end で
// 完成したスパンツリー (親が先の順) を返し、それ以外は nil を返す。
func (p *Projector) Apply(ev run.Event) ([]Span, error) {
	switch ev.Type {
	case run.EventNodeStart:
		return nil, p.applyNodeStart(ev)
	case run.EventStepsEnumerated:
		// Pending の列挙はスパンにしない (step は step_end で完成する)
		return nil, nil
	case run.EventStepEnd:
		return nil, p.applyStepEnd(ev)
	case run.EventNodeEnd:
		return p.applyNodeEnd(ev)
	default:
		return nil, fmt.Errorf("%s is not journal event type", ev.Type)
	}
}

func (p *Projector) applyNodeStart(ev run.Event) error {
	start, err := parseTS(ev.TS)
	if err != nil {
		return fmt.Errorf("node %s: %w", ev.NodeID, err)
	}
	n := &nodeState{
		id:       ev.NodeID,
		parentID: ev.ParentID,
		nodeType: ev.NodeType,
		attrs:    ev.Attrs,
		start:    start,
	}
	if ev.NodeType == run.NodeTypeRun {
		// run はルートスパン (親を持たない)
		n.parentID = ""
		p.runID = ev.Attrs["run_id"]
	}
	n.name = spanName(n)
	p.nodes[ev.NodeID] = n
	p.order = append(p.order, ev.NodeID)
	return nil
}

func (p *Projector) applyStepEnd(ev run.Event) error {
	n, ok := p.nodes[ev.NodeID]
	if !ok {
		return fmt.Errorf("node %s is not started", ev.NodeID)
	}
	span := Span{
		ID:       ev.NodeID + "+" + ev.Step,
		ParentID: ev.NodeID,
		Name:     ev.Step,
		Attrs: []Attr{
			{Key: AttrRunID, Value: p.runID},
			{Key: AttrNodeType, Value: "step"},
			{Key: AttrNodeID, Value: ev.NodeID + "+" + ev.Step},
			{Key: AttrStepStatus, Value: ev.Status},
		},
	}
	switch run.StepStatus(ev.Status) {
	case run.StepBlocked:
		// 未実行のため開始・終了時刻を持たない: イベント時刻を点として記録し、
		// スパンステータスは Unset のまま stfw.step.status=Blocked で表現する
		ts, err := parseTS(ev.TS)
		if err != nil {
			return fmt.Errorf("step %s: %w", span.ID, err)
		}
		span.Start, span.End = ts, ts
		span.Status = SpanStatusUnset
	case run.StepSuccess, run.StepWarn, run.StepError:
		start, err := parseTS(ev.StartTS)
		if err != nil {
			return fmt.Errorf("step %s: %w", span.ID, err)
		}
		end, err := parseTS(ev.EndTS)
		if err != nil {
			return fmt.Errorf("step %s: %w", span.ID, err)
		}
		span.Start, span.End = start, end
		exitCode := int64(0)
		if ev.ExitCode != nil {
			exitCode = int64(*ev.ExitCode)
		}
		span.Attrs = append(span.Attrs, Attr{Key: AttrStepExit, Value: exitCode})
		span.Status = SpanStatusOK
		if run.StepStatus(ev.Status) == run.StepError {
			span.Status = SpanStatusError
			span.StatusMessage = fmt.Sprintf("step %s failed with exit_code %d", ev.Step, exitCode)
		}
	default:
		return fmt.Errorf("step %s: %s is not step status", span.ID, ev.Status)
	}
	n.steps = append(n.steps, span)
	return nil
}

func (p *Projector) applyNodeEnd(ev run.Event) ([]Span, error) {
	n, ok := p.nodes[ev.NodeID]
	if !ok {
		return nil, fmt.Errorf("node %s is not started", ev.NodeID)
	}
	end, err := parseTS(ev.TS)
	if err != nil {
		return nil, fmt.Errorf("node %s: %w", ev.NodeID, err)
	}
	n.end = end
	n.ended = true
	n.status = SpanStatusOK
	switch run.NodeStatus(ev.Status) {
	case run.NodeError:
		n.status = SpanStatusError
		n.message = fmt.Sprintf("%s %s finished with status Error", n.nodeType, n.name)
	case run.NodeWarn:
		// OTel に Warn 相当が無いため Ok + 属性 stfw.node.status=Warn で表現する
		n.warn = true
	}
	if n.nodeType != run.NodeTypeRun {
		return nil, nil
	}
	return p.tree(), nil
}

// tree は開始順 (親が先) にスパンツリー全体を組み立てる。
// step スパンは親 process スパンの直後に並べる。
func (p *Projector) tree() []Span {
	var spans []Span
	for _, id := range p.order {
		n := p.nodes[id]
		if !n.ended {
			continue
		}
		spans = append(spans, Span{
			ID:            n.id,
			ParentID:      n.parentID,
			Name:          n.name,
			Start:         n.start,
			End:           n.end,
			Status:        n.status,
			StatusMessage: n.message,
			Attrs:         nodeAttrs(p.runID, n),
		})
		spans = append(spans, n.steps...)
	}
	return spans
}

// spanName はスパン名を導出する。run は "stfw run"、
// scenario はシナリオ名、bizdate / process はディレクトリ名。
func spanName(n *nodeState) string {
	switch n.nodeType {
	case run.NodeTypeRun:
		return "stfw run"
	case run.NodeTypeScenario:
		return n.attrs["name"]
	default:
		return n.attrs["dirname"]
	}
}

// nodeAttrs は階層スパンの属性を組み立てる (該当階層にあるもののみ)。
func nodeAttrs(runID string, n *nodeState) []Attr {
	attrs := []Attr{
		{Key: AttrRunID, Value: runID},
		{Key: AttrNodeType, Value: string(n.nodeType)},
		{Key: AttrNodeID, Value: n.id},
	}
	if n.warn {
		attrs = append(attrs, Attr{Key: AttrNodeStatus, Value: string(run.NodeWarn)})
	}
	switch n.nodeType {
	case run.NodeTypeRun:
		// run_mode は実行契約の "--run" | "--dry-run" を run | dry-run に正規化する
		attrs = append(attrs, Attr{Key: AttrRunMode, Value: strings.TrimPrefix(n.attrs["run_mode"], "--")})
	case run.NodeTypeBizdate:
		attrs = append(attrs,
			Attr{Key: AttrBizdate, Value: n.attrs["bizdate"]},
			Attr{Key: AttrSeq, Value: n.attrs["seq"]},
		)
	case run.NodeTypeProcess:
		attrs = append(attrs,
			Attr{Key: AttrSeq, Value: n.attrs["seq"]},
			Attr{Key: AttrGroup, Value: n.attrs["group"]},
			Attr{Key: AttrProcessType, Value: n.attrs["process_type"]},
		)
	}
	return attrs
}

// parseTS はジャーナルのタイムスタンプ (run.TSFormat) をスパンの時刻へ変換する。
func parseTS(ts string) (time.Time, error) {
	t, err := time.Parse(run.TSFormat, ts)
	if err != nil {
		return time.Time{}, fmt.Errorf("ts %s: %w", ts, err)
	}
	return t, nil
}
