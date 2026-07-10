package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// metadataFileName は各階層に置くメタ情報ファイル名 (v0.2 の FILENAME_META と同じ)。
const metadataFileName = "metadata.yml"

// metadataContent は `stfw new` (空 scaffold) が生成する metadata.yml の初期内容
// (v0.2 の metadata_repository と同じ)。
const metadataContent = "description:\n\nrequirement_specifications:\n\n"

// Metadata は scenario / bizdate / process 階層の metadata.yml の内容。
// `stfw new` は空スタブしか書かないため、これまで参照する側がいなかった
// (description / requirement_specifications は生成専用だった)。doc / spec (tree → doc/spec の
// 投影) がこのフィールドを読む最初の consumer になる。
type Metadata struct {
	Description               string   `yaml:"description"`
	RequirementSpecifications []string `yaml:"requirement_specifications"`
}

// ReadNodeMetadata は dir 直下の metadata.yml を読み取る。
// ファイル不在・空値は許容し、ゼロ値を返す (`stfw new` が生成する空スタブとの互換)。
func ReadNodeMetadata(dir string) (Metadata, error) {
	path := filepath.Join(dir, metadataFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Metadata{}, nil
		}
		return Metadata{}, err
	}
	var meta Metadata
	if err := yaml.Unmarshal(raw, &meta); err != nil {
		return Metadata{}, fmt.Errorf("%s: %w", path, err)
	}
	if len(meta.RequirementSpecifications) == 0 {
		// 空リストと未設定 (nil) を同一視する (往復での spec.RequirementSpecifications の
		// 同一性を保つため。ReadProcessConfigSubtree の空 map 正規化と同じ考え方)。
		meta.RequirementSpecifications = nil
	}
	return meta, nil
}

// WriteNodeMetadata は dir 直下へ metadata.yml を書き出す。
// `stfw new` の空スタブ (metadataContent) とは異なり、spec の description /
// requirement_specifications をそのまま書き出す (`scenario scaffold` の入口)。
func WriteNodeMetadata(dir string, meta Metadata) error {
	raw, err := yaml.Marshal(meta)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, metadataFileName), raw, 0o644)
}

// CreateSpecNode は dir を作成し metadata.yml を書き出す (`scenario scaffold` の骨格生成)。
// 既存ディレクトリの上書きも許容する冪等な操作 (`--force` 再生成時に data/scripts/expect
// 等の葉ディレクトリを巻き込んで消さないよう、削除は一切行わない)。
func CreateSpecNode(dir string, meta Metadata) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return WriteNodeMetadata(dir, meta)
}
