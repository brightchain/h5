package wechabot

import (
	"bytes"
	"fmt"
	"h5/pkg/config"
	"io"
	"time"
	"mime/multipart"
	"path/filepath"
	"net/http"
	"os"
	"encoding/json"
)

type WechaBot struct {
	Token string
}

// UploadResponse 存储上传响应
type UploadResponse struct {
	ErrorCode    int    `json:"errcode"`
	ErrorMsg     string `json:"errmsg"`
	Type        string `json:"type"`
	MediaID     string `json:"media_id"`
	CreatedAt   string `json:"created_at"`
}

// FileMessage 定义文件消息结构
type FileMessage struct {
	MsgType string `json:"msgtype"`
	File    struct {
		MediaID string `json:"media_id"`
	} `json:"file"`
}

func NewWechaBot(bot string) *WechaBot {
	bots := config.GetStringMap("wechatBot")
	return &WechaBot{Token: bots[bot].(string)}
}

func (w *WechaBot) Upload(fileName string) (*UploadResponse, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/upload_media?key=%s&type=file", w.Token)
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	fmt.Printf("%v", url)
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// 创建文件表单字段
	part, err := writer.CreateFormFile("media", filepath.Base(fileName))
	if err != nil {
		return nil, fmt.Errorf("创建表单字段失败: %v", err)
	}

	// 写入文件内容
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("文件写入失败: %v", err)
	}

	// 添加额外的表单字段
	_ = writer.WriteField("filename", filepath.Base(fileName))
	_ = writer.WriteField("filelength", fmt.Sprintf("%d", fileInfo.Size()))

	// 关闭multipart writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("文件关闭失败: %v", err)

	}

	// 创建HTTP请求
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("请求创建失败: %v", err)
	}

	// 设置请求头
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %v", err)
	}
	defer response.Body.Close()

	// 读取响应
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应JSON
	var uploadResponse UploadResponse
	
	err = json.Unmarshal(responseBody, &uploadResponse)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 检查错误码
	if uploadResponse.ErrorCode != 0 {
		return nil, fmt.Errorf("上传失败: %s", uploadResponse.ErrorMsg)
	}

	return &uploadResponse, nil
}

func (w *WechaBot) SendFile(mediaID string) error {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", w.Token)
	// 构建消息内容
	message := FileMessage{
		MsgType: "file",
		File: struct {
			MediaID string `json:"media_id"`
		}{
			MediaID: mediaID,
		},
	}

	// 将消息转换为JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("JSON编码失败: %v", err)
	}

	// 创建HTTP请求
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	request.Header.Set("Content-Type", "application/json")

	// 创建HTTP客户端并发送请求
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("发送请求失败: %v", err)
	}
	defer response.Body.Close()

	// 读取响应
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}
	fmt.Printf("读取响应: %v", body)
	return nil;
}
