package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Hasher struct {
	key string
}

func New(key string) (*Hasher, error) {
	if len(key) == 0 {
		return nil, errors.New("key is empty")
	}
	return &Hasher{
		key: key,
	}, nil
}

func (h *Hasher) Check(data []byte, acceptedHash []byte) (bool, error) {
	hash, err := h.Hash(data)
	if err != nil {
		return false, err
	}

	if !hmac.Equal([]byte(hash), acceptedHash) {
		return false, nil
	}

	return true, nil
}

func (h *Hasher) Hash(data []byte) (string, error) {
	hc := hmac.New(sha256.New, []byte(h.key))
	_, err := hc.Write(data)
	if err != nil {
		return "", err
	}
	dst := hc.Sum(nil)
	return hex.EncodeToString(dst), nil
}
