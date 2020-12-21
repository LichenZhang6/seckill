package main

import (
	"fmt"
	"seckill/RabbitMQ"
	"strconv"
	"time"
)

func main() {
	rabbitmq := RabbitMQ.NewRabbitMQSimple("simpleMQ")

	for i := 1; i <= 100; i++ {
		rabbitmq.PublishSimple("Hello World " + strconv.Itoa(i))
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}
}
