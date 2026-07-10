package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadNodeMetadata(t *testing.T) {
	dir := t.TempDir()
	raw := "description: |\n  line1\n  line2\n\nrequirement_specifications:\n  - SPEC-1\n  - SPEC-2\n"
	if err := os.WriteFile(filepath.Join(dir, metadataFileName), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadNodeMetadata(dir)
	if err != nil {
		t.Fatalf("ReadNodeMetadata: %v", err)
	}
	want := Metadata{Description: "line1\nline2\n", RequirementSpecifications: []string{"SPEC-1", "SPEC-2"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ReadNodeMetadata = %#v, want %#v", got, want)
	}
}

// stfw new が生成する空スタブ (metadataContent) は description / requirement_specifications
// のいずれも欠損値であり、ゼロ値として読めなければならない (既存生成物との互換)。
func TestReadNodeMetadataEmptyStub(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, metadataFileName), []byte(metadataContent), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadNodeMetadata(dir)
	if err != nil {
		t.Fatalf("ReadNodeMetadata: %v", err)
	}
	if !reflect.DeepEqual(got, Metadata{}) {
		t.Errorf("ReadNodeMetadata(empty stub) = %#v, want zero value", got)
	}
}

// metadata.yml が存在しないディレクトリはゼロ値・エラー無しで返す。
func TestReadNodeMetadataMissing(t *testing.T) {
	dir := t.TempDir()

	got, err := ReadNodeMetadata(dir)
	if err != nil {
		t.Fatalf("ReadNodeMetadata: %v", err)
	}
	if !reflect.DeepEqual(got, Metadata{}) {
		t.Errorf("ReadNodeMetadata(missing) = %#v, want zero value", got)
	}
}

func TestWriteNodeMetadataRoundtrip(t *testing.T) {
	dir := t.TempDir()
	want := Metadata{Description: "line1\nline2\n", RequirementSpecifications: []string{"SPEC-1", "SPEC-2"}}

	if err := WriteNodeMetadata(dir, want); err != nil {
		t.Fatalf("WriteNodeMetadata: %v", err)
	}
	got, err := ReadNodeMetadata(dir)
	if err != nil {
		t.Fatalf("ReadNodeMetadata: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("roundtrip = %#v, want %#v", got, want)
	}
}

func TestCreateSpecNode(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "scenario", "test")
	meta := Metadata{Description: "desc", RequirementSpecifications: []string{"SPEC-1"}}

	if err := CreateSpecNode(dir, meta); err != nil {
		t.Fatalf("CreateSpecNode: %v", err)
	}
	got, err := ReadNodeMetadata(dir)
	if err != nil {
		t.Fatalf("ReadNodeMetadata: %v", err)
	}
	if !reflect.DeepEqual(got, meta) {
		t.Errorf("CreateSpecNode wrote %#v, want %#v", got, meta)
	}

	// 冪等 (再実行で既存の葉ディレクトリを削除しない)。
	leafFile := filepath.Join(dir, "data", "keep.txt")
	if err := os.MkdirAll(filepath.Dir(leafFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(leafFile, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CreateSpecNode(dir, Metadata{Description: "updated"}); err != nil {
		t.Fatalf("CreateSpecNode (2nd): %v", err)
	}
	if _, err := os.Stat(leafFile); err != nil {
		t.Errorf("leaf file was removed by CreateSpecNode: %v", err)
	}
}
