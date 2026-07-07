// Package runscenario は stfw run (内蔵ランナー) のビジネスフローを制御する。
// ScenarioTree を深さ優先で走査し、各階層で setup フック → 子の逐次実行 →
// teardown フックを Go プロセス内で実行する。エラー時は後続の兄弟ノードを
// 実行せず停止する (v0.2 の digdag ワークフロー実行の置き換え)。
// 実行イベントは JSONL ジャーナル (.stfw/runs/{run_id}/journal.jsonl) へ記録する。
package runscenario

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Run はシナリオを内蔵ランナーで実行する。
// 実行前にディレクトリ規約を静的検証する (v0.2 では dig 生成時に行われていた検証に相当)。
// now は採番・ジャーナル記録に使う時刻源 (テスト容易性のため引数で受け取る)。
func Run(log *slog.Logger, out, errOut io.Writer, projDir string, cfg *repository.Config, version string, names []string, dryRun bool, now func() time.Time) error {
	projDir, err := filepath.Abs(projDir)
	if err != nil {
		return err
	}

	// 構造検証
	tree, err := repository.LoadScenarioTree(projDir, names)
	if err != nil {
		return err
	}
	installed, err := repository.ListProcessPlugins(projDir)
	if err != nil {
		return err
	}
	if violations := tree.Validate(installed); violations.HasError() {
		for _, v := range violations {
			fmt.Fprintln(out, v.String())
		}
		errs, warns := violations.Count()
		return fmt.Errorf("validation failed: %d error(s), %d warning(s)", errs, warns)
	}

	// run_id 採番 + ジャーナル作成。同一秒・同一プロセスの再実行 (テスト等) で
	// run_id が衝突した場合は採番時刻をずらして再採番する。
	runID := run.NewRunID(now(), os.Getpid())
	var journal *repository.Journal
	for i := 0; ; i++ {
		journal, err = repository.CreateJournal(projDir, runID)
		if err == nil {
			break
		}
		if !errors.Is(err, fs.ErrExist) || i >= 10 {
			return err
		}
		runID = run.NewRunID(now().Add(time.Duration(i+1)*time.Second), os.Getpid())
	}
	defer journal.Close()

	// webhook 通知は run 終了時に全送信の完了を待つ (エラー時も待つ)
	notifier := newWebhookNotifier(log, cfg, projDir, version, runID, now)
	defer notifier.wait()

	r := &runner{
		log:      log,
		out:      out,
		errOut:   errOut,
		projDir:  projDir,
		dryRun:   dryRun,
		now:      now,
		journal:  journal,
		agg:      run.NewRun(runID),
		baseEnv:  baseEnv(cfg, projDir, version, runID, dryRun),
		notifier: notifier,
		reporter: newReporter(log, projDir, runID),
	}

	fmt.Fprintf(out, "run_id: %s\n", runID)
	status, err := r.runRun(runID, tree, names)
	if err != nil {
		return err
	}
	if status != run.NodeSuccess {
		return fmt.Errorf("run %s finished with status %s", runID, status)
	}
	log.Info("run finished", "run_id", runID.String(), "status", string(status))
	return nil
}

// runner は 1 回の実行のオーケストレーション状態を保持する。
type runner struct {
	log      *slog.Logger
	out      io.Writer
	errOut   io.Writer
	projDir  string
	dryRun   bool
	now      func() time.Time
	journal  *repository.Journal
	agg      *run.Run
	baseEnv  map[string]string
	notifier *webhookNotifier
	reporter *reporter
}

// emit は生成時検証 (リプレイと同一の状態遷移検証) を通してジャーナルへ追記する。
// webhook 通知と HTML レポートはジャーナルイベントの投影のため、追記成功後に
// 連動させる (投影の失敗はログのみで実行結果へは影響しない)。
func (r *runner) emit(ev run.Event) error {
	if err := r.agg.Apply(ev); err != nil {
		return err
	}
	if err := r.journal.Append(ev); err != nil {
		return err
	}
	r.notifier.onEvent(ev)
	r.reporter.onEvent(ev)
	return nil
}

// runRun は run 階層を実行する: setup フック → シナリオの逐次実行 → teardown フック。
func (r *runner) runRun(runID run.RunID, tree *scenario.ScenarioTree, names []string) (run.NodeStatus, error) {
	nodeID := run.NewRunNodeID(runID)
	attrs := map[string]string{
		"run_id":   runID.String(),
		"run_mode": r.baseEnv["run_mode"],
		"params":   strings.Join(names, " "),
	}
	if err := r.emit(run.NewNodeStartEvent(r.now(), nodeID, run.NodeTypeRun, attrs)); err != nil {
		return "", err
	}

	env := cloneEnv(r.baseEnv)
	status := run.NodeSuccess
	if !r.runHooks(run.NodeTypeRun, "setup", env) {
		status = run.NodeError
	} else {
		for _, name := range targetOrder(tree, names) {
			view, ok := tree.ScenarioView(name)
			if !ok {
				return "", fmt.Errorf("scenario: %s is not exist", name)
			}
			st, err := r.runScenario(nodeID, view, env)
			if err != nil {
				return "", err
			}
			// エラー時は後続の兄弟ノードを実行せず停止する
			if st != run.NodeSuccess {
				status = run.NodeError
				break
			}
		}
	}

	// teardown フックはエラー時も実行する (v0.2 の _error ハンドラ相当)
	env["stfw_run_status"] = string(status)
	if !r.runHooks(run.NodeTypeRun, "teardown", env) {
		status = run.NodeError
	}
	if err := r.emit(run.NewNodeEndEvent(r.now(), nodeID, status)); err != nil {
		return "", err
	}
	return status, nil
}

// runScenario はシナリオ 1 件を実行する
// (逐次のみ。将来の --parallel 対応のためシナリオ単位で関数分離している)。
func (r *runner) runScenario(parent run.NodeID, view scenario.ScenarioView, parentEnv map[string]string) (run.NodeStatus, error) {
	nodeID, err := parent.Child(view.Name)
	if err != nil {
		return "", err
	}
	attrs := map[string]string{"name": view.Name}
	if err := r.emit(run.NewNodeStartEvent(r.now(), nodeID, run.NodeTypeScenario, attrs)); err != nil {
		return "", err
	}

	scenarioDir := filepath.Join(r.projDir, scenario.RootDirName, view.Name)
	env := cloneEnv(parentEnv)
	env["stfw_scenario_dir"] = scenarioDir
	env["stfw_scenario_name"] = view.Name

	status := run.NodeSuccess
	if !r.runHooks(run.NodeTypeScenario, "setup", env) {
		status = run.NodeError
	} else {
		for _, bizdate := range view.Bizdates {
			st, err := r.runBizdate(nodeID, scenarioDir, bizdate, env)
			if err != nil {
				return "", err
			}
			if st != run.NodeSuccess {
				status = run.NodeError
				break
			}
		}
	}

	env["stfw_run_status"] = string(status)
	if !r.runHooks(run.NodeTypeScenario, "teardown", env) {
		status = run.NodeError
	}
	if err := r.emit(run.NewNodeEndEvent(r.now(), nodeID, status)); err != nil {
		return "", err
	}
	return status, nil
}

// runBizdate は業務日付 1 件を実行する。
func (r *runner) runBizdate(parent run.NodeID, scenarioDir string, view scenario.BizdateView, parentEnv map[string]string) (run.NodeStatus, error) {
	nodeID, err := parent.Child(view.DirName)
	if err != nil {
		return "", err
	}
	attrs := map[string]string{
		"dirname": view.DirName,
		"seq":     view.Seq,
		"bizdate": view.Bizdate,
	}
	if err := r.emit(run.NewNodeStartEvent(r.now(), nodeID, run.NodeTypeBizdate, attrs)); err != nil {
		return "", err
	}

	bizdateDir := filepath.Join(scenarioDir, view.DirName)
	env := cloneEnv(parentEnv)
	env["stfw_bizdate_dir"] = bizdateDir
	env["stfw_bizdate_dirname"] = view.DirName
	env["stfw_bizdate_seq"] = view.Seq
	env["stfw_bizdate"] = view.Bizdate

	status := run.NodeSuccess
	if !r.runHooks(run.NodeTypeBizdate, "setup", env) {
		status = run.NodeError
	} else {
		for _, process := range view.Processes {
			st, err := r.runProcess(nodeID, bizdateDir, process, env)
			if err != nil {
				return "", err
			}
			if st != run.NodeSuccess {
				status = run.NodeError
				break
			}
		}
	}

	env["stfw_run_status"] = string(status)
	if !r.runHooks(run.NodeTypeBizdate, "teardown", env) {
		status = run.NodeError
	}
	if err := r.emit(run.NewNodeEndEvent(r.now(), nodeID, status)); err != nil {
		return "", err
	}
	return status, nil
}

// runHooks は階層フック plugins/{level}/_common/{phase}/ を昇順逐次実行する。
// エラー発生時は後続を実行せず false を返す。フック未定義は true (正常)。
// 作業ディレクトリはフック配置ディレクトリ (v0.2 の stfw.bulk_exec_scripts と同じ)。
func (r *runner) runHooks(level run.NodeType, phase string, env map[string]string) bool {
	scripts, err := repository.ListHookScripts(r.projDir, level, phase)
	if err != nil {
		r.log.Error(err.Error())
		return false
	}
	envList := envList(env)
	for _, script := range scripts {
		r.log.Info("hook start", "level", string(level), "phase", phase, "script", filepath.Base(script))
		code, err := gateway.RunScript(filepath.Dir(script), script, envList, r.out, r.errOut)
		if err != nil {
			r.log.Error(err.Error())
			return false
		}
		r.log.Info("hook end", "level", string(level), "phase", phase, "script", filepath.Base(script), "exit_code", code)
		if code != 0 {
			return false
		}
	}
	return true
}

// targetOrder は実行対象シナリオの順序を決める。
// 指定あり: 指定順 (重複は除去)。指定なし: 走査順 (名前昇順)。
// v0.2 の run.dig がコマンド引数順にタスクを生成していたことに対応する。
func targetOrder(tree *scenario.ScenarioTree, names []string) []string {
	if len(names) == 0 {
		return tree.Scenarios()
	}
	seen := map[string]bool{}
	var ordered []string
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		ordered = append(ordered, name)
	}
	return ordered
}
