// Package runscenario は stfw run (内蔵ランナー) のビジネスフローを制御する。
// ScenarioTree を深さ優先で走査し、各階層で setup フック → 子の逐次実行 →
// teardown フックを Go プロセス内で実行する。Error 時は後続の兄弟ノードを
// 実行せず停止し、Warn は記録して続行する (Error > Warn > Success で上位へ集約。
// AS-BUILT §4.6)。実行イベントは JSONL ジャーナル
// (.stfw/runs/{run_id}/journal.jsonl) へ記録する。
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
	"sync"
	"time"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Options は stfw run の実行オプション。
type Options struct {
	DryRun bool
	From   string // 部分実行: 指定ノード以降を実行 (AS-BUILT §3.4)
	Only   string // 部分実行: 指定サブツリーのみ実行 (AS-BUILT §3.4)
}

// Run はシナリオを内蔵ランナーで実行する。
// 実行前にディレクトリ規約を静的検証する (v0.2 では dig 生成時に行われていた検証に相当)。
// now は採番・ジャーナル記録に使う時刻源 (テスト容易性のため引数で受け取る)。
func Run(log *slog.Logger, out, errOut io.Writer, projDir string, cfg *repository.Config, version string, names []string, opts Options, now func() time.Time) error {
	projDir, err := filepath.Abs(projDir)
	if err != nil {
		return err
	}

	// stfw.yml の設定値を環境へ反映し、後続の config チェーン (${...}) から
	// 参照できるようにする (v0.2 export_yaml 互換。§8.2)。
	if err := exportConfigEnv(cfg); err != nil {
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

	// プラグインのランタイム依存 (plugin.yml requires) の実行前ゲート。
	// 実行環境に前提コマンドが無ければプラグインは必ず失敗するため、
	// 実行を開始せず fail-fast する (validate では警告に留めるのと対の扱い)。
	missing, err := repository.CheckPluginRequires(projDir, tree.ProcessTypes())
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		for _, m := range missing {
			v := scenario.Violation{Path: m.ProcessType, Level: scenario.ViolationError,
				Message: fmt.Sprintf("required command not found: %s", m.Command)}
			fmt.Fprintln(out, v.String())
		}
		return fmt.Errorf("missing required command(s) for %d plugin dependency(ies)", len(missing))
	}

	// 接続情報の直書き禁止 (グループ名参照の徹底) を実行前に静的検証する。
	forbidden, err := repository.CheckForbiddenConnConfig(projDir, tree.ScenarioViews())
	if err != nil {
		return err
	}
	if len(forbidden) > 0 {
		for _, f := range forbidden {
			v := scenario.Violation{Path: f.ProcessPath, Level: scenario.ViolationError,
				Message: fmt.Sprintf("config で接続情報を直書きしています (%s)", f.Key)}
			fmt.Fprintln(out, v.String())
		}
		return fmt.Errorf("forbidden connection config in %d place(s)", len(forbidden))
	}

	// 部分実行フィルタ (--from / --only) の解決。指定ノードの不存在は
	// ハウスキープ・ジャーナル作成前に fail-fast する (AS-BUILT §3.4)。
	filter, err := scenario.NewRunFilter(opts.From, opts.Only)
	if err != nil {
		return err
	}
	// 契約 (§3.4) は「シナリオ 1 つのみ」。重複除去前の指定個数で判定する
	// (`demo demo` のような重複指定も部分実行では受け付けない)。
	if filter.Active() && len(names) != 1 {
		return fmt.Errorf("--from / --only requires exactly one scenario")
	}
	targets := targetOrder(tree, names)
	views := make([]scenario.ScenarioView, 0, len(targets))
	for _, name := range targets {
		view, ok := tree.ScenarioView(name)
		if !ok {
			return fmt.Errorf("scenario: %s is not exist", name)
		}
		view, err := filter.Apply(view)
		if err != nil {
			return err
		}
		views = append(views, view)
	}

	// run 開始時のハウスキープ (REQ-019): 保存日数を過ぎた過去の実行結果を削除する。
	// 検証ゲート通過後に行う (誤ったコマンドで削除だけが走ることを防ぐ)。
	housekeep(log, projDir, cfg, now)

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
	defer func() { _ = journal.Close() }()

	// OTLP トレースは run 終了時に flush の完了を待つ (エラー時も待つ)
	notifier := newOTelNotifier(log, cfg, version)
	defer notifier.close()

	r := &runner{
		log:      log,
		out:      out,
		errOut:   errOut,
		projDir:  projDir,
		dryRun:   opts.DryRun,
		filter:   filter,
		now:      now,
		journal:  journal,
		agg:      run.NewRun(runID),
		baseEnv:  baseEnv(cfg, projDir, version, runID, opts.DryRun),
		notifier: notifier,
		reporter: newReporter(log, projDir, runID),
	}

	fmt.Fprintf(out, "run_id: %s\n", runID)
	status, err := r.runRun(runID, views, names)
	if err != nil {
		return err
	}
	if status != run.NodeSuccess {
		return &StatusError{RunID: runID, Status: status}
	}
	log.Info("run finished", "run_id", runID.String(), "status", string(status))
	return nil
}

// StatusError は run が Success 以外のステータスで完走したことを表す。
// インフラ障害 (ジャーナル書き込み失敗等) と区別し、presentation 層が
// Warn=exit 3 / Error=exit 6 へ変換できるようにする (SPEC-023-03)。
type StatusError struct {
	RunID  run.RunID
	Status run.NodeStatus
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("run %s finished with status %s", e.RunID, e.Status)
}

// runner は 1 回の実行のオーケストレーション状態を保持する。
type runner struct {
	log      *slog.Logger
	out      io.Writer
	errOut   io.Writer
	projDir  string
	dryRun   bool
	filter   scenario.RunFilter
	now      func() time.Time
	journal  *repository.Journal
	agg      *run.Run
	baseEnv  map[string]string
	notifier *otelNotifier
	reporter *reporter
	emitMu   sync.Mutex // 並走する parallel の子からの emit を直列化する (AS-BUILT §4.14)
}

// emit は生成時検証 (リプレイと同一の状態遷移検証) を通してジャーナルへ追記する。
// OTLP トレースと HTML レポートはジャーナルイベントの投影のため、追記成功後に
// 連動させる (投影の失敗はログのみで実行結果へは影響しない)。
func (r *runner) emit(ev run.Event) error {
	r.emitMu.Lock()
	defer r.emitMu.Unlock()
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
// views は部分実行フィルタ適用済みの実行計画 (AS-BUILT §3.4)。
func (r *runner) runRun(runID run.RunID, views []scenario.ScenarioView, names []string) (run.NodeStatus, error) {
	nodeID := run.NewRunNodeID(runID)
	attrs := map[string]string{
		"run_id":   runID.String(),
		"run_mode": r.baseEnv["run_mode"],
		"params":   strings.Join(names, " "),
	}
	// 部分実行時はフィルタ指定を attrs へ記録する (この run が全体実行でないことの証跡)
	if key, value := r.filter.Attr(); key != "" {
		attrs[key] = value
	}
	if err := r.emit(run.NewNodeStartEvent(r.now(), nodeID, run.NodeTypeRun, attrs)); err != nil {
		return "", err
	}

	env := cloneEnv(r.baseEnv)
	status := run.NodeSuccess
	if !r.runHooks(run.NodeTypeRun, "setup", env) {
		status = run.NodeError
	} else {
		for _, view := range views {
			st, err := r.runScenario(nodeID, view, env)
			if err != nil {
				return "", err
			}
			// Error 時は後続の兄弟ノードを実行せず停止する。
			// Warn は記録して続行する (Error > Warn > Success で集約。§4.6)
			status = run.WorstStatus(status, st)
			if st == run.NodeError {
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
			status = run.WorstStatus(status, st)
			if st == run.NodeError {
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
	startTS := r.now()
	if err := r.emit(run.NewNodeStartEvent(startTS, nodeID, run.NodeTypeBizdate, attrs)); err != nil {
		return "", err
	}

	bizdateDir := filepath.Join(scenarioDir, view.DirName)
	env := cloneEnv(parentEnv)
	env["stfw_bizdate_dir"] = bizdateDir
	env["stfw_bizdate_dirname"] = view.DirName
	env["stfw_bizdate_seq"] = view.Seq
	env["stfw_bizdate"] = view.Bizdate
	// stfw_bizdate_start_ts は業務日付ノードの node_start 時刻 (RFC3339)。
	// 収集系プラグインが「この業務日付の実行開始以降に発生したログ」を
	// 絞り込む基準として使う (プラグイン env 契約の一部)。
	env["stfw_bizdate_start_ts"] = startTS.Format(time.RFC3339)

	status := run.NodeSuccess
	if !r.runHooks(run.NodeTypeBizdate, "setup", env) {
		status = run.NodeError
	} else {
		for _, process := range view.Processes {
			st, err := r.runProcess(nodeID, bizdateDir, process, env)
			if err != nil {
				return "", err
			}
			status = run.WorstStatus(status, st)
			if st == run.NodeError {
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
