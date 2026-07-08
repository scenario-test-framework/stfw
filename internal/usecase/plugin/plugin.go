// Package plugin は stfw plugin (list / install) のビジネスフローを制御する。
package plugin

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/scenario"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// ErrAlreadyInstalled はインストール済みプラグインへの再インストール。
// v0.2 互換で警告 (exit 3) として扱う。
var ErrAlreadyInstalled = errors.New("already installed")

// List はプロセスプラグイン名の一覧を出力する。
// プロジェクト + 同梱の和集合を昇順で表示する (v0.2 の process -l と同じ)。
func List(out io.Writer, projDir string) error {
	names, err := repository.ListProcessPlugins(projDir)
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Fprintln(out, name)
	}
	return nil
}

// Install はプロセスプラグインの依存モジュールをインストールする
// (v0.2 の process -I 相当)。同梱プラグインは .stfw/ 配下へ展開してから
// bin/install/install を実行する。
func Install(log *slog.Logger, out, errOut io.Writer, projDir, processType string) error {
	if err := scenario.ValidateProcessType(processType); err != nil {
		return err
	}

	// プラグイン解決 (プロジェクト plugins/ → 同梱の順)
	loc, err := repository.ResolveProcessPlugin(projDir, processType)
	if err != nil {
		return err
	}

	dir, err := repository.MaterializePlugin(projDir, loc)
	if err != nil {
		return fmt.Errorf("plugin materialize: %w", err)
	}

	env, err := provisionEnv(projDir, loc, processType)
	if err != nil {
		return err
	}
	if repository.IsPluginInstalled(dir, env, errOut) {
		return fmt.Errorf("%s is %w", dir, ErrAlreadyInstalled)
	}

	code, err := repository.RunPluginInstall(dir, env, out, errOut)
	if err != nil {
		return err
	}
	if code != 0 {
		return fmt.Errorf("plugin install failed: %s (exit %d)", processType, code)
	}
	log.Info("plugin installed", "type", processType, "dir", dir)
	return nil
}

// InitAll は解決可能な全プロセスプラグインの install (プロビジョニング) を
// 実行する (stfw init から呼ばれる)。インストール済みはスキップし、個々の
// 失敗は warn に留めて後続を継続する (1 プラグインの DL 失敗で init 全体を
// 失敗させない。利用者は stfw plugin install <type> で個別に再実行できる)。
func InitAll(log *slog.Logger, out, errOut io.Writer, projDir string) error {
	names, err := repository.ListProcessPlugins(projDir)
	if err != nil {
		return err
	}
	for _, name := range names {
		err := Install(log, out, errOut, projDir, name)
		switch {
		case err == nil:
			// 実行済み
		case errors.Is(err, ErrAlreadyInstalled):
			log.Info("plugin already provisioned", "type", name)
		default:
			log.Warn("plugin provisioning failed (retry with `stfw plugin install`)", "type", name, "error", err.Error())
		}
	}
	return nil
}

// provisionEnv は install / is_installed へ渡すプロビジョニング用 env を返す。
// プラグインはダウンロードしたバイナリ等を stfw_plugin_cache_dir (実行を
// またいで保持される永続キャッシュ) に配置する。プラグイン設定 (config.yml +
// プロジェクト上書き) も公開し、install が version / arches 等を参照できるようにする。
func provisionEnv(projDir string, loc repository.PluginLocation, processType string) ([]string, error) {
	cacheDir := repository.PluginCacheDir(projDir, processType)
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	m := map[string]string{
		"STFW_PROJ_DIR":         projDir,
		"STFW_PROJ_DIR_DATA":    filepath.Join(projDir, project.DataDirName),
		"stfw_process_type":     processType,
		"stfw_plugin_cache_dir": cacheDir,
	}
	conf, err := repository.PluginConfigEnv(projDir, loc, processType)
	if err != nil {
		return nil, err
	}
	for k, v := range conf {
		if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
	env := make([]string, 0, len(m))
	for k, v := range m {
		env = append(env, k+"="+v)
	}
	return env, nil
}
