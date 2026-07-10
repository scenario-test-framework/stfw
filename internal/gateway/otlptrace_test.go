package gateway

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"

	"github.com/scenario-test-framework/stfw/internal/domain/notify"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// exportEvents はジャーナルイベント列を投影して InMemoryExporter へエクスポートする。
func exportEvents(t *testing.T, events []run.Event) tracetest.SpanStubs {
	t.Helper()
	projector := notify.NewProjector()
	var spans []notify.Span
	for i, ev := range events {
		s, err := projector.Apply(ev)
		if err != nil {
			t.Fatalf("event %d (%s): %v", i, ev.Type, err)
		}
		spans = append(spans, s...)
	}
	if len(spans) == 0 {
		t.Fatal("projector returned no spans")
	}

	exp := tracetest.NewInMemoryExporter()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	exporter := NewTraceExporter(log, exp, "1.0.0-test")
	exporter.Export(spans)
	// InMemoryExporter は Shutdown で記録を破棄するため flush 後に取得する
	if err := exporter.tp.ForceFlush(context.Background()); err != nil {
		t.Fatal(err)
	}
	stubs := exp.GetSpans()
	exporter.Shutdown()
	return stubs
}

// testEvents は run > scenario(demo) > bizdate x1 > process x1 x steps x2 のイベント列を組み立てる。
// failStep=true は step1 エラー + step2 Blocked + 全階層 Error の系列。
func testEvents(t *testing.T, failStep bool) []run.Event {
	t.Helper()
	runID, err := run.ParseRunID("_20200101120000_99")
	if err != nil {
		t.Fatal(err)
	}
	runNode := run.NewRunNodeID(runID)
	scenario, err := runNode.Child("demo")
	if err != nil {
		t.Fatal(err)
	}
	bizdate, err := scenario.Child("_10_99990101")
	if err != nil {
		t.Fatal(err)
	}
	process, err := bizdate.Child("_10_pre_scripts")
	if err != nil {
		t.Fatal(err)
	}

	events := []run.Event{
		run.NewNodeStartEvent(ts(0), runNode, run.NodeTypeRun,
			map[string]string{"run_id": "_20200101120000_99", "run_mode": "--run", "params": "demo"}),
		run.NewNodeStartEvent(ts(1), scenario, run.NodeTypeScenario,
			map[string]string{"name": "demo"}),
		run.NewNodeStartEvent(ts(2), bizdate, run.NodeTypeBizdate,
			map[string]string{"dirname": "_10_99990101", "seq": "10", "bizdate": "99990101"}),
		run.NewNodeStartEvent(ts(3), process, run.NodeTypeProcess,
			map[string]string{"dirname": "_10_pre_scripts", "seq": "10", "group": "pre", "process_type": "scripts"}),
		run.NewStepsEnumeratedEvent(ts(3), process, []string{"100_step1", "200_step2"}),
	}
	status := run.NodeSuccess
	if failStep {
		status = run.NodeError
		events = append(events,
			run.NewStepEndEvent(ts(4), process, "100_step1", run.StepError, 6, ts(3), ts(4)),
			run.NewStepBlockedEvent(ts(5), process, "200_step2"),
		)
	} else {
		events = append(events,
			run.NewStepEndEvent(ts(4), process, "100_step1", run.StepSuccess, 0, ts(3), ts(4)),
			run.NewStepEndEvent(ts(5), process, "200_step2", run.StepSuccess, 0, ts(4), ts(5)),
		)
	}
	return append(events,
		run.NewNodeEndEvent(ts(6), process, status),
		run.NewNodeEndEvent(ts(7), bizdate, status),
		run.NewNodeEndEvent(ts(8), scenario, status),
		run.NewNodeEndEvent(ts(9), runNode, status),
	)
}

func ts(sec int) time.Time {
	return time.Date(2020, 1, 1, 12, 0, sec, 0, time.FixedZone("JST", 9*60*60))
}

// stubByName は名前でスパンを取得する。
func stubByName(t *testing.T, stubs tracetest.SpanStubs, name string) tracetest.SpanStub {
	t.Helper()
	for _, s := range stubs {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("span %s is not exported", name)
	return tracetest.SpanStub{}
}

// stubAttr はスパン属性の値を返す (未設定は zero Value)。
func stubAttr(s tracetest.SpanStub, key string) attribute.Value {
	for _, kv := range s.Attributes {
		if string(kv.Key) == key {
			return kv.Value
		}
	}
	return attribute.Value{}
}

func TestTraceExporterSuccessTree(t *testing.T) {
	t.Run("TraceExporter_全Success系列の場合_同一トレースの親子ツリーとOkステータスになること", func(t *testing.T) {
		// Arrange
		events := testEvents(t, false)

		// Act
		stubs := exportEvents(t, events)

		// Assert
		if len(stubs) != 6 {
			t.Fatalf("exported spans = %d, want 6", len(stubs))
		}

		root := stubByName(t, stubs, "stfw run")
		scenario := stubByName(t, stubs, "demo")
		bizdate := stubByName(t, stubs, "_10_99990101")
		process := stubByName(t, stubs, "_10_pre_scripts")
		step1 := stubByName(t, stubs, "100_step1")
		step2 := stubByName(t, stubs, "200_step2")

		// 1 run = 1 トレース: 全スパンが同一 TraceID
		traceID := root.SpanContext.TraceID()
		for _, s := range stubs {
			if s.SpanContext.TraceID() != traceID {
				t.Errorf("span %s trace_id = %s, want %s", s.Name, s.SpanContext.TraceID(), traceID)
			}
		}

		// 親子関係: run をルートに scenario > bizdate > process > step
		if root.Parent.IsValid() {
			t.Errorf("run span must be root, parent = %+v", root.Parent)
		}
		pairs := []struct {
			child, parent tracetest.SpanStub
		}{
			{scenario, root}, {bizdate, scenario}, {process, bizdate}, {step1, process}, {step2, process},
		}
		for _, p := range pairs {
			if p.child.Parent.SpanID() != p.parent.SpanContext.SpanID() {
				t.Errorf("span %s parent = %s, want %s (%s)",
					p.child.Name, p.child.Parent.SpanID(), p.parent.SpanContext.SpanID(), p.parent.Name)
			}
		}

		// 開始・終了時刻はジャーナルイベントの時刻と一致する
		if !root.StartTime.Equal(ts(0)) || !root.EndTime.Equal(ts(9)) {
			t.Errorf("run span time = %v-%v, want %v-%v", root.StartTime, root.EndTime, ts(0), ts(9))
		}
		if !step1.StartTime.Equal(ts(3)) || !step1.EndTime.Equal(ts(4)) {
			t.Errorf("step1 span time = %v-%v, want %v-%v", step1.StartTime, step1.EndTime, ts(3), ts(4))
		}

		// 全 Success → 全スパン Ok
		for _, s := range stubs {
			if s.Status.Code != codes.Ok {
				t.Errorf("span %s status = %s, want Ok", s.Name, s.Status.Code)
			}
		}

		// スパン属性 (実行コンテキスト)
		if v := stubAttr(root, notify.AttrRunMode).AsString(); v != "run" {
			t.Errorf("run %s = %q, want run", notify.AttrRunMode, v)
		}
		if v := stubAttr(bizdate, notify.AttrBizdate).AsString(); v != "99990101" {
			t.Errorf("bizdate %s = %q", notify.AttrBizdate, v)
		}
		if v := stubAttr(process, notify.AttrProcessType).AsString(); v != "scripts" {
			t.Errorf("process %s = %q", notify.AttrProcessType, v)
		}
		if v := stubAttr(step1, notify.AttrStepExit).AsInt64(); v != 0 {
			t.Errorf("step1 %s = %d, want 0", notify.AttrStepExit, v)
		}
		for _, s := range stubs {
			if v := stubAttr(s, notify.AttrRunID).AsString(); v != "_20200101120000_99" {
				t.Errorf("span %s %s = %q", s.Name, notify.AttrRunID, v)
			}
		}

		// リソース属性: service.name / service.version
		res := root.Resource.Attributes()
		found := map[attribute.Key]string{}
		for _, kv := range res {
			found[kv.Key] = kv.Value.AsString()
		}
		if found[semconv.ServiceNameKey] != "stfw" {
			t.Errorf("service.name = %q, want stfw", found[semconv.ServiceNameKey])
		}
		if found[semconv.ServiceVersionKey] != "1.0.0-test" {
			t.Errorf("service.version = %q, want 1.0.0-test", found[semconv.ServiceVersionKey])
		}
	})
}

func TestTraceExporterErrorTree(t *testing.T) {
	t.Run("TraceExporter_stepエラーとBlockedを含む系列の場合_ErrorとUnsetを正しく表現すること", func(t *testing.T) {
		// Arrange
		events := testEvents(t, true)

		// Act
		stubs := exportEvents(t, events)

		// Assert
		if len(stubs) != 6 {
			t.Fatalf("exported spans = %d, want 6", len(stubs))
		}

		// Error 終了の階層はスパンステータス Error (メッセージ付き)
		for _, name := range []string{"stfw run", "demo", "_10_99990101", "_10_pre_scripts"} {
			s := stubByName(t, stubs, name)
			if s.Status.Code != codes.Error || s.Status.Description == "" {
				t.Errorf("span %s status = %s (%q), want Error with message", name, s.Status.Code, s.Status.Description)
			}
		}

		// エラーステップはスパンステータス Error
		step1 := stubByName(t, stubs, "100_step1")
		if step1.Status.Code != codes.Error {
			t.Errorf("step1 status = %s, want Error", step1.Status.Code)
		}
		if v := stubAttr(step1, notify.AttrStepExit).AsInt64(); v != 6 {
			t.Errorf("step1 %s = %d, want 6", notify.AttrStepExit, v)
		}

		// Blocked ステップはスパンステータス Unset のまま属性で表現する
		step2 := stubByName(t, stubs, "200_step2")
		if step2.Status.Code != codes.Unset {
			t.Errorf("step2 status = %s, want Unset", step2.Status.Code)
		}
		if v := stubAttr(step2, notify.AttrStepStatus).AsString(); v != "Blocked" {
			t.Errorf("step2 %s = %q, want Blocked", notify.AttrStepStatus, v)
		}
	})
}
