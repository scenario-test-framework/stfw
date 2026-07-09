package runscenario

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/scenario-test-framework/stfw/internal/repository"
)

// retentionDaysKey は実行結果の保存日数設定 (stfw.housekeep.retention_days) のフラット化キー。
const retentionDaysKey = "stfw_housekeep_retention_days"

// housekeep は保存日数を過ぎた過去の実行結果 (実行ジャーナル・HTML レポート) を削除する。
// stfw run の開始時の振る舞いとして実行される (専用サブコマンド・常駐ジョブは設けない)。
// ハウスキープは補助処理のため、失敗しても本体の実行は止めず警告に留める。
// 保存日数が未設定・0 の場合は無効 (無期限保存)。
func housekeep(log *slog.Logger, projDir string, cfg *repository.Config, now func() time.Time) {
	raw := cfg.Get(retentionDaysKey)
	if raw == "" || raw == "0" {
		return
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days < 0 {
		log.Warn("housekeep: invalid retention_days (must be a non-negative integer)", "value", raw)
		return
	}
	if days == 0 {
		return
	}

	cutoff := now().AddDate(0, 0, -days)
	deleted, err := repository.HousekeepRuns(projDir, repository.ReportsDir(projDir), cutoff)
	for _, id := range deleted {
		log.Info("housekeep: removed expired run", "run_id", id, "retention_days", days)
	}
	if err != nil {
		log.Warn("housekeep: failed to remove some expired runs", "message", err.Error())
	}
}
