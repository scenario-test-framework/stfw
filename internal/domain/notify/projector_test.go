package notify

import (
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// testEvents は run > scenario > bizdate > process の開始イベント列を組み立てる。
func testProjector(t *testing.T) (*Projector, map[string]run.NodeID) {
	t.Helper()
	ctx := Context{
		Host:           "192.0.2.1",
		User:           "tester",
		Version:        "1.0.0-test",
		ProjectVersion: "0.1.0",
		ProjectHome:    "/proj",
		WorkspaceDir:   "/proj/.stfw/runs/_20200101120000_99",
	}
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
	return NewProjector(ctx), ids
}

func ts(sec int) time.Time {
	return time.Date(2020, 1, 1, 12, 0, sec, 0, time.FixedZone("JST", 9*60*60))
}

func TestProjectorHierarchy(t *testing.T) {
	p, ids := testProjector(t)
	now := ts(0)

	// run start: 即時に通知 1 件
	notifs, err := p.Project(run.NewNodeStartEvent(ts(0), ids["run"], run.NodeTypeRun,
		map[string]string{"run_id": "_20200101120000_99", "run_mode": "--run", "params": "demo"}), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(notifs) != 1 || notifs[0].Event != EventStart {
		t.Fatalf("run start notifs = %+v, want 1 start", notifs)
	}
	body := notifs[0].Payload.Payload
	if body.ID != "_20200101120000_99+run" || body.ParentID != "_20200101120000_99" {
		t.Errorf("id/parent_id = %v/%v", body.ID, body.ParentID)
	}
	if body.Type != "run" || body.Status != "Started" {
		t.Errorf("type/status = %s/%s", body.Type, body.Status)
	}
	// payload の時刻は v0.2 の timestamp_to_iso と同じコロン無しタイムゾーン
	if body.StartTime != "2020-01-01T12:00:00+0900" {
		t.Errorf("start_time = %s", body.StartTime)
	}
	if body.EndTime != "" || body.ProcessingTime != "" {
		t.Errorf("start payload must have empty end_time/processing_time: %+v", body)
	}
	if body.Run == nil || body.Run.RunID != "_20200101120000_99" || body.Run.Params != "demo" {
		t.Errorf("run attrs = %+v", body.Run)
	}
	if body.Run.Scenario != nil {
		t.Errorf("run payload must not contain scenario: %+v", body.Run.Scenario)
	}

	// scenario / bizdate start
	if _, err := p.Project(run.NewNodeStartEvent(ts(1), ids["scenario"], run.NodeTypeScenario,
		map[string]string{"name": "demo"}), now); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Project(run.NewNodeStartEvent(ts(2), ids["bizdate"], run.NodeTypeBizdate,
		map[string]string{"dirname": "_10_99990101", "seq": "10", "bizdate": "99990101"}), now); err != nil {
		t.Fatal(err)
	}

	// process start はステップ列挙まで保留される (v0.2 の Pending 列挙付き start payload)
	notifs, err = p.Project(run.NewNodeStartEvent(ts(3), ids["process"], run.NodeTypeProcess,
		map[string]string{"dirname": "_10_pre_scripts", "seq": "10", "group": "pre", "process_type": "scripts"}), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(notifs) != 0 {
		t.Fatalf("process start must be deferred, got %+v", notifs)
	}
	notifs, err = p.Project(run.NewStepsEnumeratedEvent(ts(3), ids["process"], []string{"100_step1", "200_step2"}), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(notifs) != 1 || notifs[0].Event != EventStart {
		t.Fatalf("steps_enumerated notifs = %+v, want deferred start", notifs)
	}
	process := notifs[0].Payload.Payload.Run.Scenario.Bizdate.Process
	// 階層別属性の unquoted 値は yaml2json (PyYAML) と同じ型になる
	if process.Seq != int64(10) {
		t.Errorf("process seq = %v (%T), want int64(10)", process.Seq, process.Seq)
	}
	if bizdate := notifs[0].Payload.Payload.Run.Scenario.Bizdate; bizdate.Bizdate != int64(99990101) {
		t.Errorf("bizdate = %v (%T), want int64(99990101)", bizdate.Bizdate, bizdate.Bizdate)
	}
	if process.Plugin == nil || process.Plugin.Type != "scripts" {
		t.Fatalf("plugin = %+v", process.Plugin)
	}
	if len(process.Plugin.Targets) != 2 {
		t.Fatalf("targets = %+v", process.Plugin.Targets)
	}
	if target := process.Plugin.Targets[0]["100_step1"]; target.Result != "Pending" || target.StartTime != "" {
		t.Errorf("pending target = %+v", target)
	}

	// step_end: Success + Error、後続 Blocked
	if _, err := p.Project(run.NewStepEndEvent(ts(5), ids["process"], "100_step1", run.StepError, 6, ts(4), ts(5)), now); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Project(run.NewStepBlockedEvent(ts(5), ids["process"], "200_step2"), now); err != nil {
		t.Fatal(err)
	}

	// process end: ステップ詳細と処理時間を含む
	notifs, err = p.Project(run.NewNodeEndEvent(ts(6), ids["process"], run.NodeError), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(notifs) != 1 || notifs[0].Event != EventEnd || notifs[0].Status != "Error" {
		t.Fatalf("process end notifs = %+v", notifs)
	}
	body = notifs[0].Payload.Payload
	if body.EndTime != "2020-01-01T12:00:06+0900" || body.ProcessingTime != "00:00:03" {
		t.Errorf("end_time/processing_time = %s/%s", body.EndTime, body.ProcessingTime)
	}
	targets := body.Run.Scenario.Bizdate.Process.Plugin.Targets
	if target := targets[0]["100_step1"]; target.Result != "Error" || target.ProcessingTime != "00:00:01" {
		t.Errorf("error target = %+v", target)
	}
	if target := targets[1]["200_step2"]; target.Result != "Blocked" || target.StartTime != "" {
		t.Errorf("blocked target = %+v", target)
	}
}

func TestProjectorZeroStepProcess(t *testing.T) {
	// ステップ 0 件の scripts プロセス: steps_enumerated が来ないまま
	// node_end に到達した時点で保留中の start を flush する (targets は null)
	p, ids := testProjector(t)
	now := ts(0)
	seed := []run.Event{
		run.NewNodeStartEvent(ts(0), ids["run"], run.NodeTypeRun, map[string]string{"run_id": "_20200101120000_99", "params": "demo"}),
		run.NewNodeStartEvent(ts(1), ids["scenario"], run.NodeTypeScenario, map[string]string{"name": "demo"}),
		run.NewNodeStartEvent(ts(2), ids["bizdate"], run.NodeTypeBizdate, map[string]string{"dirname": "_10_99990101", "seq": "10", "bizdate": "99990101"}),
	}
	for _, ev := range seed {
		if _, err := p.Project(ev, now); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := p.Project(run.NewNodeStartEvent(ts(3), ids["process"], run.NodeTypeProcess,
		map[string]string{"dirname": "_10_pre_scripts", "seq": "10", "group": "pre", "process_type": "scripts"}), now); err != nil {
		t.Fatal(err)
	}
	notifs, err := p.Project(run.NewNodeEndEvent(ts(4), ids["process"], run.NodeSuccess), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(notifs) != 2 || notifs[0].Event != EventStart || notifs[1].Event != EventEnd {
		t.Fatalf("notifs = %+v, want flushed start + end", notifs)
	}
	if targets := notifs[0].Payload.Payload.Run.Scenario.Bizdate.Process.Plugin.Targets; targets != nil {
		t.Errorf("zero-step targets = %+v, want nil", targets)
	}
}

func TestYamlScalar(t *testing.T) {
	// v0.2 の yaml2json (PyYAML) の unquoted スカラー解決を再現する
	tests := []struct {
		in   string
		want any
	}{
		{in: "", want: nil},
		{in: "10", want: int64(10)},
		{in: "0", want: int64(0)},
		{in: "99990101", want: int64(99990101)},
		{in: "05", want: int64(5)}, // 先頭 0 は 8 進
		{in: "09", want: "09"},     // 8 進として不正なら文字列
		{in: "_10_99990101", want: "_10_99990101"},
		{in: "demo", want: "demo"},
	}
	for _, tt := range tests {
		if got := yamlScalar(tt.in); got != tt.want {
			t.Errorf("yamlScalar(%q) = %v (%T), want %v", tt.in, got, got, tt.want)
		}
	}
}
