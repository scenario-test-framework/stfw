// Package secret は stfw secret (keygen / set / show / migrate) のビジネスフローを制御する。
package secret

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Keygen は age 暗号化キーペアを生成する。
// 既存キーの再生成抑止 (--force 必須) は v0.2 の gen-encrypt-key 互換の条件。
func Keygen(log *slog.Logger, out io.Writer, projDir string, force bool) error {
	exists := repository.AgeKeyExists(projDir)
	if err := project.ValidateKeygen(exists, force); err != nil {
		return err
	}

	recipient, err := repository.GenerateAgeKey(projDir)
	if err != nil {
		return fmt.Errorf("keygen: %w", err)
	}

	fmt.Fprintln(out, recipient)
	log.Info("encrypt key generated",
		"key", repository.AgeKeyPath(projDir),
		"recipient", repository.AgeRecipientPath(projDir))
	return nil
}

// Set はシークレットを age で暗号化して config/passwd/{host}-{user} へ保存する。
// 重複登録禁止 (--force 必須) は v0.2 の passwd 互換の条件。
// シークレット値はログへ出力しない。
func Set(log *slog.Logger, projDir, host, user, secret string, force bool) error {
	if secret == "" {
		return fmt.Errorf("password must not be empty")
	}
	name, err := project.SecretFileName(host, user)
	if err != nil {
		return err
	}

	exists := repository.SecretExists(projDir, name)
	if err := project.ValidateSecretSave(exists, force); err != nil {
		return fmt.Errorf("%s: %w", repository.SecretPath(projDir, name), err)
	}

	if err := repository.SaveSecret(projDir, name, secret); err != nil {
		return fmt.Errorf("secret save: %w", err)
	}
	log.Info("secret saved", "file", repository.SecretPath(projDir, name))
	return nil
}

// Show はシークレットを復号して標準出力へ表示する (v0.2 の passwd --show 相当)。
// show は意図的な表示のため、復号値を Masker には登録しない。
func Show(out io.Writer, projDir, host, user string) error {
	name, err := project.SecretFileName(host, user)
	if err != nil {
		return err
	}
	plain, err := repository.LoadSecret(projDir, name)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, plain)
	return nil
}

// Migrate は config/passwd/ 配下の v0.2 形式 (S/MIME PEM) ファイルを
// 旧 RSA キーペアで復号し、age で再暗号化する。変換元は {name}.bak へ退避する。
// 復号値は register (Masker への登録) を経由させ、ログへの漏洩を防ぐ。
func Migrate(log *slog.Logger, out io.Writer, projDir string, register func(secret string)) error {
	if !repository.AgeKeyExists(projDir) {
		return fmt.Errorf("age encrypt key is not generated (run `stfw secret keygen` first)")
	}

	names, err := repository.ListSecretNames(projDir)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	migrated := 0
	for _, name := range names {
		raw, err := repository.ReadSecretFile(projDir, name)
		if err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
		if !repository.IsLegacySecret(raw) {
			log.Debug("skip (not a v0.2 format secret)", "file", name)
			continue
		}

		plain, err := repository.DecryptLegacySecret(projDir, raw)
		if err != nil {
			return fmt.Errorf("migrate: %s: %w", name, err)
		}
		// v0.2 は `echo | openssl smime -encrypt` で保存するため末尾に改行が付く。
		// 再暗号化ではシークレット本体のみを保存する。
		plain = strings.TrimSuffix(plain, "\n")
		register(plain)

		if err := repository.BackupSecretFile(projDir, name); err != nil {
			return fmt.Errorf("migrate: %s: %w", name, err)
		}
		if err := repository.SaveSecret(projDir, name, plain); err != nil {
			return fmt.Errorf("migrate: %s: %w", name, err)
		}
		fmt.Fprintln(out, name)
		log.Info("secret migrated", "file", name, "backup", name+".bak")
		migrated++
	}

	log.Info("migrate completed", "migrated", migrated, "total", len(names))
	return nil
}
