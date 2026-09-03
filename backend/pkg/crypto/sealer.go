package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

type Sealer struct {
	aead cipher.AEAD
}

func NewSealer(key string) (*Sealer, error) {
	if key == "" {
		return nil, errors.New("encryption key is empty")
	}

	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &Sealer{aead: aead}, nil
}

func (s *Sealer) Seal(plaintext string) ([]byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return s.aead.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (s *Sealer) Open(ciphertext []byte) (string, error) {
	if len(ciphertext) <= s.aead.NonceSize() {
		return "", ErrInvalidCiphertext
	}

	nonce := ciphertext[:s.aead.NonceSize()]
	plaintext, err := s.aead.Open(nil, nonce, ciphertext[s.aead.NonceSize():], nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	return string(plaintext), nil
}
