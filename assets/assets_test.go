package assets

import (
	"io/fs"
	"regexp"
	"testing"
)

// TestPlugins は同梱プラグインのスクリプトが契約上の禁止事項を含まないことを確かめる。
func TestPlugins(t *testing.T) {
	t.Run("Plugins_同梱スクリプトの場合_sshpassにパスワードをargvで渡さないこと", func(t *testing.T) {
		// Arrange
		// sshpass -p はパスワードをコマンドの引数に載せ、同じホストの他の利用者から ps で見える。
		// 秘匿値は env 経由 (SSHPASS + sshpass -e) で渡す (CLAUDE.md 実装規約・#54)。
		argvPassword := regexp.MustCompile(`sshpass\s+-p`)
		var scanned int

		// Act
		err := fs.WalkDir(Plugins, "plugins", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			b, err := fs.ReadFile(Plugins, path)
			if err != nil {
				return err
			}
			scanned++
			if argvPassword.Match(b) {
				t.Errorf("%s passes the SSH password on the command line (sshpass -p); export SSHPASS and use sshpass -e", path)
			}
			return nil
		})

		// Assert
		if err != nil {
			t.Fatalf("walk embedded plugins: %v", err)
		}
		if scanned == 0 {
			t.Fatal("no embedded plugin files were scanned")
		}
	})
}
