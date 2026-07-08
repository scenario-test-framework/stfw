package repository

import (
	"strings"
	"testing"
)

func encodeRow(t *testing.T, key, typ, ttl, rawJSON string) string {
	t.Helper()
	var b strings.Builder
	if err := RedisEncodeRow(&b, key, typ, ttl, []byte(rawJSON)); err != nil {
		t.Fatalf("RedisEncodeRow(%s): %v", typ, err)
	}
	return b.String()
}

func TestRedisEncodeRow(t *testing.T) {
	cases := []struct {
		name, key, typ, ttl, raw, want string
	}{
		// string は生値をそのまま value 列へ。
		{"string", "k", "string", "-1", `"hello"`, "k,string,-1,hello\n"},
		// list は順序保持の JSON 配列。CSV クオートされる。
		{"list", "k", "list", "-1", `["a","b"]`, "k,list,-1,\"[\"\"a\"\",\"\"b\"\"]\"\n"},
		// set は要素ソート後の JSON 配列。
		{"set", "k", "set", "-1", `["y","x"]`, "k,set,-1,\"[\"\"x\"\",\"\"y\"\"]\"\n"},
		// hash は field ソート後の JSON オブジェクト (json.Marshal がキー昇順)。
		{"hash", "k", "hash", "10", `["f2","v2","f1","v1"]`, "k,hash,10,\"{\"\"f1\"\":\"\"v1\"\",\"\"f2\"\":\"\"v2\"\"}\"\n"},
		// zset は member ソート後の [[m,s],...]。
		{"zset", "k", "zset", "-1", `["m2","2","m1","1"]`, "k,zset,-1,\"[[\"\"m1\"\",\"\"1\"\"],[\"\"m2\"\",\"\"2\"\"]]\"\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := encodeRow(t, c.key, c.typ, c.ttl, c.raw)
			if got != c.want {
				t.Errorf("encode %s:\n got=%q\nwant=%q", c.name, got, c.want)
			}
		})
	}
}

// 値にカンマ・改行を含む string も CSV クオートで壊れない。
func TestRedisEncodeRowStringSpecial(t *testing.T) {
	got := encodeRow(t, "k", "string", "-1", `"a,b\nc"`)
	want := "k,string,-1,\"a,b\nc\"\n"
	if got != want {
		t.Errorf("got=%q want=%q", got, want)
	}
}

func TestRedisDecode(t *testing.T) {
	csvIn := "key,type,ttl,value\n" +
		"s,string,-1,hello\n" +
		"l,list,-1,\"[\"\"a\"\",\"\"b\"\"]\"\n" +
		"st,set,-1,\"[\"\"x\"\",\"\"y\"\"]\"\n" +
		"h,hash,100,\"{\"\"f1\"\":\"\"v1\"\",\"\"f2\"\":\"\"v2\"\"}\"\n" +
		"z,zset,-1,\"[[\"\"m1\"\",\"\"1\"\"],[\"\"m2\"\",\"\"2\"\"]]\"\n"

	var out strings.Builder
	if err := RedisDecode(strings.NewReader(csvIn), &out); err != nil {
		t.Fatalf("RedisDecode: %v", err)
	}
	got := out.String()

	want := "DEL \"s\"\n" + "SET \"s\" \"hello\"\n" +
		"DEL \"l\"\n" + "RPUSH \"l\" \"a\" \"b\"\n" +
		"DEL \"st\"\n" + "SADD \"st\" \"x\" \"y\"\n" +
		"DEL \"h\"\n" + "HSET \"h\" \"f1\" \"v1\" \"f2\" \"v2\"\n" + "EXPIRE \"h\" 100\n" +
		"DEL \"z\"\n" + "ZADD \"z\" \"1\" \"m1\" \"2\" \"m2\"\n"

	if got != want {
		t.Errorf("decode:\n got=%q\nwant=%q", got, want)
	}
}

func TestRedisQuoteBinary(t *testing.T) {
	// 非表示文字は \xNN、" と \ はエスケープ。
	got := redisQuote("a\"b\\c\n")
	want := `"a\"b\\c\x0a"`
	if got != want {
		t.Errorf("redisQuote got=%q want=%q", got, want)
	}
}

func TestTTLSeconds(t *testing.T) {
	cases := map[string]int{"-1": 0, "-2": 0, "0": 0, "100": 100, "": 0, "abc": 0, "12x": 0}
	for in, want := range cases {
		if got := ttlSeconds(in); got != want {
			t.Errorf("ttlSeconds(%q)=%d want %d", in, got, want)
		}
	}
}
