package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeWorkspaceFixture はテスト用のシナリオツリーをプロジェクト配下に作る。
func writeWorkspaceFixture(t *testing.T, projDir string, rel string, content string, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(projDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

// TestCopyScenarioToWorkspace は scenario/ 正本の実行ワークスペースへの複製
// (AS-BUILT §5.7) を固定する。
func TestCopyScenarioToWorkspace(t *testing.T) {
	const runID = "_20260723120000_100"

	t.Run("CopyScenarioToWorkspace_通常ツリーの場合_パーミッションを保持して複製されること", func(t *testing.T) {
		// Arrange: 実行ファイル (0755) と秘匿寄りの入力 (0600)
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/metadata.yml", "description:\n", 0o644)
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_step", "#!/bin/bash\n", 0o755)
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/input.csv", "a,b\n", 0o600)

		// Act
		dest, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(WorkspaceDir(projDir, runID), "demo")
		if dest != want {
			t.Errorf("複製先 = %s, want %s", dest, want)
		}
		raw, err := os.ReadFile(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "data", "input.csv"))
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != "a,b\n" {
			t.Errorf("data/input.csv の内容 = %q, want %q", raw, "a,b\n")
		}
		info, err := os.Stat(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "data", "input.csv"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("0600 の入力は 0600 のまま複製されるべき (group/other へ権限を広げない): got %o", info.Mode().Perm())
		}
		info, err = os.Stat(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "scripts", "100_step"))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("スクリプトの実行権限が複製で保持されるべき: got %o", info.Mode().Perm())
		}
	})

	t.Run("CopyScenarioToWorkspace_予約出力ディレクトリがある場合_複製されないこと", func(t *testing.T) {
		// Arrange: 旧バージョンの実行残骸 (evidence/actual/result) を正本側に置く
		projDir := t.TempDir()
		proc := "scenario/demo/_10_99990101/_20_asrt_compare"
		writeWorkspaceFixture(t, projDir, proc+"/expect/_10_col_scripts/host1/data.txt", "expected\n", 0o644)
		writeWorkspaceFixture(t, projDir, proc+"/evidence/host1/data.txt", "stale\n", 0o644)
		writeWorkspaceFixture(t, projDir, proc+"/actual/_10_col_scripts/host1/data.txt", "stale\n", 0o644)
		writeWorkspaceFixture(t, projDir, proc+"/result/CompareSummary.csv", "stale\n", 0o644)

		// Act
		dest, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		destProc := filepath.Join(dest, "_10_99990101", "_20_asrt_compare")
		if _, err := os.Stat(filepath.Join(destProc, "expect", "_10_col_scripts", "host1", "data.txt")); err != nil {
			t.Errorf("expect/ (git 管理の入力) は複製されるべき: %v", err)
		}
		for _, reserved := range []string{"evidence", "actual", "result"} {
			if _, err := os.Stat(filepath.Join(destProc, reserved)); !os.IsNotExist(err) {
				t.Errorf("予約出力ディレクトリ %s/ は複製されないべき", reserved)
			}
		}
	})

	t.Run("CopyScenarioToWorkspace_シンボリックリンクがある場合_リンク先の実体が複製されること", func(t *testing.T) {
		// Arrange: ファイル symlink と、シナリオ外の共有ディレクトリへの symlink
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/real.txt", "real\n", 0o644)
		if err := os.Symlink("real.txt", filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/link.txt")); err != nil {
			t.Fatal(err)
		}
		writeWorkspaceFixture(t, projDir, "shared/fixtures/common.txt", "shared\n", 0o644)
		if err := os.Symlink(filepath.Join(projDir, "shared", "fixtures"), filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/fixtures")); err != nil {
			t.Fatal(err)
		}

		// Act
		dest, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		destLink := filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "data", "link.txt")
		if info, err := os.Lstat(destLink); err != nil || info.Mode()&os.ModeSymlink != 0 {
			t.Errorf("ファイル symlink はリンク先の実体として複製されるべき: info=%v err=%v", info, err)
		}
		raw, err := os.ReadFile(destLink)
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != "real\n" {
			t.Errorf("複製内容 = %q, want %q", raw, "real\n")
		}
		raw, err = os.ReadFile(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "fixtures", "common.txt"))
		if err != nil {
			t.Fatalf("ディレクトリ symlink はリンク先の実体として複製されるべき: %v", err)
		}
		if string(raw) != "shared\n" {
			t.Errorf("複製内容 = %q, want %q", raw, "shared\n")
		}
	})

	t.Run("CopyScenarioToWorkspace_シナリオ自体がシンボリックリンクの場合_実体が複製されること", func(t *testing.T) {
		// Arrange: scenario/demo -> 実体ディレクトリへの symlink
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "real-demo/_10_99990101/_10_pre_scripts/data/input.txt", "input\n", 0o644)
		if err := os.MkdirAll(filepath.Join(projDir, "scenario"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(projDir, "real-demo"), filepath.Join(projDir, "scenario", "demo")); err != nil {
			t.Fatal(err)
		}

		// Act
		dest, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "data", "input.txt")); err != nil {
			t.Errorf("symlink シナリオも実体として複製されるべき: %v", err)
		}
	})

	t.Run("CopyScenarioToWorkspace_リンク先が複製先の祖先の場合_エラーであること", func(t *testing.T) {
		// Arrange: プロジェクトルートへの symlink (走査が生成中のワークスペース
		// 自身へ到達して自己複製が無限に続く経路)
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/input.txt", "input\n", 0o644)
		if err := os.Symlink(projDir, filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/proj")); err != nil {
			t.Fatal(err)
		}

		// Act
		_, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err == nil {
			t.Fatal("複製先の祖先を指す symlink は自己複製になるため複製時エラーになるべき")
		}
	})

	t.Run("CopyScenarioToWorkspace_リンク先が大文字小文字違いで複製先の祖先の場合_エラーであること", func(t *testing.T) {
		// Arrange: case-insensitive ファイルシステムではパス文字列比較の
		// ガードをすり抜ける表記 (全大文字) でプロジェクトルートを指す symlink。
		// case-sensitive 環境ではリンク先不在のためどちらでもエラーになる。
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/input.txt", "input\n", 0o644)
		if err := os.Symlink(strings.ToUpper(projDir), filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/proj")); err != nil {
			t.Fatal(err)
		}

		// Act
		_, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err == nil {
			t.Fatal("大文字小文字違いで複製先の祖先を指す symlink もエラーになるべき (os.SameFile による物理同一性判定)")
		}
	})

	t.Run("CopyScenarioToWorkspace_リンク循環がある場合_エラーであること", func(t *testing.T) {
		// Arrange: プロセスディレクトリ配下に祖先への symlink (循環)
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/input.txt", "input\n", 0o644)
		if err := os.Symlink(filepath.Join(projDir, "scenario", "demo"), filepath.Join(projDir, "scenario/demo/_10_99990101/_10_pre_scripts/loop")); err != nil {
			t.Fatal(err)
		}

		// Act
		_, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err == nil {
			t.Fatal("リンク循環は複製時エラーになるべき")
		}
	})

	t.Run("CopyScenarioToWorkspace_プロセスディレクトリ直下以外の予約名の場合_複製されること", func(t *testing.T) {
		// Arrange: 入力ディレクトリ (data/) 配下の evidence という名前は予約対象外。
		// data/ 配下にプロセス形式の名前を持つ入力ディレクトリがあっても同様
		// (プロセス位置は bizdate 直下と parallel 子のみ)。
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/evidence/input.txt", "input\n", 0o644)
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/_20_fixture_scripts/evidence/input.txt", "input\n", 0o644)

		// Act
		dest, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "data", "evidence", "input.txt")); err != nil {
			t.Errorf("プロセスディレクトリ直下以外の予約名は複製されるべき: %v", err)
		}
		if _, err := os.Stat(filepath.Join(dest, "_10_99990101", "_10_pre_scripts", "data", "_20_fixture_scripts", "evidence", "input.txt")); err != nil {
			t.Errorf("プロセス形式名の入力ディレクトリ配下の evidence も複製されるべき: %v", err)
		}
	})

	t.Run("CopyScenarioToWorkspace_parallel子プロセス直下の予約名の場合_複製されないこと", func(t *testing.T) {
		// Arrange: parallel の子プロセスディレクトリ直下の evidence は予約出力
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_par_parallel/_10_a_scripts/scripts/100_step", "#!/bin/bash\n", 0o755)
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_par_parallel/_10_a_scripts/evidence/stale.txt", "stale\n", 0o644)

		// Act
		dest, err := CopyScenarioToWorkspace(projDir, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(dest, "_10_99990101", "_10_par_parallel", "_10_a_scripts", "evidence")); !os.IsNotExist(err) {
			t.Error("parallel 子プロセス直下の予約出力ディレクトリは複製されないべき")
		}
	})

	t.Run("CopyScenarioToWorkspace_シナリオが存在しない場合_エラーであること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()

		// Act
		_, err := CopyScenarioToWorkspace(projDir, runID, "missing")

		// Assert
		if err == nil {
			t.Fatal("存在しないシナリオの複製はエラーになるべき")
		}
	})
}

// TestMergeRunWorkspace は resume の前 run ワークスペース引き継ぎ (AS-BUILT §5.8) を固定する。
func TestMergeRunWorkspace(t *testing.T) {
	const fromRunID = "_20260722120000_100"
	const runID = "_20260723120000_200"

	t.Run("MergeRunWorkspace_前runにのみある生成物の場合_取り込まれること", func(t *testing.T) {
		// Arrange: 正本を複製した新ワークスペース + 前 run のワークスペース
		// (evidence と実行時生成の data が残っている)
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_step", "#!/bin/bash\n", 0o755)
		if _, err := CopyScenarioToWorkspace(projDir, runID, "demo"); err != nil {
			t.Fatal(err)
		}
		fromProc := filepath.Join(WorkspaceDir(projDir, fromRunID), "demo", "_10_99990101", "_10_pre_scripts")
		for rel, content := range map[string]string{
			"evidence/out.txt": "collected\n",
			"data/gen.csv":     "generated\n",
		} {
			path := filepath.Join(fromProc, rel)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		// Act
		err := MergeRunWorkspace(projDir, fromRunID, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		destProc := filepath.Join(WorkspaceDir(projDir, runID), "demo", "_10_99990101", "_10_pre_scripts")
		for _, rel := range []string{"evidence/out.txt", "data/gen.csv"} {
			if _, err := os.Stat(filepath.Join(destProc, rel)); err != nil {
				t.Errorf("前 run にのみある %s は取り込まれるべき: %v", rel, err)
			}
		}
	})

	t.Run("MergeRunWorkspace_同名ファイルがある場合_正本が優先されること", func(t *testing.T) {
		// Arrange: 正本 (新しい内容) と前 run (古い内容) に同名ファイル
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_step", "new version\n", 0o755)
		if _, err := CopyScenarioToWorkspace(projDir, runID, "demo"); err != nil {
			t.Fatal(err)
		}
		oldStep := filepath.Join(WorkspaceDir(projDir, fromRunID), "demo", "_10_99990101", "_10_pre_scripts", "scripts", "100_step")
		if err := os.MkdirAll(filepath.Dir(oldStep), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(oldStep, []byte("old version\n"), 0o755); err != nil {
			t.Fatal(err)
		}

		// Act
		err := MergeRunWorkspace(projDir, fromRunID, runID, "demo")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(WorkspaceDir(projDir, runID), "demo", "_10_99990101", "_10_pre_scripts", "scripts", "100_step"))
		if err != nil {
			t.Fatal(err)
		}
		if string(raw) != "new version\n" {
			t.Errorf("マージは正本 (今回の複製) を優先すべき: got %q", raw)
		}
	})

	t.Run("MergeRunWorkspace_同一パスの種別が異なる場合_正本が優先されること", func(t *testing.T) {
		// Arrange: 正本ではファイルの config、前 run では同名がディレクトリ (型競合)。
		// 逆方向 (正本がディレクトリ・前 run がファイル) も同時に検証する。
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/notes", "file in source\n", 0o644)
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/data/input.txt", "dir in source\n", 0o644)
		if _, err := CopyScenarioToWorkspace(projDir, runID, "demo"); err != nil {
			t.Fatal(err)
		}
		fromProc := filepath.Join(WorkspaceDir(projDir, fromRunID), "demo", "_10_99990101", "_10_pre_scripts")
		if err := os.MkdirAll(filepath.Join(fromProc, "notes"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fromProc, "notes", "old.txt"), []byte("dir in prior\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fromProc, "data"), []byte("file in prior\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		err := MergeRunWorkspace(projDir, fromRunID, runID, "demo")

		// Assert
		if err != nil {
			t.Fatalf("型競合があってもマージは正本優先で成功すべき: %v", err)
		}
		destProc := filepath.Join(WorkspaceDir(projDir, runID), "demo", "_10_99990101", "_10_pre_scripts")
		info, err := os.Stat(filepath.Join(destProc, "notes"))
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			t.Error("正本のファイル notes が前 run のディレクトリで置き換わるべきでない")
		}
		info, err = os.Stat(filepath.Join(destProc, "data"))
		if err != nil {
			t.Fatal(err)
		}
		if !info.IsDir() {
			t.Error("正本のディレクトリ data が前 run のファイルで置き換わるべきでない")
		}
	})

	t.Run("MergeRunWorkspace_引き継ぎ元にシナリオが無い場合_エラーであること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		writeWorkspaceFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_step", "#!/bin/bash\n", 0o755)
		if _, err := CopyScenarioToWorkspace(projDir, runID, "demo"); err != nil {
			t.Fatal(err)
		}

		// Act
		err := MergeRunWorkspace(projDir, fromRunID, runID, "demo")

		// Assert
		if err == nil {
			t.Fatal("引き継ぎ元ワークスペースに対象シナリオが無い場合はエラーになるべき")
		}
	})
}
