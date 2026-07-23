package runscenario

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// scriptsPluginType は Go ネイティブ実装で実行する組込みプロセスタイプ。
const scriptsPluginType = "scripts"

// runProcess はプロセス 1 件を実行する。
// 組込み scripts タイプは Go ネイティブ、それ以外は exec 契約
// (env 注入 + リターンコード) でプラグインのエントリポイントを呼ぶ。
func (r *runner) runProcess(parent run.NodeID, bizdateDir string, view scenario.ProcessView, parentEnv map[string]string) (run.NodeStatus, error) {
	nodeID, err := parent.Child(view.DirName)
	if err != nil {
		return "", err
	}
	attrs := map[string]string{
		"dirname":      view.DirName,
		"seq":          view.Seq,
		"group":        view.Group,
		"process_type": view.ProcessType,
	}
	if err := r.emit(run.NewNodeStartEvent(r.now(), nodeID, run.NodeTypeProcess, attrs)); err != nil {
		return "", err
	}

	processDir := filepath.Join(bizdateDir, view.DirName)
	env := cloneEnv(parentEnv)
	env["stfw_process_type"] = view.ProcessType
	env["stfw_process_dir"] = processDir
	env["stfw_process_dirname"] = view.DirName
	env["stfw_process_seq"] = view.Seq
	env["stfw_process_group"] = view.Group

	status, err := r.execProcess(nodeID, processDir, view, env)
	if err != nil {
		return "", err
	}
	if err := r.emit(run.NewNodeEndEvent(r.now(), nodeID, status)); err != nil {
		return "", err
	}
	return status, nil
}

// execProcess はプロセスの実体を実行し、階層ステータスを返す。
func (r *runner) execProcess(nodeID run.NodeID, processDir string, view scenario.ProcessView, env map[string]string) (run.NodeStatus, error) {
	// プラグイン解決 (プロジェクト plugins/ → 同梱の順)
	loc, err := repository.ResolveProcessPlugin(r.projDir, view.ProcessType)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}

	// プラグイン設定 (config.yml の上書きチェーン) を env へ注入
	// (v0.2 の export_config + scripts プラグイン execute の export_yaml 相当)
	confEnv, err := repository.ProcessConfigEnv(r.projDir, loc, view.ProcessType, processDir)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}
	for k, v := range confEnv {
		env[k] = v
	}
	// プラグインが install でプロビジョニングした資産の永続キャッシュ
	// (collectLog の logfilter バイナリ等) を実行時にも参照できるようにする。
	env["stfw_plugin_cache_dir"] = repository.PluginCacheDir(r.projDir, view.ProcessType)

	if loc.Embedded && view.ProcessType == scriptsPluginType {
		return r.runScriptsProcess(nodeID, processDir, env)
	}
	if loc.Embedded && view.ProcessType == scenario.ParallelProcessType {
		return r.runParallelProcess(nodeID, processDir, view, env)
	}
	return r.runPluginProcess(processDir, loc, env)
}

// runScriptsProcess は組込み scripts タイプを Go ネイティブで実行する。
// v0.2 の scripts プラグイン (bin/run/{pre_execute,execute,post_execute}) の移植:
//
//	pre_execute  = scripts/ 直下への実行権限付与
//	execute      = scripts/ 直下を昇順に逐次実行 (retcode 0 → Success / 3 → Warn 続行 /
//	               その他非 0 → Error / 先行 Error 後は Blocked)。dry-run 時はスキップ
//	post_execute = 何もしない (v0.2 も exit 0 のみ)。dry-run 時はスキップ
func (r *runner) runScriptsProcess(nodeID run.NodeID, processDir string, env map[string]string) (run.NodeStatus, error) {
	// 計画列挙: 全ステップを Pending で登録する (dry-run でも行う)
	steps, err := repository.ListScriptSteps(processDir)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}
	if len(steps) > 0 {
		if err := r.emit(run.NewStepsEnumeratedEvent(r.now(), nodeID, steps)); err != nil {
			return "", err
		}
	}

	// setup フック (v0.2 互換: setup 失敗時は teardown も実行しない)
	if !r.runHooks(run.NodeTypeProcess, "setup", env) {
		return run.NodeError, nil
	}

	status := run.NodeSuccess

	// pre_execute
	if err := repository.EnsureStepScriptsExecutable(processDir); err != nil {
		r.log.Error(err.Error())
		status = run.NodeError
	}

	// execute (dry-run 時はスキップ)
	if status == run.NodeSuccess && !r.dryRun {
		status, err = r.execSteps(nodeID, processDir, steps, env)
		if err != nil {
			return "", err
		}
	}

	// post_execute: 何もしない

	r.runProcessTeardown(processDir, status, env)
	return status, nil
}

// execSteps はステップスクリプトを昇順に逐次実行する。
// Error 発生後の後続ステップは実行せず Blocked として記録する
// (v0.2 の plugin.process.scripts.bulk_exec_scripts と同じ規則)。
// exit 3 は Warn として記録して続行する (AS-BUILT §4.6)。
func (r *runner) execSteps(nodeID run.NodeID, processDir string, steps []string, env map[string]string) (run.NodeStatus, error) {
	scriptsDir := filepath.Join(processDir, "scripts")
	envList := envList(env)
	failed := false
	warned := false
	for _, step := range steps {
		if failed {
			if err := r.emit(run.NewStepBlockedEvent(r.now(), nodeID, step)); err != nil {
				return "", err
			}
			continue
		}

		start := r.now()
		r.log.Info("step start", "script", step)
		code, err := gateway.RunScript(scriptsDir, filepath.Join(scriptsDir, step), envList, r.out, r.errOut)
		if err != nil {
			// 実行不能 (実行形式エラー等) はエラー扱い
			r.log.Error(err.Error())
			code = run.ExitError.Int()
		}
		end := r.now()

		status := run.StepSuccess
		switch {
		case code == run.ExitSuccess.Int():
		case code == run.ExitWarn.Int():
			status = run.StepWarn
			warned = true
		default:
			status = run.StepError
			failed = true
		}
		r.log.Info("step end", "script", step, "status", string(status), "exit_code", code)
		if err := r.emit(run.NewStepEndEvent(end, nodeID, step, status, code, start, end)); err != nil {
			return "", err
		}
	}
	if failed {
		return run.NodeError, nil
	}
	if warned {
		return run.NodeWarn, nil
	}
	return run.NodeSuccess, nil
}

// runPluginProcess はプロジェクト独自プラグインを exec 契約で実行する。
// フェーズは v0.2 の process_service と同じ setup → pre_execute → execute →
// post_execute → teardown。dry-run 時は execute / post_execute をスキップする。
// 作業ディレクトリはプロセスディレクトリ (v0.2 の execute_service と同じ)。
func (r *runner) runPluginProcess(processDir string, loc repository.PluginLocation, env map[string]string) (run.NodeStatus, error) {
	// 同梱プラグインは run 単位の展開先へ展開する (並走 run 間の衝突防止。AS-BUILT §5.7)
	pluginDir, err := repository.MaterializePlugin(r.runDir, loc)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}

	envList := envList(env)
	// プロセス実行の前提条件: プラグインがインストール済み (is_installed=true) であること
	if !repository.IsPluginInstalled(pluginDir, envList, r.errOut) {
		r.log.Error(fmt.Sprintf("%s is not installed", pluginDir))
		return run.NodeError, nil
	}

	// setup フック (v0.2 互換: setup 失敗時は teardown も実行しない)
	if !r.runHooks(run.NodeTypeProcess, "setup", env) {
		return run.NodeError, nil
	}

	status := r.execPluginPhases(processDir, pluginDir, env, r.out, r.errOut)

	r.runProcessTeardown(processDir, status, env)
	return status, nil
}

// execPluginPhases は exec 契約のフェーズ (pre_execute → execute → post_execute) を
// 実行し、階層ステータスを返す。dry-run 時は pre_execute のみ実行する。
// 通常プロセス (runPluginProcess) と parallel の子 (execChildProcess) が共用する。
func (r *runner) execPluginPhases(processDir, pluginDir string, env map[string]string, out, errOut io.Writer) run.NodeStatus {
	envList := envList(env)
	status := run.NodeSuccess
	phases := []string{"pre_execute"}
	if !r.dryRun {
		phases = append(phases, "execute", "post_execute")
	}
	for _, phase := range phases {
		script := filepath.Join(pluginDir, "bin", "run", phase)
		r.log.Info("process phase start", "phase", phase, "process_dir", filepath.Base(processDir))
		code, err := gateway.RunScript(processDir, script, envList, out, errOut)
		if err != nil {
			r.log.Error(err.Error())
			code = run.ExitError.Int()
		}
		r.log.Info("process phase end", "phase", phase, "process_dir", filepath.Base(processDir), "exit_code", code)
		// exit 3 は Warn として記録して後続フェーズを続行する (AS-BUILT §4.6)
		if code == run.ExitWarn.Int() {
			status = run.NodeWarn
			continue
		}
		if code != 0 {
			status = run.NodeError
			break
		}
	}
	return status
}

// runProcessTeardown はプロセスの teardown フックを実行する。
// フックの失敗はプロセスの結果に影響しない (v0.2 の run_requested は
// teardown のリターンコードを結果へ反映していなかった)。
func (r *runner) runProcessTeardown(processDir string, status run.NodeStatus, env map[string]string) {
	env = cloneEnv(env)
	env["stfw_run_status"] = string(status)
	retcode := run.ExitSuccess
	switch status {
	case run.NodeWarn:
		retcode = run.ExitWarn
	case run.NodeSuccess:
	default:
		retcode = run.ExitError
	}
	env["stfw_process_retcode"] = strconv.Itoa(retcode.Int())
	if !r.runHooks(run.NodeTypeProcess, "teardown", env) {
		r.log.Warn("process teardown hook failed", "process_dir", processDir)
	}
}
