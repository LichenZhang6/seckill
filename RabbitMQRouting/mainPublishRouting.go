package main

import (
	"fmt"
	"seckill/RabbitMQ"
	"strconv"
	"time"
)

func main() {
	rabbitmq1 := RabbitMQ.NewRabbitMQRouting("routingMQ", "routing1")
	rabbitmq2 := RabbitMQ.NewRabbitMQRouting("routingMQ", "routing2")
	for i := 1; i <= 100; i++ {
		rabbitmq1.PublishRouting("Hello from Routing MQ 1 " + strconv.Itoa(i))
		rabbitmq2.PublishRouting("Hello from Routing MQ 2 " + strconv.Itoa(i))
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}
}
