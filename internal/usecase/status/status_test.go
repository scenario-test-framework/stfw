package status

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// writeWarnJournal は Warn を含む最小の実ジャーナル (run 階層のみ) を projDir へ書き込む
// (I/O 境界は実体でテストする)。
func writeWarnJournal(t *testing.T, projDir string) run.RunID {
	t.Helper()
	ts := time.Date(2026, 7, 12, 12, 0, 0, 0, time.Local)
	runID := run.NewRunID(ts, 99)
	journal, err := repository.CreateJournal(projDir, runID)
	if err != nil {
		t.Fatal(err)
	}
	node := run.NewRunNodeID(runID)
	attrs := map[string]string{"run_id": runID.String(), "run_mode": "--run", "params": "demo"}
	if err := journal.Append(run.NewNodeStartEvent(ts, node, run.NodeTypeRun, attrs)); err != nil {
		t.Fatal(err)
	}
	if err := journal.Append(run.NewNodeEndEvent(ts.Add(time.Second), node, run.NodeWarn)); err != nil {
		t.Fatal(err)
	}
	if err := journal.Close(); err != nil {
		t.Fatal(err)
	}
	return runID
}

func TestShowColor(t *testing.T) {
	t.Run("Show_color有効でWarnを含む場合_黄系のANSIカラーで表示されること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		runID := writeWarnJournal(t, projDir)
		var out bytes.Buffer

		// Act
		err := Show(&out, projDir, runID.String(), true)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), "[\x1b[33mWarn\x1b[0m]") {
			t.Errorf("output = %q, want yellow ANSI Warn", out.String())
		}
	})
	t.Run("Show_color無効の場合_ANSIエスケープを含まないこと", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		runID := writeWarnJournal(t, projDir)
		var out bytes.Buffer

		// Act
		err := Show(&out, projDir, runID.String(), false)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out.String(), "\x1b[") {
			t.Errorf("output = %q, want no ANSI escape", out.String())
		}
		if !strings.Contains(out.String(), "run [Warn]") {
			t.Errorf("output = %q, want plain 'run [Warn]'", out.String())
		}
	})
}
