package auth

import (
	"testing"
)

func TestHashPasswordAndCheck(t *testing.T) {
	password := "admin123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "" || hash == password {
		t.Errorf("expected hashed value different from plain password")
	}
	if !CheckPassword(hash, password) {
		t.Error("CheckPassword should succeed for correct password")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("CheckPassword should fail for wrong password")
	}
	if CheckPassword("", password) {
		t.Error("CheckPassword should fail for empty hash")
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	if h1 == h2 {
		t.Error("bcrypt should produce different hashes (salt)")
	}
}
