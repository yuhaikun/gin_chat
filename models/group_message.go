package models

import (
	"context"
	"encoding/json"
	"fmt"
	"gin_chat/utils"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
	"log"
	"strconv"
)

//const (
//	ExchangePrefix = "group_"
//	QueuePrefix    = "group_queue_"
//)
//
//var ch *amqp.Channel
//var exchangeName string
//var queueName string
//
//func InitGroupMQ(targetId int64) {
//	var err error
//	ch, err = utils.MQ.Channel()
//	if err != nil {
//		log.Fatalf("Failed to open a channel: %s", err)
//	}
//	exchangeName = ExchangePrefix + strconv.Itoa(int(targetId))
//	queueName = QueuePrefix + strconv.Itoa(int(targetId))
//
//	// 声明交换机
//	err = ch.ExchangeDeclare(
//		exchangeName,
//		amqp.ExchangeFanout,
//		true,
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Println("Failed to declare exchange:", err)
//		return
//	}
//	// 创建队列
//	_, err = ch.QueueDeclare(
//		queueName,
//		true,
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Fatalf("Failed to declare a queue: %s", err)
//	}
//	// 将队列绑定到交换机
//	err = ch.QueueBind(
//		queueName,
//		"",
//		exchangeName,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Fatalf("Failed to bind a queue to the exchange: %s", err)
//	}
//}
//
//// 生产者
//func SendGroupMessage(msg []byte) (err error) {
//
//	// 发送消息到群
//	err = ch.Publish(
//		exchangeName,
//		"",
//		false,
//		false,
//		amqp.Publishing{
//			ContentType: "text/plain",
//			Body:        msg,
//		},
//	)
//	if err != nil {
//		log.Println("Failed to publish message:", err)
//	}
//	return
//}
//
//func ReceiveGroupMessage(node *Node) (err error) {
//
//	// 消费消息
//	msgs, err := ch.Consume(
//		queueName,
//		"",
//		true,
//		false,
//		false,
//		false,
//		nil,
//	)
//	if err != nil {
//		log.Println("Failed to consume messages:", err)
//		return
//	}
//	// 处理接收到的私聊消息
//
//	for message := range msgs {
//		// 处理收到的私聊消息
//
//		err = node.Conn.WriteMessage(websocket.TextMessage, message.Body)
//		if err != nil {
//			log.Println("Failed to send message to websocket:", err)
//		}
//	}
//	return
//}

type ChatRoom struct {
	Channel *amqp.Channel
	Queue   amqp.Queue
}

const preFixExchangeName = "group_"

var chatRooms = make(map[int64]*ChatRoom)

// 创建群聊通道和队列，并绑定到交换机
func createChatRoom(roomId int64) *ChatRoom {
	ch, err := utils.MQ.Channel()
	if err != nil {
		log.Fatalf("Failed to create channel:%v", err)
	}
	queueExchangeName := preFixExchangeName + strconv.Itoa(int(roomId))
	// 创建交换机
	err = ch.ExchangeDeclare(
		queueExchangeName,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	queue, err := ch.QueueDeclare(
		"room_queue_"+strconv.Itoa(int(roomId)),
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name,
		"",
		queueExchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue to exchange: %v", err)
	}

	return &ChatRoom{
		Channel: ch,
		Queue:   queue,
	}
}

// 加入群聊
func JoinGroupChat(roomId int64, userId int64) {
	chatRoom, exits := chatRooms[roomId]
	if !exits {
		// 创建新的群聊通道和队列
		chatRoom = createChatRoom(roomId)
		chatRooms[roomId] = chatRoom
	}
	log.Printf("joined the group chat %d", roomId)

	consumer, err := chatRoom.Channel.Consume(
		chatRoom.Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create consumer:%v", err)
	}

	// 处理接收到群聊消息
	go func() {
		for delivery := range consumer {
			msg := Message{}
			err = json.Unmarshal(delivery.Body, &msg)
			if err != nil {
				fmt.Println(err)
				return
			}
			sendId := msg.UserId
			log.Printf("%d Received message in group chat %d: %v", userId, roomId, msg)

			userIds := SearchUserByGroupId(uint(roomId))
			for i := 0; i < len(userIds); i++ {
				// 排除给自己的
				if sendId != int64(userIds[i]) {
					sendMessage(int64(userIds[i]), delivery.Body)
				}
			}
		}
	}()
}
func sendMessage(userId int64, msg []byte) {
	clientMap.lock.RLock()
	node, ok := clientMap.connection[userId]
	clientMap.lock.RUnlock()

	ctx := context.Background()
	r, err := utils.Red.Get(ctx, "online_"+strconv.Itoa(int(userId))).Result()
	if err != nil {
		log.Fatalf("获取在线状态错误：%v", err)
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>>> userID: ", userId, " msg:", string(msg))
			node.Conn.WriteMessage(websocket.TextMessage, msg)
		}
	}

}

// 发送群聊消息
func sendGroupMessage(roomId int64, message []byte) {
	chatRoom, exists := chatRooms[roomId]
	if !exists {
		// 创建新的群聊通道和队列
		chatRoom = createChatRoom(roomId)
		chatRooms[roomId] = chatRoom
		//log.Printf("Group chat %d does not exist", roomId)
		//return
	}
	queueExchangeName := preFixExchangeName + strconv.Itoa(int(roomId))
	err := chatRoom.Channel.Publish(
		queueExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			Body:         message,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish message: %v", err)
	}
	log.Printf("Sent message in group chat %d: %s", roomId, string(message))
}
