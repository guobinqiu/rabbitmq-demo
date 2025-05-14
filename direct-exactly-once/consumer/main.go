package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
	queueName := "sample_direct_eos_queue"
	bindingKey := "log.info" // direct 交换机要求消费端的 binding key 要完全匹配生产端的 routing key

	// 声明 direct 类型交换机
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"direct",     // kind 关键配置 交换机类型为 direct 点对点模式
		true,         // durable
		false,        // autoDelete
		false,        // internal
		false,        // noWait
		nil,          // args
	)
	failOnError(err, "声明交换机失败")

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

	// 初始化 SQLite 数据库
	db, err := sql.Open("sqlite3", "./consumer_state.db")
	if err != nil {
		failOnError(err, "Failed to open database")
	}
	defer db.Close()

	// 创建幂等性校验表（如果不存在）
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS consumed_messages (
		id TEXT PRIMARY KEY
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		failOnError(err, "Failed to create table")
	}

	// 处理消息
	for msg := range msgCh {
		var m Message

		if err := json.Unmarshal(msg.Body, &m); err != nil {
			log.Printf("JSON error: %v", err)
			_ = msg.Nack(false, false) // 丢弃
			continue
		}

		// ID 作为唯一标识
		if m.ID == "" {
			log.Println("Empty message ID, skipped")
			_ = msg.Nack(false, false) // 丢弃
			continue
		}

		// 幂等校验
		var existingID string
		err := db.QueryRow(`SELECT id FROM consumed_messages WHERE id = ?`, m.ID).Scan(&existingID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("DB query error: %v", err)
			_ = msg.Nack(false, true) // 重新入队尾
			continue
		}
		if existingID != "" {
			log.Printf("Duplicate message skipped (ID=%s, content=%s)", m.ID, m.Content)
			_ = msg.Ack(false) // 很关键，告诉 rabbit 这条处理完了
			continue
		}

		// 开启事务
		tx, err := db.Begin()
		if err != nil {
			log.Printf("Failed to begin transaction: %v", err)
			_ = msg.Nack(false, true)
			continue
		}

		// 处理消息
		log.Printf("Consumed message: %s", m.Content)
		time.Sleep(time.Second)

		// 处理成功后再写入幂等表
		_, err = tx.Exec(`INSERT INTO consumed_messages (id) VALUES (?)`, m.ID)
		if err != nil {
			log.Printf("DB insert error: %v", err)
			_ = tx.Rollback()
			_ = msg.Nack(false, true)
			continue
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			log.Printf("Commit error: %v", err)
			_ = msg.Nack(false, true)
			continue
		}

		// 消息确认
		_ = msg.Ack(false)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
