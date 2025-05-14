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

## TODO List

- [x] 单机安装
- [x] 集群安装
- [x] direct
- [x] fanout
- [x] topic
- [x] simulate-kafka-multi-partiions 模拟kafka的多分区
- [x] direct-exactly-once topic 消息精确一次投递

## 消息传递语义

- 最多一次（At Most Once）
  - 每条消息最多被处理一次，**可能丢失，但不会重复**。
  - 发送消息后不等待确认，不做重试。
  - **性能最高，可靠性最低**。
  - 适用场景：日志采集、监控数据等可容忍丢失的业务。

- 最少一次（At Least Once）
  - 每条消息至少被处理一次，**不会丢失，但可能重复**。
  - 生产端或消费者失败会触发重试。
  - **需要消费者具备幂等性处理能力**（如去重逻辑）。
  - 适用场景：数据采集、任务处理等对完整性要求较高但可容忍重复的业务。

- 精确一次（Exactly Once）
  - 每条消息被处理 **且仅处理一次**，**既不重复也不丢失**。
  - Kafka 提供端到端事务机制实现。
  - **性能开销大，实现复杂**。
  - 适用场景：订单系统、金融交易、库存更新等对一致性要求极高的系统。

RabbitMQ 里的精确一次性要点如下:

- 生产端
  - 启用`publisher confirms`
  - 设置队列(`durable`)和消息(`DeliveryMode`)为持久化
  - 增加失败重试机制(如死信队列)
  - 处理反馈`Ack`的超时
- 消费端
  - 消费逻辑`幂等`(通过唯一的消息ID做去重)
  - 设置`autoAck=false`并手动确认`Ack`
  - 增加失败重试机制`Nack(false, true)`

## Install

### docker

单节点配置

```
services:
  rabbitmq:
    image: "rabbitmq:3-management" # 使用带有管理插件的镜像
    container_name: "rabbitmq"
    ports:
      - "5672:5672" # AMQP 端口
      - "15672:15672" # 管理界面端口
    environment:
      RABBITMQ_DEFAULT_USER: "admin" # 设置默认用户
      RABBITMQ_DEFAULT_PASS: "111111" # 设置默认密码
```

集群配置 (高可用)

```
services:
  rabbitmq1:
    image: "rabbitmq:3-management" # 使用带有管理插件的镜像
    container_name: "rabbitmq1"
    ports:
      - "5672:5672" # AMQP 端口
      - "15672:15672" # 管理界面端口
    environment:
      RABBITMQ_DEFAULT_USER: "admin" # 设置默认用户
      RABBITMQ_DEFAULT_PASS: "111111" # 设置默认密码
      RABBITMQ_ERLANG_COOKIE: "secretcookie"
      RABBITMQ_NODENAME: "rabbit@rabbitmq1"

  rabbitmq2:
    image: "rabbitmq:3" # 使用不带管理插件的镜像
    container_name: "rabbitmq2"
    ports:
      - "5673:5672"
    environment:
      RABBITMQ_DEFAULT_USER: "admin"
      RABBITMQ_DEFAULT_PASS: "111111"
      RABBITMQ_ERLANG_COOKIE: "secretcookie"
      RABBITMQ_NODENAME: "rabbit@rabbitmq2"
    depends_on:
      - rabbitmq1

  rabbitmq3:
    image: "rabbitmq:3" # 使用不带管理插件的镜像
    container_name: "rabbitmq3"
    ports:
      - "5674:5672"
    environment:
      RABBITMQ_DEFAULT_USER: "admin"
      RABBITMQ_DEFAULT_PASS: "111111"
      RABBITMQ_ERLANG_COOKIE: "secretcookie"
      RABBITMQ_NODENAME: "rabbit@rabbitmq3"
    depends_on:
      - rabbitmq1
```

进入 rabbitmq2 容器运行命令

```
docker exec -it rabbitmq2 bash
rabbitmqctl stop_app
rabbitmqctl join_cluster rabbit@rabbitmq1
rabbitmqctl start_app
```

进入 rabbitmq3 容器运行命令

```
docker exec -it rabbitmq3 bash
rabbitmqctl stop_app
rabbitmqctl join_cluster rabbit@rabbitmq1
rabbitmqctl start_app
```

验证节点是否成功加入集群

```
rabbitmqctl cluster_status
```

配置镜像队列策略(Mirrored Queue)

```
rabbitmqctl set_policy ha-all "^mirror\." '{"ha-mode":"all"}' --apply-to queues
```

只是做备份和故障转移，性能不会提高，读写只能在激活的`master`节点上

查看已设置的策略

```
rabbitmqctl list_policies
```

## 管理界面

http://localhost:15672

用户名admin
密码111111

## 它的邻居Kafka

> https://github.com/guobinqiu/kafka-demo