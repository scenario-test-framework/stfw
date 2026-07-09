package repository

import (
	"bufio"
	"encoding/csv"
	"io"
	"strings"
)

// mysqlNullToken は mysql クライアントの --batch (非表形式) 出力が SQL NULL を
// 表す文字列。mysql-server client/mysql.cc の safe_put_field は NULL フィールドを
// リテラル "NULL" として出力する (`\N` ではない)。
const mysqlNullToken = "NULL"

// csvNullMarker はエクスポート CSV での NULL 表現。mysqldump の慣行に合わせ、
// 空文字と区別できるよう `\N` を用いる (SPEC-013-01)。
const csvNullMarker = `\N`

// MySQLBatchTSVToCSV は `mysql --batch` (タブ区切り・ヘッダー付き・特殊文字
// エスケープ) の出力を RFC 4180 準拠の CSV へ変換する。
//
// mysql --batch の仕様 (man mysql / client/mysql.cc で確認):
//   - 列区切りは TAB、行区切りは LF。
//   - フィールド内の特殊文字は NEWLINE→`\n`, TAB→`\t`, NUL→`\0`, BACKSLASH→`\\`
//     とエスケープされる (よって TAB/LF はフィールドを跨がない)。
//   - SQL NULL はリテラル "NULL"、空文字は空フィールドとして出力される。
//
// 変換規則:
//   - 1 行目 (ヘッダー) は列名。NULL 判定は行わず un-escape のみ。
//   - データ行の各フィールドが "NULL" (完全一致) なら NULL とみなし `\N` を出力。
//     それ以外は mysql エスケープを復元してから CSV クオートする。
//
// 既知の制約 (mysql クライアント出力の本質的な非可逆性):
//   - SQL NULL と文字列 "NULL" は --batch 出力上どちらも "NULL" となり区別できない。
//     本変換は両者を NULL (`\N`) として扱う。
//   - mysqldump の `\N` 慣行と同様、値そのものが `\N` の文字列は NULL と区別できない。
//     厳密な NULL 忠実性が要る場合はドライバ直結が必要 (本フレームワークのスコープ外)。
func MySQLBatchTSVToCSV(in io.Reader, out io.Writer) error {
	w := csv.NewWriter(out)
	// mysql --batch の行区切りは LF。bufio.Scanner の既定 (ScanLines) は
	// 末尾 CR を除去してしまうため、CR を含む値を壊さないよう自前で LF 分割する。
	br := bufio.NewReader(in)
	header := true
	for {
		line, err := br.ReadString('\n')
		if len(line) > 0 {
			line = strings.TrimSuffix(line, "\n")
			fields := strings.Split(line, "\t")
			record := make([]string, len(fields))
			for i, f := range fields {
				if !header && f == mysqlNullToken {
					record[i] = csvNullMarker
					continue
				}
				record[i] = unescapeMySQLBatch(f)
			}
			if werr := w.Write(record); werr != nil {
				return werr
			}
			header = false
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// unescapeMySQLBatch は mysql --batch のエスケープ (`\n` `\t` `\0` `\\`) を
// 元の文字へ復元する。未知の `\x` はバックスラッシュを保持して x をそのまま残す
// (mysql は上記 4 種以外をエスケープしないため通常発生しない防御的処理)。
func unescapeMySQLBatch(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
			case 't':
				b.WriteByte('\t')
				i++
			case '0':
				b.WriteByte(0)
				i++
			case '\\':
				b.WriteByte('\\')
				i++
			default:
				b.WriteByte('\\')
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
