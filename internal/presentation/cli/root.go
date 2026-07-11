// Package cli は cobra による CLI 表面 (presentation 層) を提供する。
package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/presentation/logger"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Version はビルド時に -ldflags で注入される。
var Version = "1.1.1-dev"

type app struct {
	log     *slog.Logger
	masker  *logger.Masker
	config  *repository.Config
	projDir string
}

// exitError は終了コード付きエラー。RunE から返すと Execute が
// プロセス終了コードへ変換する。
type exitError struct {
	code run.ExitCode
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

// Execute は CLI を実行し、プロセス終了コードを返す。
func Execute() int {
	a := &app{}
	rootCmd := newRootCmd(a)
	if err := rootCmd.Execute(); err != nil {
		var coded *exitError
		if errors.As(err, &coded) {
			return coded.code.Int()
		}
		// exitError 以外 (引数パースエラー等) はコマンド側で未出力のためここで出力する
		fmt.Fprintln(os.Stderr, err)
		return run.ExitError.Int()
	}
	return run.ExitSuccess.Int()
}

func newRootCmd(a *app) *cobra.Command {
	var logLevel string
	cmd := &cobra.Command{
		Use:          "stfw",
		Short:        "scenario test framework cli",
		Version:      Version,
		SilenceUsage: true,
		// エラーはコマンド側で slog へ出力するため cobra の二重出力を抑止する
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			a.projDir = discoverProjectDir()
			cfg, warns, err := repository.LoadConfig(a.projDir)
			if err != nil {
				return err
			}
			a.config = cfg

			level := cfg.LogLevel()
			if logLevel != "" {
				level = logLevel
			}
			a.log, a.masker = logger.New(cmd.ErrOrStderr(), logger.ParseLevel(level))
			for _, w := range warns {
				a.log.Warn(w)
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "", "log level [error, warn, info, debug, trace] (default: info)")
	cmd.AddCommand(
		newInitCmd(a),
		newNewCmd(a),
		newScenarioCmd(a),
		newValidateCmd(a),
		newRunCmd(a),
		newStatusCmd(a),
		newReportCmd(a),
		newInventoryCmd(a),
		newSecretCmd(a),
		newSSHCmd(a),
		newPluginCmd(a),
	)
	return cmd
}

// discoverProjectDir はプロジェクトディレクトリを決定する。
// 優先順: STFW_PROJ_DIR 環境変数 → カレントから上位への stfw.yml 探索 →
// カレントディレクトリ (未初期化とみなす)。v0.2 の bin/stfw と同じ規則。
func discoverProjectDir() string {
	if dir := os.Getenv("STFW_PROJ_DIR"); dir != "" {
		return dir
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, project.ConfigFileName)); err == nil {
			return dir
		}
		if dir == filepath.Dir(dir) {
			break
		}
	}
	return cwd
}
