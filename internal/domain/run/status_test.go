package run

import "testing"

func TestNodeStatusTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    NodeStatus
		to      NodeStatus
		wantErr bool
	}{
		{"Transition_Started→Successの場合_成功すること", NodeStarted, NodeSuccess, false},
		{"Transition_Started→Warnの場合_成功すること", NodeStarted, NodeWarn, false},
		{"Transition_Started→Errorの場合_成功すること", NodeStarted, NodeError, false},
		{"Transition_終了状態Warn→Errorの場合_エラーであること", NodeWarn, NodeError, true},
		{"Transition_Started→Startedの場合_エラーであること", NodeStarted, NodeStarted, true},
		{"Transition_終了状態Success→Errorの場合_エラーであること", NodeSuccess, NodeError, true},
		{"Transition_終了状態Error→Successの場合_エラーであること", NodeError, NodeSuccess, true},
		{"Transition_不明な値への遷移の場合_エラーであること", NodeStarted, NodeStatus("Pending"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.from.Transition(tt.to)
			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Transition(%s→%s) error = %v, wantErr %v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}

func TestStepStatusTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    StepStatus
		to      StepStatus
		wantErr bool
	}{
		{"Transition_Pending→Successの場合_成功すること", StepPending, StepSuccess, false},
		{"Transition_Pending→Warnの場合_成功すること", StepPending, StepWarn, false},
		{"Transition_Pending→Errorの場合_成功すること", StepPending, StepError, false},
		{"Transition_Pending→Blockedの場合_成功すること", StepPending, StepBlocked, false},
		{"Transition_終了状態Warn→Successの場合_エラーであること", StepWarn, StepSuccess, true},
		{"Transition_Pending→Pendingの場合_エラーであること", StepPending, StepPending, true},
		{"Transition_終了状態Success→Errorの場合_エラーであること", StepSuccess, StepError, true},
		{"Transition_終了状態Blocked→Successの場合_エラーであること", StepBlocked, StepSuccess, true},
		{"Transition_不明な値への遷移の場合_エラーであること", StepPending, StepStatus("Started"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.from.Transition(tt.to)
			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Transition(%s→%s) error = %v, wantErr %v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}

func TestStepsBlockedRule(t *testing.T) {
	t.Run("MarkEnd_逐次実行でエラーが発生した場合_後続がBlockedになり終了済み・未列挙は不正になること", func(t *testing.T) {
		// Arrange
		steps, err := NewSteps([]string{"100_1st", "200_2nd", "300_3rd"})
		if err != nil {
			t.Fatal(err)
		}

		// Act
		// 1 件目 Success → 2 件目 Error → 3 件目 Blocked (逐次実行・エラー時 Blocked)
		if err := steps.MarkEnd("100_1st", StepSuccess); err != nil {
			t.Fatal(err)
		}
		if err := steps.MarkEnd("200_2nd", StepError); err != nil {
			t.Fatal(err)
		}
		if err := steps.MarkEnd("300_3rd", StepBlocked); err != nil {
			t.Fatal(err)
		}

		// Assert
		views := steps.Views()
		want := []StepStatus{StepSuccess, StepError, StepBlocked}
		for i, v := range views {
			if v.Status != want[i] {
				t.Errorf("step %s = %s, want %s", v.Name, v.Status, want[i])
			}
		}

		// 終了済みステップの再遷移は不正
		if err := steps.MarkEnd("100_1st", StepError); err == nil {
			t.Error("MarkEnd on terminal step should fail")
		}
		// 未列挙のステップは不正
		if err := steps.MarkEnd("999_none", StepSuccess); err == nil {
			t.Error("MarkEnd on unknown step should fail")
		}
	})
}

func TestNewStepsValidation(t *testing.T) {
	t.Run("NewSteps_nilの場合_エラーであること", func(t *testing.T) {
		// Act
		_, err := NewSteps(nil)
		// Assert
		if err == nil {
			t.Error("NewSteps(nil) should fail")
		}
	})
	t.Run("NewSteps_名前が重複する場合_エラーであること", func(t *testing.T) {
		// Act
		_, err := NewSteps([]string{"a", "a"})
		// Assert
		if err == nil {
			t.Error("NewSteps with duplicated name should fail")
		}
	})
	t.Run("NewSteps_空文字の名前の場合_エラーであること", func(t *testing.T) {
		// Act
		_, err := NewSteps([]string{""})
		// Assert
		if err == nil {
			t.Error("NewSteps with empty name should fail")
		}
	})
}
