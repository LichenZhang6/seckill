package main

import "seckill/RabbitMQ"

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQPubSub("newProduct")
	rabbitmq.RecieveSub()
}
