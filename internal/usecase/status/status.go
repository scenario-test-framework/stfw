// Package status は stfw status のビジネスフローを制御する。
// ジャーナルをリプレイして実行ツリーとステータスを表示する
// (v0.2 の digdag Web UI の代替)。
package status

import (
	"fmt"
	"io"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Show は run_id のジャーナルをリプレイして階層ツリーとステータスを表示する。
// runID が空の場合は最新の run を対象とする。
// リプレイは生成時と同じ状態遷移検証を通す (不正なジャーナルはエラー)。
func Show(out io.Writer, projDir, runID string) error {
	if runID == "" {
		latest, err := repository.LatestRunID(projDir)
		if err != nil {
			return err
		}
		runID = latest
	}
	id, err := run.ParseRunID(runID)
	if err != nil {
		return err
	}

	events, err := repository.ReadJournal(projDir, runID)
	if err != nil {
		return err
	}
	r, err := run.Replay(id, events)
	if err != nil {
		return fmt.Errorf("journal replay: %w", err)
	}

	fmt.Fprintf(out, "run_id: %s\n", runID)
	for _, node := range r.NodeViews() {
		indent := strings.Repeat("  ", node.Depth)
		fmt.Fprintf(out, "%s%s [%s]\n", indent, node.Name, node.Status)
		for _, step := range node.Steps {
			fmt.Fprintf(out, "%s  - %s [%s]\n", indent, step.Name, step.Status)
		}
	}
	return nil
}
