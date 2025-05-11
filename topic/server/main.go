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
	queueName := "sample_topic_queue"
	bindingKey := "log.*" // 匹配 log.info、log.error 等

	// 声明交换机
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

	// 声明队列
	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "声明队列失败")

	// 绑定队列到交换机，使用通配符匹配 routing key
	err = ch.QueueBind(
		queueName,
		bindingKey,
		exchangeName,
		false,
		nil,
	)
	failOnError(err, "绑定队列失败")

	// 消费消息
	msgCh, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "注册消费者失败")

	for msg := range msgCh {
		log.Printf("收到消息: %s", msg.Body)
		msg.Ack(false)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
