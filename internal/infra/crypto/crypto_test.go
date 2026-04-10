package crypto

import (
	"bytes"
	"testing"
)

var testKey = bytes.Repeat([]byte("k"), 32)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	plaintext := []byte("hello, world")

	ciphertext, err := Encrypt(testKey, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	got, err := Decrypt(testKey, ciphertext)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("want %q, got %q", plaintext, got)
	}
}

func TestEncryptProducesNonDeterministicOutput(t *testing.T) {
	plaintext := []byte("same input")

	ct1, err := Encrypt(testKey, plaintext)
	if err != nil {
		t.Fatalf("first encrypt: %v", err)
	}
	ct2, err := Encrypt(testKey, plaintext)
	if err != nil {
		t.Fatalf("second encrypt: %v", err)
	}
	if bytes.Equal(ct1, ct2) {
		t.Fatal("expected different ciphertexts for the same plaintext (nonce reuse)")
	}
}

func TestEncryptEmptyReturnsNil(t *testing.T) {
	ct, err := Encrypt(testKey, []byte{})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if ct != nil {
		t.Fatalf("expected nil, got %v", ct)
	}
}

func TestDecryptNilReturnsNil(t *testing.T) {
	got, err := Decrypt(testKey, nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestDecryptEmptyReturnsNil(t *testing.T) {
	got, err := Decrypt(testKey, []byte{})
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	ct, err := Encrypt(testKey, []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	wrongKey := bytes.Repeat([]byte("x"), 32)
	_, err = Decrypt(wrongKey, ct)
	if err == nil {
		t.Fatal("expected error decrypting with wrong key")
	}
}

func TestDecryptTamperedCiphertextFails(t *testing.T) {
	ct, err := Encrypt(testKey, []byte("secret"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	ct[len(ct)-1] ^= 0xff // flip last byte
	_, err = Decrypt(testKey, ct)
	if err == nil {
		t.Fatal("expected error decrypting tampered ciphertext")
	}
}

func TestDecryptTooShortFails(t *testing.T) {
	_, err := Decrypt(testKey, []byte("short"))
	if err == nil {
		t.Fatal("expected error for ciphertext shorter than nonce size")
	}
}

func TestEncryptInvalidKeyFails(t *testing.T) {
	_, err := Encrypt([]byte("bad key"), []byte("data"))
	if err == nil {
		t.Fatal("expected error for invalid key size")
	}
}

func TestDecryptInvalidKeyFails(t *testing.T) {
	_, err := Decrypt([]byte("bad key"), []byte("some ciphertext bytes here!!"))
	if err == nil {
		t.Fatal("expected error for invalid key size")
	}
}
