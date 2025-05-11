package main

import (
	"log"

	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://admin:111111@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	exchangeName := "direct_logs"
	queueName := "log_info"
	routingKey := "info" // direct交换机要求生产端的 routing key 和消费端的 binding key 要完全匹配

	// 声明交换机
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // kind 关键配置 交换机类型为direct 点对点收发消息
		true,         // durable
		false,        // autoDelete
		false,        // internal
		false,        // noWait
		nil,          // args
	)
	failOnError(err, "声明交换机失败")

	// 声明队列
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // autoDelete 临时队列可以设置成true
		false,     // exclusive
		false,     // noWait
		nil,       // args
	)
	failOnError(err, "声明队列失败")

	// 绑定队列到交换机
	err = ch.QueueBind(
		queueName,    // name
		routingKey,   // key 关键配置
		exchangeName, // exchange
		false,        // noWait
		nil,          // args
	)
	failOnError(err, "绑定队列失败")

	// 发送消息
	body := "Hello Direct!"
	err = ch.Publish(
		exchangeName, // exchange
		routingKey,   // key
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
