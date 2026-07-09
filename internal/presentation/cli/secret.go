package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/scenario-test-framework/stfw/internal/domain/run"
	"github.com/scenario-test-framework/stfw/internal/usecase/secret"
)

func newSecretCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "manage encrypted credentials",
	}
	cmd.AddCommand(
		newSecretKeygenCmd(a),
		newSecretSetCmd(a),
		newSecretShowCmd(a),
		newSecretMigrateCmd(a),
	)
	return cmd
}

func newSecretKeygenCmd(a *app) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "generate encrypt key pair (age X25519)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := secret.Keygen(a.log, cmd.OutOrStdout(), a.projDir, force); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force regenerate")
	return cmd
}

func newSecretSetCmd(a *app) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "set <host> <user> [password]",
		Short: "encrypt and save credential",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			password, err := resolvePassword(cmd, args)
			if err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			// 復号値と同様にログへの漏洩を防ぐため Masker へ登録する
			a.masker.Register(password)

			if err := secret.Set(a.log, a.projDir, args[0], args[1], password, force); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "force overwrite")
	return cmd
}

func newSecretShowCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "show <host> <user>",
		Short: "decrypt and print credential",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// show は意図的な表示のため復号値を Masker には登録しない
			if err := secret.Show(cmd.OutOrStdout(), a.projDir, args[0], args[1]); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

func newSecretMigrateCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "migrate v0.2 secrets (openssl S/MIME) to age",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := secret.Migrate(a.log, cmd.OutOrStdout(), a.projDir, a.masker.Register); err != nil {
				a.log.Error(err.Error())
				return &exitError{code: run.ExitError, err: err}
			}
			return nil
		},
	}
}

// resolvePassword はパスワードを引数 → 対話入力の順で解決する。
// 端末から実行された場合は非エコーで入力を受け付け、パイプ・リダイレクトの
// 場合は標準入力から 1 行読み込む。
func resolvePassword(cmd *cobra.Command, args []string) (string, error) {
	if len(args) == 3 {
		return args[2], nil
	}

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		fmt.Fprint(cmd.ErrOrStderr(), "password: ")
		raw, err := term.ReadPassword(fd)
		fmt.Fprintln(cmd.ErrOrStderr())
		if err != nil {
			return "", fmt.Errorf("password input: %w", err)
		}
		return string(raw), nil
	}

	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && line == "" {
		return "", fmt.Errorf("password input: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}
