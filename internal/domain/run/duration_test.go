package run

import "testing"

func TestElapsedString(t *testing.T) {
	// v0.2 の private.calc_processing_time と同じ HH:MM:SS 形式
	tests := []struct {
		name    string
		startTS string
		endTS   string
		want    string
	}{
		{name: "秒", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-01T12:00:03+09:00", want: "00:00:03"},
		{name: "分秒", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-01T12:01:30+09:00", want: "00:01:30"},
		{name: "同時刻", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-01T12:00:00+09:00", want: "00:00:00"},
		{name: "24時間超は時に加算", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-02T14:00:05+09:00", want: "26:00:05"},
	}
	for _, tt := range tests {
		got, err := ElapsedString(tt.startTS, tt.endTS)
		if err != nil {
			t.Errorf("%s: ElapsedString() error = %v", tt.name, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s: ElapsedString() = %s, want %s", tt.name, got, tt.want)
		}
	}
}

func TestElapsedStringInvalid(t *testing.T) {
	// 不正なタイムスタンプ・逆転した時刻は error を返す (panic しない)
	cases := [][2]string{
		{"invalid", "2020-01-01T12:00:00+09:00"},
		{"2020-01-01T12:00:00+09:00", "invalid"},
		{"2020-01-01T12:00:01+09:00", "2020-01-01T12:00:00+09:00"},
	}
	for _, c := range cases {
		if _, err := ElapsedString(c[0], c[1]); err == nil {
			t.Errorf("ElapsedString(%q, %q) should fail", c[0], c[1])
		}
	}
}
