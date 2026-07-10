package evidence

import "testing"

func TestHostFilePath(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		src     string
		want    string
		wantErr bool
	}{
		{name: "HostFilePath_絶対パスの場合_パスを再現すること", host: "web01", src: "/var/log/app/access.log", want: "evidence/web01/var/log/app/access.log"},
		{name: "HostFilePath_末尾スラッシュの場合_正規化すること", host: "web01", src: "/var/log/", want: "evidence/web01/var/log"},
		{name: "HostFilePath_相対パスの場合_エラーであること", host: "web01", src: "var/log/app.log", wantErr: true},
		{name: "HostFilePath_ルートの場合_エラーであること", host: "web01", src: "/", wantErr: true},
		{name: "HostFilePath_hostに区切りを含む場合_エラーであること", host: "web/01", src: "/var/log/a.log", wantErr: true},
		{name: "HostFilePath_host空の場合_エラーであること", host: "", src: "/var/log/a.log", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got, err := HostFilePath(tc.host, tc.src)
			// Assert
			if tc.wantErr {
				if err == nil {
					t.Fatalf("エラーを期待したが nil (got=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if got != tc.want {
				t.Fatalf("want=%q got=%q", tc.want, got)
			}
		})
	}
}

func TestHostFilePathTraversalStaysUnderEvidence(t *testing.T) {
	t.Run("HostFilePath_パストラバーサルの場合_evidence配下に収まること", func(t *testing.T) {
		// Arrange
		// 絶対パスの Clean は `..` がルートを越えないため、常に evidence/{host}/ 配下に収まる。
		// Act
		got, err := HostFilePath("web01", "/var/log/../../../etc/passwd")
		// Assert
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if got != "evidence/web01/etc/passwd" {
			t.Fatalf("パストラバーサルが evidence 配下に収まっていない: %q", got)
		}
	})
}

func TestDatabaseTablePath(t *testing.T) {
	t.Run("DatabaseTablePath_基本形の場合_csvパスを返すこと", func(t *testing.T) {
		// Act
		got, err := DatabaseTablePath("appdb", "orders")
		// Assert
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if got != "evidence/appdb/orders.csv" {
			t.Fatalf("want=evidence/appdb/orders.csv got=%q", got)
		}
	})

	t.Run("DatabaseTablePath_tableに区切りを含む場合_エラーであること", func(t *testing.T) {
		// Act
		_, err := DatabaseTablePath("appdb", "ord/ers")
		// Assert
		if err == nil {
			t.Fatal("table に区切りを含む場合はエラーを期待")
		}
	})

	t.Run("DatabaseTablePath_database空の場合_エラーであること", func(t *testing.T) {
		// Act
		_, err := DatabaseTablePath("", "orders")
		// Assert
		if err == nil {
			t.Fatal("database が空の場合はエラーを期待")
		}
	})
}
