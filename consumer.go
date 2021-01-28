package main

import (
	"fmt"
	"seckill/common"
	"seckill/rabbitmq"
	"seckill/repositories"
	"seckill/services"
)

func main() {
	db, err := common.NewMysqlConn()
	if err != nil {
		fmt.Println(err)
	}
	// 创建product数据库操作实例
	product := repositories.NewProductManager("product", db)
	// 创建product service
	productService := services.NewProductService(product)
	// 创建order数据库
	order := repositories.NewOrderManager("order", db)
	// 创建order service
	orderService := services.NewOrderService(order)

	rabbitmqConsumeSimple := rabbitmq.NewRabbitMQSimple("product")
	rabbitmqConsumeSimple.ConsumeSimple(orderService, productService)
}
