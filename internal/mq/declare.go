package mq

import  amqp "github.com/rabbitmq/amqp091-go"

const (
	ExchangeOrder = "order_exchange"

	QueueOrderCancel      = "order_cancel_queue"
	QueueOrderCancelRetry = "order_cancel_retry_queue"
)

func Declare(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(
		ExchangeOrder,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	// 主队列
	_, err := ch.QueueDeclare(
		QueueOrderCancel,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// 重试队列（TTL + DLX）
	_, err = ch.QueueDeclare(
		QueueOrderCancelRetry,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    ExchangeOrder,
			"x-dead-letter-routing-key": "cancel",
		},
	)
	if err != nil {
		return err
	}

	// bind
	return ch.QueueBind(
		QueueOrderCancel,
		"cancel",
		ExchangeOrder,
		false,
		nil,
	)
}