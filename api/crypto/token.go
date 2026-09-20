package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const byteLength = 32

func GenerateRandomToken() (string, error) {
	tokenBytes := make([]byte, byteLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(tokenBytes), nil
}

func HashToken(token string) string {
	alg := sha256.New()
	alg.Write([]byte(token))
	tokenBytes := alg.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(tokenBytes)
}
