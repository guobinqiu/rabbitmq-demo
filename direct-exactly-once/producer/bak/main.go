package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func main() {
	conn, err := amqp.Dial("amqp://admin:111111@localhost:5672/")
	failOnError(err, "连接失败")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "打开通道失败")
	defer ch.Close()

	exchangeName := "sample_direct_eos_exchange"
	routingKey := "log.info" // direct 交换机要求消费端的 binding key 要完全匹配生产端的 routing key

	// 声明 direct 类型交换机
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // kind 关键配置 交换机类型为direct 点对点模式
		true,         // durable
		false,        // autoDelete
		false,        // internal
		false,        // noWait
		nil,          // args
	)
	failOnError(err, "声明交换机失败")

	// 开启 publisher confirm 模式
	err = ch.Confirm(false)
	failOnError(err, "开启 Confirm 模式失败")

	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 10))

	var wg sync.WaitGroup

	// 发送消息
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			message := Message{
				ID:      fmt.Sprintf("%d", i+1),          //消息唯一ID, 给消费端做幂等用
				Content: fmt.Sprintf("Message #%d", i+1), //消息内容
			}

			body, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshalling message: %v", err)
				return
			}

			err = ch.Publish(
				exchangeName, // exchange
				routingKey,   // key
				false,        // mandatory
				false,        // immediate
				amqp.Publishing{
					DeliveryMode: amqp.Persistent,
					ContentType:  "text/plain",
					Body:         body,
				}, // msg
			)
			failOnError(err, "发送消息失败")
			log.Printf(" [x] Sent %s", body)
		}(i)
	}

	// 监听并处理消息确认
	// 但是你不知道哪条失败了，所以最好串行发
	go func() {
		for confirm := range confirms {
			select {
			case <-time.After(3 * time.Second):
				// TODO 这里你可以发送到死信队列 或做业务层面重试逻辑 或其他兜底机制
			default:
				if confirm.Ack {
					log.Printf("Message %d sent successfully\n", confirm.DeliveryTag)
				} else {
					log.Printf("Message %d failed to deliver\n", confirm.DeliveryTag)
					// TODO 这里你可以发送到死信队列 或做业务层面重试逻辑 或其他兜底机制
				}
			}
		}
	}()

	wg.Wait()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
