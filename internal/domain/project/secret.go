package project

import (
	"errors"
	"fmt"
	"strings"
)

// ErrKeyAlreadyExists は暗号化キー再生成の抑止ルール違反 (v0.2 互換の条件)。
var ErrKeyAlreadyExists = errors.New("encrypt key already exists (use --force to regenerate)")

// ErrSecretAlreadyExists はパスワード重複登録禁止ルール違反 (v0.2 互換の条件)。
var ErrSecretAlreadyExists = errors.New("secret already exists (use --force to overwrite)")

// ValidateKeygen は暗号化キー生成可否を判定する。
// 既存キーがある場合は --force 指定時のみ再生成を許可する
// (v0.2 の gen-encrypt-key と同じ抑止条件)。
func ValidateKeygen(keyExists, force bool) error {
	if keyExists && !force {
		return ErrKeyAlreadyExists
	}
	return nil
}

// ValidateSecretSave はシークレット保存可否を判定する。
// 同一ホスト × ユーザーの登録済みエントリは --force 指定時のみ上書きを許可する
// (v0.2 の passwd と同じ重複登録禁止条件)。
func ValidateSecretSave(entryExists, force bool) error {
	if entryExists && !force {
		return ErrSecretAlreadyExists
	}
	return nil
}

// SecretFileName はシークレットファイル名 `{host}-{user}` を導出する。
// `:` は `_` へ置換する (v0.2 の passwd_spec と同じ命名規則)。
// ファイル名として安全でない入力 (空・パス区切り・`..`) は拒否する。
func SecretFileName(host, user string) (string, error) {
	for _, v := range []struct{ name, value string }{
		{"host", host},
		{"user", user},
	} {
		if v.value == "" {
			return "", fmt.Errorf("%s must not be empty", v.name)
		}
		if strings.ContainsAny(v.value, "/\\") || v.value == "." || v.value == ".." {
			return "", fmt.Errorf("%s contains invalid characters: %s", v.name, v.value)
		}
	}
	return strings.ReplaceAll(host, ":", "_") + "-" + strings.ReplaceAll(user, ":", "_"), nil
}
