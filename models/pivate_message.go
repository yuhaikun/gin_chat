package models

import (
	"fmt"
	"gin_chat/utils"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"log"
)

// var priCh *amqp.Channel
//var exchangePriName string
//var queueSendPriName string
//var queueRecvPriName string

//func InitPrivateMQ(node *Node, sendNode *Node, userId, targetId int64) {
//	//var err error
//	//priCh, err = utils.MQ.Channel()
//	//if err != nil {
//	//	log.Println("Failed to open a channel:", err)
//	//	return
//	//}
//	// 打印通道状态
//	if utils.MQ.IsClosed() == true {
//		utils.InitRabbitMQ()
//	}
//	log.Println("connection:", utils.MQ.IsClosed())
//	log.Println("66666")
//	queueSendPriName = fmt.Sprintf("queue_%d", targetId)
//	queueRecvPriName = fmt.Sprintf("queue_%d", userId)
//	// 创建队列
//	_, err := sendNode.Ch.QueueDeclare(
//		queueSendPriName,
//		true,
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Fatalf("Failed to declare a queue: %s", err)
//	}
//	// 创建队列
//	_, err = node.Ch.QueueDeclare(
//		queueRecvPriName,
//		true,
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Fatalf("Failed to declare a queue: %s", err)
//	}
//}

func SendPrivateMessage(node *Node, targetId int64, msg []byte) {

	queueSendPriName := fmt.Sprintf("queue_%d", targetId)
	var err error
	// 发送消息到私聊队列
	err = node.Ch.Publish(
		"",
		queueSendPriName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msg,
		},
	)
	if err != nil {
		log.Println("Failed to publish message:", err)
		return
	}
	return
}

func ReceivePrivateMessage(node *Node, userId int64) (err error) {
	//if utils.PrivChan == nil {
	//	log.Println("Channel is nil")
	//	// 执行适当的错误处理或恢复操作
	//	return
	//}
	queueRecvPriName := fmt.Sprintf("queue_%d", userId)
	if utils.MQ.IsClosed() == true {
		utils.InitRabbitMQ()
	}
	// 消费消息
	msgs, err := node.Ch.Consume(
		queueRecvPriName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to consume messages:", err)
		log.Println("connection:", utils.MQ.IsClosed())

		//// 尝试重新初始化 RabbitMQ 连接
		//err = utils.InitRabbitMQ()
		//if err != nil {
		//	log.Println("Failed to initialize RabbitMQ connection:", err)
		//	// 执行适当的错误处理或恢复操作
		//}
		return
	}
	// 处理接收到的私聊消息

	for message := range msgs {
		// 处理收到的消息
		node.Lock.Lock()
		err = node.Conn.WriteMessage(websocket.TextMessage, message.Body)
		node.Lock.Unlock()

		if err != nil {
			log.Println("Failed to send message to websocket:", err)
		}
	}

	return
}
