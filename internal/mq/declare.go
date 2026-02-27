package mq

import  amqp "github.com/rabbitmq/amqp091-go"

const (
	ExchangeOrder = "order_exchange"
	ExchangeSms   = "sms_exchange"

	QueueOrderCancel      = "order_cancel_queue"
	QueueOrderCancelRetry = "order_cancel_retry_queue"
	QueueSmsNotify        = "sms_queue"
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

	// SMS 通知队列
	if err := ch.ExchangeDeclare(
		ExchangeSms,
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
	if err := ch.QueueBind(
		QueueOrderCancel,
		"cancel",
		ExchangeOrder,
		false,
		nil,
	); err != nil {
		return err
	}

	// SMS 通知队列
	_, err = ch.QueueDeclare(
		QueueSmsNotify,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// bind SMS 队列
	return ch.QueueBind(
		QueueSmsNotify,
		"send",
		ExchangeSms,
		false,
		nil,
	)
}