package cli

import (
	"github.com/spf13/cobra"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/inventory"
)

func newInventoryCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inventory",
		Short: "read inventory settings",
	}
	cmd.AddCommand(
		newInventoryListCmd(a),
		newInventoryExistsCmd(a),
		newInventoryArchCmd(a),
	)
	return cmd
}

func newInventoryArchCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "arch <host>",
		Short: "print arch configured for host (empty if unset)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := inventory.Arch(cmd.OutOrStdout(), a.projDir, a.config.Get("stfw_inventory"), args[0]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newInventoryListCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "list [group]",
		Short: "list hosts belonging to group (default: all)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			group := project.InventoryGroupAll
			if len(args) == 1 {
				group = args[0]
			}
			if err := inventory.List(cmd.OutOrStdout(), a.projDir, a.config.Get("stfw_inventory"), group); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newInventoryExistsCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "exists <group>",
		Short: "check existence of group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := inventory.Exists(cmd.OutOrStdout(), a.projDir, a.config.Get("stfw_inventory"), args[0]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}
