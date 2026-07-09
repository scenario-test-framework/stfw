package run

import "testing"

func TestNodeStatusTransition(t *testing.T) {
	tests := []struct {
		name    string
		from    NodeStatus
		to      NodeStatus
		wantErr bool
	}{
		{"Started→Success", NodeStarted, NodeSuccess, false},
		{"Started→Error", NodeStarted, NodeError, false},
		{"Started→Started は不正", NodeStarted, NodeStarted, true},
		{"Success→Error は不正 (終了状態)", NodeSuccess, NodeError, true},
		{"Error→Success は不正 (終了状態)", NodeError, NodeSuccess, true},
		{"Started→不明な値 は不正", NodeStarted, NodeStatus("Pending"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.from.Transition(tt.to)
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
		{"Pending→Success", StepPending, StepSuccess, false},
		{"Pending→Error", StepPending, StepError, false},
		{"Pending→Blocked", StepPending, StepBlocked, false},
		{"Pending→Pending は不正", StepPending, StepPending, true},
		{"Success→Error は不正 (終了状態)", StepSuccess, StepError, true},
		{"Blocked→Success は不正 (終了状態)", StepBlocked, StepSuccess, true},
		{"Pending→不明な値 は不正", StepPending, StepStatus("Started"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.from.Transition(tt.to)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transition(%s→%s) error = %v, wantErr %v", tt.from, tt.to, err, tt.wantErr)
			}
		})
	}
}

func TestStepsBlockedRule(t *testing.T) {
	steps, err := NewSteps([]string{"100_1st", "200_2nd", "300_3rd"})
	if err != nil {
		t.Fatal(err)
	}
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
}

func TestNewStepsValidation(t *testing.T) {
	if _, err := NewSteps(nil); err == nil {
		t.Error("NewSteps(nil) should fail")
	}
	if _, err := NewSteps([]string{"a", "a"}); err == nil {
		t.Error("NewSteps with duplicated name should fail")
	}
	if _, err := NewSteps([]string{""}); err == nil {
		t.Error("NewSteps with empty name should fail")
	}
}
