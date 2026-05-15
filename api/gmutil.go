// Package handler contains Gin HTTP handlers that expose the GM crypto
// utilities as a REST API.  Each handler follows the same pattern:
//   - bind JSON request
//   - call crypto package
//   - return JSON response
package api

import (
	"encoding/base64"
	"net/http"

	"h5/pkg/crypto"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────
// Shared response helper
// ─────────────────────────────────────────────

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "ok", "data": data})
}

func fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": -1, "msg": msg})
}

// ─────────────────────────────────────────────
// POST /api/gm/keygen
// ─────────────────────────────────────────────

// KeyGen generates a new SM2 key pair and returns both keys as hex strings.
// Mirrors Java's generSM2Key().
func KeyGen(c *gin.Context) {
	pub, priv, err := crypto.GenerateKeyPair()
	if err != nil {
		fail(c, "keygen failed: "+err.Error())
		return
	}
	ok(c, gin.H{"publicKey": pub, "privateKey": priv})
}

// ─────────────────────────────────────────────
// POST /api/gm/encrypt
// Body: { "data": "<base64 plaintext>", "publicKey": "<hex>" }
// ─────────────────────────────────────────────

// Encrypt encrypts arbitrary bytes (e.g. an SM4 key) with a SM2 public key.
// Mirrors Java's GmUtil.encrypt(sm4Key, outPublicKey).
func Encrypt(c *gin.Context) {
	var req struct {
		Data      string `json:"data" binding:"required"` // base64
		PublicKey string `json:"publicKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}
	plainBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		fail(c, "data must be base64: "+err.Error())
		return
	}
	pub, err := crypto.HexToPublicKey(req.PublicKey)
	if err != nil {
		fail(c, "invalid publicKey: "+err.Error())
		return
	}
	cipher, err := crypto.Encrypt(plainBytes, pub)
	if err != nil {
		fail(c, "encrypt failed: "+err.Error())
		return
	}
	ok(c, gin.H{"encryptedData": base64.StdEncoding.EncodeToString(cipher)})
}

// ─────────────────────────────────────────────
// POST /api/gm/decrypt
// Body: { "data": "<base64 ciphertext>", "privateKey": "<hex>" }
// ─────────────────────────────────────────────

// Decrypt decrypts SM2-encrypted bytes with a private key.
// Mirrors Java's GmUtil.decrypt(Base64.decode(encKey), outPrivateKey).
func Decrypt(c *gin.Context) {
	var req struct {
		Data       string `json:"data" binding:"required"`
		PrivateKey string `json:"privateKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}
	cipherBytes, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		fail(c, "data must be base64: "+err.Error())
		return
	}
	priv, err := crypto.HexToPrivateKey(req.PrivateKey)
	if err != nil {
		fail(c, "invalid privateKey: "+err.Error())
		return
	}
	plain, err := crypto.Decrypt(cipherBytes, priv)
	if err != nil {
		fail(c, "decrypt failed: "+err.Error())
		return
	}
	ok(c, gin.H{"plainData": base64.StdEncoding.EncodeToString(plain)})
}

// ─────────────────────────────────────────────
// POST /api/gm/sign
// Body: { "data": "<string to sign>", "privateKey": "<hex>" }
// ─────────────────────────────────────────────

// Sign signs a message with SM3withSM2 and returns an ASN.1 Base64 signature.
// Mirrors Java's generateResponseSign().
func Sign(c *gin.Context) {
	var req struct {
		Data       string `json:"data" binding:"required"`
		PrivateKey string `json:"privateKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}
	priv, err := crypto.HexToPrivateKey(req.PrivateKey)
	if err != nil {
		fail(c, "invalid privateKey: "+err.Error())
		return
	}
	sig, err := crypto.Sign([]byte(req.Data), priv)
	if err != nil {
		fail(c, "sign failed: "+err.Error())
		return
	}
	ok(c, gin.H{"sign": base64.StdEncoding.EncodeToString(sig)})
}

// ─────────────────────────────────────────────
// POST /api/gm/verify
// Body: { "data": "<signed string>", "sign": "<base64>", "publicKey": "<hex>" }
// ─────────────────────────────────────────────

// Verify checks an SM3withSM2 signature.
// Mirrors Java's verifyRequestSign().
func Verify(c *gin.Context) {
	var req struct {
		Data      string `json:"data" binding:"required"`
		Sign      string `json:"sign" binding:"required"`
		PublicKey string `json:"publicKey" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}
	sigBytes, err := base64.StdEncoding.DecodeString(req.Sign)
	if err != nil {
		fail(c, "sign must be base64: "+err.Error())
		return
	}
	pub, err := crypto.HexToPublicKey(req.PublicKey)
	if err != nil {
		fail(c, "invalid publicKey: "+err.Error())
		return
	}
	valid := crypto.Verify([]byte(req.Data), sigBytes, pub)
	ok(c, gin.H{"valid": valid})
}

// ─────────────────────────────────────────────
// POST /api/gm/sm4/encrypt
// Body: { "plainText": "...", "sm4Key": "<base64>", "iv": "<base64>" }
// ─────────────────────────────────────────────

// SM4Encrypt encrypts a UTF-8 string body using SM4-CBC/PKCS7.
// Mirrors Java's encryptResponseBody().
func SM4Encrypt(c *gin.Context) {
	var req struct {
		PlainText string `json:"plainText" binding:"required"`
		SM4Key    string `json:"sm4Key" binding:"required"` // base64
		IV        string `json:"iv" binding:"required"`     // base64
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}
	key, err := base64.StdEncoding.DecodeString(req.SM4Key)
	if err != nil {
		fail(c, "sm4Key must be base64")
		return
	}
	iv, err := base64.StdEncoding.DecodeString(req.IV)
	if err != nil {
		fail(c, "iv must be base64")
		return
	}
	cipher, err := crypto.SM4EncryptCBC(key, iv, []byte(req.PlainText))
	if err != nil {
		fail(c, "SM4 encrypt failed: "+err.Error())
		return
	}
	ok(c, gin.H{"encryptedBody": base64.StdEncoding.EncodeToString(cipher)})
}

// ─────────────────────────────────────────────
// POST /api/gm/sm4/decrypt
// Body: { "encryptedBody": "<base64>", "sm4Key": "<base64>", "iv": "<base64>" }
// ─────────────────────────────────────────────

// SM4Decrypt decrypts a SM4-CBC body.
// Mirrors Java's decryptRequestBody().
func SM4Decrypt(c *gin.Context) {
	var req struct {
		EncryptedBody string `json:"encryptedBody" binding:"required"`
		SM4Key        string `json:"sm4Key" binding:"required"`
		IV            string `json:"iv" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}
	key, err := base64.StdEncoding.DecodeString(req.SM4Key)
	if err != nil {
		fail(c, "sm4Key must be base64")
		return
	}
	iv, err := base64.StdEncoding.DecodeString(req.IV)
	if err != nil {
		fail(c, "iv must be base64")
		return
	}
	cipher, err := base64.StdEncoding.DecodeString(req.EncryptedBody)
	if err != nil {
		fail(c, "encryptedBody must be base64")
		return
	}

	plain, err := crypto.SM4DecryptCBC(key, iv, cipher)
	if err != nil {
		fail(c, "SM4 decrypt failed: "+err.Error())
		return
	}
	ok(c, gin.H{"plainText": string(plain)})
}

// ─────────────────────────────────────────────
// POST /api/gm/full/send
// Mirrors Java's main() "拼接请求报文" section.
// Body: {
//   "privateKey":    "<sender priv hex>",
//   "outPublicKey":  "<receiver pub hex>",
//   "body":          "<plain JSON string>"
// }
// ─────────────────────────────────────────────

// FullSend performs the complete sender-side flow:
//  1. generate SM4 key + IV
//  2. SM4-CBC encrypt the request body
//  3. SM3withSM2 sign the encrypted body
//  4. SM2 encrypt the SM4 key with the receiver's public key
//  5. return the assembled request envelope
func FullSend(c *gin.Context) {
	var req struct {
		PrivateKey   string `json:"privateKey" binding:"required"`
		OutPublicKey string `json:"outPublicKey" binding:"required"`
		Body         string `json:"body" binding:"required"`
		// optional head fields
		SysCode  string `json:"sysCode"`
		TranCode string `json:"tranCode"`
		TranDate string `json:"tranDate"`
		TranNo   string `json:"tranNo"`
		TranTime string `json:"tranTime"`
		Version  string `json:"version"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}

	// Step 1: sender keys
	priv, err := crypto.HexToPrivateKey(req.PrivateKey)
	if err != nil {
		fail(c, "invalid privateKey: "+err.Error())
		return
	}
	outPub, err := crypto.HexToPublicKey(req.OutPublicKey)
	if err != nil {
		fail(c, "invalid outPublicKey: "+err.Error())
		return
	}

	// Step 2: generate SM4 key and IV
	sm4Key, err := crypto.GenerateSM4Key()
	if err != nil {
		fail(c, "generateSM4Key: "+err.Error())
		return
	}
	iv, err := crypto.GenerateIV()
	if err != nil {
		fail(c, "generateIV: "+err.Error())
		return
	}

	// Step 3: SM4 encrypt body
	cipherBody, err := crypto.SM4EncryptCBC(sm4Key, iv, []byte(req.Body))
	if err != nil {
		fail(c, "SM4 encrypt: "+err.Error())
		return
	}
	encBody := base64.StdEncoding.EncodeToString(cipherBody)

	// Step 4: sign encrypted body with sender's private key
	sigBytes, err := crypto.Sign([]byte(encBody), priv)
	if err != nil {
		fail(c, "sign: "+err.Error())
		return
	}
	sign := base64.StdEncoding.EncodeToString(sigBytes)

	// Step 5: SM2 encrypt SM4 key with receiver's public key
	encKeyBytes, err := crypto.Encrypt(sm4Key, outPub)
	if err != nil {
		fail(c, "SM2 encrypt SM4 key: "+err.Error())
		return
	}
	encKey := base64.StdEncoding.EncodeToString(encKeyBytes)
	ivB64 := base64.StdEncoding.EncodeToString(iv)

	ok(c, gin.H{
		"head": gin.H{
			"sign":     sign,
			"sysCode":  req.SysCode,
			"tranCode": req.TranCode,
			"tranDate": req.TranDate,
			"tranNo":   req.TranNo,
			"tranTime": req.TranTime,
			"version":  req.Version,
			"IV":       ivB64,
			"encKey":   encKey,
		},
		"body": encBody,
	})
}

// ─────────────────────────────────────────────
// POST /api/gm/full/recv
// Mirrors Java's main() "接收响应报文验签解密" section.
// Body: {
//   "sign":          "<base64>",
//   "encBody":       "<base64>",
//   "senderPubKey":  "<hex>",
//   "recvPrivKey":   "<hex>",
//   "encKey":        "<base64>",
//   "iv":            "<base64>"
// }
// ─────────────────────────────────────────────

// FullRecv performs the complete receiver-side flow:
//  1. verify the signature with the sender's public key
//  2. SM2 decrypt the SM4 key with the receiver's private key
//  3. SM4 decrypt the body
func FullRecv(c *gin.Context) {
	var req struct {
		Sign         string `json:"sign" binding:"required"`
		EncBody      string `json:"encBody" binding:"required"`
		SenderPubKey string `json:"senderPubKey" binding:"required"`
		RecvPrivKey  string `json:"recvPrivKey" binding:"required"`
		EncKey       string `json:"encKey" binding:"required"`
		IV           string `json:"iv" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, err.Error())
		return
	}

	// Step 1: verify signature
	sigBytes, err := base64.StdEncoding.DecodeString(req.Sign)
	if err != nil {
		fail(c, "sign base64 decode: "+err.Error())
		return
	}
	senderPub, err := crypto.HexToPublicKey(req.SenderPubKey)
	if err != nil {
		fail(c, "invalid senderPubKey: "+err.Error())
		return
	}
	if !crypto.Verify([]byte(req.EncBody), sigBytes, senderPub) {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "❌ 验签失败，拒绝处理",
		})
		return
	}

	// Step 2: SM2 decrypt SM4 key
	recvPriv, err := crypto.HexToPrivateKey(req.RecvPrivKey)
	if err != nil {
		fail(c, "invalid recvPrivKey: "+err.Error())
		return
	}
	encKeyBytes, err := base64.StdEncoding.DecodeString(req.EncKey)
	if err != nil {
		fail(c, "encKey base64 decode: "+err.Error())
		return
	}
	sm4Key, err := crypto.Decrypt(encKeyBytes, recvPriv)
	if err != nil {
		fail(c, "SM2 decrypt SM4 key: "+err.Error())
		return
	}

	// Step 3: SM4 decrypt body
	iv, err := base64.StdEncoding.DecodeString(req.IV)
	if err != nil {
		fail(c, "iv base64 decode: "+err.Error())
		return
	}
	cipherBody, err := base64.StdEncoding.DecodeString(req.EncBody)
	if err != nil {
		fail(c, "encBody base64 decode: "+err.Error())
		return
	}
	plainBody, err := crypto.SM4DecryptCBC(sm4Key, iv, cipherBody)
	if err != nil {
		fail(c, "SM4 decrypt: "+err.Error())
		return
	}

	ok(c, gin.H{
		"verifyResult": true,
		"plainBody":    string(plainBody),
	})
}
