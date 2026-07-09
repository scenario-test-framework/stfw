// Package report は stfw report のビジネスフローを制御する。
// ジャーナルから HTML レポート (run 一覧 + run 詳細) をオンデマンド再生成する。
package report

import (
	"fmt"
	"io"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Generate は run_id のレポートと run 一覧を outDir へ生成する。
// runID が空の場合は最新の run を対象とする。
// outDir が空の場合はデフォルト (.stfw/reports) へ出力する。
func Generate(out io.Writer, projDir, runID, outDir string) error {
	if outDir == "" {
		outDir = repository.ReportsDir(projDir)
	}
	if runID == "" {
		latest, err := repository.LatestRunID(projDir)
		if err != nil {
			return err
		}
		runID = latest
	}
	if _, err := run.ParseRunID(runID); err != nil {
		return err
	}

	if err := repository.WriteRunReport(projDir, outDir, runID); err != nil {
		return err
	}
	if err := repository.WriteReportIndex(projDir, outDir); err != nil {
		return err
	}
	fmt.Fprintf(out, "report: %s\n", repository.RunReportPath(outDir, runID))
	return nil
}
