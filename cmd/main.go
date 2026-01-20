package main

import (
	"flag"
	"log"

	"h5/bootstrap"
	"h5/config"
	"h5/internal/consumer"
)

func init() {
	config.Initialize()
}

func main() {
	bootstrap.SetupSlog()
	bootstrap.SetupModel()
	once := flag.Bool("once", false, "consume once and exit")
	flag.Parse()

	log.Println("OrderCancelConsumer start")
	err := consumer.ConsumeOrderCancel(*once)
	if err != nil {
		log.Fatal(err)
	}
}
