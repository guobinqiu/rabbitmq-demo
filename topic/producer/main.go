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

	exchangeName := "sample_topic_exchange"
	routingKey := "log.info" // topic交换机要求消费端的 binding key 通配符匹配生产端的 routing key

	// 声明 topic 类型交换机
	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "声明交换机失败")

	// 发送消息
	body := "Hello Topic!"
	err = ch.Publish(
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		},
	)
	failOnError(err, "发送消息失败")
	log.Printf(" [x] Sent %s", body)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
