package project

import (
	"errors"
	"testing"
)

func TestValidateKeygen(t *testing.T) {
	tests := []struct {
		name      string
		keyExists bool
		force     bool
		wantErr   error
	}{
		{"ValidateKeygen_新規生成の場合_許可されること", false, false, nil},
		{"ValidateKeygen_既存キーの場合_抑止されること", true, false, ErrKeyAlreadyExists},
		{"ValidateKeygen_既存キーでforceの場合_許可されること", true, true, nil},
		{"ValidateKeygen_新規生成でforceの場合_許可されること", false, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidateKeygen(tt.keyExists, tt.force)
			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateKeygen(%v, %v) = %v, want %v", tt.keyExists, tt.force, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSecretSave(t *testing.T) {
	tests := []struct {
		name        string
		entryExists bool
		force       bool
		wantErr     error
	}{
		{"ValidateSecretSave_新規登録の場合_許可されること", false, false, nil},
		{"ValidateSecretSave_重複登録の場合_禁止されること", true, false, ErrSecretAlreadyExists},
		{"ValidateSecretSave_重複登録でforceの場合_許可されること", true, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := ValidateSecretSave(tt.entryExists, tt.force)
			// Assert
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateSecretSave(%v, %v) = %v, want %v", tt.entryExists, tt.force, err, tt.wantErr)
			}
		})
	}
}

func TestSecretFileName(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		user    string
		want    string
		wantErr bool
	}{
		{"SecretFileName_基本形の場合_host-user形式を返すこと", "127.0.0.1", "some_user", "127.0.0.1-some_user", false},
		{"SecretFileName_コロンを含む場合_アンダースコアへ置換すること", "host:22", "user:x", "host_22-user_x", false},
		{"SecretFileName_host空の場合_拒否すること", "", "user", "", true},
		{"SecretFileName_user空の場合_拒否すること", "host", "", "", true},
		{"SecretFileName_パス区切りを含む場合_拒否すること", "../etc", "user", "", true},
		{"SecretFileName_バックスラッシュを含む場合_拒否すること", "host", "a\\b", "", true},
		{"SecretFileName_相対参照の場合_拒否すること", "..", "user", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got, err := SecretFileName(tt.host, tt.user)
			// Assert
			if (err != nil) != tt.wantErr {
				t.Fatalf("SecretFileName(%q, %q) error = %v, wantErr %v", tt.host, tt.user, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("SecretFileName(%q, %q) = %q, want %q", tt.host, tt.user, got, tt.want)
			}
		})
	}
}
