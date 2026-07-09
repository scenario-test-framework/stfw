package cli

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/repository"
	"github.com/scenario-test-framework/stfw/internal/usecase/plugin"
)

func newPluginCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "manage process plugins",
	}
	cmd.AddCommand(
		newPluginListCmd(a),
		newPluginInstallCmd(a),
		newPluginMysqlCSVCmd(a),
		newPluginRedisEncodeCmd(a),
		newPluginRedisDecodeCmd(a),
	)
	return cmd
}

// newPluginRedisEncodeCmd は組込み exportRedis が利用する内部ヘルパ。
// redis-cli -2 --json の値 (stdin) を正規化し、CSV の 1 行 (key,type,ttl,value) を
// stdout へ書き出す。エンドユーザー向けではないため隠しコマンドとする。
func newPluginRedisEncodeCmd(a *app) *cobra.Command {
	var key, typ, ttl string
	c := &cobra.Command{
		Use:    "redis-encode-row",
		Short:  "encode a redis value into one CSV row (internal plugin helper)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			if err := repository.RedisEncodeRow(cmd.OutOrStdout(), key, typ, ttl, raw); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	c.Flags().StringVar(&key, "key", "", "redis key")
	c.Flags().StringVar(&typ, "type", "", "redis type (string/list/set/hash/zset)")
	c.Flags().StringVar(&ttl, "ttl", "-1", "ttl seconds (-1 no expire)")
	return c
}

// newPluginRedisDecodeCmd は組込み importRedis が利用する内部ヘルパ。
// export 形式の CSV (stdin) を、キーを再現する redis-cli コマンド列 (stdout) へ変換する。
func newPluginRedisDecodeCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:    "redis-decode",
		Short:  "decode export CSV into redis-cli commands (internal plugin helper)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := repository.RedisDecode(cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

// newPluginMysqlCSVCmd は組込み RDBMS プラグイン (exportMysql) が利用する
// 内部ヘルパ。`mysql --batch` のタブ区切り出力を stdin で受け、RFC 4180 準拠の
// CSV を stdout へ書き出す。エンドユーザー向けではないため隠しコマンドとする。
func newPluginMysqlCSVCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:    "mysql-tsv-to-csv",
		Short:  "convert `mysql --batch` output to RFC4180 CSV (internal plugin helper)",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := repository.MySQLBatchTSVToCSV(cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newPluginListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list process plugins",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := plugin.List(cmd.OutOrStdout(), a.projDir); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newPluginInstallCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "install <type>",
		Short: "install process plugin dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := plugin.Install(a.log, cmd.OutOrStdout(), cmd.ErrOrStderr(), a.projDir, args[0])
			if err == nil {
				return nil
			}
			// インストール済みは警告 (exit 3) として扱う (v0.2 互換)
			if errors.Is(err, plugin.ErrAlreadyInstalled) {
				a.log.Warn(err.Error())
				return &exitError{code: run.ExitWarn, err: err}
			}
			a.log.Error(err.Error())
			return &exitError{code: run.ExitError, err: err}
		},
	}
}
