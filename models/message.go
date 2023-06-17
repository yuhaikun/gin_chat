package models

import (
	"context"
	"encoding/json"
	"fmt"
	"gin_chat/utils"
	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"strconv"
	"time"
)

//type Message struct {
//	gorm.Model
//	UserId     int64  // 发送者
//	TargetId   int64  // 接收者
//	Type       int64  // 发送类型  1.私聊  2.群聊 3.广播
//	Media      int    // 消息类型  1.文字  2.表情包 3.图片 3.音频
//	Content    string // 消息内容
//	CreateTime uint64 // 创建时间
//	ReadTime   uint64 // 读取时间
//	Pic        string
//	Url        string
//	Desc       string
//	Amount     int //其他数字统计
//}
//
//func (table *Message) TableName() string {
//	return "message"
//}

//var clientMap *ClientMap
//
//func init() {
//	clientMap = &ClientMap{
//		lock:       sync.RWMutex{},
//		connection: make(map[int64]*Node, 0),
//	}
//
//}

//	type Node struct {
//		Conn          *websocket.Conn
//		Addr          string        // 客户端地址
//		FirstTime     uint64        //首次连接时间
//		HeartbeatTime uint64        // 心跳时间
//		LoginTime     uint64        //  登录时间
//		DataQueue     chan []byte   // 消息
//		GroupSets     set.Interface // 好友/群
//	}
//
// // 映射关系
// var clientMap map[int64]*Node = make(map[int64]*Node, 0)
//
// // 读写锁
// var rwLocker sync.RWMutex

// Chat 需要: 发送者ID，接受者ID，消息类型，发送的内容，发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//	1.获取参数并检验token 等合法性
	query := request.URL.Query()
	id := query.Get("userId")
	userId, _ := strconv.ParseInt(id, 10, 64)
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
	node := &Node{
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
	go sendProc(node)
	//	6. 完成接收逻辑
	go recvProc(node)
	// 单独用一个协程来检测心跳是否超时
	go ClearConnection(node)

	//sendMsg(userId, []byte("欢迎进入聊天系统:"))
	// 7. 加入在线用户到缓存
	SetUserOnlineInfo("online_"+id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}
func recvProc(node *Node) {

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
			dispatch(data)
			broadMsg(data) //todo消息广播到局域网
			fmt.Println("[ws]<<<<<< ", string(data))
		}
		//dispatch(data)
		//broadMsg(data) //todo消息广播到局域网
		//fmt.Println("[ws]<<<<<< ", string(data))
	}
}

var udpsendChan = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func Init() {
	go udpSendProc()
	go udpRecvProc()
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255),
		Port: viper.GetInt("port.udp"),
	})
	defer con.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case data := <-udpsendChan:
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 完成udp数据接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: viper.GetInt("port.udp"),
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()

	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n])
	}
}

func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: //私信
		sendMsg(msg.TargetId, data)
	case 2: //群发
		sendGroupMsg(msg.TargetId, data)
		//	sendGroupMsg()
		//case 3: //广播
		//	sendAllMsg()
	}
}

func sendMsg(userId int64, msg []byte) {
	//rwLocker.RLock()
	//node, ok := clientMap[userId]
	//rwLocker.RUnlock()
	//if ok {
	//	node.DataQueue <- msg
	//}
	//rwLocker.RLock()
	//node, ok := clientMap[userId]
	////conn, ok := clientMap.Load(userId)
	////node := conn.(*Node)
	//rwLocker.RUnlock()
	clientMap.lock.RLock()
	node, ok := clientMap.connection[userId]
	clientMap.lock.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>>> userID: ", userId, " msg:", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, redis.Z{Score: score, Member: msg}).Result() //jsonMsg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		// 排除给自己的
		if targetId != int64(userIds[i]) {
			sendMsg(int64(userIds[i]), msg)
		}

	}
}

// 获取缓存里的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	//rwLocker.RLock()
	//node,ok := clientMap[userIdA]
	//rwLocker.RUnlock()
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}

	if err != nil {
		fmt.Println(err)
	}
	return rels
}

// 更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// 清理超时连接
func ClearConnection(node *Node) {

	//fmt.Println("1111111")
	for {
		//defer func() {
		//	//	recover函数用于捕获程序运行时的panic异常，并在函数内部进行处理，避免程序崩溃
		//	if r := recover(); r != nil {
		//		fmt.Println("cleanConection err", r)
		//	}
		//}()
		currentTime := uint64(time.Now().Unix())
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时。。。自动下线", node)
			node.Conn.Close()
			close(node.DataQueue)
			return
		}
		//for i := range clientMap {
		//	node := clientMap[i]
		//	if node.IsHeartbeatTimeOut(currentTime) {
		//		fmt.Println("心跳超时。。。自动下线", node)
		//		node.Conn.Close()
		//		return
		//	}
		//}
	}

}
