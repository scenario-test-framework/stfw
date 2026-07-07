package run

import (
	"testing"
	"time"
)

func TestNewRunID(t *testing.T) {
	// 採番規則は v0.2 の `_{yyyymmddhhmmss}_{pid}` と同一
	ts := time.Date(2020, 1, 1, 12, 34, 56, 0, time.Local)
	id := NewRunID(ts, 123)
	if id.String() != "_20200101123456_123" {
		t.Errorf("NewRunID() = %s, want _20200101123456_123", id)
	}
	if _, err := ParseRunID(id.String()); err != nil {
		t.Errorf("ParseRunID(%s) error = %v", id, err)
	}
}

func TestParseRunIDInvalid(t *testing.T) {
	for _, s := range []string{"", "20200101123456_123", "_2020_123", "_20200101123456_", "_20200101123456_12a"} {
		if _, err := ParseRunID(s); err == nil {
			t.Errorf("ParseRunID(%q) should fail", s)
		}
	}
}

func TestNodeIDDerivation(t *testing.T) {
	// webhook_id 導出規則 (v0.2 の `}` バグは修正済み) と同一の階層 ID
	runID := NewRunID(time.Date(2020, 1, 1, 12, 0, 0, 0, time.Local), 99)

	runNode := NewRunNodeID(runID)
	if runNode.String() != "_20200101120000_99+run" {
		t.Errorf("run node = %s", runNode)
	}
	if runNode.Parent() != "_20200101120000_99" {
		t.Errorf("run node parent = %s, want run_id", runNode.Parent())
	}

	scenarioNode, err := runNode.Child("scenario1")
	if err != nil {
		t.Fatal(err)
	}
	bizdateNode, err := scenarioNode.Child("_10_99990101")
	if err != nil {
		t.Fatal(err)
	}
	processNode, err := bizdateNode.Child("_10_pre_scripts")
	if err != nil {
		t.Fatal(err)
	}

	want := "_20200101120000_99+run+scenario1+_10_99990101+_10_pre_scripts"
	if processNode.String() != want {
		t.Errorf("process node = %s, want %s", processNode, want)
	}
	if processNode.Parent() != bizdateNode.String() {
		t.Errorf("process parent = %s, want %s", processNode.Parent(), bizdateNode)
	}
}

func TestNodeIDChildInvalidSegment(t *testing.T) {
	runNode := NewRunNodeID(NewRunID(time.Now(), 1))
	for _, seg := range []string{"", "a+b", "a^b", "a/b", `a\b`} {
		if _, err := runNode.Child(seg); err == nil {
			t.Errorf("Child(%q) should fail", seg)
		}
	}
}

func TestParseNodeID(t *testing.T) {
	runID, _ := ParseRunID("_20200101120000_99")
	tests := []struct {
		name     string
		id       string
		nodeType NodeType
		wantErr  bool
	}{
		{"run 階層", "_20200101120000_99+run", NodeTypeRun, false},
		{"process 階層", "_20200101120000_99+run+s1+_10_99990101+_10_pre_scripts", NodeTypeProcess, false},
		{"他 run 配下", "_20200101120001_99+run", NodeTypeRun, true},
		{"深さと種別の不一致", "_20200101120000_99+run+s1", NodeTypeRun, true},
		{"先頭セグメントが run でない", "_20200101120000_99+abc", NodeTypeRun, true},
		{"空セグメント", "_20200101120000_99+run++x", NodeTypeBizdate, true},
		{"未知の種別", "_20200101120000_99+run", NodeType("job"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseNodeID(runID, tt.id, tt.nodeType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNodeID(%s, %s) error = %v, wantErr %v", tt.id, tt.nodeType, err, tt.wantErr)
			}
		})
	}
}
