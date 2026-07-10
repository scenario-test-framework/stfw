package run

import (
	"testing"
	"time"
)

func TestNewRunID(t *testing.T) {
	t.Run("NewRunID_採番時刻とpidを渡した場合_v0.2互換のID文字列でありParseできること", func(t *testing.T) {
		// Arrange
		// 採番規則は v0.2 の `_{yyyymmddhhmmss}_{pid}` と同一
		ts := time.Date(2020, 1, 1, 12, 34, 56, 0, time.Local)
		// Act
		id := NewRunID(ts, 123)
		// Assert
		if id.String() != "_20200101123456_123" {
			t.Errorf("NewRunID() = %s, want _20200101123456_123", id)
		}
		if _, err := ParseRunID(id.String()); err != nil {
			t.Errorf("ParseRunID(%s) error = %v", id, err)
		}
	})
}

func TestParseRunIDInvalid(t *testing.T) {
	t.Run("ParseRunID_不正な書式の場合_エラーであること", func(t *testing.T) {
		// Arrange
		invalid := []string{"", "20200101123456_123", "_2020_123", "_20200101123456_", "_20200101123456_12a"}
		for _, s := range invalid {
			// Act
			_, err := ParseRunID(s)
			// Assert
			if err == nil {
				t.Errorf("ParseRunID(%q) should fail", s)
			}
		}
	})
}

func TestRunIDTime(t *testing.T) {
	t.Run("Time_採番時刻を埋め込んだRunIDの場合_ラウンドトリップで復元されること", func(t *testing.T) {
		// Arrange
		// run_id 埋め込みの採番時刻がラウンドトリップで復元される (ハウスキープの保存期間判定)
		ts := time.Date(2020, 1, 1, 12, 34, 56, 0, time.Local)
		// Act
		got, err := NewRunID(ts, 123).Time()
		// Assert
		if err != nil {
			t.Fatalf("Time() error = %v", err)
		}
		if !got.Equal(ts) {
			t.Errorf("Time() = %v, want %v", got, ts)
		}
	})

	t.Run("Time_ゼロ値RunIDの場合_panicせずエラーを返すこと", func(t *testing.T) {
		// Act
		// ゼロ値 RunID は panic せずエラーを返す
		_, err := (RunID{}).Time()
		// Assert
		if err == nil {
			t.Error("zero RunID Time() should fail")
		}
	})
}

func TestNodeIDDerivation(t *testing.T) {
	t.Run("NodeID_run配下に子ノードを導出する場合_階層IDとParentが一致すること", func(t *testing.T) {
		// Arrange
		// webhook_id 導出規則 (v0.2 の `}` バグは修正済み) と同一の階層 ID
		runID := NewRunID(time.Date(2020, 1, 1, 12, 0, 0, 0, time.Local), 99)

		// Act
		runNode := NewRunNodeID(runID)

		// Assert
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
	})
}

func TestNodeIDChildInvalidSegment(t *testing.T) {
	t.Run("Child_不正な文字を含むセグメントの場合_エラーであること", func(t *testing.T) {
		// Arrange
		runNode := NewRunNodeID(NewRunID(time.Now(), 1))
		for _, seg := range []string{"", "a+b", "a^b", "a/b", `a\b`} {
			// Act
			_, err := runNode.Child(seg)
			// Assert
			if err == nil {
				t.Errorf("Child(%q) should fail", seg)
			}
		}
	})
}

func TestParseNodeID(t *testing.T) {
	runID, _ := ParseRunID("_20200101120000_99")
	tests := []struct {
		name     string
		id       string
		nodeType NodeType
		wantErr  bool
	}{
		{"ParseNodeID_run階層のIDの場合_成功すること", "_20200101120000_99+run", NodeTypeRun, false},
		{"ParseNodeID_process階層のIDの場合_成功すること", "_20200101120000_99+run+s1+_10_99990101+_10_pre_scripts", NodeTypeProcess, false},
		{"ParseNodeID_他runのIDの場合_エラーであること", "_20200101120001_99+run", NodeTypeRun, true},
		{"ParseNodeID_深さと種別が不一致の場合_エラーであること", "_20200101120000_99+run+s1", NodeTypeRun, true},
		{"ParseNodeID_先頭セグメントがrunでない場合_エラーであること", "_20200101120000_99+abc", NodeTypeRun, true},
		{"ParseNodeID_空セグメントを含む場合_エラーであること", "_20200101120000_99+run++x", NodeTypeBizdate, true},
		{"ParseNodeID_未知の種別の場合_エラーであること", "_20200101120000_99+run", NodeType("job"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := ParseNodeID(runID, tt.id, tt.nodeType)
			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNodeID(%s, %s) error = %v, wantErr %v", tt.id, tt.nodeType, err, tt.wantErr)
			}
		})
	}
}
