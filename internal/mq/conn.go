package mq

import (
	"h5/pkg/config"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	conn *amqp.Connection
	once sync.Once
)

func GetConn()( *amqp.Connection, error) {
	var err error
	once.Do(func() {
		conn, err = amqp.Dial(
			config.GetString("mq.amqp.url"),
		)
	})
	return conn, err
}

