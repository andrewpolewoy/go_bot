package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

var (
	ErrMissingSignature  = errors.New("missing X-Hub-Signature-256")
	ErrBadSignatureHex   = errors.New("invalid signature hex")
	ErrSignatureMismatch = errors.New("signature mismatch")
)

func validateGitHubSignature(body []byte, sigHeader string, secret []byte) error {
	if len(secret) == 0 {
		// без секрета в конфиге — в проде лучше фейлить, но для локалки можно пропустить
		return nil
	}

	const prefix = "sha256="
	if !strings.HasPrefix(sigHeader, prefix) {
		return ErrMissingSignature
	}

	sigHex := strings.TrimPrefix(sigHeader, prefix)
	provided, err := hex.DecodeString(sigHex)
	if err != nil {
		return ErrBadSignatureHex
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	expected := mac.Sum(nil)

	if !hmac.Equal(provided, expected) {
		return ErrSignatureMismatch
	}
	return nil
}
