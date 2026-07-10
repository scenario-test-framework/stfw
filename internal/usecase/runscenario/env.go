package runscenario

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// exportConfigEnv は stfw.yml (+ 同梱デフォルト) のフラット化結果を stfw 自身の
// 環境変数へ反映する (v0.2 の export_yaml 相当)。
// 子スクリプトへは baseEnv 経由で既に渡るが、プラグイン/プロセス config チェーンの
// ${...} 展開 (flattenValue の os.Getenv) は stfw 自身の環境しか見ないため、
// stfw.yml の設定値 (例: stfw.db.database → ${stfw_db_database}) を config から
// 参照できるようにするにはこの明示的な export が必要。
func exportConfigEnv(cfg *repository.Config) error {
	for k, v := range cfg.Flat() {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

// baseEnv は全スクリプト共通の env (実行契約) を組み立てる。
// v0.2 の setenv (STFW_PROJ_DIR_* の export) + dig の _export (run_id / run_mode) +
// stfw.yml のフラット化 (export_yaml) に対応する。
// STFW_HOME はバイナリ化により廃止 (配置ディレクトリが存在しないため代替なし)。
func baseEnv(cfg *repository.Config, projDir, version string, runID run.RunID, dryRun bool) map[string]string {
	env := cfg.Flat()
	env["STFW_PROJ_DIR"] = projDir
	env["STFW_PROJ_DIR_CONFIG"] = filepath.Join(projDir, "config")
	env["STFW_PROJ_DIR_PLUGIN"] = filepath.Join(projDir, "plugins")
	env["STFW_PROJ_DIR_DATA"] = filepath.Join(projDir, project.DataDirName)
	env["STFW_VERSION"] = version
	env["run_id"] = runID.String()
	env["run_mode"] = "--run"
	if dryRun {
		env["run_mode"] = "--dry-run"
	}
	return env
}

// cloneEnv は env マップのコピーを返す (階層コンテキストの追記用)。
func cloneEnv(env map[string]string) map[string]string {
	clone := make(map[string]string, len(env))
	for k, v := range env {
		clone[k] = v
	}
	return clone
}

// envList は env マップをキー昇順の KEY=VALUE リストへ変換する。
func envList(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	list := make([]string, 0, len(keys))
	for _, k := range keys {
		list = append(list, k+"="+env[k])
	}
	return list
}
