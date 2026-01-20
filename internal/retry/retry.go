package retry

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"h5/internal/mq"
	"h5/pkg/logger"

	amqp "github.com/rabbitmq/amqp091-go"
)

var delays = []int{5, 10, 20}
const maxRetry = 3

func Retry(ch *amqp.Channel, body []byte, retryCount int) {
	if retryCount >= maxRetry {
		logger.LogError(fmt.Sprintf("处理订单重试失败: %s", string(body)), nil)
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
