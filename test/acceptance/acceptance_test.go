// Package acceptance は integration_test.sh (v0.2) を翻訳した受け入れテスト。
// testscript で実際の stfw CLI をエンドツーエンドに検証する。
package acceptance

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/scenario-test-framework/stfw/internal/presentation/cli"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"stfw": cli.Execute,
	}))
}

func TestAcceptance(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}
