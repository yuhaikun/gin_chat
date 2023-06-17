package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var (
	DB       *gorm.DB
	Red      *redis.Client
	MQ       *amqp.Connection
	PrivChan *amqp.Channel
)

//type RabbitMQ struct {
//	User     string `mapstructure:"user"`
//	Password string `mapstructure:"password"`
//	Addr     string `mapstructure:"addr"`
//}

func InitConfig() {
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app invited")
	//fmt.Println("config mysql:", viper.Get("mysql"))
}
func InitMysql() (err error) {
	//自定义日志模板 打印SQL语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, //慢SQl阈值
			LogLevel:      logger.Info, //级别
			Colorful:      true,        //彩色
		},
	)

	DB, err = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	if err != nil {
		return
	}
	//user := models.UserBasic{}
	//DB.Find(&user)
	//fmt.Println(user)

	return nil
}

func InitRedis() (err error) {
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConn"),
	})
	pong, err := Red.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("init redis....", err)
	} else {
		fmt.Println("Redis inited....", pong)
	}
	return nil
}

func InitRabbitMQ() (err error) {
	dns := viper.GetString("rabbitmq.dns")
	MQ, err = amqp.DialConfig(dns, amqp.Config{
		Heartbeat: 10 * time.Second,
	})
	if err != nil {
		fmt.Println("init rabbitmq...", err)
	}

	PrivChan, err = MQ.Channel()
	if err != nil {
		fmt.Println("init rabbitmq Channel", err)
	}
	return
}

const (
	PublishKey = "websocket"
)

// Publish 发布消息到Redis
func Publish(ctx context.Context, channel string, msg string) (err error) {
	err = Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println(err)
	}
	return
}

// Subscribe 订阅Redis消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)
	fmt.Println("Subscribe ....", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	fmt.Println("消息1111为：", msg)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	fmt.Println("Subscribe ....", msg.Payload)
	return msg.Payload, err

}
