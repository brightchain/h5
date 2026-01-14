package main

import (
	"flag"
	"log"

	"h5/internal/consumer"
)

func main() {
	once := flag.Bool("once", false, "consume once and exit")
	flag.Parse()

	log.Println("OrderCancelConsumer start")
	consumer.ConsumeOrderCancel(*once)
}
