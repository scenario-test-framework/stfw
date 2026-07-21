package runscenario

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/gateway"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// runParallelProcess は組込み parallel タイプを Go ネイティブで実行する (AS-BUILT §4.14)。
// 子プロセスを goroutine で並走させ、子 1 件をジャーナル上の 1 ステップとして記録する。
// 親のステータスは全子の悪い方 (Error > Warn > Success)。子同士に Blocked は無い
// (並走に順序依存が無いため、先行子の Error 後も全子を必ず実行する)。
// process 階層フック (setup / teardown) は親で 1 回のみ実行し、子では実行しない。
func (r *runner) runParallelProcess(nodeID run.NodeID, processDir string, view scenario.ProcessView, env map[string]string) (run.NodeStatus, error) {
	children := view.Children
	if len(children) == 0 {
		// validate が実行前に検出するため通常は到達しない (防御)
		r.log.Error(fmt.Sprintf("parallel process %s has no child process", view.DirName))
		return run.NodeError, nil
	}

	maxParallel, err := parseMaxParallel(env["stfw_process_parallel_max_parallel"])
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError, nil
	}

	// 計画列挙: 子ディレクトリ名 (昇順) を全件 Pending で登録する (dry-run でも行う)
	steps := make([]string, 0, len(children))
	for _, c := range children {
		steps = append(steps, c.DirName)
	}
	if err := r.emit(run.NewStepsEnumeratedEvent(r.now(), nodeID, steps)); err != nil {
		return "", err
	}

	// setup フック (v0.2 互換: setup 失敗時は teardown も実行しない)
	if !r.runHooks(run.NodeTypeProcess, "setup", env) {
		return run.NodeError, nil
	}

	status, err := r.execChildren(nodeID, processDir, children, env, maxParallel)
	if err != nil {
		return "", err
	}

	r.runProcessTeardown(processDir, status, env)
	return status, nil
}

// childPlugin は並走開始前に解決・展開した子プラグインの実体。
// 同梱プラグインの展開 (MaterializePlugin) は展開先を毎回ワイプするため、
// 同一タイプの子が並走中に展開し合わないよう事前にタイプ単位で 1 回だけ行う。
type childPlugin struct {
	loc       repository.PluginLocation
	pluginDir string // exec 契約の実体パス (組込み scripts は使用しない)
	err       error  // 解決・展開の失敗 (該当タイプの子を Error にする)
}

// execChildren は子プロセスを並走実行し、全子の悪い方のステータスを返す。
// ワーカープールで起動順 (seq 昇順) を保証し、maxParallel (0 = 上限なし) で
// 同時実行数を制限する。エラー戻り値はジャーナル追記失敗 (実行継続不能) のみ。
func (r *runner) execChildren(nodeID run.NodeID, processDir string, children []scenario.ProcessView, env map[string]string, maxParallel int) (run.NodeStatus, error) {
	plugins := r.resolveChildPlugins(children)

	workers := len(children)
	if maxParallel > 0 && maxParallel < workers {
		workers = maxParallel
	}
	queue := make(chan scenario.ProcessView, len(children))
	for _, c := range children {
		queue <- c
	}
	close(queue)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		status  = run.NodeSuccess
		emitErr error
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for child := range queue {
				// ジャーナル追記が失敗した場合は run を継続できないため、
				// 未着手の子は起動せずに終了へ向かう (逐次経路の即時中断と対称)
				mu.Lock()
				aborted := emitErr != nil
				mu.Unlock()
				if aborted {
					continue
				}

				start := r.now()
				st := r.execChildProcess(processDir, child, plugins[child.ProcessType], env)
				end := r.now()

				stepStatus := run.StepSuccess
				code := run.ExitSuccess
				switch st {
				case run.NodeWarn:
					stepStatus = run.StepWarn
					code = run.ExitWarn
				case run.NodeError:
					stepStatus = run.StepError
					code = run.ExitError
				}

				mu.Lock()
				status = run.WorstStatus(status, st)
				if err := r.emit(run.NewStepEndEvent(end, nodeID, child.DirName, stepStatus, code.Int(), start, end)); err != nil && emitErr == nil {
					emitErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if emitErr != nil {
		return "", emitErr
	}
	return status, nil
}

// resolveChildPlugins は子のプロセスタイプ単位でプラグインを解決・展開する。
// 失敗はタイプ単位で保持し、該当タイプの子のみ Error にする (他の子は実行する)。
func (r *runner) resolveChildPlugins(children []scenario.ProcessView) map[string]childPlugin {
	plugins := map[string]childPlugin{}
	for _, c := range children {
		if _, ok := plugins[c.ProcessType]; ok {
			continue
		}
		loc, err := repository.ResolveProcessPlugin(r.projDir, c.ProcessType)
		if err != nil {
			plugins[c.ProcessType] = childPlugin{err: err}
			continue
		}
		p := childPlugin{loc: loc}
		if !loc.Embedded || c.ProcessType != scriptsPluginType {
			p.pluginDir, p.err = repository.MaterializePlugin(r.projDir, loc)
		}
		plugins[c.ProcessType] = p
	}
	return plugins
}

// execChildProcess は parallel の子プロセス 1 件を実行し、階層ステータスを返す。
// 実行内容は通常プロセスと同じ (config チェーン注入 → exec 契約 / scripts ネイティブ)
// だが、process 階層フックは親側で 1 回だけ実行するため子では実行しない (AS-BUILT §4.14)。
func (r *runner) execChildProcess(parentDir string, view scenario.ProcessView, plugin childPlugin, parentEnv map[string]string) run.NodeStatus {
	if plugin.err != nil {
		r.log.Error(plugin.err.Error())
		return run.NodeError
	}

	childDir := filepath.Join(parentDir, view.DirName)
	env := cloneEnv(parentEnv)
	env["stfw_process_type"] = view.ProcessType
	env["stfw_process_dir"] = childDir
	env["stfw_process_dirname"] = view.DirName
	env["stfw_process_seq"] = view.Seq
	env["stfw_process_group"] = view.Group

	confEnv, err := repository.ProcessConfigEnv(r.projDir, plugin.loc, view.ProcessType, childDir)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError
	}
	for k, v := range confEnv {
		env[k] = v
	}
	env["stfw_plugin_cache_dir"] = repository.PluginCacheDir(r.projDir, view.ProcessType)

	// 子ごとの行バッファ: 並走する子の出力断片が同一行に混ざらないようにする
	// (行の間の交錯は許容する。AS-BUILT §4.14)
	out := &lineWriter{w: r.out}
	errOut := &lineWriter{w: r.errOut}
	defer func() {
		_ = out.Flush()
		_ = errOut.Flush()
	}()

	if plugin.loc.Embedded && view.ProcessType == scriptsPluginType {
		return r.execChildScripts(childDir, env, out, errOut)
	}

	envList := envList(env)
	// プロセス実行の前提条件: プラグインがインストール済み (is_installed=true) であること
	if !repository.IsPluginInstalled(plugin.pluginDir, envList, errOut) {
		r.log.Error(fmt.Sprintf("%s is not installed", plugin.pluginDir))
		return run.NodeError
	}
	return r.execPluginPhases(childDir, plugin.pluginDir, env, out, errOut)
}

// execChildScripts は parallel の子の組込み scripts タイプを Go ネイティブで実行する。
// 粒度は「子 1 件 = 1 ステップ」のため、子内スクリプトのステップイベントは記録しない
// (昇順逐次・Error 後は後続を実行しない規則は §4.3 と同じで、結果はログにのみ現れる)。
func (r *runner) execChildScripts(processDir string, env map[string]string, out, errOut io.Writer) run.NodeStatus {
	steps, err := repository.ListScriptSteps(processDir)
	if err != nil {
		r.log.Error(err.Error())
		return run.NodeError
	}

	// pre_execute 相当: scripts/ 直下への実行権限付与
	if err := repository.EnsureStepScriptsExecutable(processDir); err != nil {
		r.log.Error(err.Error())
		return run.NodeError
	}

	// execute (dry-run 時はスキップ)
	if r.dryRun {
		return run.NodeSuccess
	}
	scriptsDir := filepath.Join(processDir, "scripts")
	dirname := filepath.Base(processDir)
	envList := envList(env)
	status := run.NodeSuccess
	for _, step := range steps {
		// Error 後の後続スクリプトは実行しない (§4.3 と同じ規則。ジャーナルには
		// 記録しないため、ログで blocked を明示する)
		if status == run.NodeError {
			r.log.Info("step blocked", "process_dir", dirname, "script", step)
			continue
		}
		r.log.Info("step start", "process_dir", dirname, "script", step)
		code, err := gateway.RunScript(scriptsDir, filepath.Join(scriptsDir, step), envList, out, errOut)
		if err != nil {
			r.log.Error(err.Error())
			code = run.ExitError.Int()
		}
		r.log.Info("step end", "process_dir", dirname, "script", step, "exit_code", code)
		switch {
		case code == run.ExitSuccess.Int():
		case code == run.ExitWarn.Int():
			status = run.WorstStatus(status, run.NodeWarn)
		default:
			status = run.NodeError
		}
	}
	return status
}

// parseMaxParallel は max_parallel 設定値 (env 経由) を検証する。
// 未設定・空は 0 (上限なし)。負数・非整数は設定不正としてエラーを返す。
func parseMaxParallel(v string) (int, error) {
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("stfw.process.parallel.max_parallel must be a non-negative integer: %q", v)
	}
	return n, nil
}

// lineWriter は並走する子プロセスの出力を行単位で下位 Writer へ転送する。
// 子ごとに独立したバッファを持つことで、複数の子の出力断片が同一行に混ざらない
// ようにする (下位 Writer への書き込みは完了行単位。行の間の交錯は許容する)。
type lineWriter struct {
	w   io.Writer
	buf []byte
}

// Write は p を蓄積し、改行までの完了行を下位 Writer へ転送する。
// p はバッファへ取り込み済みのため、下位 Writer の失敗時も消費バイト数は
// len(p) を返す (short write と誤認させない。Masker.Write と同じ判断)。
func (l *lineWriter) Write(p []byte) (int, error) {
	l.buf = append(l.buf, p...)
	idx := bytes.LastIndexByte(l.buf, '\n')
	if idx < 0 {
		return len(p), nil
	}
	line := l.buf[:idx+1]
	if _, err := l.w.Write(line); err != nil {
		return len(p), err
	}
	l.buf = append(l.buf[:0], l.buf[idx+1:]...)
	return len(p), nil
}

// Flush は未改行の残りを転送する (gateway.RunScript がスクリプト完了時に呼ぶ)。
// 改行で終わらない最終出力には改行を補う: 下位 Writer が行バッファ (Masker) の場合、
// 未改行断片がバッファに残ると後続の別の子の行と連結されてしまうため
// (「各ストリームの行内は壊れない」契約の担保。AS-BUILT §4.14)。
func (l *lineWriter) Flush() error {
	if len(l.buf) == 0 {
		return nil
	}
	if l.buf[len(l.buf)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.w.Write(l.buf)
	l.buf = l.buf[:0]
	return err
}
