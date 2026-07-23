//go:build !windows

package repository

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// TestCopyScenarioToWorkspaceSpecialFile は通常ファイル以外 (FIFO 等) が
// 複製時エラーになること (無期限ブロックしないこと) を固定する。
func TestCopyScenarioToWorkspaceSpecialFile(t *testing.T) {
	t.Run("CopyScenarioToWorkspace_FIFOがある場合_エラーであること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/input.txt", "input\n", 0o644)
		if err := syscall.Mkfifo(filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/pipe"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		_, err := CopyScenarioToWorkspace(projDir, "_20260723120000_100", "demo")

		// Assert
		if err == nil {
			t.Fatal("通常ファイル以外は複製時エラーになるべき (FIFO の Open でブロックしない)")
		}
		if _, statErr := os.Stat(filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/pipe")); statErr != nil {
			t.Fatal(statErr)
		}
	})
}
