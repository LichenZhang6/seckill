package main

import (
	"fmt"
	"seckill/RabbitMQ"
	"strconv"
	"time"
)

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQPubSub("newProduct")

	for i := 1; i <= 100; i++ {
		rabbitmq.PublishPub(fmt.Sprintf("this is %v message", strconv.Itoa(i)))
		fmt.Println(fmt.Sprintf("this is %v message", strconv.Itoa(i)))
		time.Sleep(1 * time.Second)
	}
}
