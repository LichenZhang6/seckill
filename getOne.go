package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	sum int64 = 0

	// 预存商品数量
	productNum int64 = 5

	// 互斥锁
	mutex sync.Mutex

	// 计数
	count int64 = 0
)

// 获取秒杀商品
func GetOneProduct() bool {
	// 加锁
	mutex.Lock()
	defer mutex.Unlock()
	count += 1
	// 判断数据是否超限
	if sum < productNum {
		sum += 1
		fmt.Println(sum)
		return true
	}
	return false
}

func GetProduct(w http.ResponseWriter, req *http.Request) {
	if GetOneProduct() {
		w.Write([]byte("true"))
		return
	}
	w.Write([]byte("false"))
	return
}

func main() {
	fmt.Println("预存商品数量", productNum)
	http.HandleFunc("/getOne", GetProduct)
	err := http.ListenAndServe(":8084", nil)
	if err != nil {
		log.Fatal("Error:", err)
	}
}
