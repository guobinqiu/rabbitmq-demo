package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// 模拟 Kafka 的分区机制：
// 使用 direct 类型的 Exchange
// 创建多个 queue （如 queue-0、queue-1、queue-2）模拟 kafka 多分区
// 通过取余的方式实现 Round-Robin，将消息轮流发送到不同 routing key （如 partition.0、partition.1、partition.2）
// 假设发送 a,b,c 3条消息
// a 将通过 partition.0 进入 queue-0
// b 将通过 partition.1 进入 queue-1
// c 将通过 partition.2 进入 queue-2
func main() {
	conn, err := amqp.Dial("amqp://admin:111111@localhost:5672/")
	failOnError(err, "连接失败")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "打开通道失败")
	defer ch.Close()

	exchangeName := "sample_simulate_kafka_exchange"
	partitionNum := 3 // 设置分区数

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

	// 发送消息
	for i := range 10 {
		body := fmt.Sprintf("Hello Direct! - %d", i)
		routingKey := fmt.Sprintf("partition.%d", i%partitionNum) // direct 交换机要求消费端的 binding key 要完全匹配生产端的 routing key
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
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
