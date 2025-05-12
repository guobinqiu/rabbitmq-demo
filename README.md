# RabbitMQ Demo

- 生产者
  - 声明`Exchange`
  - 指定消息的`Exchange`
  - 指定消息的`routing key`
- 消费者
  - 声明`Exchange`
  - 声明`Queue`
  - 设置`QueueBind`
    - 将`Queue`绑定到`Exchange`
    - 设置`binding key`

| Exchange 类型 | 设置binding key                           |
| ------------- | ----------------------------------------- |
| direct        | `binding key`要完全匹配`routing key`      |
| fanout        | 不需要`binding key`和`routing key`的参与  |
| topic         | `bingding key`通过通配符匹配`routing key` |

> 如果你希望在没有消费者运行时也能让生产者“预发消息”，就需要生产者在发送前确保 Exchange、Queue 都已经存在并绑定好。你可以在生产端主动声明 Queue 并绑定，这样消息不会丢失，而是会暂存在队列中，等消费者上线后再被消费。

此时生产端将如下设置
```
ch.ExchangeDeclare(...)
ch.QueueDeclare(...)
ch.QueueBind(...)
ch.Publish(...)
```

TODO List

- [x] 单机安装
- [ ] 集群安装
- [x] direct
- [x] fanout
- [x] topic
- [x] simulate-kafka-multi-partiions 模拟kafka的多分区
- [ ] topic-base-exactly-once topic 消息精确一次投递
