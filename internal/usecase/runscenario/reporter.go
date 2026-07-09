package runscenario

import (
	"log/slog"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// reporter は実行中の HTML レポートをインクリメンタル生成する。
// 生成タイミングは run 開始時 / process の node_end ごと / run 終了時。
// 実行中のページには meta refresh が入り、run 終了時の再生成で外れる。
// 生成失敗はログのみで実行結果へは影響しない。
type reporter struct {
	log     *slog.Logger
	projDir string
	outDir  string
	runID   string
}

// newReporter はデフォルト出力先 (.stfw/reports) のレポーターを生成する。
func newReporter(log *slog.Logger, projDir string, runID run.RunID) *reporter {
	return &reporter{
		log:     log,
		projDir: projDir,
		outDir:  repository.ReportsDir(projDir),
		runID:   runID.String(),
	}
}

// onEvent は生成タイミングに該当するイベントでレポートを再生成する。
func (r *reporter) onEvent(ev run.Event) {
	if !r.shouldRefresh(ev) {
		return
	}
	if err := repository.WriteRunReport(r.projDir, r.outDir, r.runID); err != nil {
		r.log.Warn("[report] run page generation failed", "run_id", r.runID, "message", err.Error())
	}
	if err := repository.WriteReportIndex(r.projDir, r.outDir); err != nil {
		r.log.Warn("[report] index generation failed", "message", err.Error())
	}
}

// shouldRefresh は再生成タイミング (run の開始・終了 / process の終了) を判定する。
func (r *reporter) shouldRefresh(ev run.Event) bool {
	switch ev.Type {
	case run.EventNodeStart:
		return ev.NodeType == run.NodeTypeRun
	case run.EventNodeEnd:
		depth := r.nodeDepth(ev.NodeID)
		return depth == run.NodeTypeRun || depth == run.NodeTypeProcess
	}
	return false
}

// nodeDepth は NodeID のセグメント数から階層種別を判定する。
func (r *reporter) nodeDepth(nodeID string) run.NodeType {
	segments := strings.Split(strings.TrimPrefix(nodeID, r.runID+"+"), "+")
	switch len(segments) {
	case 1:
		return run.NodeTypeRun
	case 2:
		return run.NodeTypeScenario
	case 3:
		return run.NodeTypeBizdate
	case 4:
		return run.NodeTypeProcess
	}
	return ""
}
