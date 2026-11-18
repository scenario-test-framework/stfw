package repository

import (
	"os"
	"reflect"
	"testing"
)

func TestListRunIDsOrder(t *testing.T) {
	t.Run("ListRunIDs_同一採番時刻でpidの桁数が異なる場合_pidの数値順であること", func(t *testing.T) {
		// Arrange
		// CI で発生した組み合わせ。文字列順では 9855 が 10006 より後になる
		projDir := t.TempDir()
		for _, id := range []string{"_20261117235333_10006", "_20261117235334_7", "_20261117235333_9855"} {
			if err := os.MkdirAll(RunDir(projDir, id), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// Act
		ids, err := ListRunIDs(projDir)
		// Assert
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"_20261117235333_9855", "_20261117235333_10006", "_20261117235334_7"}
		if !reflect.DeepEqual(ids, want) {
			t.Errorf("ListRunIDs() = %v, want %v", ids, want)
		}
	})

	t.Run("LatestRunID_同一採番時刻でpidの桁数が異なる場合_pidの数値が大きい方であること", func(t *testing.T) {
		// Arrange
		projDir := t.TempDir()
		for _, id := range []string{"_20261117235333_9855", "_20261117235333_10006"} {
			if err := os.MkdirAll(RunDir(projDir, id), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		// Act
		latest, err := LatestRunID(projDir)
		// Assert
		if err != nil {
			t.Fatal(err)
		}
		if latest != "_20261117235333_10006" {
			t.Errorf("LatestRunID() = %s, want _20261117235333_10006", latest)
		}
	})
}
