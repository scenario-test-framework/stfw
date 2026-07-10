package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadNodeMetadata(t *testing.T) {
	t.Run("ReadNodeMetadata_有効なmetadataがある場合_構造体に読めること", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		raw := "description: |\n  line1\n  line2\n\nrequirement_specifications:\n  - SPEC-1\n  - SPEC-2\n"
		if err := os.WriteFile(filepath.Join(dir, metadataFileName), []byte(raw), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		got, err := ReadNodeMetadata(dir)

		// Assert
		if err != nil {
			t.Fatalf("ReadNodeMetadata: %v", err)
		}
		want := Metadata{Description: "line1\nline2\n", RequirementSpecifications: []string{"SPEC-1", "SPEC-2"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ReadNodeMetadata = %#v, want %#v", got, want)
		}
	})
}

// stfw new が生成する空スタブ (metadataContent) は description / requirement_specifications
// のいずれも欠損値であり、ゼロ値として読めなければならない (既存生成物との互換)。
func TestReadNodeMetadataEmptyStub(t *testing.T) {
	t.Run("ReadNodeMetadata_空スタブの場合_ゼロ値になること", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, metadataFileName), []byte(metadataContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		got, err := ReadNodeMetadata(dir)

		// Assert
		if err != nil {
			t.Fatalf("ReadNodeMetadata: %v", err)
		}
		if !reflect.DeepEqual(got, Metadata{}) {
			t.Errorf("ReadNodeMetadata(empty stub) = %#v, want zero value", got)
		}
	})
}

// metadata.yml が存在しないディレクトリはゼロ値・エラー無しで返す。
func TestReadNodeMetadataMissing(t *testing.T) {
	t.Run("ReadNodeMetadata_metadataが存在しない場合_ゼロ値かつエラー無しであること", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()

		// Act
		got, err := ReadNodeMetadata(dir)

		// Assert
		if err != nil {
			t.Fatalf("ReadNodeMetadata: %v", err)
		}
		if !reflect.DeepEqual(got, Metadata{}) {
			t.Errorf("ReadNodeMetadata(missing) = %#v, want zero value", got)
		}
	})
}

func TestWriteNodeMetadataRoundtrip(t *testing.T) {
	t.Run("WriteNodeMetadata_書き込み後に読み戻す場合_同一値になること", func(t *testing.T) {
		// Arrange
		dir := t.TempDir()
		want := Metadata{Description: "line1\nline2\n", RequirementSpecifications: []string{"SPEC-1", "SPEC-2"}}

		// Act
		if err := WriteNodeMetadata(dir, want); err != nil {
			t.Fatalf("WriteNodeMetadata: %v", err)
		}
		got, err := ReadNodeMetadata(dir)

		// Assert
		if err != nil {
			t.Fatalf("ReadNodeMetadata: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("roundtrip = %#v, want %#v", got, want)
		}
	})
}

func TestCreateSpecNode(t *testing.T) {
	t.Run("CreateSpecNode_新規ディレクトリの場合_metadataが書き込まれること", func(t *testing.T) {
		// Arrange
		root := t.TempDir()
		dir := filepath.Join(root, "scenario", "test")
		meta := Metadata{Description: "desc", RequirementSpecifications: []string{"SPEC-1"}}

		// Act
		err := CreateSpecNode(dir, meta)

		// Assert
		if err != nil {
			t.Fatalf("CreateSpecNode: %v", err)
		}
		got, err := ReadNodeMetadata(dir)
		if err != nil {
			t.Fatalf("ReadNodeMetadata: %v", err)
		}
		if !reflect.DeepEqual(got, meta) {
			t.Errorf("CreateSpecNode wrote %#v, want %#v", got, meta)
		}
	})

	t.Run("CreateSpecNode_再実行する場合_既存の葉ファイルを削除しないこと", func(t *testing.T) {
		// Arrange
		root := t.TempDir()
		dir := filepath.Join(root, "scenario", "test")
		meta := Metadata{Description: "desc", RequirementSpecifications: []string{"SPEC-1"}}
		if err := CreateSpecNode(dir, meta); err != nil {
			t.Fatalf("CreateSpecNode: %v", err)
		}
		// 冪等 (再実行で既存の葉ディレクトリを削除しない)。
		leafFile := filepath.Join(dir, "data", "keep.txt")
		if err := os.MkdirAll(filepath.Dir(leafFile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(leafFile, []byte("keep"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Act
		err := CreateSpecNode(dir, Metadata{Description: "updated"})

		// Assert
		if err != nil {
			t.Fatalf("CreateSpecNode (2nd): %v", err)
		}
		if _, err := os.Stat(leafFile); err != nil {
			t.Errorf("leaf file was removed by CreateSpecNode: %v", err)
		}
	})
}
