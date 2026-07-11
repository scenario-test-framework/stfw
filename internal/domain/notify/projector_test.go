package notify

import (
	"strings"
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// testNodeIDs は run > scenario > bizdate > process の NodeID を組み立てる。
func testNodeIDs(t *testing.T) map[string]run.NodeID {
	t.Helper()
	runID, err := run.ParseRunID("_20200101120000_99")
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]run.NodeID{}
	ids["run"] = run.NewRunNodeID(runID)
	scenario, err := ids["run"].Child("demo")
	if err != nil {
		t.Fatal(err)
	}
	ids["scenario"] = scenario
	bizdate, err := scenario.Child("_10_99990101")
	if err != nil {
		t.Fatal(err)
	}
	ids["bizdate"] = bizdate
	process, err := bizdate.Child("_10_pre_scripts")
	if err != nil {
		t.Fatal(err)
	}
	ids["process"] = process
	return ids
}

func ts(sec int) time.Time {
	return time.Date(2020, 1, 1, 12, 0, sec, 0, time.FixedZone("JST", 9*60*60))
}

// applyAll はイベント列を投影し、途中のイベントがスパンを返さないことを検証する。
func applyAll(t *testing.T, p *Projector, events []run.Event) []Span {
	t.Helper()
	for i, ev := range events {
		spans, err := p.Apply(ev)
		if err != nil {
			t.Fatalf("event %d (%s): %v", i, ev.Type, err)
		}
		if i < len(events)-1 && spans != nil {
			t.Fatalf("event %d (%s): spans must be deferred until run end, got %+v", i, ev.Type, spans)
		}
		if i == len(events)-1 {
			return spans
		}
	}
	return nil
}

// attrValue はスパン属性の値を返す (未設定は nil)。
func attrValue(s Span, key string) any {
	for _, a := range s.Attrs {
		if a.Key == key {
			return a.Value
		}
	}
	return nil
}

func TestProjectorTreeSuccess(t *testing.T) {
	t.Run("Projector_全ノード成功の場合_階層スパンを全てOkで生成すること", func(t *testing.T) {
		// Arrange
		ids := testNodeIDs(t)
		p := NewProjector()
		events := []run.Event{
			run.NewNodeStartEvent(ts(0), ids["run"], run.NodeTypeRun,
				map[string]string{"run_id": "_20200101120000_99", "run_mode": "--run", "params": "demo"}),
			run.NewNodeStartEvent(ts(1), ids["scenario"], run.NodeTypeScenario,
				map[string]string{"name": "demo"}),
			run.NewNodeStartEvent(ts(2), ids["bizdate"], run.NodeTypeBizdate,
				map[string]string{"dirname": "_10_99990101", "seq": "10", "bizdate": "99990101"}),
			run.NewNodeStartEvent(ts(3), ids["process"], run.NodeTypeProcess,
				map[string]string{"dirname": "_10_pre_scripts", "seq": "10", "group": "pre", "process_type": "scripts"}),
			run.NewStepsEnumeratedEvent(ts(3), ids["process"], []string{"100_step1", "200_step2"}),
			run.NewStepEndEvent(ts(4), ids["process"], "100_step1", run.StepSuccess, 0, ts(3), ts(4)),
			run.NewStepEndEvent(ts(5), ids["process"], "200_step2", run.StepSuccess, 0, ts(4), ts(5)),
			run.NewNodeEndEvent(ts(6), ids["process"], run.NodeSuccess),
			run.NewNodeEndEvent(ts(7), ids["bizdate"], run.NodeSuccess),
			run.NewNodeEndEvent(ts(8), ids["scenario"], run.NodeSuccess),
			run.NewNodeEndEvent(ts(9), ids["run"], run.NodeSuccess),
		}

		// Act
		spans := applyAll(t, p, events)

		// Assert
		// スパンは開始順 (親が先)、step は親 process の直後
		wantNames := []string{"stfw run", "demo", "_10_99990101", "_10_pre_scripts", "100_step1", "200_step2"}
		if len(spans) != len(wantNames) {
			t.Fatalf("spans = %d, want %d", len(spans), len(wantNames))
		}
		for i, want := range wantNames {
			if spans[i].Name != want {
				t.Errorf("spans[%d].Name = %s, want %s", i, spans[i].Name, want)
			}
		}

		// 親子関係: run をルートに scenario > bizdate > process > step
		if spans[0].ParentID != "" {
			t.Errorf("run span must be root, parent = %s", spans[0].ParentID)
		}
		for i := 1; i < 4; i++ {
			if spans[i].ParentID != spans[i-1].ID {
				t.Errorf("spans[%d].ParentID = %s, want %s", i, spans[i].ParentID, spans[i-1].ID)
			}
		}
		for _, step := range spans[4:] {
			if step.ParentID != ids["process"].String() {
				t.Errorf("step %s parent = %s, want %s", step.Name, step.ParentID, ids["process"])
			}
		}

		// 開始・終了時刻はジャーナルイベントの時刻と一致する
		if !spans[0].Start.Equal(ts(0)) || !spans[0].End.Equal(ts(9)) {
			t.Errorf("run span time = %v-%v, want %v-%v", spans[0].Start, spans[0].End, ts(0), ts(9))
		}
		if !spans[3].Start.Equal(ts(3)) || !spans[3].End.Equal(ts(6)) {
			t.Errorf("process span time = %v-%v, want %v-%v", spans[3].Start, spans[3].End, ts(3), ts(6))
		}
		if !spans[4].Start.Equal(ts(3)) || !spans[4].End.Equal(ts(4)) {
			t.Errorf("step1 span time = %v-%v, want %v-%v", spans[4].Start, spans[4].End, ts(3), ts(4))
		}

		// 全 Success → 全スパン Ok
		for _, s := range spans {
			if s.Status != SpanStatusOK || s.StatusMessage != "" {
				t.Errorf("span %s status = %s (%s), want Ok", s.Name, s.Status, s.StatusMessage)
			}
		}

		// 階層別属性 (該当階層にあるもののみ)
		for _, s := range spans {
			if attrValue(s, AttrRunID) != "_20200101120000_99" {
				t.Errorf("span %s %s = %v", s.Name, AttrRunID, attrValue(s, AttrRunID))
			}
		}
		if v := attrValue(spans[0], AttrNodeType); v != "run" {
			t.Errorf("run %s = %v", AttrNodeType, v)
		}
		if v := attrValue(spans[0], AttrRunMode); v != "run" {
			t.Errorf("run %s = %v, want run", AttrRunMode, v)
		}
		if v := attrValue(spans[0], AttrNodeID); v != ids["run"].String() {
			t.Errorf("run %s = %v", AttrNodeID, v)
		}
		if v := attrValue(spans[1], AttrNodeType); v != "scenario" {
			t.Errorf("scenario %s = %v", AttrNodeType, v)
		}
		if v := attrValue(spans[2], AttrBizdate); v != "99990101" {
			t.Errorf("bizdate %s = %v", AttrBizdate, v)
		}
		if v := attrValue(spans[2], AttrSeq); v != "10" {
			t.Errorf("bizdate %s = %v", AttrSeq, v)
		}
		if v := attrValue(spans[3], AttrGroup); v != "pre" {
			t.Errorf("process %s = %v", AttrGroup, v)
		}
		if v := attrValue(spans[3], AttrProcessType); v != "scripts" {
			t.Errorf("process %s = %v", AttrProcessType, v)
		}
		if v := attrValue(spans[4], AttrNodeType); v != "step" {
			t.Errorf("step %s = %v", AttrNodeType, v)
		}
		if v := attrValue(spans[4], AttrStepStatus); v != "Success" {
			t.Errorf("step %s = %v", AttrStepStatus, v)
		}
		if v := attrValue(spans[4], AttrStepExit); v != int64(0) {
			t.Errorf("step %s = %v (%T), want int64(0)", AttrStepExit, v, v)
		}
	})
}

func TestProjectorTreeError(t *testing.T) {
	t.Run("Projector_ステップエラーの場合_後続をBlockedにし全階層をErrorにすること", func(t *testing.T) {
		// Arrange
		ids := testNodeIDs(t)
		p := NewProjector()
		// step1 がエラー終了 → step2 は Blocked、全階層の end が Error
		events := []run.Event{
			run.NewNodeStartEvent(ts(0), ids["run"], run.NodeTypeRun,
				map[string]string{"run_id": "_20200101120000_99", "run_mode": "--dry-run", "params": "demo"}),
			run.NewNodeStartEvent(ts(1), ids["scenario"], run.NodeTypeScenario,
				map[string]string{"name": "demo"}),
			run.NewNodeStartEvent(ts(2), ids["bizdate"], run.NodeTypeBizdate,
				map[string]string{"dirname": "_10_99990101", "seq": "10", "bizdate": "99990101"}),
			run.NewNodeStartEvent(ts(3), ids["process"], run.NodeTypeProcess,
				map[string]string{"dirname": "_10_pre_scripts", "seq": "10", "group": "pre", "process_type": "scripts"}),
			run.NewStepsEnumeratedEvent(ts(3), ids["process"], []string{"100_step1", "200_step2"}),
			run.NewStepEndEvent(ts(4), ids["process"], "100_step1", run.StepError, 6, ts(3), ts(4)),
			run.NewStepBlockedEvent(ts(5), ids["process"], "200_step2"),
			run.NewNodeEndEvent(ts(6), ids["process"], run.NodeError),
			run.NewNodeEndEvent(ts(7), ids["bizdate"], run.NodeError),
			run.NewNodeEndEvent(ts(8), ids["scenario"], run.NodeError),
			run.NewNodeEndEvent(ts(9), ids["run"], run.NodeError),
		}

		// Act
		spans := applyAll(t, p, events)

		// Assert
		if len(spans) != 6 {
			t.Fatalf("spans = %d, want 6", len(spans))
		}

		// dry-run は stfw.run.mode=dry-run
		if v := attrValue(spans[0], AttrRunMode); v != "dry-run" {
			t.Errorf("run %s = %v, want dry-run", AttrRunMode, v)
		}

		// Error 終了の階層はスパンステータス Error (メッセージ付き)
		for _, i := range []int{0, 1, 2, 3} {
			if spans[i].Status != SpanStatusError {
				t.Errorf("span %s status = %s, want Error", spans[i].Name, spans[i].Status)
			}
			if !strings.Contains(spans[i].StatusMessage, "Error") {
				t.Errorf("span %s message = %q", spans[i].Name, spans[i].StatusMessage)
			}
		}

		// エラーステップはスパンステータス Error + exit_code 属性
		step1 := spans[4]
		if step1.Status != SpanStatusError {
			t.Errorf("step1 status = %s, want Error", step1.Status)
		}
		if !strings.Contains(step1.StatusMessage, "exit_code 6") {
			t.Errorf("step1 message = %q", step1.StatusMessage)
		}
		if v := attrValue(step1, AttrStepExit); v != int64(6) {
			t.Errorf("step1 %s = %v", AttrStepExit, v)
		}

		// Blocked ステップはスパンステータス Unset のまま属性で表現する
		step2 := spans[5]
		if step2.Status != SpanStatusUnset {
			t.Errorf("step2 status = %s, want Unset", step2.Status)
		}
		if v := attrValue(step2, AttrStepStatus); v != "Blocked" {
			t.Errorf("step2 %s = %v, want Blocked", AttrStepStatus, v)
		}
		if v := attrValue(step2, AttrStepExit); v != nil {
			t.Errorf("step2 %s = %v, want unset", AttrStepExit, v)
		}
		// 未実行のためイベント時刻を点として持つ
		if !step2.Start.Equal(ts(5)) || !step2.End.Equal(ts(5)) {
			t.Errorf("step2 time = %v-%v, want %v", step2.Start, step2.End, ts(5))
		}
	})
}

func TestProjectorTreeWarn(t *testing.T) {
	t.Run("Projector_ステップWarnの場合_スパンはOkのまま属性でWarnを表現すること", func(t *testing.T) {
		// Arrange
		ids := testNodeIDs(t)
		p := NewProjector()
		// step1 が Warn 終了 → 後続 step2 は実行され Success、全階層の end が Warn
		events := []run.Event{
			run.NewNodeStartEvent(ts(0), ids["run"], run.NodeTypeRun,
				map[string]string{"run_id": "_20200101120000_99", "run_mode": "--run", "params": "demo"}),
			run.NewNodeStartEvent(ts(1), ids["scenario"], run.NodeTypeScenario,
				map[string]string{"name": "demo"}),
			run.NewNodeStartEvent(ts(2), ids["bizdate"], run.NodeTypeBizdate,
				map[string]string{"dirname": "_10_99990101", "seq": "10", "bizdate": "99990101"}),
			run.NewNodeStartEvent(ts(3), ids["process"], run.NodeTypeProcess,
				map[string]string{"dirname": "_10_pre_scripts", "seq": "10", "group": "pre", "process_type": "scripts"}),
			run.NewStepsEnumeratedEvent(ts(3), ids["process"], []string{"100_step1", "200_step2"}),
			run.NewStepEndEvent(ts(4), ids["process"], "100_step1", run.StepWarn, 3, ts(3), ts(4)),
			run.NewStepEndEvent(ts(5), ids["process"], "200_step2", run.StepSuccess, 0, ts(4), ts(5)),
			run.NewNodeEndEvent(ts(6), ids["process"], run.NodeWarn),
			run.NewNodeEndEvent(ts(7), ids["bizdate"], run.NodeWarn),
			run.NewNodeEndEvent(ts(8), ids["scenario"], run.NodeWarn),
			run.NewNodeEndEvent(ts(9), ids["run"], run.NodeWarn),
		}

		// Act
		spans := applyAll(t, p, events)

		// Assert
		if len(spans) != 6 {
			t.Fatalf("spans = %d, want 6", len(spans))
		}

		// Warn 終了の階層はスパンステータス Ok のまま stfw.node.status=Warn (メッセージなし)
		for _, i := range []int{0, 1, 2, 3} {
			if spans[i].Status != SpanStatusOK {
				t.Errorf("span %s status = %s, want Ok", spans[i].Name, spans[i].Status)
			}
			if v := attrValue(spans[i], AttrNodeStatus); v != "Warn" {
				t.Errorf("span %s %s = %v, want Warn", spans[i].Name, AttrNodeStatus, v)
			}
			if spans[i].StatusMessage != "" {
				t.Errorf("span %s message = %q, want empty", spans[i].Name, spans[i].StatusMessage)
			}
		}

		// Warn ステップはスパンステータス Ok + stfw.step.status=Warn + exit_code 3
		step1 := spans[4]
		if step1.Status != SpanStatusOK {
			t.Errorf("step1 status = %s, want Ok", step1.Status)
		}
		if v := attrValue(step1, AttrStepStatus); v != "Warn" {
			t.Errorf("step1 %s = %v, want Warn", AttrStepStatus, v)
		}
		if v := attrValue(step1, AttrStepExit); v != int64(3) {
			t.Errorf("step1 %s = %v, want 3", AttrStepExit, v)
		}
		// Warn 後の Success ステップは実行されている (Ok + exit_code 0)
		step2 := spans[5]
		if step2.Status != SpanStatusOK {
			t.Errorf("step2 status = %s, want Ok", step2.Status)
		}
		if v := attrValue(step2, AttrStepExit); v != int64(0) {
			t.Errorf("step2 %s = %v, want 0", AttrStepExit, v)
		}
	})
}

func TestProjectorUnknownNode(t *testing.T) {
	t.Run("Projector_未開始ノードへのstep_endの場合_エラーであること", func(t *testing.T) {
		// Arrange
		ids := testNodeIDs(t)
		p := NewProjector()
		// Act
		_, err := p.Apply(run.NewStepEndEvent(ts(1), ids["process"], "100_step1", run.StepSuccess, 0, ts(0), ts(1)))
		// Assert
		if err == nil {
			t.Error("step_end for unknown node must fail")
		}
	})
	t.Run("Projector_未開始ノードへのnode_endの場合_エラーであること", func(t *testing.T) {
		// Arrange
		ids := testNodeIDs(t)
		p := NewProjector()
		// Act
		_, err := p.Apply(run.NewNodeEndEvent(ts(1), ids["run"], run.NodeSuccess))
		// Assert
		if err == nil {
			t.Error("node_end for unknown node must fail")
		}
	})
	t.Run("Projector_未知のイベント種別の場合_エラーであること", func(t *testing.T) {
		// Arrange
		p := NewProjector()
		// Act
		_, err := p.Apply(run.Event{Type: "unknown"})
		// Assert
		if err == nil {
			t.Error("unknown event type must fail")
		}
	})
}
