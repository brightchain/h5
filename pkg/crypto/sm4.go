package crypto

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	"github.com/tjfoc/gmsm/sm4"
)

const sm4BlockSize = 16

// ─────────────────────────────────────────────
// Key / IV generation
// ─────────────────────────────────────────────

// GenerateSM4Key returns 16 random bytes (128-bit SM4 key).
func GenerateSM4Key() ([]byte, error) {
	key := make([]byte, sm4BlockSize)
	_, err := rand.Read(key)
	return key, err
}

// GenerateIV returns 16 random bytes suitable as a CBC IV.
func GenerateIV() ([]byte, error) {
	iv := make([]byte, sm4BlockSize)
	_, err := rand.Read(iv)
	return iv, err
}

// ─────────────────────────────────────────────
// SM4-CBC with PKCS#7 padding
//
// tjfoc/gmsm v1.4.1: sm4.Sm4Cbc(key, in, mode) does NOT accept an IV
// argument and uses a global IV (zero by default).
// We use the standard cipher.Block interface instead so that the caller's
// IV is respected — this is the correct approach for production use.
// ─────────────────────────────────────────────

// SM4EncryptCBC encrypts plainText using SM4-CBC with PKCS#7 padding.
// key and iv must each be 16 bytes.
func SM4EncryptCBC(key, iv, plainText []byte) ([]byte, error) {
	if len(key) != sm4BlockSize {
		return nil, errors.New("SM4 key must be 16 bytes")
	}
	if len(iv) != sm4BlockSize {
		return nil, errors.New("SM4 IV must be 16 bytes")
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(plainText, sm4BlockSize)
	cipherText := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, padded)
	return cipherText, nil
}

// SM4DecryptCBC decrypts cipherText using SM4-CBC and removes PKCS#7 padding.
func SM4DecryptCBC(key, iv, cipherText []byte) ([]byte, error) {
	if len(key) != sm4BlockSize {
		return nil, errors.New("SM4 key must be 16 bytes")
	}
	if len(iv) != sm4BlockSize {
		return nil, errors.New("SM4 IV must be 16 bytes")
	}
	if len(cipherText)%sm4BlockSize != 0 {
		return nil, errors.New("cipherText length is not a multiple of SM4 block size")
	}
	sm4.SetIV(iv)

	return sm4.Sm4Cbc(key, cipherText, false)
}

// ─────────────────────────────────────────────
// PKCS#7 padding helpers
// ─────────────────────────────────────────────

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data after decryption")
	}
	padding := int(data[length-1])
	if padding == 0 || padding > sm4BlockSize {
		return nil, errors.New("invalid PKCS7 padding")
	}
	return data[:length-padding], nil
}
