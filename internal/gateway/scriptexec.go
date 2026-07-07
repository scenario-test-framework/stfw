// Package gateway は外部プロセス・ネットワーク等の Driven 側 I/O を提供する。
package gateway

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RunScript はスクリプトを作業ディレクトリ workDir で実行し、終了コードを返す。
// env は os.Environ に追記する KEY=VALUE のリスト。
// プラグイン実行契約 (任意言語スクリプト + リターンコード) に従い、
// スクリプトはファイルとして直接実行する。
func RunScript(workDir, script string, env []string, stdout, stderr io.Writer) (int, error) {
	cmd := exec.Command(script)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return -1, fmt.Errorf("script exec: %s: %w", script, err)
	}
	return 0, nil
}
