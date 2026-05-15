// Package crypto wraps tjfoc/gmsm to provide SM2 / SM3 / SM4 utilities
// that mirror the Java GmUtil class.
package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/tjfoc/gmsm/sm2"
)

// SM2 is a thin wrapper; exported so callers can reference the type.
type SM2 struct{}

// NewSM2 returns an SM2 helper (stateless).
func NewSM2() *SM2 { return &SM2{} }

// SM2UserID mirrors Java's SM2_USER_ID = "BENE".getBytes()
var SM2UserID = []byte("BENE")

// cipherMode controls the SM2 ciphertext byte order.
// Java BouncyCastle 1.60/1.61 (Cipher.getInstance("SM2","BC")) outputs
// raw C1||C2||C3.  tjfoc constant: C1C3C2=0, C1C2C3=1.
var cipherMode = sm2.C1C2C3

// ─────────────────────────────────────────────
// Key generation
// ─────────────────────────────────────────────

// GenerateKeyPair creates a new SM2 key pair and returns the
// uncompressed public key (04 || X || Y) and private key as
// upper-case hex strings, matching the Java output format.
func GenerateKeyPair() (pubHex, privHex string, err error) {
	priv, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return
	}
	pubHex = PublicKeyToHex(&priv.PublicKey)
	privHex = PrivateKeyToHex(priv)
	return
}

// ─────────────────────────────────────────────
// Key serialisation helpers
// ─────────────────────────────────────────────

// PublicKeyToHex returns 04 || X(32B) || Y(32B) upper-case hex.
func PublicKeyToHex(pub *sm2.PublicKey) string {
	xBytes := zeroPad(pub.X.Bytes(), 32)
	yBytes := zeroPad(pub.Y.Bytes(), 32)
	raw := append([]byte{0x04}, xBytes...)
	raw = append(raw, yBytes...)
	return strings.ToUpper(hex.EncodeToString(raw))
}

// PrivateKeyToHex returns the 32-byte scalar d as 64 upper-case hex chars.
func PrivateKeyToHex(priv *sm2.PrivateKey) string {
	return fmt.Sprintf("%064X", priv.D)
}

// HexToPublicKey parses a 04||X||Y hex string into an SM2 public key.
func HexToPublicKey(hexStr string) (*sm2.PublicKey, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	if len(b) != 65 || b[0] != 0x04 {
		return nil, errors.New("invalid SM2 public key: expected 65 bytes starting with 04")
	}
	curve := sm2.P256Sm2()
	x := new(big.Int).SetBytes(b[1:33])
	y := new(big.Int).SetBytes(b[33:65])
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("public key point is not on SM2 curve")
	}
	return &sm2.PublicKey{Curve: curve, X: x, Y: y}, nil
}

// HexToPrivateKey parses a 64-char hex string into an SM2 private key.
func HexToPrivateKey(hexStr string) (*sm2.PrivateKey, error) {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	curve := sm2.P256Sm2()
	d := new(big.Int).SetBytes(b)
	priv := new(sm2.PrivateKey)
	priv.Curve = curve
	priv.D = d
	priv.PublicKey.Curve = curve
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d.Bytes())
	return priv, nil
}

// ─────────────────────────────────────────────
// SM2 Encrypt / Decrypt
//
// Java BC 1.60/1.61: Cipher.getInstance("SM2","BC") outputs raw C1||C2||C3.
// Java BC 1.62+:     use SM2Engine.Mode.C1C3C2 for new standard.
//
// This implementation matches BC 1.60/1.61 (C1C2C3) by default.
// If your Java side uses BC 1.62+ with C1C3C2 mode, change cipherMode above.
// ─────────────────────────────────────────────

// Encrypt encrypts plainText with the recipient's SM2 public key.
func Encrypt(plainText []byte, pub *sm2.PublicKey) ([]byte, error) {
	return sm2.Encrypt(pub, plainText, rand.Reader, cipherMode)
}

// Decrypt decrypts cipherText with the SM2 private key.
func Decrypt(cipherText []byte, priv *sm2.PrivateKey) ([]byte, error) {
	return sm2.Decrypt(priv, cipherText, cipherMode)
}

// ─────────────────────────────────────────────
// SM3withSM2 Sign / Verify
//
// tjfoc/gmsm v1.4.1 API:
//   sm2.Sm2Sign(priv, msg, uid, random) -> (r, s *big.Int, err)
//   sm2.Sm2Verify(pub, msg, uid, r, s)  -> bool
//   sm2.SignDigitToSignData(r, s)        -> ([]byte, error)  ASN.1 DER
//   sm2.SignDataToSignDigit(sign)        -> (r, s *big.Int, err)
// ─────────────────────────────────────────────

// Sign produces an ASN.1 DER-encoded (r, s) signature using SM3withSM2,
// matching Java's signSm3WithSm2Asn1Rs.
func Sign(msg []byte, priv *sm2.PrivateKey) ([]byte, error) {
	r, s, err := sm2.Sm2Sign(priv, msg, SM2UserID, rand.Reader)
	if err != nil {
		return nil, err
	}
	// Convert (r, s) big.Int pair -> ASN.1 DER bytes
	return sm2.SignDigitToSignData(r, s)
}

// Verify checks an ASN.1 DER signature produced by Sign.
func Verify(msg, sig []byte, pub *sm2.PublicKey) bool {
	r, s, err := sm2.SignDataToSignDigit(sig)
	if err != nil {
		return false
	}
	return sm2.Sm2Verify(pub, msg, SM2UserID, r, s)
}

// ─────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────

func zeroPad(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	padded := make([]byte, size)
	copy(padded[size-len(b):], b)
	return padded
}