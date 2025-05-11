# RabbitMQ Demo

交换机类型及路由匹配规则

| Exchange 类型 | 生产端声明                 | 消费端声明                 | 路由匹配规则                              |
| ------------- | -------------------------- | -------------------------- | ----------------------------------------- |
| direct        | Exchange, Queue, QueueBind | Exchange, Queue, QueueBind | `binding key`要完全匹配`routing key`      |
| fanout        | Exchange                   | Exchange, Queue, QueueBind | 不需要`binding key`和`routing key`        |
| topic         | Exchange                   | Exchange, Queue, QueueBind | `bingding key`通过通配符匹配`routing key` |
