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
		{"新規生成は許可", false, false, nil},
		{"既存キーは抑止", true, false, ErrKeyAlreadyExists},
		{"既存キーも force で許可", true, true, nil},
		{"force は新規生成にも影響しない", false, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeygen(tt.keyExists, tt.force)
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
		{"新規登録は許可", false, false, nil},
		{"重複登録は禁止", true, false, ErrSecretAlreadyExists},
		{"重複登録も force で許可", true, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretSave(tt.entryExists, tt.force)
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
		{"基本形", "127.0.0.1", "some_user", "127.0.0.1-some_user", false},
		{"コロンは _ へ置換 (v0.2 互換)", "host:22", "user:x", "host_22-user_x", false},
		{"host 空は拒否", "", "user", "", true},
		{"user 空は拒否", "host", "", "", true},
		{"パス区切りは拒否", "../etc", "user", "", true},
		{"バックスラッシュは拒否", "host", "a\\b", "", true},
		{"相対参照は拒否", "..", "user", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SecretFileName(tt.host, tt.user)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SecretFileName(%q, %q) error = %v, wantErr %v", tt.host, tt.user, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("SecretFileName(%q, %q) = %q, want %q", tt.host, tt.user, got, tt.want)
			}
		})
	}
}
