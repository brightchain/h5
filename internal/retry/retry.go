package retry

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"h5/internal/mq"
)

var delays = []int{5, 10, 20}
const maxRetry = 3

func Retry(ch *amqp.Channel, body []byte, retryCount int) {
	if retryCount >= maxRetry {
		log.Println("Retry exceeded, give up")
		return
	}

	var data map[string]interface{}
	_ = json.Unmarshal(body, &data)

	data["retry_count"] = retryCount + 1
	data["last_retry_time"] = time.Now().Unix()

	delay := delays[retryCount]

	msg, _ := json.Marshal(data)

	ch.Publish(
		"",
		mq.QueueOrderCancelRetry,
		false,
		false,
		amqp.Publishing{
			Body:       msg,
			Expiration: strconv.Itoa(delay * 1000),
		},
	)
}
