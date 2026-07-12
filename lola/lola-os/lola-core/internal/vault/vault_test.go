package vault

import (
	"path/filepath"
	"testing"
)

func TestVault_CreateOpenRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v, err := Create(path, "correct-passphrase", DefaultParams())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := v.Set("deployer", "0xprivatekeyhex"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	v.Close()

	v2, err := Open(path, "correct-passphrase")
	if err != nil {
		t.Fatalf("Open with correct passphrase failed: %v", err)
	}
	defer v2.Close()

	secret, err := v2.Get("deployer")
	if err != nil || secret != "0xprivatekeyhex" {
		t.Fatalf("expected to retrieve stored secret, got %q err=%v", secret, err)
	}
}

func TestVault_WrongPassphraseFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v, err := Create(path, "correct-passphrase", DefaultParams())
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	_ = v.Set("k", "v")
	v.Close()

	_, err = Open(path, "wrong-passphrase")
	if err == nil {
		t.Fatalf("expected wrong passphrase to fail")
	}
}

func TestVault_ListNeverExposesValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v, err := Create(path, "pw", DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	_ = v.Set("alice", "secret1")
	_ = v.Set("bob", "secret2")

	names := v.List()
	if len(names) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(names))
	}
	for _, n := range names {
		if n == "secret1" || n == "secret2" {
			t.Fatalf("List leaked a secret value")
		}
	}
}

func TestVault_DeleteRemovesEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v, err := Create(path, "pw", DefaultParams())
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	_ = v.Set("k", "v")

	if err := v.Delete("k"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if _, err := v.Get("k"); err != ErrEntryNotFound {
		t.Fatalf("expected ErrEntryNotFound after delete, got %v", err)
	}
}

func TestVault_VerifyIntegrity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	v, err := OpenOrCreate(path, "pw")
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	if err := v.VerifyIntegrity(); err != nil {
		t.Fatalf("expected fresh vault to pass integrity check, got %v", err)
	}
}
