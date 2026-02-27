package consumer

import (
	"encoding/json"
	"fmt"
	"h5/internal/handler"
	"h5/internal/mq"
	"h5/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SmsNotifyMessage struct {
	Mobile     string            `json:"mobile"`
	Content    string            `json:"content"`
	TemplateID int               `json:"template_id"`
	TpContent  map[string]string `json:"tpContent"`
	SendTime   int64             `json:"send_time"`
}

func ConsumeSmsNotify(once bool) error {
	logger.LogError("[SMS] Connecting to RabbitMQ...", nil)
	conn, err := mq.GetConn()
	if err != nil {
		logger.LogError(fmt.Sprintf("[SMS] GetConn error: %v", err), err)
		return err
	}
	logger.LogError(fmt.Sprintf("[SMS] Connected, conn closed: %v", conn.IsClosed()), nil)

	ch, err := conn.Channel()
	if err != nil {
		logger.LogError(fmt.Sprintf("[SMS] Channel error: %v", err), err)
		return err
	}
	logger.LogError("[SMS] Channel created", nil)

	ch.Qos(1, 0, false)
	defer ch.Close()

	mq.Declare(ch)

	msgs, err := ch.Consume(
		mq.QueueSmsNotify,
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
		if err := handleSmsMessage(msg); err != nil {
			msg.Nack(false, false)
			logger.LogError(fmt.Sprintf("处理短信通知消息失败: %s", err.Error()), err)
		} else {
			msg.Ack(false)
		}
		if once {
			break
		}
	}
	return nil
}

func handleSmsMessage(msg amqp.Delivery) error {
	var data SmsNotifyMessage
	if err := json.Unmarshal(msg.Body, &data); err != nil {
		return err
	}
	if data.TemplateID == 0 {
		return handler.HandleSmsNotify(data.Mobile, data.Content, data.SendTime)
	}
	return handler.HandleSmsNotify1(data.Mobile, data.TpContent, int64(data.TemplateID))
}
