package main

import (
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

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

	var wg sync.WaitGroup
	wg.Add(partitionNum)

	// 每个队列由独立消费者处理，并行消费
	for i := 0; i < partitionNum; i++ {
		queueName := fmt.Sprintf("sample_simulate_kafka_queue%d", i)
		bindingKey := fmt.Sprintf("partition.%d", i) // direct 交换机要求消费端的 binding key 要完全匹配生产端的 routing key

		// 声明队列
		q, err := ch.QueueDeclare(
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
			bindingKey,   // key 关键配置
			exchangeName, // exchange
			false,        // noWait
			nil,          // args
		)
		failOnError(err, "绑定队列失败")

		// 接收消息
		msgCh, err := ch.Consume(
			q.Name, // queue
			"",     // consumer 让RabbitMQ自动生成一个唯一的消费者标签
			false,  // autoAck 消费者手动提交
			false,  // exclusive
			false,  // noLocal
			false,  // noWait
			nil,    // args
		)
		failOnError(err, "注册消费者失败")

		// 处理消息
		go func() {
			defer wg.Done()
			for msg := range msgCh {
				log.Printf("收到消息: %s, queue: %s, routing key: %s ", msg.Body, queueName, bindingKey)
				// TODO 消费端处理逻辑
				msg.Ack(false)
			}
		}()
	}

	wg.Wait()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
