package models

import (
	"encoding/json"
	"fmt"
	"gin_chat/utils"
	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Message struct {
	gorm.Model
	UserId     int64  // 发送者
	TargetId   int64  // 接收者
	Type       int64  // 发送类型  1.私聊  2.群聊 3.广播
	Media      int    // 消息类型  1.文字  2.表情包 3.图片 3.音频
	Content    string // 消息内容
	CreateTime uint64 // 创建时间
	ReadTime   uint64 // 读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

var clientMap *ClientMap

var queuePriName string

func init() {
	clientMap = &ClientMap{
		lock:       sync.RWMutex{},
		connection: make(map[int64]*Node, 0),
	}

}
func Chat1(writer http.ResponseWriter, request *http.Request) {
	//	1.获取参数并检验token 等合法性
	query := request.URL.Query()
	id := query.Get("userId")
	userId, _ := strconv.ParseInt(id, 10, 64)
	queuePriName = fmt.Sprintf("queue_%d", userId)
	//if err2 != nil {
	//	panic(err2)
	//}
	//msgType := query.Get("type")
	//targetId := query.Get("targetId")
	//context := query.Get("context")

	isValid := true // checkToken()
	conn, err := (&websocket.Upgrader{
		//	token校验
		CheckOrigin: func(r *http.Request) bool {
			return isValid
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//	2.获取conn
	currentTime := uint64(time.Now().Unix())
	ch1, _ := utils.MQ.Channel()
	privQueue, err := ch1.QueueDeclare(
		queuePriName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %s", err)
	}

	node := &Node{
		Lock:          sync.Mutex{},
		Ch:            ch1,
		PrivQueue:     &privQueue,
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(),
		HeartbeatTime: currentTime,
		LoginTime:     currentTime,
		DataQueue:     make(chan []byte),
		GroupSets:     set.New(set.ThreadSafe),
	}
	//	3. 用户关系
	//	4. userid和node绑定起来并加锁
	//rwLocker.Lock()
	//clientMap[userId] = node
	//rwLocker.Unlock()
	clientMap.lock.Lock()
	clientMap.connection[userId] = node
	clientMap.lock.Unlock()
	//clientMap.Store(userId, node)
	//	5. 完成发送逻辑
	//go sendProc(node)
	go func() {
		for {
			err = ReceivePrivateMessage(node, userId)
			if err != nil {
				// 处理消费者出现的错误，可以进行日志记录或其他处理方式
				log.Println("Consumer error:", err)
				return
			}
		}
	}()
	//	6. 完成接收逻辑
	go recvProc1(node)
	// 单独用一个协程来检测心跳是否超时
	go ClearConnection1(node)

	//sendMsg(userId, []byte("欢迎进入聊天系统:"))
	// 7. 加入在线用户到缓存
	SetUserOnlineInfo("online_"+id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
}

func recvProc1(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println(err)
		}
		// 心跳检测 msg.Media == -1 || msg.Type == 3
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
			fmt.Println("[ws]<<<<<< ", string(data))
		} else {
			dispatch2(node, data)
			//broadMsg(data) //todo消息广播到局域网
			fmt.Println("[ws]<<<<<< ", string(data))
		}
	}
}
func dispatch2(node *Node, data []byte) (err error) {
	msg := Message{}
	err = json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch msg.Type {
	case 1: //私信
		handlePrivateMessage(msg.TargetId, data)
	case 2: //群发
		handGroupMessage(node, msg.UserId, msg.TargetId, data)
		//	sendGroupMsg()
		//case 3: //广播
		//	sendAllMsg()
	}
	return
}

// 清理超时连接
func ClearConnection1(node *Node) {

	for {
		// 每10s进行一次心跳检测
		time.Sleep(time.Second * 10)

		currentTime := uint64(time.Now().Unix())
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时。。。自动下线", node)
			node.Conn.Close()
			close(node.DataQueue)
			return
		}
	}

}

func handlePrivateMessage(targetId int64, msg []byte) {

	clientMap.lock.Lock()
	sendNode := clientMap.connection[targetId]
	clientMap.lock.Unlock()

	//initCh := make(chan struct{})
	//go func() {
	//	InitPrivateMQ(node, sendNode, userId, targetId)
	//	close(initCh)
	//}()
	//<-initCh
	SendPrivateMessage(sendNode, targetId, msg)
	//go func() {
	//	for {
	//		err := ReceivePrivateMessage(node, userId)
	//		if err != nil {
	//			// 处理消费者出现的错误，可以进行日志记录或其他处理方式
	//			log.Println("Consumer error:", err)
	//			return
	//		}
	//	}
	//}()

}
func handGroupMessage(node *Node, userId, targetId int64, msg []byte) {

	//initCh := make(chan struct{})
	//go func() {
	//	InitGroupMQ(targetId)
	//	close(initCh)
	//}()
	//
	//<-initCh
	//go SendGroupMessage(msg)
	//
	//go func() {
	//	for {
	//		err := ReceiveGroupMessage(node)
	//		if err != nil {
	//			// 处理消费者出现的错误，可以进行日志记录或其他处理方式
	//			log.Println("Consumer error:", err)
	//		}
	//	}
	//}()
	//joinGroupChat(targetId)
	sendGroupMessage(targetId, msg)

}
