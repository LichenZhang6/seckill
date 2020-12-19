package main

import "seckill/RabbitMQ"

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("simpleMQ")
	rabbitmq.ConsumeSimple()
}