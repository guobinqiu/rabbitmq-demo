# RabbitMQ Demo

默认情况下，RabbitMQ如果没有消费者绑定队列，消息会直接被丢弃，所以最好生产端和消费端都完整的声明以下三项:

- Exchange
- Queue
- QueueBind (绑定 Queue 到 Exchange)
