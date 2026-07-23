package runscenario

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/repository"
)

// writeRunFixture はテスト用プロジェクトのファイルを作る。
func writeRunFixture(t *testing.T, projDir, rel, content string, mode os.FileMode) {
	t.Helper()
	path := filepath.Join(projDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

// TestResolveResumeFrom は resume の引き継ぎ元解決 (AS-BUILT §5.8) を固定する。
// 実行中 (未完了) の run はスキップ・拒否されること。
func TestResolveResumeFrom(t *testing.T) {
	// Arrange (共通): 完了済み run を 1 つ作り、その後に「実行中」を模した
	// 未完了ジャーナル (node_start のみ) の run ディレクトリを作る
	setup := func(t *testing.T) (projDir, finishedID, unfinishedID string) {
		t.Helper()
		projDir = t.TempDir()
		writeRunFixture(t, projDir, "stfw.yml", "stfw:\n  project_version: 0.1.0\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/metadata.yml", "description:\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/metadata.yml", "description:\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/metadata.yml", "description:\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/config/config.yml", "stfw:\n  process:\n    scripts: {}\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_ok", "#!/bin/bash\nexit 0\n", 0o755)
		cfg, _, err := repository.LoadConfig(projDir)
		if err != nil {
			t.Fatal(err)
		}
		log := slog.New(slog.NewTextHandler(io.Discard, nil))
		if err := Run(log, io.Discard, io.Discard, projDir, cfg, "test", []string{"demo"}, Options{}, time.Now); err != nil {
			t.Fatal(err)
		}
		ids, err := repository.ListRunIDs(projDir)
		if err != nil || len(ids) != 1 {
			t.Fatalf("完了済み run が 1 つあるべき: ids=%v err=%v", ids, err)
		}
		finishedID = ids[0]

		// 完了済みジャーナルの node_start 行を流用し、run_id を差し替えた
		// 未完了ジャーナル (より新しい run_id) を作る
		raw, err := os.ReadFile(repository.JournalPath(projDir, finishedID))
		if err != nil {
			t.Fatal(err)
		}
		firstLine := strings.SplitN(string(raw), "\n", 2)[0]
		unfinishedID = "_99990101000000_1"
		unfinishedDir := repository.RunDir(projDir, unfinishedID)
		if err := os.MkdirAll(filepath.Join(unfinishedDir, "workspace", "demo"), 0o755); err != nil {
			t.Fatal(err)
		}
		line := strings.ReplaceAll(firstLine, finishedID, unfinishedID)
		if err := os.WriteFile(filepath.Join(unfinishedDir, "journal.jsonl"), []byte(line+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return projDir, finishedID, unfinishedID
	}

	t.Run("resolveResumeFrom_latestで最新runが実行中の場合_完了済みrunへ遡ること", func(t *testing.T) {
		// Arrange
		projDir, finishedID, _ := setup(t)

		// Act
		got, err := resolveResumeFrom(projDir, "latest")

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if got != finishedID {
			t.Errorf("latest は実行中 run をスキップして完了済み run を返すべき: got %s, want %s", got, finishedID)
		}
	})

	t.Run("resolveResumeFrom_実行中runを明示指定した場合_エラーであること", func(t *testing.T) {
		// Arrange
		projDir, _, unfinishedID := setup(t)

		// Act
		_, err := resolveResumeFrom(projDir, unfinishedID)

		// Assert
		if err == nil {
			t.Fatal("未完了 run の明示指定はエラーになるべき (書き込み途中ワークスペースの取り込み防止)")
		}
	})
}

// TestRunConcurrent は同一シナリオを含む複数 run の並走 (AS-BUILT §5.7) を固定する。
// ステップは他方の run のランデブーマーカーを待つため、2 つの run が同時に
// 実行されていなければタイムアウトで失敗する (逐次化の退行を検出する)。
func TestRunConcurrent(t *testing.T) {
	t.Run("Run_同一シナリオを2つ同時に実行する場合_両方Successで生成物がrunごとに分離されること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		writeRunFixture(t, projDir, "stfw.yml", "stfw:\n  project_version: 0.1.0\n  loglevel: \"info\"\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/metadata.yml", "description:\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/metadata.yml", "description:\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/metadata.yml", "description:\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/config/config.yml", "stfw:\n  process:\n    scripts: {}\n", 0o644)
		writeRunFixture(t, projDir, "scenario/demo/_10_99990101/_10_pre_scripts/scripts/100_rendezvous",
			`#!/bin/bash
# 自分のマーカーを置き、もう一方の run のマーカーが現れるまで待つ (並走の証明)
touch "${STFW_PROJ_DIR}/rendezvous_${run_id}"
mkdir -p "${stfw_process_dir}/evidence"
echo "${run_id}" >"${stfw_process_dir}/evidence/marker.txt"
for _ in $(seq 1 100); do
  n=$(ls "${STFW_PROJ_DIR}" | grep -c '^rendezvous_')
  [ "${n}" -ge 2 ] && exit 0
  sleep 0.1
done
exit 6
`, 0o755)
		cfg, _, err := repository.LoadConfig(projDir)
		if err != nil {
			t.Fatal(err)
		}
		log := slog.New(slog.NewTextHandler(io.Discard, nil))

		// Act: 同一シナリオの run を 2 つ同時に開始する
		var wg sync.WaitGroup
		errs := make([]error, 2)
		outs := make([]bytes.Buffer, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				errs[i] = Run(log, &outs[i], io.Discard, projDir, cfg, "test", []string{"demo"}, Options{}, time.Now)
			}(i)
		}
		wg.Wait()

		// Assert
		for i, err := range errs {
			if err != nil {
				t.Fatalf("run %d は Success になるべき (並走できず相互待ちがタイムアウトした可能性): %v\n%s", i, err, outs[i].String())
			}
		}
		runIDs, err := repository.ListRunIDs(projDir)
		if err != nil {
			t.Fatal(err)
		}
		if len(runIDs) != 2 {
			t.Fatalf("run ディレクトリは 2 つ作られるべき: got %v", runIDs)
		}
		for _, id := range runIDs {
			marker := filepath.Join(repository.WorkspaceDir(projDir, id), "demo", "_10_99990101", "_10_pre_scripts", "evidence", "marker.txt")
			raw, err := os.ReadFile(marker)
			if err != nil {
				t.Fatalf("run %s のワークスペースに evidence が分離されるべき: %v", id, err)
			}
			if strings.TrimSpace(string(raw)) != id {
				t.Errorf("run %s の marker = %q (他 run の生成物と混ざっている)", id, raw)
			}
		}
	})
}
