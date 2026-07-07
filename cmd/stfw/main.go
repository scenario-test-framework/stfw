package main

import (
	"os"

	"github.com/scenario-test-framework/stfw/internal/presentation/cli"
)

func main() {
	os.Exit(cli.Execute())
}
