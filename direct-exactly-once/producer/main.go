package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func main() {
	servers := []string{
		"amqp://admin:111111@localhost:5672/",
		"amqp://admin:111111@localhost:5673/",
		"amqp://admin:111111@localhost:5674/",
	}
	conn, err := connect(servers)
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

	confirm := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

	// confirm并行发送ack失败后不能确定出问题的是哪一条消息，所以最好串行发，每发一条确认一条
	// 发送消息
	for i := 0; i < 10; i++ {
		message := Message{
			ID:      fmt.Sprintf("%d", i+1),          //消息唯一ID, 给消费端做幂等用
			Content: fmt.Sprintf("Message #%d", i+1), //消息内容
		}

		body, err := json.Marshal(message)
		if err != nil {
			log.Printf("Error marshalling message: %v", err)
			return
		}

		// 发一条
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

		// 确认一条
		select {
		case confirmResult := <-confirm:
			if confirmResult.Ack {
				log.Println("消息确认成功", string(body))
			} else {
				log.Println("消息确认失败", string(body))
				// TODO 这里你可以发送到死信队列 或做业务层面重试逻辑 或其他兜底机制
			}
		case <-time.After(3 * time.Second):
			log.Println("消息确认超时", string(body))
			// TODO 这里你可以发送到死信队列 或做业务层面重试逻辑 或其他兜底机制
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func connect(servers []string) (*amqp.Connection, error) {
	var errors []string
	for _, server := range servers {
		conn, err := amqp.Dial(server)
		if err == nil {
			return conn, nil
		}
		errors = append(errors, err.Error())
	}
	return nil, fmt.Errorf("所有连接尝试失败:\n%s", strings.Join(errors, "\n"))
}
