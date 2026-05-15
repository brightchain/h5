package routers

import (
	"h5/api"
	"h5/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterApiRouters(r *gin.Engine) {
	r.POST("/zip", api.Zip)
	r.POST("/redis", api.Redis)
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AesDecrypt())
	{
		apiGroup.POST("/downzip", api.PhotoOrderCy)
	}

	// 国密接口（SM2/SM3/SM4），不经过 AES 中间件
	gmGroup := r.Group("/api/gm")
	{
		gmGroup.POST("/keygen", api.KeyGen)         // 生成 SM2 密钥对
		gmGroup.POST("/encrypt", api.Encrypt)       // SM2 加密 SM4 key
		gmGroup.POST("/decrypt", api.Decrypt)       // SM2 解密 SM4 key
		gmGroup.POST("/sign", api.Sign)             // SM3withSM2 签名
		gmGroup.POST("/verify", api.Verify)         // SM3withSM2 验签
		gmGroup.POST("/sm4/encrypt", api.SM4Encrypt) // SM4-CBC 加密 body
		gmGroup.POST("/sm4/decrypt", api.SM4Decrypt) // SM4-CBC 解密 body
		gmGroup.POST("/full/send", api.FullSend)    // 完整发送方流程（演示）
		gmGroup.POST("/full/recv", api.FullRecv)    // 完整接收方流程（演示）
	}
}
