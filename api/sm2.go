package api

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

const (
	PublicKey  = "MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEN9MT1eQC81Zv0is6JlUPYSRLRmGUq1eB++gCeVFTZf4dXygjat+US0braeKn3QpKN6XeinjWjroj3SXWdZikxA=="
	PrivateKey = "MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgLFbffPP+lXseaACB81n/ACYIod8gWi0uzaM7BiNoqFOgCgYIKoEcz1UBgi2hRANCAAQ30xPV5ALzVm/SKzomVQ9hJEtGYZSrV4H76AJ5UVNl/h1fKCNq35RLRutp4qfdCko3pd6KeNaOuiPdJdZ1mKTE"
)

// KeyPair 密钥对结构
type KeyPair struct {
	PublicKey  string
	PrivateKey string
}

// GenerateSmKey 生成国密公私钥对
func GenerateSmKey() (*KeyPair, error) {
	// 生成SM2密钥对
	privateKey, err := sm2.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("生成SM2密钥对失败: %w", err)
	}

	// 获取公钥
	publicKey := &privateKey.PublicKey

	// 将私钥转换为PKCS8格式并编码为Base64
	privateKeyBytes, err := x509.MarshalSm2PrivateKey(privateKey, nil)
	if err != nil {
		return nil, fmt.Errorf("序列化私钥失败: %w", err)
	}
	privateKeyStr := base64.StdEncoding.EncodeToString(privateKeyBytes)

	// 将公钥转换为X509格式并编码为Base64
	publicKeyBytes, err := x509.MarshalSm2PublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("序列化公钥失败: %w", err)
	}
	publicKeyStr := base64.StdEncoding.EncodeToString(publicKeyBytes)

	return &KeyPair{
		PublicKey:  publicKeyStr,
		PrivateKey: privateKeyStr,
	}, nil
}

// CreatePublicKey 将Base64编码的公钥字符串转换为公钥对象
func CreatePublicKey(publicKeyStr string) (*sm2.PublicKey, error) {
	// Base64解码
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, fmt.Errorf("Base64解码公钥失败: %w", err)
	}

	// 解析公钥
	publicKey, err := x509.ParseSm2PublicKey(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("解析公钥失败: %w", err)
	}

	return publicKey, nil
}

// CreatePrivateKey 将Base64编码的私钥字符串转换为私钥对象
func CreatePrivateKey(privateKeyStr string) (*sm2.PrivateKey, error) {
	// Base64解码
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("Base64解码私钥失败: %w", err)
	}

	// 尝试使用 PKCS8 解析（无密码）
	priv, err := x509.ParsePKCS8PrivateKey(privateKeyBytes, nil)
	if err != nil {
		// 如果失败，尝试使用 SM2 专用解析
		priv, err = x509.ParseSm2PrivateKey(privateKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
	}

	return priv, nil
}

// Encrypt 使用SM2公钥加密数据
func Encrypt(data string, publicKeyStr string) (string, error) {
	if data == "" {
		return "", errors.New("待加密数据不能为空")
	}

	// 创建公钥对象
	publicKey, err := CreatePublicKey(publicKeyStr)
	if err != nil {
		return "", fmt.Errorf("创建公钥对象失败: %w", err)
	}

	// SM2加密
	ciphertext, err := sm2.Encrypt(publicKey, []byte(data), rand.Reader, sm2.C1C3C2)
	if err != nil {
		return "", fmt.Errorf("SM2加密失败: %w", err)
	}

	// Base64编码
	encryptedStr := base64.StdEncoding.EncodeToString(ciphertext)
	return encryptedStr, nil
}

// Decrypt 使用SM2私钥解密数据
func Decrypt(encryptedData string, privateKeyStr string) (string, error) {
	if encryptedData == "" {
		return "", errors.New("待解密数据不能为空")
	}

	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 创建私钥对象
	privateKey, err := CreatePrivateKey(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("创建私钥对象失败: %w", err)
	}

	// SM2解密
	plaintext, err := sm2.Decrypt(privateKey, ciphertext, sm2.C1C3C2)
	if err != nil {
		return "", fmt.Errorf("SM2解密失败: %w", err)
	}

	return string(plaintext), nil
}

// SignByPrivateKey 使用私钥对数据进行签名
func SignByPrivateKey(data string, privateKeyStr string) (string, error) {
	if data == "" {
		return "", errors.New("待签名数据不能为空")
	}

	// 创建私钥对象
	privateKey, err := CreatePrivateKey(privateKeyStr)
	if err != nil {
		return "", fmt.Errorf("创建私钥对象失败: %w", err)
	}

	// SM2签名（使用SM3作为哈希算法）
	signature, err := privateKey.Sign(rand.Reader, []byte(data), nil)
	if err != nil {
		return "", fmt.Errorf("SM2签名失败: %w", err)
	}

	// Base64编码
	signatureStr := base64.StdEncoding.EncodeToString(signature)
	return signatureStr, nil
}

// VerifyByPublicKey 使用公钥验证签名
func VerifyByPublicKey(data string, publicKeyStr string, signatureStr string) (bool, error) {
	if data == "" || signatureStr == "" {
		return false, errors.New("数据或签名不能为空")
	}

	// 创建公钥对象
	publicKey, err := CreatePublicKey(publicKeyStr)
	if err != nil {
		return false, fmt.Errorf("创建公钥对象失败: %w", err)
	}

	// Base64解码签名
	signature, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		return false, fmt.Errorf("Base64解码签名失败: %w", err)
	}

	// 验证签名
	valid := publicKey.Verify([]byte(data), signature)
	return valid, nil
}

// ExportPublicKeyToPEM 导出公钥为PEM格式
func ExportPublicKeyToPEM(publicKeyStr string) (string, error) {
	publicKey, err := CreatePublicKey(publicKeyStr)
	if err != nil {
		return "", err
	}

	publicKeyBytes, err := x509.MarshalSm2PublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("序列化公钥失败: %w", err)
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// ExportPrivateKeyToPEM 导出私钥为PEM格式
func ExportPrivateKeyToPEM(privateKeyStr string) (string, error) {
	privateKey, err := CreatePrivateKey(privateKeyStr)
	if err != nil {
		return "", err
	}

	privateKeyBytes, err := x509.MarshalSm2PrivateKey(privateKey, nil)
	if err != nil {
		return "", fmt.Errorf("序列化私钥失败: %w", err)
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return string(pem.EncodeToMemory(block)), nil
}

// ==================== API 接口 ====================

// GenerateKeysRequest 生成密钥请求
type GenerateKeysRequest struct{}

// GenerateKeysResponse 生成密钥响应
type GenerateKeysResponse struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data KeyPair `json:"data"`
}

// EncryptRequest 加密请求
type EncryptRequest struct {
	PlainText string `json:"plainText" binding:"required"`
	PublicKey string `json:"publicKey" binding:"required"`
}

// EncryptResponse 加密响应
type EncryptResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// DecryptRequest 解密请求
type DecryptRequest struct {
	CipherText string `json:"cipherText" binding:"required"`
	PrivateKey string `json:"privateKey" binding:"required"`
}

// DecryptResponse 解密响应
type DecryptResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// SignRequest 签名请求
type SignRequest struct {
	Data       string `json:"data" binding:"required"`
	PrivateKey string `json:"privateKey" binding:"required"`
}

// SignResponse 签名响应
type SignResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

// VerifyRequest 验签请求
type VerifyRequest struct {
	Data      string `json:"data" binding:"required"`
	PublicKey string `json:"publicKey" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// VerifyResponse 验签响应
type VerifyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data bool   `json:"data"`
}

// GenerateKeys 生成SM2密钥对
// @Summary 生成SM2公私钥对
// @Description 生成SM2公私钥对
// @Tags SM2
// @Accept json
// @Produce json
// @Success 200 {object} GenerateKeysResponse
// @Router /api/sm2/generate-keys [post]
func GenerateKeys(c *gin.Context) {
	keyPair, err := GenerateSmKey()
	if err != nil {
		c.JSON(200, GenerateKeysResponse{
			Code: 500,
			Msg:  fmt.Sprintf("生成密钥对失败: %v", err),
			Data: KeyPair{},
		})
		return
	}

	c.JSON(200, GenerateKeysResponse{
		Code: 0,
		Msg:  "success",
		Data: *keyPair,
	})
}

// EncryptData SM2加密
// @Summary 使用SM2公钥加密数据
// @Description 使用SM2公钥对明文进行加密，返回Base64编码的密文
// @Tags SM2
// @Accept json
// @Produce json
// @Param request body EncryptRequest true "加密请求"
// @Success 200 {object} EncryptResponse
// @Router /api/sm2/encrypt [post]
func EncryptData(c *gin.Context) {
	var req EncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, EncryptResponse{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: "",
		})
		return
	}

	cipherText, err := Encrypt(req.PlainText, req.PublicKey)
	if err != nil {
		c.JSON(200, EncryptResponse{
			Code: 500,
			Msg:  "加密失败: " + err.Error(),
			Data: "",
		})
		return
	}

	c.JSON(200, EncryptResponse{
		Code: 0,
		Msg:  "success",
		Data: cipherText,
	})
}

// DecryptData SM2解密
// @Summary 使用SM2私钥解密数据
// @Description 使用SM2私钥对Base64编码的密文进行解密
// @Tags SM2
// @Accept json
// @Produce json
// @Param request body DecryptRequest true "解密请求"
// @Success 200 {object} DecryptResponse
// @Router /api/sm2/decrypt [post]
func DecryptData(c *gin.Context) {
	var req DecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, DecryptResponse{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: "",
		})
		return
	}

	plainText, err := Decrypt(req.CipherText, req.PrivateKey)
	if err != nil {
		c.JSON(200, DecryptResponse{
			Code: 500,
			Msg:  "解密失败: " + err.Error(),
			Data: "",
		})
		return
	}

	c.JSON(200, DecryptResponse{
		Code: 0,
		Msg:  "success",
		Data: plainText,
	})
}

// SignData SM2签名
// @Summary 使用SM2私钥签名
// @Description 使用SM2私钥对数据进行签名，返回Base64编码的签名
// @Tags SM2
// @Accept json
// @Produce json
// @Param request body SignRequest true "签名请求"
// @Success 200 {object} SignResponse
// @Router /api/sm2/sign [post]
func SignData(c *gin.Context) {
	var req SignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, SignResponse{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: "",
		})
		return
	}

	signature, err := SignByPrivateKey(req.Data, req.PrivateKey)
	if err != nil {
		c.JSON(200, SignResponse{
			Code: 500,
			Msg:  "签名失败: " + err.Error(),
			Data: "",
		})
		return
	}

	c.JSON(200, SignResponse{
		Code: 0,
		Msg:  "success",
		Data: signature,
	})
}

// VerifySignature SM2验签
// @Summary 使用SM2公钥验签
// @Description 使用SM2公钥验证签名的有效性
// @Tags SM2
// @Accept json
// @Produce json
// @Param request body VerifyRequest true "验签请求"
// @Success 200 {object} VerifyResponse
// @Router /api/sm2/verify [post]
func VerifySignature(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, VerifyResponse{
			Code: 400,
			Msg:  "参数错误: " + err.Error(),
			Data: false,
		})
		return
	}

	valid, err := VerifyByPublicKey(req.Data, req.PublicKey, req.Signature)
	if err != nil {
		c.JSON(200, VerifyResponse{
			Code: 500,
			Msg:  "验签失败: " + err.Error(),
			Data: false,
		})
		return
	}

	c.JSON(200, VerifyResponse{
		Code: 0,
		Msg:  "success",
		Data: valid,
	})
}
