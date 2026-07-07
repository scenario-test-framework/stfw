package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/sshtrust"
)

func newSSHCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "manage SSH server keys",
	}
	cmd.AddCommand(newSSHTrustCmd(a))
	return cmd
}

func newSSHTrustCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "trust <host|group>",
		Short: "register SSH server keys to known_hosts",
		Long: "register SSH server keys to ~/.ssh/known_hosts.\n" +
			"if the argument is an inventory group, all hosts in the group are registered.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := sshtrust.Trust(a.log, a.projDir, a.config.Get("stfw_inventory"), args[0]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}
