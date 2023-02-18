package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

type HashGenerator struct {
	Hash hash.Hash
}

func NewHashGenerator(key string) *HashGenerator {
	h := hmac.New(sha256.New, []byte(key))
	return &HashGenerator{h}
}

func (h *HashGenerator) Generate(stringToHash string) (string, error) {
	h.Hash.Reset()
	_, err := h.Hash.Write([]byte(stringToHash))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Hash.Sum(nil)), nil
}

func (h *HashGenerator) Equal(hash1 string, hash2 string) bool {
	mac1, err := hex.DecodeString(hash1)
	if err != nil {
		return false
	}
	mac2, err := hex.DecodeString(hash2)
	if err != nil {
		return false
	}

	return hmac.Equal(mac1, mac2)
}
