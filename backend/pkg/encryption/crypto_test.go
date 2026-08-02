package encryption

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	inputs := [][]byte{
		[]byte("hello world"),
		[]byte("postgres://user:pass@host:5432/db?sslmode=require"),
		[]byte{},
		bytes.Repeat([]byte("a"), 4096),
	}

	for _, input := range inputs {
		encoded, err := Encrypt(input, key)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		if encoded == "" {
			t.Fatal("Encrypt() returned empty string")
		}

		decoded, err := Decrypt(encoded, key)
		if err != nil {
			t.Fatalf("Decrypt() error = %v", err)
		}
		if !bytes.Equal(decoded, input) {
			t.Errorf("round trip mismatch: got %q, want %q", decoded, input)
		}
	}
}

func TestEncryptProducesUniqueCiphertexts(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 7)
	}

	a, err := Encrypt([]byte("same data"), key)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Encrypt([]byte("same data"), key)
	if err != nil {
		t.Fatal(err)
	}

	if a == b {
		t.Error("Encrypt() produced identical ciphertext for same input (nonce reuse)")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	t.Parallel()

	keyA := bytes.Repeat([]byte{1}, 32)
	keyB := bytes.Repeat([]byte{2}, 32)

	encoded, err := Encrypt([]byte("secret"), keyA)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Decrypt(encoded, keyB); err == nil {
		t.Error("Decrypt() with wrong key = nil error, want error")
	}
}

func TestEncryptWithInvalidKey(t *testing.T) {
	t.Parallel()

	if _, err := Encrypt([]byte("x"), []byte("short")); err == nil {
		t.Error("Encrypt() with 5-byte key = nil error, want error")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{3}, 32)

	tests := []struct {
		name  string
		input string
	}{
		{name: "not base64", input: "!!not-base64!!"},
		{name: "empty string", input: ""},
		{name: "too short ciphertext", input: strings.Repeat("A", 8)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Decrypt(tt.input, key); err == nil {
				t.Errorf("Decrypt(%q) = nil error, want error", tt.input)
			}
		})
	}
}
