// Package plugin は stfw plugin (list / install) のビジネスフローを制御する。
package plugin

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

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

	env := []string{"STFW_PROJ_DIR=" + projDir}
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
