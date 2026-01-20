package consumer

import (
	"encoding/json"
	"fmt"
	"h5/internal/handler"
	"h5/internal/mq"
	"h5/internal/retry"
	"h5/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderCancelMessage struct {
	OrderNo       string `json:"order_no"`
	RetryCount    int    `json:"retry_count"`
	LastRetryTime int64  `json:"last_retry_time"`
}

func ConsumeOrderCancel(once bool) error {
	conn, err := mq.GetConn()
	if err != nil {
		return err
	}
	ch, err := conn.Channel()
	
	if err != nil {
		return err
	}

	ch.Qos(1, 0, false)
	defer ch.Close()

	mq.Declare(ch)

	msgs, err := ch.Consume(
		mq.QueueOrderCancel,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		err := handleMessage(ch,msg)
		if err != nil {
			msg.Nack(false, false)
			logger.LogError(fmt.Sprintf("处理订单取消消息失败: %s", err.Error()), err)
		} else {
			msg.Ack(false)
		}
		if once {
			break // ✔ 消费一次就退出（适合 cron）
		}
	}
	return nil
}

func handleMessage(ch *amqp.Channel, msg amqp.Delivery) error {
	var data OrderCancelMessage
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		return err
	}

	err := handler.HandleOrderCancel(data.OrderNo)
	if err != nil {
		retry.Retry(ch, msg.Body, data.RetryCount)
		return err
	}

	return nil
}