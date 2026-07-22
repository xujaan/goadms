package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Signer creates HMAC-SHA256 signatures for webhook payload verification.
type Signer struct{}

func NewSigner() *Signer {
	return &Signer{}
}

// Sign returns hex-encoded HMAC-SHA256 of payload using secret.
func (s *Signer) Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks that the signature matches the payload.
func (s *Signer) Verify(payload []byte, secret, signature string) bool {
	expected := s.Sign(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}
