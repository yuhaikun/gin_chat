package main

import (
	"gin_chat/router"
	"gin_chat/utils"
	"github.com/spf13/viper"
	"log"
)

func main() {
	utils.InitConfig()
	err := utils.InitRedis()
	if err != nil {
		return
	}
	err = utils.InitMysql()
	if err != nil {
		return
	}
	err = utils.InitRabbitMQ()
	if err != nil {
		// 处理连接错误
		log.Println("Failed to initialize RabbitMQ:", err)
		return
	}
	//defer utils.MQ.Close()
	r := router.Router()

	r.Run(viper.GetString("port.server"))
}
