package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/initialize"
	"github.com/scenario-test-framework/stfw/internal/usecase/plugin"
)

func newInitCmd(a *app) *cobra.Command {
	var skipPluginInit bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize stfw project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := initialize.Run(a.log, cmd.OutOrStdout(), a.projDir); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			// 組み込みプラグインのプロビジョニング (例: collectLog の logfilter DL)。
			// 個々の失敗は warn 継続 (stfw plugin install <type> で再実行可)。
			if !skipPluginInit {
				if err := plugin.InitAll(a.log, cmd.OutOrStdout(), cmd.ErrOrStderr(), a.projDir); err != nil {
					a.log.Error(err.Error())
					return &exitError{code: run.ExitError, err: err}
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&skipPluginInit, "skip-plugin-init", false, "skip per-plugin provisioning (install)")
	return cmd
}
