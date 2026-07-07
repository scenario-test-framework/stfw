// Package initialize は stfw init のビジネスフローを制御する。
package initialize

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/scenario-test-framework/stfw/internal/domain/project"
	"github.com/scenario-test-framework/stfw/internal/repository"
)

// Run はプロジェクトを初期化する。
// 再初期化禁止 (stfw.yml 既存時はエラー) は v0.2 互換の条件。
func Run(log *slog.Logger, out io.Writer, projDir string) error {
	exists := repository.ProjectConfigExists(projDir, project.ConfigFileName)
	if err := project.ValidateInit(projDir, exists); err != nil {
		return err
	}

	created, err := repository.MaterializeTemplate(projDir)
	if err != nil {
		return fmt.Errorf("template materialize: %w", err)
	}

	for _, rel := range created {
		fmt.Fprintln(out, rel)
	}
	log.Info("initialized", "dir", projDir, "files", len(created))
	return nil
}
