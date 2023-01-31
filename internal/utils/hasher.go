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
