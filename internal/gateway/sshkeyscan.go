package gateway

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KnownHostsPath は ~/.ssh/known_hosts のパスを返す。
func KnownHostsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("known_hosts: %w", err)
	}
	return filepath.Join(home, ".ssh", "known_hosts"), nil
}

// KnownHostsContains は known_hosts に host のエントリが存在するかを判定する。
// v0.2 の `grep "${_ip}" known_hosts` と同じ部分一致で判定する。
// ファイルが無い場合は false を返す。
func KnownHostsContains(path, host string) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.Contains(line, host) {
			return true, nil
		}
	}
	return false, nil
}

// RemoveKnownHostKey は `ssh-keygen -R <host>` で既存のサーバキーを削除する。
// 出力は破棄する (v0.2 の gen_ssh_server_key と同じ)。
func RemoveKnownHostKey(host string) error {
	cmd := exec.Command("ssh-keygen", "-R", host)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete SSH server key for %s. cmd: ssh-keygen -R %s: %w", host, host, err)
	}
	return nil
}

// ScanHostKey は `ssh-keyscan <host>` の出力を knownHostsPath へ追記する。
// 標準エラーは破棄する (v0.2 の gen_ssh_server_key と同じ)。
func ScanHostKey(host, knownHostsPath string) error {
	var stdout bytes.Buffer
	cmd := exec.Command("ssh-keyscan", host)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add SSH server key for %s. cmd: ssh-keyscan %s: %w", host, host, err)
	}

	if err := os.MkdirAll(filepath.Dir(knownHostsPath), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(stdout.Bytes()); err != nil {
		return fmt.Errorf("failed to add SSH server key for %s: %w", host, err)
	}
	return nil
}
