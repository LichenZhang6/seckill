package main

import (
	"fmt"
	"seckill/RabbitMQ"
)

func  main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("simpleMQ")
	rabbitmq.PublishSimple("Hello world!")
	fmt.Println("sent successfully!")
}
