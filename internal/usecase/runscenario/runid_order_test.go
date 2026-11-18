package runscenario

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/scenario-test-framework/stfw/internal/repository"
)

// TestRunIssuesAfterLatest は、同一秒に続けて実行した場合も後の run が最新になることを固定する
// (AS-BUILT §5.2 採番時刻)。pid は時刻順にならないため、pid の比較だけでは新旧を決められない。
func TestRunIssuesAfterLatest(t *testing.T) {
	t.Run("Run_同一秒にpidの大きい既存runがある場合_新しいrunが最新runであること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
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
		fixed := time.Now().Truncate(time.Second)
		now := func() time.Time { return fixed }

		// 同じ秒に別プロセス (pid が大きく桁数も多い) が完了させた run を模す。
		// 実際に 1 回実行し、その run を pid 99999999 の run_id に付け替える
		if err := Run(log, io.Discard, io.Discard, projDir, cfg, "test", []string{"demo"}, Options{}, now); err != nil {
			t.Fatal(err)
		}
		ids, err := repository.ListRunIDs(projDir)
		if err != nil || len(ids) != 1 {
			t.Fatalf("run が 1 つあるべき: ids=%v err=%v", ids, err)
		}
		orig := ids[0]
		other := orig[:16] + "99999999"
		raw, err := os.ReadFile(repository.JournalPath(projDir, orig))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(repository.RunDir(projDir, orig), repository.RunDir(projDir, other)); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(repository.RunDir(projDir, other), "journal.jsonl"), []byte(strings.ReplaceAll(string(raw), orig, other)), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		var out bytes.Buffer
		err = Run(log, &out, io.Discard, projDir, cfg, "test", []string{"demo"}, Options{}, now)

		// Assert
		if err != nil {
			t.Fatal(err)
		}
		m := regexp.MustCompile(`run_id: (\S+)`).FindStringSubmatch(out.String())
		if m == nil {
			t.Fatalf("run_id が出力されていない: %q", out.String())
		}
		latest, err := repository.LatestRunID(projDir)
		if err != nil {
			t.Fatal(err)
		}
		if latest != m[1] {
			t.Errorf("LatestRunID() = %s, want 新しい run %s (既存 run %s)", latest, m[1], other)
		}
	})

	t.Run("Run_同一秒に同じpidで衝突の再採番回数を超えて続けて実行する場合_すべて開始でき最後のrunが最新runであること", func(t *testing.T) {
		// Arrange
		// コンテナ内では毎回 pid 1 になるため、同一秒の連続実行は同じ pid で起きる。
		// 衝突時の再採番 (最大 10 回) だけに頼ると 12 回目以降が開始できない
		projDir := t.TempDir()
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
		fixed := time.Now().Truncate(time.Second)
		now := func() time.Time { return fixed }
		const runs = 15
		re := regexp.MustCompile(`run_id: (\S+)`)

		// Act
		var issued []string
		for i := 0; i < runs; i++ {
			var out bytes.Buffer
			if err := Run(log, &out, io.Discard, projDir, cfg, "test", []string{"demo"}, Options{}, now); err != nil {
				t.Fatalf("%d 回目の run が開始できない: %v", i+1, err)
			}
			m := re.FindStringSubmatch(out.String())
			if m == nil {
				t.Fatalf("%d 回目の run_id が出力されていない: %q", i+1, out.String())
			}
			issued = append(issued, m[1])
		}

		// Assert
		ids, err := repository.ListRunIDs(projDir)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Join(ids, ",") != strings.Join(issued, ",") {
			t.Errorf("ListRunIDs() = %v, want 実行順 %v", ids, issued)
		}
	})
}
