package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func signBody(secret, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestValidateGitHubSignature_OK(t *testing.T) {
	secret := []byte("secret")
	body := []byte(`{"hello":"world"}`)

	sig := signBody(secret, body)
	if err := validateGitHubSignature(body, sig, secret); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestValidateGitHubSignature_EmptySecret(t *testing.T) {
	err := validateGitHubSignature([]byte(`{}`), "sha256=00", nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestValidateGitHubSignature_MissingPrefix(t *testing.T) {
	secret := []byte("secret")
	err := validateGitHubSignature([]byte(`{}`), "nope", secret)
	if err != ErrMissingSignature {
		t.Fatalf("expected %v, got %v", ErrMissingSignature, err)
	}
}

func TestValidateGitHubSignature_BadHex(t *testing.T) {
	secret := []byte("secret")
	err := validateGitHubSignature([]byte(`{}`), "sha256=ZZZ", secret)
	if err != ErrBadSignatureHex {
		t.Fatalf("expected %v, got %v", ErrBadSignatureHex, err)
	}
}

func TestValidateGitHubSignature_Mismatch(t *testing.T) {
	secret := []byte("secret")
	err := validateGitHubSignature([]byte(`{}`), "sha256=00", secret)
	if err != ErrSignatureMismatch {
		t.Fatalf("expected %v, got %v", ErrSignatureMismatch, err)
	}
}
