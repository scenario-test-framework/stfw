package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
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
	)
	return cmd
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
