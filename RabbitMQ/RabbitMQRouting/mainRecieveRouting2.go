package main

import "seckill/RabbitMQ"

func main() {
	rabbitmq2 := RabbitMQ.NewRabbitMQRouting("routingMQ", "routing2")
	rabbitmq2.RecieveRouting()
}
