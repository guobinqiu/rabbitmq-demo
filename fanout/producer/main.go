package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://admin:111111@localhost:5672/")
	failOnError(err, "连接失败")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "打开通道失败")
	defer ch.Close()

	exchangeName := "sample_fanout_exchange"

	// 声明 fanout 类型交换机
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"fanout",     // kind 关键配置 交换机类型为fanout 发布订阅类型
		true,         // durable
		false,        // autoDelete
		false,        // internal
		false,        // noWait
		nil,          // args
	)
	failOnError(err, "声明交换机失败")

	// 发送消息
	body := "Hello Fanout!"
	err = ch.Publish(
		exchangeName, // exchange
		"",           // key fanout 不需要 routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		}, // msg
	)
	failOnError(err, "发送消息失败")
	log.Printf(" [x] Sent %s", body)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
