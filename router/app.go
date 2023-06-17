package router

import (
	"gin_chat/docs"
	"gin_chat/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()

	// swagger
	//docs.SwaggerInfo.BasePath = "" 将基础路径设置为空，这意味着 API 的根路径将是域名或 IP 地址。例如，如果您的域名是 example.com，那么 API 的根路径将是 example.com。
	//r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) 创建了一个 GET 路由，用于访问 Swagger UI。在这里，/swagger/*any 是通配符路由，它将匹配任何以 /swagger/ 开头的路径。
	//ginSwagger.WrapHandler(swaggerFiles.Handler) 将 Swagger UI 的处理程序包装成 Gin 处理程序，以便在路由中使用。这将自动提供 Swagger UI 界面，允许用户浏览 API 文档。
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 静态资源
	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("views/**/*")

	// 首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)

	r.POST("/searchFriends", service.SearchFriends)

	r.GET("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.GET("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	r.POST("/user/find", service.FindByID)

	// 发送消息
	r.GET("/user/sendMsg", service.SendMsg)
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	// 上传文件
	r.POST("/attach/upload", service.Upload)
	// 添加好友
	r.POST("/contact/addfriend", service.AddFriend)
	// 创建群
	r.POST("/contact/createCommunity", service.CreateCommunity)
	// 群列表
	r.POST("/contact/loadCommunity", service.LoadCommunity)
	// 加群
	r.POST("/contact/joinGroup", service.JoinGroup)

	// 心跳续命 不合适 因为node 所以前端发过来的消息在recvProc里面处理
	// r.POST("/user/heartbeart",service.Heartbeat)
	r.POST("/user/redisMsg", service.RedisMsg)

	r.POST("/joinGroupChat", service.JoinGroupChat)
	return r

}
