package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"h5/pkg/config"
	"h5/pkg/logger"
)

// HandleSmsNotify 短信通知处理
func HandleSmsNotify(mobile, content string, sendTime int64) error {
	if mobile == "" {
		return errors.New("手机号不能为空")
	}

	// 调用短信发送接口
	return sendSms(mobile, content)
}

// sendSms 发送短信
func sendSms(tel, content string) error {
	if tel == "" {
		return errors.New("手机号不能为空")
	}

	// 从配置获取短信服务配置
	smsURL := config.Env("sms.url", "").(string)
	account := config.Env("sms.account", "").(string)
	password := config.Env("sms.password", "").(string)
	signature := config.Env("sms.signature", "物会网络").(string)

	if smsURL == "" || account == "" || password == "" {
		logger.LogError("短信服务配置未完善", nil)
		return errors.New("短信服务配置未完善")
	}

	// 添加签名
	fullContent := "【" + signature + "】" + content

	// 构建请求参数
	postArr := map[string]string{
		"account":  account,
		"password": password,
		"msg":      url.QueryEscape(fullContent),
		"phone":    tel,
		"report":   "true",
	}

	jsonData, err := json.Marshal(postArr)
	if err != nil {
		return fmt.Errorf("序列化请求参数失败: %w", err)
	}

	req, err := http.NewRequest("POST", smsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送短信请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 记录日志
	logger.LogError(fmt.Sprintf("短信发送请求: tel=%s, content=%s", tel, fullContent), nil)
	logger.LogError(fmt.Sprintf("短信发送响应: %s", string(body)), nil)

	if resp.StatusCode != http.StatusOK {
		logger.LogError(fmt.Sprintf("短信发送失败1 status=%d, body=%s", resp.StatusCode, string(body)), nil)
		return fmt.Errorf("短信发送失败 status=%d", resp.StatusCode)
	}

	// 解析响应，检查是否成功 code == 0 为成功
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 判断成功: code == 0 (兼容int、float64和string)
	var code float64
	var ok bool
	switch v := result["code"].(type) {
	case float64:
		code = v
		ok = true
	case int:
		code = float64(v)
		ok = true
	case string:
		// 尝试解析字符串数字
		var err error
		code, err = strconv.ParseFloat(v, 64)
		ok = err == nil
	}
	if !ok || code != 0 {
		logger.LogError(fmt.Sprintf("短信发送失败 result=%v", result), nil)
		return fmt.Errorf("短信发送失败: result=%v", result)
	}

	logger.LogError(fmt.Sprintf("短信发送成功: tel=%s", tel), nil)
	return nil
}

func HandleSmsNotify1(mobile string, content map[string]string, template_id int64) error {
	if mobile == "" {
		return errors.New("手机号不能为空")
	}

	// 调用短信发送接口
	return sendSms1(mobile, content, template_id)
}

// sendSms1 模板短信发送
func sendSms1(mobile string, content map[string]string, tpId int64) error {
	// 从配置获取短信服务配置
	smsURL := config.Env("zthysms.url", "").(string)
	account := config.Env("zthysms.account", "").(string)
	password := config.Env("zthysms.password", "").(string)
	signature := config.Env("sms.signature", "物会网络").(string)

	if account == "" || password == "" {
		logger.LogError("短信服务配置未完善", nil)
		return errors.New("短信服务配置未完善")
	}

	// tKey为时间戳
	tKey := time.Now().Unix()

	// password = md5(md5(password) + tKey)
	firstMd5 := fmt.Sprintf("%x", md5.Sum([]byte(password)))
	passwordHash := fmt.Sprintf("%x", md5.Sum([]byte(firstMd5+fmt.Sprintf("%d", tKey))))

	// 构建records
	records := []map[string]interface{}{
		{"mobile": mobile},
	}
	if content != nil && len(content) > 0 {
		records[0]["tpContent"] = content
	}

	// 构建请求参数
	postArr := map[string]interface{}{
		"tpId":      strconv.FormatInt(tpId, 10),
		"username":  account,
		"password":  passwordHash,
		"tKey":      tKey,
		"signature": fmt.Sprintf("【%s】", signature),
		"records":   records,
	}

	jsonData, err := json.Marshal(postArr)
	if err != nil {
		return fmt.Errorf("序列化请求参数失败: %w", err)
	}

	req, err := http.NewRequest("POST", smsURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送短信请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 记录日志
	logger.LogError(fmt.Sprintf("模板短信发送请求: mobile=%s, tpId=%d, tpContent=%v", mobile, tpId, content), nil)
	logger.LogError(fmt.Sprintf("模板短信发送响应: %s", string(body)), nil)

	// 解析响应，检查是否成功 code == 200 为成功
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.LogError(fmt.Sprintf("解析响应失败: %v, body=%s", err, string(body)), nil)
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 判断成功: code == 200 (兼容int、float64和string)
	var codeVal float64
	var ok bool
	switch v := result["code"].(type) {
	case float64:
		codeVal = v
		ok = true
	case int:
		codeVal = float64(v)
		ok = true
	case string:
		var err error
		codeVal, err = strconv.ParseFloat(v, 64)
		ok = err == nil
	}
	if !ok || codeVal != 200 {
		logger.LogError(fmt.Sprintf("模板短信发送失败 result=%v", result), nil)
		return fmt.Errorf("模板短信发送失败: result=%v", result)
	}

	logger.LogError(fmt.Sprintf("模板短信发送成功: mobile=%s", mobile), nil)
	return nil
}
