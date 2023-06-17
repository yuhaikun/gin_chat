package models

import (
	"fmt"
	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"sync"
)

type Node struct {
	Lock          sync.Mutex
	Ch            *amqp.Channel
	PrivQueue     *amqp.Queue
	Conn          *websocket.Conn
	Addr          string        // 客户端地址
	FirstTime     uint64        //首次连接时间
	HeartbeatTime uint64        // 心跳时间
	LoginTime     uint64        //  登录时间
	DataQueue     chan []byte   // 消息
	GroupSets     set.Interface // 好友/群
}

// 映射关系
//var clientMap syncmap.Map

// 映射关系
//var clientMap map[int64]*Node = make(map[int64]*Node, 0)
//
//// 读写锁
//var rwLocker sync.RWMutex

type ClientMap struct {
	lock       sync.RWMutex
	connection map[int64]*Node
}

// 用户心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	//fmt.Println("节点时间为：", node.HeartbeatTime)
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		fmt.Println("心跳超时。。。自动下线", node)
		timeout = true
	}
	return
}
