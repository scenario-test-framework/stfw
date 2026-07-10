package run

import (
	"testing"
	"time"
)

var testTS = time.Date(2020, 1, 1, 12, 0, 0, 0, time.Local)

// validEvents は 1 scenario / 1 bizdate / 1 process (2 steps) の正常イベント列。
func validEvents(runID RunID) []Event {
	runNode := NewRunNodeID(runID)
	scenarioNode, _ := runNode.Child("s1")
	bizdateNode, _ := scenarioNode.Child("_10_99990101")
	processNode, _ := bizdateNode.Child("_10_pre_scripts")

	return []Event{
		NewNodeStartEvent(testTS, runNode, NodeTypeRun, map[string]string{"run_id": runID.String()}),
		NewNodeStartEvent(testTS, scenarioNode, NodeTypeScenario, map[string]string{"name": "s1"}),
		NewNodeStartEvent(testTS, bizdateNode, NodeTypeBizdate, nil),
		NewNodeStartEvent(testTS, processNode, NodeTypeProcess, nil),
		NewStepsEnumeratedEvent(testTS, processNode, []string{"100_1st", "200_2nd"}),
		NewStepEndEvent(testTS, processNode, "100_1st", StepSuccess, 0, testTS, testTS),
		NewStepEndEvent(testTS, processNode, "200_2nd", StepSuccess, 0, testTS, testTS),
		NewNodeEndEvent(testTS, processNode, NodeSuccess),
		NewNodeEndEvent(testTS, bizdateNode, NodeSuccess),
		NewNodeEndEvent(testTS, scenarioNode, NodeSuccess),
		NewNodeEndEvent(testTS, runNode, NodeSuccess),
	}
}

func TestReplayValidJournal(t *testing.T) {
	t.Run("Replay_正常なジャーナルの場合_全ノードがSuccessで深さとステップ数が一致すること", func(t *testing.T) {
		// Arrange
		runID := NewRunID(testTS, 1)

		// Act
		r, err := Replay(runID, validEvents(runID))

		// Assert
		if err != nil {
			t.Fatalf("Replay() error = %v", err)
		}

		views := r.NodeViews()
		if len(views) != 4 {
			t.Fatalf("NodeViews() = %d nodes, want 4", len(views))
		}
		for _, v := range views {
			if v.Status != NodeSuccess {
				t.Errorf("node %s = %s, want Success", v.ID, v.Status)
			}
		}
		// 深さは run=0, scenario=1, bizdate=2, process=3
		for i, wantDepth := range []int{0, 1, 2, 3} {
			if views[i].Depth != wantDepth {
				t.Errorf("node %s depth = %d, want %d", views[i].ID, views[i].Depth, wantDepth)
			}
		}
		// ステップは process ノードのみに 2 件
		if len(views[3].Steps) != 2 {
			t.Errorf("process steps = %d, want 2", len(views[3].Steps))
		}
	})
}

// mutate はイベント列の一部を差し替えて不正ジャーナルを作る。
func TestReplayInvalidJournal(t *testing.T) {
	runID := NewRunID(testTS, 1)
	runNode := NewRunNodeID(runID)
	scenarioNode, _ := runNode.Child("s1")
	bizdateNode, _ := scenarioNode.Child("_10_99990101")
	processNode, _ := bizdateNode.Child("_10_pre_scripts")

	tests := []struct {
		name   string
		mutate func(events []Event) []Event
	}{
		{"Replay_node_startが重複する場合_エラーであること", func(evs []Event) []Event {
			return append(evs[:1:1], append([]Event{evs[0]}, evs[1:]...)...)
		}},
		{"Replay_親未開始のnode_startの場合_エラーであること", func(evs []Event) []Event {
			return evs[1:] // run の node_start を欠落させる
		}},
		{"Replay_parent_idが改ざんされた場合_エラーであること", func(evs []Event) []Event {
			evs[1].ParentID = runID.String()
			return evs
		}},
		{"Replay_終了済みノードへのnode_endの場合_エラーであること", func(evs []Event) []Event {
			return append(evs, NewNodeEndEvent(testTS, runNode, NodeError))
		}},
		{"Replay_node_endが不正ステータスの場合_エラーであること", func(evs []Event) []Event {
			evs[10].Status = "Pending"
			return evs
		}},
		{"Replay_列挙前のstep_endの場合_エラーであること", func(evs []Event) []Event {
			return append(evs[:4:4], evs[5:]...) // steps_enumerated を欠落させる
		}},
		{"Replay_未列挙ステップのstep_endの場合_エラーであること", func(evs []Event) []Event {
			evs[5].Step = "999_unknown"
			return evs
		}},
		{"Replay_終了済みステップへのstep_endの場合_エラーであること", func(evs []Event) []Event {
			evs[6] = NewStepEndEvent(testTS, processNode, "100_1st", StepError, 6, testTS, testTS)
			return evs
		}},
		{"Replay_steps_enumeratedが重複する場合_エラーであること", func(evs []Event) []Event {
			return append(evs[:5:5], append([]Event{evs[4]}, evs[5:]...)...)
		}},
		{"Replay_process以外へのsteps_enumeratedの場合_エラーであること", func(evs []Event) []Event {
			evs[4] = NewStepsEnumeratedEvent(testTS, bizdateNode, []string{"100_1st"})
			return evs
		}},
		{"Replay_終了済みノードへのstep_endの場合_エラーであること", func(evs []Event) []Event {
			return append(evs, NewStepEndEvent(testTS, processNode, "200_2nd", StepError, 6, testTS, testTS))
		}},
		{"Replay_不明なイベント種別の場合_エラーであること", func(evs []Event) []Event {
			evs[0].Type = "node_pause"
			return evs
		}},
		{"Replay_別runのnode_idの場合_エラーであること", func(evs []Event) []Event {
			other := NewRunNodeID(NewRunID(testTS, 2))
			evs[0] = NewNodeStartEvent(testTS, other, NodeTypeRun, nil)
			return evs
		}},
		{"Replay_深さと種別が不一致の場合_エラーであること", func(evs []Event) []Event {
			evs[1].NodeType = NodeTypeBizdate
			return evs
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			events := tt.mutate(validEvents(runID))
			// Act
			_, err := Replay(runID, events)
			// Assert
			if err == nil {
				t.Error("Replay() should fail")
			}
		})
	}
}

func TestApplyRejectsInvalidTransitionOnGeneration(t *testing.T) {
	t.Run("Apply_終了済みノードへの再終了や配下追加の場合_エラーであること", func(t *testing.T) {
		// Arrange
		// 生成経路 (Apply) がリプレイと同じ検証を通すことを確認する。
		runID := NewRunID(testTS, 1)
		r := NewRun(runID)
		runNode := NewRunNodeID(runID)

		// Act
		if err := r.Apply(NewNodeStartEvent(testTS, runNode, NodeTypeRun, nil)); err != nil {
			t.Fatal(err)
		}
		if err := r.Apply(NewNodeEndEvent(testTS, runNode, NodeSuccess)); err != nil {
			t.Fatal(err)
		}

		// Assert
		// Success 終了後の再終了は不正
		if err := r.Apply(NewNodeEndEvent(testTS, runNode, NodeError)); err == nil {
			t.Error("Apply(node_end twice) should fail")
		}
		// 終了済み run 配下への node_start は不正 (親が Started ではない)
		scenarioNode, _ := runNode.Child("s1")
		if err := r.Apply(NewNodeStartEvent(testTS, scenarioNode, NodeTypeScenario, nil)); err == nil {
			t.Error("Apply(node_start under terminal parent) should fail")
		}
	})
}
