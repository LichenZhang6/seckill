package main

import (
	"fmt"
	"seckill/common"
	"seckill/datamodels"
)

func main() {
	data := map[string]string{
		"ID":           "1",
		"productName":  "test",
		"productNum":   "1",
		"productImage": "test.com",
		"productUrl":   "test.com",
	}
	product := &datamodels.Product{}
	common.DataToStructByTagSql(data, product)
	fmt.Println(product)
}
