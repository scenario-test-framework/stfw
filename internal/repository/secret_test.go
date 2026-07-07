package repository

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateAgeKeyAndSecretRoundTrip(t *testing.T) {
	projDir := t.TempDir()

	if AgeKeyExists(projDir) {
		t.Fatal("AgeKeyExists = true before keygen")
	}
	recipient, err := GenerateAgeKey(projDir)
	if err != nil {
		t.Fatalf("GenerateAgeKey: %v", err)
	}
	if !strings.HasPrefix(recipient, "age1") {
		t.Errorf("recipient = %q, want age1... prefix", recipient)
	}
	if !AgeKeyExists(projDir) {
		t.Error("AgeKeyExists = false after keygen")
	}

	// 秘密鍵は 0600 で保存される
	info, err := os.Stat(AgeKeyPath(projDir))
	if err != nil {
		t.Fatalf("stat key: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("key perm = %o, want 600", perm)
	}

	// 保存 → 復号のラウンドトリップ
	const secret = "p@ssw0rd!"
	if err := SaveSecret(projDir, "127.0.0.1-some_user", secret); err != nil {
		t.Fatalf("SaveSecret: %v", err)
	}
	got, err := LoadSecret(projDir, "127.0.0.1-some_user")
	if err != nil {
		t.Fatalf("LoadSecret: %v", err)
	}
	if got != secret {
		t.Errorf("LoadSecret = %q, want %q", got, secret)
	}

	// 暗号化ファイルは armor 形式 (テキスト) で、平文を含まない
	raw, err := ReadSecretFile(projDir, "127.0.0.1-some_user")
	if err != nil {
		t.Fatalf("ReadSecretFile: %v", err)
	}
	if !strings.HasPrefix(string(raw), "-----BEGIN AGE ENCRYPTED FILE-----") {
		t.Errorf("secret file is not armored: %q", raw[:40])
	}
	if strings.Contains(string(raw), secret) {
		t.Error("secret file contains plaintext")
	}
}

func TestLoadSecretRejectsLegacyFormat(t *testing.T) {
	projDir := t.TempDir()
	if _, err := GenerateAgeKey(projDir); err != nil {
		t.Fatalf("GenerateAgeKey: %v", err)
	}
	if err := os.MkdirAll(SecretDir(projDir), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := "-----BEGIN PKCS7-----\nMIIB\n-----END PKCS7-----\n"
	if err := os.WriteFile(SecretPath(projDir, "host-user"), []byte(legacy), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSecret(projDir, "host-user")
	if err == nil || !strings.Contains(err.Error(), "stfw secret migrate") {
		t.Errorf("LoadSecret legacy = %v, want migrate guidance error", err)
	}
}

func TestIsLegacySecret(t *testing.T) {
	if !IsLegacySecret([]byte("-----BEGIN PKCS7-----\nAAA\n-----END PKCS7-----\n")) {
		t.Error("IsLegacySecret(PKCS7 PEM) = false, want true")
	}
	if IsLegacySecret([]byte("-----BEGIN AGE ENCRYPTED FILE-----\nAAA\n-----END AGE ENCRYPTED FILE-----\n")) {
		t.Error("IsLegacySecret(age armor) = true, want false")
	}
}

func TestListSecretNames(t *testing.T) {
	projDir := t.TempDir()

	// ディレクトリが無い場合は空
	names, err := ListSecretNames(projDir)
	if err != nil {
		t.Fatalf("ListSecretNames: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("ListSecretNames = %v, want empty", names)
	}

	if err := os.MkdirAll(SecretDir(projDir), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"host2-user", "host1-user", "host1-user.bak"} {
		if err := os.WriteFile(SecretPath(projDir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	names, err = ListSecretNames(projDir)
	if err != nil {
		t.Fatalf("ListSecretNames: %v", err)
	}
	// *.bak は除外され昇順で返る
	want := []string{"host1-user", "host2-user"}
	if len(names) != len(want) || names[0] != want[0] || names[1] != want[1] {
		t.Errorf("ListSecretNames = %v, want %v", names, want)
	}
}

func TestBackupSecretFile(t *testing.T) {
	projDir := t.TempDir()
	if err := os.MkdirAll(SecretDir(projDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(SecretPath(projDir, "host-user"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := BackupSecretFile(projDir, "host-user"); err != nil {
		t.Fatalf("BackupSecretFile: %v", err)
	}
	if _, err := os.Stat(SecretPath(projDir, "host-user")); !os.IsNotExist(err) {
		t.Error("original file still exists")
	}
	if _, err := os.Stat(SecretPath(projDir, "host-user") + ".bak"); err != nil {
		t.Errorf("backup file: %v", err)
	}
}
