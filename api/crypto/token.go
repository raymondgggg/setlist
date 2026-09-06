package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

const byteLength = 32

func GenerateRandomToken() (string, error) {
	tokenBytes := make([]byte, byteLength)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	return token, nil
}
