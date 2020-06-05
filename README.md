# Go Chat

使用 Go 语言开发，kafka作为消息队列，websocket 进行客户端/服务端通信，然后使用vue来完成web界面显示。


使用方法，需要先将kafka服务启动，然后再运行代码。

#### 1. 下载kafka
```bash
$ wget http://apache.mirrors.hoobly.com/kafka/2.5.0/kafka_2.12-2.5.0.tgz
$ tar -xzf kafka_2.12-2.5.0.tgz
$ cd kafka_2.12-2.5.0
```

#### 2. 启动kafka服务
```bash
$ bin/zookeeper-server-start.sh config/zookeeper.properties # kafka 服务依赖于zookeeper
$ bin/kafka-server-start.sh config/server.properties
```

#### 3. 运行此Go程序
```
$ cd ./src
$ go run main.go
```

#### 4. 打开浏览器，访问 http://localhost:8000

#### 5. 效果图
![即时通讯效果图](./public/myGoChat.png)
