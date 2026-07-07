package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeInventory(t *testing.T, projDir, fileName, content string) {
	t.Helper()
	dir := filepath.Join(projDir, "config", "inventory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadInventory(t *testing.T) {
	projDir := t.TempDir()
	writeInventory(t, projDir, "staging.yml", `stfw_inventory:
  - web:
    - 127.0.0.1
    - localhost
  - ap:
    - 127.0.0.1
  - db:
    - localhost
`)

	groups, err := LoadInventory(projDir, "staging.yml")
	if err != nil {
		t.Fatalf("LoadInventory: %v", err)
	}
	want := map[string][]string{
		"web": {"127.0.0.1", "localhost"},
		"ap":  {"127.0.0.1"},
		"db":  {"localhost"},
	}
	if !reflect.DeepEqual(groups, want) {
		t.Errorf("LoadInventory = %v, want %v", groups, want)
	}
}

func TestLoadInventoryMergesDuplicateGroups(t *testing.T) {
	projDir := t.TempDir()
	writeInventory(t, projDir, "staging.yml", `stfw_inventory:
  - web:
    - host1
  - web:
    - host2
`)

	groups, err := LoadInventory(projDir, "staging.yml")
	if err != nil {
		t.Fatalf("LoadInventory: %v", err)
	}
	if !reflect.DeepEqual(groups["web"], []string{"host1", "host2"}) {
		t.Errorf("groups[web] = %v, want [host1 host2]", groups["web"])
	}
}

func TestLoadInventoryFileNotFound(t *testing.T) {
	projDir := t.TempDir()
	if _, err := LoadInventory(projDir, "missing.yml"); err == nil {
		t.Error("LoadInventory(missing) = nil, want error")
	}
}
