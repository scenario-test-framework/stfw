// Package project はプロジェクト環境管理 (Supporting BC) のドメインルールを持つ。
package project

import (
	"errors"
	"fmt"
)

// ConfigFileName はプロジェクトルートを識別する設定ファイル名。
// このファイルの存在がプロジェクト初期化済みの判定条件になる。
const ConfigFileName = "stfw.yml"

// DataDirName はプロジェクト内部データディレクトリ名。
const DataDirName = ".stfw"

// ErrAlreadyInitialized は再初期化禁止ルール違反。
var ErrAlreadyInitialized = errors.New("already initialized")

// ValidateInit は初期化可否を判定する。configExists はプロジェクト
// ディレクトリに stfw.yml が存在するかどうか。
func ValidateInit(dir string, configExists bool) error {
	if configExists {
		return fmt.Errorf("%s is %w", dir, ErrAlreadyInitialized)
	}
	return nil
}
