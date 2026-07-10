package repository

import (
	"strings"
	"testing"
)

func TestMySQLBatchTSVToCSV(t *testing.T) {
	t.Run("MySQLBatchTSVToCSV_特殊文字を含むbatch出力の場合_CSVに正しく変換されること", func(t *testing.T) {
		// Arrange
		// mysql --batch 出力を模した入力 (TAB 区切り・LF 行区切り・エスケープ済み)。
		// 列: id, name, note
		//  1  Alice  hello                 通常値
		//  2  <empty> NULL                 空文字 と SQL NULL
		//  3  a,b    "q"                    カンマ・ダブルクオート (要 CSV クオート)
		//  4  x\ty   line1\nline2           エスケープされた TAB・改行 (復元して CSV クオート)
		//  5  back\\slash  NULL             エスケープされたバックスラッシュ
		input := "id\tname\tnote\n" +
			"1\tAlice\thello\n" +
			"2\t\tNULL\n" +
			"3\ta,b\t\"q\"\n" +
			"4\tx\\ty\tline1\\nline2\n" +
			"5\tback\\\\slash\tNULL\n"
		var out strings.Builder

		// Act
		err := MySQLBatchTSVToCSV(strings.NewReader(input), &out)

		// Assert
		if err != nil {
			t.Fatalf("MySQLBatchTSVToCSV: %v", err)
		}
		got := out.String()
		want := "id,name,note\n" +
			"1,Alice,hello\n" +
			"2,,\\N\n" + // 空文字は空フィールド、NULL は \N
			"3,\"a,b\",\"\"\"q\"\"\"\n" + // カンマ・クオートはクオート + クオート二重化
			"4,x\ty,\"line1\nline2\"\n" + // TAB は CSV 上通常文字 (非クオート)、改行はクオート
			"5,back\\slash,\\N\n" // \\ を \ に復元、NULL は \N
		if got != want {
			t.Errorf("CSV mismatch:\n got=%q\nwant=%q", got, want)
		}
	})
}

func TestMySQLBatchTSVToCSVEmptyInput(t *testing.T) {
	t.Run("MySQLBatchTSVToCSV_空入力の場合_空出力になること", func(t *testing.T) {
		// Arrange
		var out strings.Builder

		// Act
		err := MySQLBatchTSVToCSV(strings.NewReader(""), &out)

		// Assert
		if err != nil {
			t.Fatalf("MySQLBatchTSVToCSV: %v", err)
		}
		if got := out.String(); got != "" {
			t.Errorf("expected empty output, got %q", got)
		}
	})
}

// ヘッダー行は NULL 判定を行わない (列名 "NULL" はそのまま)。
func TestMySQLBatchTSVToCSVHeaderNotNullMapped(t *testing.T) {
	t.Run("MySQLBatchTSVToCSV_ヘッダーに列名NULLがある場合_ヘッダーはNULL判定されないこと", func(t *testing.T) {
		// Arrange
		input := "NULL\tval\n" + "NULL\tNULL\n"
		var out strings.Builder

		// Act
		err := MySQLBatchTSVToCSV(strings.NewReader(input), &out)

		// Assert
		if err != nil {
			t.Fatalf("MySQLBatchTSVToCSV: %v", err)
		}
		want := "NULL,val\n" + "\\N,\\N\n" // ヘッダーの "NULL" は列名、データ行は NULL
		if got := out.String(); got != want {
			t.Errorf("mismatch:\n got=%q\nwant=%q", got, want)
		}
	})
}

func TestUnescapeMySQLBatch(t *testing.T) {
	t.Run("unescapeMySQLBatch_各種エスケープ列の場合_正しくアンエスケープされること", func(t *testing.T) {
		// Arrange
		cases := map[string]string{
			"plain":        "plain",
			`a\tb`:         "a\tb",
			`a\nb`:         "a\nb",
			`a\\b`:         `a\b`,
			`a\0b`:         "a\x00b",
			`no backslash`: "no backslash",
			`trailing\`:    `trailing\`, // 末尾の孤立バックスラッシュは保持
			`\z`:           `\z`,        // 未知エスケープはバックスラッシュ保持
		}
		for in, want := range cases {
			// Act
			got := unescapeMySQLBatch(in)
			// Assert
			if got != want {
				t.Errorf("unescapeMySQLBatch(%q) = %q, want %q", in, got, want)
			}
		}
	})
}
