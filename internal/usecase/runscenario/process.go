package runscenario

import (
	"fmt"
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

	status, err := r.execProcess(nodeID, processDir, view.ProcessType, env)
	if err != nil {
		return "", err
	}
	if err := r.emit(run.NewNodeEndEvent(r.now(), nodeID, status)); err != nil {
		return "", err
	}
	return status, nil
}

// execProcess はプロセスの実体を実行し、階層ステータスを返す。
func (r *runner) execProcess(nodeID run.NodeID, processDir, processType string, env map[string]string) (run.NodeStatus, error) {
	// プラグイン解決 (プロジェクト plugins/ → 同梱の順)
	loc, err := repository.ResolveProcessPlugin(r.projDir, processType)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}

	// プラグイン設定 (config.yml の上書きチェーン) を env へ注入
	// (v0.2 の export_config + scripts プラグイン execute の export_yaml 相当)
	confEnv, err := repository.ProcessConfigEnv(r.projDir, loc, processType, processDir)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}
	for k, v := range confEnv {
		env[k] = v
	}

	if loc.Embedded && processType == scriptsPluginType {
		return r.runScriptsProcess(nodeID, processDir, env)
	}
	return r.runPluginProcess(processDir, loc, env)
}

// runScriptsProcess は組込み scripts タイプを Go ネイティブで実行する。
// v0.2 の scripts プラグイン (bin/run/{pre_execute,execute,post_execute}) の移植:
//
//	pre_execute  = scripts/ 直下への実行権限付与
//	execute      = scripts/ 直下を昇順に逐次実行 (retcode 0 → Success / 非 0 → Error /
//	               先行エラー後は Blocked)。dry-run 時はスキップ
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
// エラー発生後の後続ステップは実行せず Blocked として記録する
// (v0.2 の plugin.process.scripts.bulk_exec_scripts と同じ規則)。
func (r *runner) execSteps(nodeID run.NodeID, processDir string, steps []string, env map[string]string) (run.NodeStatus, error) {
	scriptsDir := filepath.Join(processDir, "scripts")
	envList := envList(env)
	failed := false
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
		if code != 0 {
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
	return run.NodeSuccess, nil
}

// runPluginProcess はプロジェクト独自プラグインを exec 契約で実行する。
// フェーズは v0.2 の process_service と同じ setup → pre_execute → execute →
// post_execute → teardown。dry-run 時は execute / post_execute をスキップする。
// 作業ディレクトリはプロセスディレクトリ (v0.2 の execute_service と同じ)。
func (r *runner) runPluginProcess(processDir string, loc repository.PluginLocation, env map[string]string) (run.NodeStatus, error) {
	pluginDir, err := repository.MaterializePlugin(r.projDir, loc)
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

	status := run.NodeSuccess
	phases := []string{"pre_execute"}
	if !r.dryRun {
		phases = append(phases, "execute", "post_execute")
	}
	for _, phase := range phases {
		script := filepath.Join(pluginDir, "bin", "run", phase)
		r.log.Info("process phase start", "phase", phase)
		code, err := gateway.RunScript(processDir, script, envList, r.out, r.errOut)
		if err != nil {
			r.log.Error(err.Error())
			code = run.ExitError.Int()
		}
		r.log.Info("process phase end", "phase", phase, "exit_code", code)
		if code != 0 {
			status = run.NodeError
			break
		}
	}

	r.runProcessTeardown(processDir, status, env)
	return status, nil
}

// runProcessTeardown はプロセスの teardown フックを実行する。
// フックの失敗はプロセスの結果に影響しない (v0.2 の run_requested は
// teardown のリターンコードを結果へ反映していなかった)。
func (r *runner) runProcessTeardown(processDir string, status run.NodeStatus, env map[string]string) {
	env = cloneEnv(env)
	env["stfw_run_status"] = string(status)
	retcode := run.ExitSuccess
	if status != run.NodeSuccess {
		retcode = run.ExitError
	}
	env["stfw_process_retcode"] = strconv.Itoa(retcode.Int())
	if !r.runHooks(run.NodeTypeProcess, "teardown", env) {
		r.log.Warn("process teardown hook failed", "process_dir", processDir)
	}
}
