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
// color が true のときはステータスを ANSI カラーで色付けする
// (Success=緑 / Warn=黄 / Error=赤 / Blocked=灰。SPEC-024-01)。
func Show(out io.Writer, projDir, runID string, color bool) error {
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
		fmt.Fprintf(out, "%s%s [%s]\n", indent, node.Name, colorize(string(node.Status), color))
		for _, step := range node.Steps {
			fmt.Fprintf(out, "%s  - %s [%s]\n", indent, step.Name, colorize(string(step.Status), color))
		}
	}
	return nil
}

// colorize はステータス文字列を ANSI カラーで装飾する (color=false は素通し)。
func colorize(status string, color bool) string {
	if !color {
		return status
	}
	var code string
	switch status {
	case string(run.NodeSuccess):
		code = "32" // green
	case string(run.NodeWarn):
		code = "33" // yellow
	case string(run.NodeError):
		code = "31" // red
	case string(run.StepBlocked):
		code = "90" // gray
	default:
		return status
	}
	return "\x1b[" + code + "m" + status + "\x1b[0m"
}
