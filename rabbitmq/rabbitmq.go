package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"seckill/datamodels"
	"seckill/services"
	"sync"
)

// url格式 amqp://账号:密码@rabbitmq服务器地址:端口号/vhost
const MQURL = "amqp://lic:password@127.0.0.1:5672/lic"

type RabbitMQ struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	QueueName string // 队列名称
	Exchange  string // 交换机名称
	Key       string // bind key 名称
	Mqurl     string // 连接信息
	sync.Mutex
}

//创建RAbbitMQ结构体实例
func NewRabbitMQ(queueName, exchange, key string) *RabbitMQ {
	return &RabbitMQ{QueueName: queueName, Exchange: exchange, Key: key, Mqurl: MQURL}
}

// 断开channel和connection
func (r *RabbitMQ) Destory() {
	r.channel.Close()
	r.conn.Close()
}

// 错误处理函数
func (r *RabbitMQ) failOnErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s:%s", message, err)
		panic(fmt.Sprintf("%s:%s", message, err))
	}
}

// 简单模式Step1：创建简单模式下RabbitMQ实例
func NewRabbitMQSimple(queueName string) *RabbitMQ {
	rabbitmq := NewRabbitMQ(queueName, "", "")
	var err error
	rabbitmq.conn, err = amqp.Dial(rabbitmq.Mqurl)
	rabbitmq.failOnErr(err, "failed to connect rabbitmq")
	rabbitmq.channel, err = rabbitmq.conn.Channel()
	rabbitmq.failOnErr(err, "failed to open a channel")
	return rabbitmq
}

// 简单模式Step2：简单模式下生产代码
func (r *RabbitMQ) PublishSimple(message string) error {
	r.Lock()
	defer r.Unlock()
	// 1.申请队列，如果队列不存在会自动创建，如果存在跳过创建
	// 保证队列存在，消息能发送到队列中
	_, err := r.channel.QueueDeclare(
		r.QueueName,
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否具有排他性
		false, // 是否阻塞
		nil,   // 额外属性
	)
	if err != nil {
		fmt.Println(err)
	}

	// 2.发送消息到队列中
	r.channel.Publish(
		r.Exchange,
		r.QueueName,
		false, // 如果true，根据exchange类型和routekey规则，无法找到符合条件的队列，则会将消息返还给发送者
		false, // 如果true，当exchange发送消息到队列后，发现队列没有绑定消费者，则会把消息发还给发送者
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)

	return nil
}

func (r *RabbitMQ) ConsumeSimple(OrderService services.IOrderService, productService services.IProductService) {

	// 1.申请队列，如果队列不存在会自动创建，如果存在跳过创建
	// 保证队列存在，消息能发送到队列中
	q, err := r.channel.QueueDeclare(
		r.QueueName,
		false, // 是否持久化
		false, // 是否自动删除
		false, // 是否具有排他性
		false, // 是否阻塞
		nil,   // 额外属性
	)
	if err != nil {
		fmt.Println(err)
	}

	r.channel.Qos(
		1,     // 当前消费者一次能接受的最大消息数量
		0,     // 服务器传递的最大容量（以8字节为单位）
		false, // 如果设置为true，对channel可用
	)

	// 接受消息
	msgs, err := r.channel.Consume(
		q.Name,
		"",    // 用来区分多个消费者
		false, //是否自动应答
		false, // 是否具有排他性
		false, // 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false, // 消息队列是否阻塞
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			// 实现我们要处理的逻辑函数
			log.Printf("Recieved a message: %s", d.Body)
			message := &datamodels.Message{}
			err := json.Unmarshal([]byte(d.Body), message)
			if err != nil {
				fmt.Println(err)
			}
			// 插入订单
			_, err = OrderService.InsertOrderByMessage(message)
			if err != nil {
				fmt.Println(err)
			}
			// 扣除商品shuliang
			err = productService.SubNumberOne(message.ProductID)
			if err != nil {
				fmt.Println(err)
			}
			// 如果为true表示确认所有未确认的消息，为false表示确认当前消息
			d.Ack(false)
		}
	}()

	log.Printf("[*] Waiting for messages, To exit press CTRL + C")
	<-forever
}
