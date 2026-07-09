package repository

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Redis の export/import で扱う CSV は 1 行 = 1 キーで列 key,type,ttl,value。
// value は type=string は生値、list/set/hash/zset は「キー順ソートの正規化 JSON」
// (compare の安定性のため)。値の取得は redis-cli -2 --json（RESP2 のフラット配列 +
// JSON エスケープ）を前提とし、UTF-8 テキスト値を対象とする（バイナリ値は対象外）。

// RedisEncodeRow は 1 キー分の生値 (redis-cli -2 --json の出力) を正規化し、
// RFC4180 CSV の 1 行 (key,type,ttl,value) を w へ書き出す。
func RedisEncodeRow(w io.Writer, key, typ, ttl string, rawValueJSON []byte) error {
	value, err := normalizeRedisValue(typ, rawValueJSON)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{key, typ, ttl, value}); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

// normalizeRedisValue は type に応じて redis-cli -2 --json の出力を
// 正規化した value 列へ変換する。
//   - string: JSON 文字列をデコードした生値
//   - list  : JSON 配列 (順序保持) をコンパクト JSON
//   - set   : JSON 配列を要素ソートしたコンパクト JSON
//   - hash  : フラット配列 [f,v,...] を field でソートした JSON オブジェクト
//   - zset  : フラット配列 [m,s,...] を member でソートした [[m,s],...] JSON 配列
func normalizeRedisValue(typ string, raw []byte) (string, error) {
	switch typ {
	case "string":
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return "", fmt.Errorf("redis encode: string value: %w", err)
		}
		return s, nil
	case "list":
		var a []string
		if err := json.Unmarshal(raw, &a); err != nil {
			return "", fmt.Errorf("redis encode: list value: %w", err)
		}
		return compactJSON(a)
	case "set":
		var a []string
		if err := json.Unmarshal(raw, &a); err != nil {
			return "", fmt.Errorf("redis encode: set value: %w", err)
		}
		sort.Strings(a)
		return compactJSON(a)
	case "hash":
		flat, err := decodeFlatArray(raw, "hash")
		if err != nil {
			return "", err
		}
		// フラット [f,v,f,v] → map。json.Marshal は map のキーを昇順で出力するため安定。
		m := make(map[string]string, len(flat)/2)
		for i := 0; i+1 < len(flat); i += 2 {
			m[flat[i]] = flat[i+1]
		}
		return compactJSON(m)
	case "zset":
		flat, err := decodeFlatArray(raw, "zset")
		if err != nil {
			return "", err
		}
		// フラット [m,s,m,s] → [[m,s],...] を member でソート。
		pairs := make([][2]string, 0, len(flat)/2)
		for i := 0; i+1 < len(flat); i += 2 {
			pairs = append(pairs, [2]string{flat[i], flat[i+1]})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i][0] < pairs[j][0] })
		return compactJSON(pairs)
	default:
		return "", fmt.Errorf("redis encode: unsupported type: %q", typ)
	}
}

func decodeFlatArray(raw []byte, what string) ([]string, error) {
	var a []string
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("redis encode: %s value: %w", what, err)
	}
	if len(a)%2 != 0 {
		return nil, fmt.Errorf("redis encode: %s value has odd element count", what)
	}
	return a, nil
}

func compactJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("redis encode: marshal: %w", err)
	}
	return string(b), nil
}

// RedisDecode は export 形式の CSV (ヘッダー付き) を読み、各キーを再現する
// redis-cli コマンド列を w へ書き出す (redis-cli の標準入力へ渡す想定)。
// 既存値を上書きするため各キーへ DEL を先行させ、ttl>0 は EXPIRE を付す。
// 値・キーは redis-cli のダブルクオート形式 (\xNN エスケープ) でバイナリ安全に出力する。
func RedisDecode(r io.Reader, w io.Writer) error {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	rows, err := cr.ReadAll()
	if err != nil {
		return fmt.Errorf("redis decode: read csv: %w", err)
	}
	for i, row := range rows {
		if i == 0 {
			continue // ヘッダー
		}
		if len(row) != 4 {
			return fmt.Errorf("redis decode: row %d: expected 4 columns, got %d", i, len(row))
		}
		key, typ, ttl, value := row[0], row[1], row[2], row[3]
		var b strings.Builder
		b.WriteString("DEL " + redisQuote(key) + "\n")
		if err := writeRestoreCommand(&b, key, typ, value); err != nil {
			return fmt.Errorf("redis decode: row %d: %w", i, err)
		}
		if secs := ttlSeconds(ttl); secs > 0 {
			fmt.Fprintf(&b, "EXPIRE %s %d\n", redisQuote(key), secs)
		}
		if _, err := io.WriteString(w, b.String()); err != nil {
			return err
		}
	}
	return nil
}

func writeRestoreCommand(b *strings.Builder, key, typ, value string) error {
	qk := redisQuote(key)
	switch typ {
	case "string":
		b.WriteString("SET " + qk + " " + redisQuote(value) + "\n")
	case "list":
		var a []string
		if err := json.Unmarshal([]byte(value), &a); err != nil {
			return fmt.Errorf("list value: %w", err)
		}
		if len(a) > 0 {
			b.WriteString("RPUSH " + qk + redisArgs(a) + "\n")
		}
	case "set":
		var a []string
		if err := json.Unmarshal([]byte(value), &a); err != nil {
			return fmt.Errorf("set value: %w", err)
		}
		if len(a) > 0 {
			b.WriteString("SADD " + qk + redisArgs(a) + "\n")
		}
	case "hash":
		var m map[string]string
		if err := json.Unmarshal([]byte(value), &m); err != nil {
			return fmt.Errorf("hash value: %w", err)
		}
		if len(m) > 0 {
			fields := make([]string, 0, len(m))
			for f := range m {
				fields = append(fields, f)
			}
			sort.Strings(fields)
			b.WriteString("HSET " + qk)
			for _, f := range fields {
				b.WriteString(" " + redisQuote(f) + " " + redisQuote(m[f]))
			}
			b.WriteString("\n")
		}
	case "zset":
		var pairs [][2]string
		if err := json.Unmarshal([]byte(value), &pairs); err != nil {
			return fmt.Errorf("zset value: %w", err)
		}
		if len(pairs) > 0 {
			b.WriteString("ZADD " + qk)
			for _, p := range pairs {
				// ZADD score member の順。
				b.WriteString(" " + redisQuote(p[1]) + " " + redisQuote(p[0]))
			}
			b.WriteString("\n")
		}
	default:
		return fmt.Errorf("unsupported type: %q", typ)
	}
	return nil
}

func redisArgs(vals []string) string {
	var b strings.Builder
	for _, v := range vals {
		b.WriteString(" " + redisQuote(v))
	}
	return b.String()
}

// ttlSeconds は TTL 文字列を秒へ変換する。-1 (無期限) / -2 (不在) / 非数値は 0。
func ttlSeconds(ttl string) int {
	n := 0
	neg := false
	s := strings.TrimSpace(ttl)
	if s == "" {
		return 0
	}
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		return 0
	}
	return n
}

// redisQuote は文字列を redis-cli のダブルクオート形式へエスケープする。
// 非表示文字は \xNN、" と \ はバックスラッシュエスケープする (binary-safe)。
func redisQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			b.WriteString(`\"`)
		case c == '\\':
			b.WriteString(`\\`)
		case c >= 0x20 && c < 0x7f:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, `\x%02x`, c)
		}
	}
	b.WriteByte('"')
	return b.String()
}
