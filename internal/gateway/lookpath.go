package gateway

import "os/exec"

// CommandExists はコマンドが PATH 上に存在するかを判定する。
// プラグインのランタイム依存 (plugin.yml / 組み込みレジストリの requires) を
// 実行前に検出する静的検証で使う。
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
