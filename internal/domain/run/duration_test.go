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
		{name: "ElapsedString_3秒差の場合_00:00:03であること", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-01T12:00:03+09:00", want: "00:00:03"},
		{name: "ElapsedString_1分30秒差の場合_00:01:30であること", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-01T12:01:30+09:00", want: "00:01:30"},
		{name: "ElapsedString_同時刻の場合_00:00:00であること", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-01T12:00:00+09:00", want: "00:00:00"},
		{name: "ElapsedString_24時間超の場合_時に加算されること", startTS: "2020-01-01T12:00:00+09:00", endTS: "2020-01-02T14:00:05+09:00", want: "26:00:05"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := ElapsedString(tt.startTS, tt.endTS)
			// Assert
			if err != nil {
				t.Fatalf("ElapsedString() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ElapsedString() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestElapsedStringInvalid(t *testing.T) {
	t.Run("ElapsedString_不正なタイムスタンプや逆転時刻の場合_エラーであること", func(t *testing.T) {
		// Arrange
		// 不正なタイムスタンプ・逆転した時刻は error を返す (panic しない)
		cases := [][2]string{
			{"invalid", "2020-01-01T12:00:00+09:00"},
			{"2020-01-01T12:00:00+09:00", "invalid"},
			{"2020-01-01T12:00:01+09:00", "2020-01-01T12:00:00+09:00"},
		}
		for _, c := range cases {
			// Act
			_, err := ElapsedString(c[0], c[1])
			// Assert
			if err == nil {
				t.Errorf("ElapsedString(%q, %q) should fail", c[0], c[1])
			}
		}
	})
}
