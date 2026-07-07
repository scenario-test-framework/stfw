package repository

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/smallstep/pkcs7"
)

// 暗号化キーの配置 (v0.2 の passwd_spec と同じ config/encrypt/ 配下)。
// v1.0 では age (X25519) キーペアを key.txt / key.txt.pub として保存する。
const (
	secretKeyDirName     = "encrypt"
	ageKeyFileName       = "key.txt"
	ageRecipientFileName = "key.txt.pub"

	// v0.2 の openssl RSA キーペア (migrate の読込にのみ使用)
	legacyDecryptKeyName = "decrypt_key"
	legacyEncryptKeyName = "encrypt_key"

	secretDirName = "passwd"

	// legacyPKCS7Header は v0.2 の openssl smime -outform PEM 出力の先頭行。
	legacyPKCS7Header = "-----BEGIN PKCS7-----"
)

// SecretKeyDir は暗号化キーディレクトリのパスを返す。
func SecretKeyDir(projDir string) string {
	return filepath.Join(projDir, "config", secretKeyDirName)
}

// AgeKeyPath は age 秘密鍵ファイルのパスを返す。
func AgeKeyPath(projDir string) string {
	return filepath.Join(SecretKeyDir(projDir), ageKeyFileName)
}

// AgeRecipientPath は age 受信者公開鍵ファイルのパスを返す。
func AgeRecipientPath(projDir string) string {
	return filepath.Join(SecretKeyDir(projDir), ageRecipientFileName)
}

// AgeKeyExists は age 秘密鍵の存在を判定する。
// v0.2 はキーディレクトリの存在で判定していたが、v1.0 では migrate 時に
// 旧キーと共存させるため鍵ファイル単位で判定する。
func AgeKeyExists(projDir string) bool {
	_, err := os.Stat(AgeKeyPath(projDir))
	return err == nil
}

// GenerateAgeKey は age (X25519) キーペアを生成して保存し、受信者公開鍵を返す。
// 秘密鍵 key.txt は 0600、公開鍵 key.txt.pub は 0644 で保存する。
func GenerateAgeKey(projDir string) (string, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", fmt.Errorf("generate age identity: %w", err)
	}
	recipient := identity.Recipient().String()

	if err := os.MkdirAll(SecretKeyDir(projDir), 0o755); err != nil {
		return "", err
	}

	// age-keygen と同じレイアウト (コメント行 + 秘密鍵 1 行)
	key := fmt.Sprintf("# created: %s\n# public key: %s\n%s\n",
		time.Now().Format(time.RFC3339), recipient, identity.String())
	if err := os.WriteFile(AgeKeyPath(projDir), []byte(key), 0o600); err != nil {
		return "", err
	}
	if err := os.WriteFile(AgeRecipientPath(projDir), []byte(recipient+"\n"), 0o644); err != nil {
		return "", err
	}
	return recipient, nil
}

// SecretDir はシークレットファイルディレクトリのパスを返す。
func SecretDir(projDir string) string {
	return filepath.Join(projDir, "config", secretDirName)
}

// SecretPath はシークレットファイル (config/passwd/{host}-{user}) のパスを返す。
func SecretPath(projDir, name string) string {
	return filepath.Join(SecretDir(projDir), name)
}

// SecretExists はシークレットファイルの存在を判定する。
func SecretExists(projDir, name string) bool {
	_, err := os.Stat(SecretPath(projDir, name))
	return err == nil
}

// SaveSecret はシークレットを age で暗号化して保存する。
// 出力は ASCII armor 形式 (v0.2 の PEM 同様にテキストで扱えるようにする)。
func SaveSecret(projDir, name, secret string) error {
	recipient, err := loadAgeRecipient(projDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(SecretDir(projDir), 0o755); err != nil {
		return err
	}

	var buf bytes.Buffer
	armorWriter := armor.NewWriter(&buf)
	encWriter, err := age.Encrypt(armorWriter, recipient)
	if err != nil {
		return fmt.Errorf("age encrypt: %w", err)
	}
	if _, err := io.WriteString(encWriter, secret); err != nil {
		return fmt.Errorf("age encrypt: %w", err)
	}
	if err := encWriter.Close(); err != nil {
		return fmt.Errorf("age encrypt: %w", err)
	}
	if err := armorWriter.Close(); err != nil {
		return fmt.Errorf("age encrypt: %w", err)
	}
	buf.WriteByte('\n')

	return os.WriteFile(SecretPath(projDir, name), buf.Bytes(), 0o600)
}

// LoadSecret はシークレットを復号して返す。
// v0.2 形式 (S/MIME PEM) のファイルは復号できないため migrate を促す。
func LoadSecret(projDir, name string) (string, error) {
	path := SecretPath(projDir, name)
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if IsLegacySecret(raw) {
		return "", fmt.Errorf("%s is a v0.2 format secret (run `stfw secret migrate` first)", path)
	}

	identities, err := loadAgeIdentities(projDir)
	if err != nil {
		return "", err
	}
	decReader, err := age.Decrypt(armor.NewReader(bytes.NewReader(raw)), identities...)
	if err != nil {
		return "", fmt.Errorf("age decrypt: %s: %w", path, err)
	}
	plain, err := io.ReadAll(decReader)
	if err != nil {
		return "", fmt.Errorf("age decrypt: %s: %w", path, err)
	}
	return string(plain), nil
}

// ListSecretNames は config/passwd/ 配下のシークレット名一覧を昇順で返す。
// 退避済みの *.bak は除外する。ディレクトリが無い場合は空を返す。
func ListSecretNames(projDir string) ([]string, error) {
	entries, err := os.ReadDir(SecretDir(projDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".bak") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// IsLegacySecret は v0.2 形式 (openssl smime の PKCS7 PEM) かを判定する。
func IsLegacySecret(raw []byte) bool {
	return bytes.HasPrefix(bytes.TrimSpace(raw), []byte(legacyPKCS7Header))
}

// ReadSecretFile はシークレットファイルの生データを読み込む。
func ReadSecretFile(projDir, name string) ([]byte, error) {
	return os.ReadFile(SecretPath(projDir, name))
}

// BackupSecretFile はシークレットファイルを {name}.bak へ退避する。
func BackupSecretFile(projDir, name string) error {
	path := SecretPath(projDir, name)
	return os.Rename(path, path+".bak")
}

// DecryptLegacySecret は v0.2 形式 (S/MIME CMS EnvelopedData, PEM) のデータを
// 旧 RSA キーペア (config/encrypt/{encrypt_key,decrypt_key}) で復号する。
// v0.2 の `openssl smime -decrypt -binary -inform PEM -inkey decrypt_key` 相当。
func DecryptLegacySecret(projDir string, raw []byte) (string, error) {
	block, _ := pem.Decode(raw)
	if block == nil {
		return "", fmt.Errorf("legacy secret: PEM decode failed")
	}
	p7, err := pkcs7.Parse(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("legacy secret: pkcs7 parse: %w", err)
	}

	cert, key, err := loadLegacyKeyPair(projDir)
	if err != nil {
		return "", err
	}
	plain, err := p7.Decrypt(cert, key)
	if err != nil {
		return "", fmt.Errorf("legacy secret: pkcs7 decrypt: %w", err)
	}
	return string(plain), nil
}

// loadLegacyKeyPair は v0.2 の openssl キーペアを読み込む
// (encrypt_key = X.509 証明書, decrypt_key = PKCS#8 RSA 秘密鍵)。
func loadLegacyKeyPair(projDir string) (*x509.Certificate, any, error) {
	certPath := filepath.Join(SecretKeyDir(projDir), legacyEncryptKeyName)
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("legacy encrypt key: %w", err)
	}
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, fmt.Errorf("legacy encrypt key: %s: PEM decode failed", certPath)
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("legacy encrypt key: %s: %w", certPath, err)
	}

	keyPath := filepath.Join(SecretKeyDir(projDir), legacyDecryptKeyName)
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("legacy decrypt key: %w", err)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, fmt.Errorf("legacy decrypt key: %s: PEM decode failed", keyPath)
	}
	key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		// openssl のバージョンによっては PKCS#1 (BEGIN RSA PRIVATE KEY) の場合がある
		if rsaKey, rsaErr := x509.ParsePKCS1PrivateKey(keyBlock.Bytes); rsaErr == nil {
			return cert, rsaKey, nil
		}
		return nil, nil, fmt.Errorf("legacy decrypt key: %s: %w", keyPath, err)
	}
	return cert, key, nil
}

// loadAgeRecipient は key.txt.pub から受信者公開鍵を読み込む。
func loadAgeRecipient(projDir string) (age.Recipient, error) {
	raw, err := os.ReadFile(AgeRecipientPath(projDir))
	if err != nil {
		return nil, fmt.Errorf("encrypt key is not generated (run `stfw secret keygen` first): %w", err)
	}
	recipients, err := age.ParseRecipients(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", AgeRecipientPath(projDir), err)
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("%s: no recipients", AgeRecipientPath(projDir))
	}
	return recipients[0], nil
}

// loadAgeIdentities は key.txt から秘密鍵を読み込む。
func loadAgeIdentities(projDir string) ([]age.Identity, error) {
	raw, err := os.ReadFile(AgeKeyPath(projDir))
	if err != nil {
		return nil, fmt.Errorf("encrypt key is not generated (run `stfw secret keygen` first): %w", err)
	}
	identities, err := age.ParseIdentities(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", AgeKeyPath(projDir), err)
	}
	return identities, nil
}
