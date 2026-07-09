package repository

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/gateway"
)

// reportsDirName は .stfw 配下の HTML レポート配置ディレクトリ名。
// nginx 等で共有 volume を配信する前提のため、レポートはこの配下で自己完結する。
const reportsDirName = "reports"

// ReportsDir はデフォルトのレポート出力ディレクトリ (.stfw/reports) を返す。
func ReportsDir(projDir string) string {
	return filepath.Join(projDir, project.DataDirName, reportsDirName)
}

// RunReportPath は run 詳細ページ (runs/{run_id}.html) のパスを返す。
func RunReportPath(outDir, runID string) string {
	return filepath.Join(outDir, "runs", runID+".html")
}

// IndexReportPath は run 一覧ページ (index.html) のパスを返す。
func IndexReportPath(outDir string) string {
	return filepath.Join(outDir, "index.html")
}

// runReportView は run 詳細ページのテンプレートデータ。
type runReportView struct {
	RunID          string
	Status         string
	Mode           string
	Params         string
	StartTime      string
	EndTime        string
	ProcessingTime string
	InProgress     bool
	GeneratedAt    string
	Nodes          []reportNodeView
}

// reportNodeView は階層ツリーの 1 ノード行。
type reportNodeView struct {
	Name           string
	Type           string
	Status         string
	IndentPx       int
	StepIndentPx   int
	StartTime      string
	EndTime        string
	ProcessingTime string
	Steps          []reportStepView
}

// reportStepView はステップの 1 行。未実行 (Pending / Blocked) は時刻を空で持つ。
type reportStepView struct {
	Name           string
	Status         string
	StartTime      string
	EndTime        string
	ProcessingTime string
}

// indexReportView は run 一覧ページのテンプレートデータ。
type indexReportView struct {
	GeneratedAt string
	InProgress  bool
	Runs        []runSummaryView
}

// runSummaryView は run 一覧の 1 行。ジャーナルを読めない run は Unknown として表示する。
type runSummaryView struct {
	RunID          string
	Status         string
	Mode           string
	Params         string
	StartTime      string
	EndTime        string
	ProcessingTime string
	InProgress     bool
}

// WriteRunReport はジャーナルから run 詳細ページを生成する。
// 実行中 (run 階層の node_end 未記録) のページには自動リロードの
// meta refresh を含め、終了後の再生成で外す。
func WriteRunReport(projDir, outDir, runID string) error {
	events, err := ReadJournal(projDir, runID)
	if err != nil {
		return err
	}
	id, err := run.ParseRunID(runID)
	if err != nil {
		return err
	}
	r, err := run.Replay(id, events)
	if err != nil {
		return fmt.Errorf("journal replay: %w", err)
	}
	view := buildRunReportView(runID, events, r)
	return gateway.WriteHTML(RunReportPath(outDir, runID), "run.html.tmpl", view)
}

// WriteReportIndex は全 run のジャーナルから run 一覧ページを生成する。
// 実行中の run が含まれる場合は一覧にも meta refresh を含める。
func WriteReportIndex(projDir, outDir string) error {
	ids, err := ListRunIDs(projDir)
	if err != nil {
		return err
	}
	view := indexReportView{GeneratedAt: time.Now().Format(run.TSFormat)}
	// 新しい run が先頭に来るよう降順で並べる
	for i := len(ids) - 1; i >= 0; i-- {
		summary := buildRunSummaryView(projDir, ids[i])
		if summary.InProgress {
			view.InProgress = true
		}
		view.Runs = append(view.Runs, summary)
	}
	return gateway.WriteHTML(IndexReportPath(outDir), "index.html.tmpl", view)
}

// buildRunReportView はイベント列と復元済み集約から表示データを組み立てる。
func buildRunReportView(runID string, events []run.Event, r *run.Run) runReportView {
	nodeStart, nodeEnd, stepTimes := collectTimes(events)

	view := runReportView{
		RunID:       runID,
		Status:      string(run.NodeStarted),
		GeneratedAt: time.Now().Format(run.TSFormat),
	}
	rootID := run.NewRunNodeID(r.RunID()).String()
	for _, node := range r.NodeViews() {
		row := reportNodeView{
			Name:         node.Name,
			Type:         string(node.Type),
			Status:       string(node.Status),
			IndentPx:     12 + node.Depth*24,
			StepIndentPx: 12 + (node.Depth+1)*24,
			StartTime:    nodeStart[node.ID],
			EndTime:      nodeEnd[node.ID],
		}
		row.ProcessingTime = elapsedOrEmpty(row.StartTime, row.EndTime)
		for _, step := range node.Steps {
			times := stepTimes[stepKey(node.ID, step.Name)]
			row.Steps = append(row.Steps, reportStepView{
				Name:           step.Name,
				Status:         string(step.Status),
				StartTime:      times.start,
				EndTime:        times.end,
				ProcessingTime: elapsedOrEmpty(times.start, times.end),
			})
		}
		view.Nodes = append(view.Nodes, row)

		if node.ID == rootID {
			view.Status = string(node.Status)
			view.Mode = node.Attrs["run_mode"]
			view.Params = node.Attrs["params"]
			view.StartTime = row.StartTime
			view.EndTime = row.EndTime
			view.ProcessingTime = row.ProcessingTime
		}
	}
	view.InProgress = view.Status == string(run.NodeStarted)
	return view
}

// buildRunSummaryView は run 一覧の 1 行を組み立てる。
// ジャーナルを読めない・run 階層イベントが見つからない場合は Unknown を返す。
func buildRunSummaryView(projDir, runID string) runSummaryView {
	summary := runSummaryView{RunID: runID, Status: "Unknown"}
	events, err := ReadJournal(projDir, runID)
	if err != nil {
		return summary
	}
	rootID := ""
	if id, err := run.ParseRunID(runID); err == nil {
		rootID = run.NewRunNodeID(id).String()
	}
	for _, ev := range events {
		if ev.NodeID != rootID {
			continue
		}
		switch ev.Type {
		case run.EventNodeStart:
			summary.Status = string(run.NodeStarted)
			summary.InProgress = true
			summary.Mode = ev.Attrs["run_mode"]
			summary.Params = ev.Attrs["params"]
			summary.StartTime = ev.TS
		case run.EventNodeEnd:
			summary.Status = ev.Status
			summary.InProgress = false
			summary.EndTime = ev.TS
			summary.ProcessingTime = elapsedOrEmpty(summary.StartTime, summary.EndTime)
		}
	}
	return summary
}

// stepTime はステップの実行時刻 (未実行は空)。
type stepTime struct {
	start string
	end   string
}

// collectTimes はイベント列からノード・ステップの実行時刻を集める。
func collectTimes(events []run.Event) (nodeStart, nodeEnd map[string]string, stepTimes map[string]stepTime) {
	nodeStart = map[string]string{}
	nodeEnd = map[string]string{}
	stepTimes = map[string]stepTime{}
	for _, ev := range events {
		switch ev.Type {
		case run.EventNodeStart:
			nodeStart[ev.NodeID] = ev.TS
		case run.EventNodeEnd:
			nodeEnd[ev.NodeID] = ev.TS
		case run.EventStepEnd:
			stepTimes[stepKey(ev.NodeID, ev.Step)] = stepTime{start: ev.StartTS, end: ev.EndTS}
		}
	}
	return nodeStart, nodeEnd, stepTimes
}

// stepKey はノード ID とステップ名の複合キーを返す。
// ステップ名にはパス区切りを含められないため "/" で連結する。
func stepKey(nodeID, step string) string {
	return nodeID + "/" + step
}

// elapsedOrEmpty は開始・終了が揃っている場合のみ処理時間を返す。
func elapsedOrEmpty(startTS, endTS string) string {
	if startTS == "" || endTS == "" {
		return ""
	}
	elapsed, err := run.ElapsedString(startTS, endTS)
	if err != nil {
		return ""
	}
	return elapsed
}
