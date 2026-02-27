package main

import (
	"flag"
	"fmt"

	"h5/bootstrap"
	"h5/config"
	"h5/internal/consumer"
	"h5/pkg/logger"
)

func init() {
	config.Initialize()
}

func main() {
	bootstrap.SetupSlog()
	bootstrap.SetupModel()
	once := flag.Bool("once", false, "consume once and exit")
	flag.Parse()

	// SMS 通知队列消费者（goroutine）
	go func() {
		err := consumer.ConsumeSmsNotify(*once)
		if err != nil {
			logger.LogError(fmt.Sprintf("[SMS] SmsNotifyConsumer error: %v", err), err)
		}
	}()

	// 订单取消队列消费者（主线程）
	logger.LogError("[ORDER_CANCEL] Consumer starting...", nil)
	err := consumer.ConsumeOrderCancel(*once)
	if err != nil {
		logger.LogError(fmt.Sprintf("[ORDER_CANCEL] Consumer error: %v", err), err)
	}

}
