// Package vault implements LOLA OS's encrypted private key storage.
//
// Design:
//   - A user-supplied passphrase is run through scrypt to derive a 32-byte
//     key encryption key (KEK). scrypt parameters (N, r, p) and a random
//     16-byte salt are stored alongside the ciphertext so the vault file
//     is self-describing and portable.
//   - Each secret (private key, mnemonic, etc.) is encrypted with
//     AES-256-GCM under the KEK. GCM gives us authenticated encryption,
//     so tampering with the vault file is detected on decrypt.
//   - The vault file on disk is a single JSON document containing the KDF
//     parameters, salt, and a map of named entries to base64 nonce+ciphertext.
//   - Keys are never written to disk, logs, or stdout in plaintext. The
//     only place plaintext exists is in memory for the duration of a
//     signing operation.
//
// Only standard, well-vetted primitives are used: crypto/aes, crypto/cipher
// (GCM), crypto/rand, and golang.org/x/crypto/scrypt. No custom crypto.
package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"golang.org/x/crypto/scrypt"
)

var (
	ErrEntryNotFound  = errors.New("vault: entry not found")
	ErrWrongPassword  = errors.New("vault: incorrect passphrase or corrupted vault")
	ErrEntryExists    = errors.New("vault: entry already exists (use Update to overwrite)")
)

const keyLen = 32 // AES-256

// fileFormat is the on-disk JSON structure.
type fileFormat struct {
	Version int               `json:"version"`
	Salt    string            `json:"salt"` // base64
	N       int               `json:"n"`
	R       int               `json:"r"`
	P       int               `json:"p"`
	Entries map[string]string `json:"entries"` // name -> base64(nonce||ciphertext)
}

// Params holds the scrypt cost parameters used to derive the KEK.
type Params struct {
	N int
	R int
	P int
}

// DefaultParams returns scrypt parameters offering a strong, interactive-
// login-speed balance of security and performance (~100-300ms on modern
// hardware).
func DefaultParams() Params {
	return Params{N: 1 << 15, R: 8, P: 1}
}

// Vault represents an opened (decrypted-on-demand) vault file. The KEK is
// held in memory only for the lifetime of the Vault value; callers should
// call Close (or let the Vault go out of scope) as soon as possible.
type Vault struct {
	path    string
	kek     []byte
	salt    []byte
	params  Params
	entries map[string]string // name -> base64(nonce||ciphertext)
}

// Exists reports whether a vault file already exists at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Create initializes a brand-new, empty vault at path, protected by
// passphrase. It fails if a vault already exists at that path.
func Create(path string, passphrase string, params Params) (*Vault, error) {
	if Exists(path) {
		return nil, fmt.Errorf("vault: file already exists at %s", path)
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("vault: generating salt: %w", err)
	}
	kek, err := deriveKey(passphrase, salt, params)
	if err != nil {
		return nil, err
	}
	v := &Vault{
		path:    path,
		kek:     kek,
		salt:    salt,
		params:  params,
		entries: map[string]string{},
	}
	if err := v.persist(); err != nil {
		return nil, err
	}
	return v, nil
}

// Open decrypts the vault at path using passphrase. Returns
// ErrWrongPassword if the passphrase is incorrect or the file is corrupted.
func Open(path string, passphrase string) (*Vault, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("vault: reading %s: %w", path, err)
	}
	var ff fileFormat
	if err := json.Unmarshal(raw, &ff); err != nil {
		return nil, fmt.Errorf("vault: parsing vault file: %w", err)
	}
	salt, err := base64.StdEncoding.DecodeString(ff.Salt)
	if err != nil {
		return nil, fmt.Errorf("vault: decoding salt: %w", err)
	}
	params := Params{N: ff.N, R: ff.R, P: ff.P}
	kek, err := deriveKey(passphrase, salt, params)
	if err != nil {
		return nil, err
	}

	v := &Vault{
		path:    path,
		kek:     kek,
		salt:    salt,
		params:  params,
		entries: ff.Entries,
	}

	// Verify the passphrase by attempting to decrypt a canary entry if
	// present, else by validating against any existing real entry.
	if canary, ok := ff.Entries["__canary__"]; ok {
		if _, err := v.decryptB64(canary); err != nil {
			return nil, ErrWrongPassword
		}
	} else {
		for _, enc := range ff.Entries {
			if _, err := v.decryptB64(enc); err != nil {
				return nil, ErrWrongPassword
			}
			break
		}
	}
	return v, nil
}

// OpenOrCreate opens the vault at path if it exists, or creates a new one.
func OpenOrCreate(path, passphrase string) (*Vault, error) {
	if Exists(path) {
		return Open(path, passphrase)
	}
	v, err := Create(path, passphrase, DefaultParams())
	if err != nil {
		return nil, err
	}
	// Plant a canary so future Open() calls can verify the passphrase even
	// before any real secret has been stored.
	if err := v.Set("__canary__", "lola-vault-ok"); err != nil {
		return nil, err
	}
	return v, nil
}

func deriveKey(passphrase string, salt []byte, p Params) ([]byte, error) {
	key, err := scrypt.Key([]byte(passphrase), salt, p.N, p.R, p.P, keyLen)
	if err != nil {
		return nil, fmt.Errorf("vault: deriving key: %w", err)
	}
	return key, nil
}

func (v *Vault) encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(v.kek)
	if err != nil {
		return "", fmt.Errorf("vault: creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("vault: creating GCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("vault: generating nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (v *Vault) decryptB64(b64 string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("vault: decoding entry: %w", err)
	}
	block, err := aes.NewCipher(v.kek)
	if err != nil {
		return nil, fmt.Errorf("vault: creating cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("vault: creating GCM: %w", err)
	}
	if len(data) < gcm.NonceSize() {
		return nil, ErrWrongPassword
	}
	nonce, ct := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrWrongPassword
	}
	return plaintext, nil
}

// Set stores (or overwrites) a named secret in the vault and persists to disk.
func (v *Vault) Set(name, secret string) error {
	enc, err := v.encrypt([]byte(secret))
	if err != nil {
		return err
	}
	v.entries[name] = enc
	return v.persist()
}

// Get decrypts and returns the named secret. The caller is responsible for
// zeroing/clearing the returned bytes/string as soon as it is no longer
// needed (e.g. after signing).
func (v *Vault) Get(name string) (string, error) {
	enc, ok := v.entries[name]
	if !ok {
		return "", ErrEntryNotFound
	}
	plain, err := v.decryptB64(enc)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Delete removes a named secret from the vault.
func (v *Vault) Delete(name string) error {
	if _, ok := v.entries[name]; !ok {
		return ErrEntryNotFound
	}
	delete(v.entries, name)
	return v.persist()
}

// List returns the names of all stored entries (never their values),
// excluding the internal canary entry.
func (v *Vault) List() []string {
	names := make([]string, 0, len(v.entries))
	for name := range v.entries {
		if name == "__canary__" {
			continue
		}
		names = append(names, name)
	}
	return names
}

// VerifyIntegrity decrypts the canary (or, if absent, any one entry) and
// returns nil if the vault is readable and uncorrupted. Used by `lola doctor`.
func (v *Vault) VerifyIntegrity() error {
	if canary, ok := v.entries["__canary__"]; ok {
		plain, err := v.decryptB64(canary)
		if err != nil {
			return err
		}
		if string(plain) != "lola-vault-ok" {
			return ErrWrongPassword
		}
		return nil
	}
	for _, enc := range v.entries {
		_, err := v.decryptB64(enc)
		return err
	}
	return nil // empty vault, nothing to verify
}

func (v *Vault) persist() error {
	ff := fileFormat{
		Version: 1,
		Salt:    base64.StdEncoding.EncodeToString(v.salt),
		N:       v.params.N,
		R:       v.params.R,
		P:       v.params.P,
		Entries: v.entries,
	}
	data, err := json.MarshalIndent(ff, "", "  ")
	if err != nil {
		return fmt.Errorf("vault: marshaling: %w", err)
	}
	// 0600: owner read/write only. Keys never touch disk in plaintext.
	return os.WriteFile(v.path, data, 0o600)
}

// Close clears the in-memory KEK. The Vault must not be used after Close.
func (v *Vault) Close() {
	for i := range v.kek {
		v.kek[i] = 0
	}
	v.kek = nil
}
