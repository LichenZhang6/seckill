package main

import "seckill/RabbitMQ"

func main() {
	rabbitmq1 := RabbitMQ.NewRabbitMQRouting("routingMQ", "routing1")
	rabbitmq1.RecieveRouting()
}
