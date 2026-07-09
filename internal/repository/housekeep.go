package repository

import (
	"errors"
	"os"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
)

// HousekeepRuns は採番時刻が cutoff より前の実行結果を物理削除する
// (実行ジャーナル .stfw/runs/{run_id} と HTML レポート {reportsOutDir}/runs/{run_id}.html)。
// 1 件でも削除した場合はレポート index を再生成する (削除済み run が index に残らない)。
// 個別の削除失敗は残りの処理を止めず、まとめてエラーとして返す (best-effort)。
// 削除した run_id のリストを返す。
func HousekeepRuns(projDir, reportsOutDir string, cutoff time.Time) ([]string, error) {
	ids, err := ListRunIDs(projDir)
	if err != nil {
		return nil, err
	}

	var deleted []string
	var errs []error
	for _, id := range ids {
		runID, err := run.ParseRunID(id)
		if err != nil {
			continue
		}
		ts, err := runID.Time()
		if err != nil {
			continue
		}
		if !ts.Before(cutoff) {
			continue
		}
		// レポート → 実行ジャーナルの順に消す。run ディレクトリを先に消すと、レポート削除の
		// 失敗時に run が ListRunIDs から見えなくなり、孤児レポートが再回収不能になるため。
		// レポート削除に失敗した run はスキップし、次回のハウスキープで再試行させる。
		// (レポートは生成に失敗した run では存在しないことがあるため不在は許容する)
		if err := os.Remove(RunReportPath(reportsOutDir, id)); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
			continue
		}
		if err := os.RemoveAll(runDir(projDir, id)); err != nil {
			errs = append(errs, err)
			continue
		}
		deleted = append(deleted, id)
	}

	if len(deleted) > 0 {
		if err := WriteReportIndex(projDir, reportsOutDir); err != nil {
			errs = append(errs, err)
		}
	}
	return deleted, errors.Join(errs...)
}
