package service

import (
	"fmt"
	"gin_chat/models"
	"gin_chat/utils"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// GetUserList
// @Tags 用户模块
// @Summary 所有用户
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data, err := models.GetUserList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}
	c.JSON(200, gin.H{
		"code":    0, //  0成功   -1失败
		"message": "用户名已注册！",
		"data":    data,
	})
}

// CreateUser
// @Tags 用户模块
// @Summary 新增用户
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	//err := c.BindJSON(&user)
	//fmt.Println(user)
	//if err != nil {
	//	c.JSON(http.StatusNotFound, "绑定form失败")
	//} else {
	//	c.JSON(http.StatusOK, "绑定form成功")
	//}
	//user.Name = c.Query("name")
	//password := c.Query("password")
	//repassword := c.Query("repassword")
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("Identity")
	fmt.Println(user.Name, ">>>>>>>", password, repassword)
	salt := fmt.Sprintf("%06d", rand.Int31())

	data := models.FindUserByName(user.Name)
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户名或者密码不能为空",
			"data":    user,
		})
		return
	}

	if data.Name != "" {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "用户名已经注册",
			"data":    data,
		})
		return
	}

	if password != repassword {
		c.JSON(-1, gin.H{
			"code":    -1,
			"message": "两次密码不一致",
			"data":    user,
		})
		return
	}
	//user.PassWord = password
	user.PassWord = utils.MakePassword(password, salt)
	user.Salt = salt
	fmt.Println(user.PassWord)

	models.CreateUser(&user)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "新增用户成功！",
		"data":    user,
	})
}

// DeleteUser
// @Tags 用户模块
// @Summary 删除用户
// @param id query int false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	//user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))

	//user.ID = uint(id)
	models.DeleteUser(id)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "删除用户成功！",
		"user id": id,
	})
}

// UpdateUser
// @Tags 用户模块
// @Summary 更新用户
// @param id formData int false "id"
// @param name formData string false "name"
// @param password formData int false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))

	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Avatar = c.PostForm("icon")
	user.Email = c.PostForm("email")

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功 -1 失败
			"message": "修改参数不匹配！",
			"data":    user,
		})
	} else {
		models.UpdateUser(&user)
		c.JSON(200, gin.H{
			"code":    0, // 0 成功 -1 失败
			"message": "修改用户成功！",
			"data":    user,
		})
	}

}

// FindUserByNameAndPwd
// @Tags 用户模块
// @Summary 所有用户
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}

	//name := c.Query("name")
	//password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")

	fmt.Println("dasd", name, password)
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功 -1 失败
			"message": "用户不存在",
			"data":    data,
		})
		return
	}
	fmt.Println(user)
	flag := utils.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功 -1 失败
			"message": "密码不正确",
			"data":    data,
		})
		return
	}

	pwd := utils.MakePassword(password, user.Salt)
	data = models.FindUserByNameAndPwd(name, pwd)

	c.JSON(200, gin.H{
		"code":    0, // 0 成功 -1 失败
		"message": "登陆成功",
		"data":    data,
	})
}

// 防止跨越站点伪造请求
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	//ws, err := upGrade.Upgrade(ctx.Writer, ctx.Request, nil)
	////msg := ctx.Query("message")
	//_, message, err := ws.ReadMessage()
	//fmt.Printf("%s", message)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//go func() {
	//	utils.Publish(ctx, utils.PublishKey, string(message))
	//}()
	//
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//
	////MsgHandler(ws, ctx)
	//defer func(ws *websocket.Conn) {
	//	err := ws.Close()
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}(ws)
	//MsgHandler(ws, ctx)
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	MsgHandler(c, ws)
}

func MsgHandler(c *gin.Context, ws *websocket.Conn) {
	//msg, err := utils.Subscribe(c, utils.PublishKey)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//tm := time.Now().Format("2006-01-02 15:04:05")
	//m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
	//err = ws.WriteMessage(1, []byte(m))
	//fmt.Println("消息为：", m)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//ch := make(chan string)
	//defer close(ch)
	//
	//go func() {
	//	msg, err := utils.Subscribe(c, utils.PublishKey)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	tm := time.Now().Format("2006-01-02 15:04:05")
	//	m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
	//	ch <- m
	//}()
	//
	//for {
	//	select {
	//	case msg := <-ch:
	//		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//		fmt.Println("消息为：", msg)
	//	}
	//}
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println("MsgHandler发送失败", err)
		}
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat1(c.Writer, c.Request)
}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.Request.FormValue("userIdA"))
	userIdB, _ := strconv.Atoi(c.Request.FormValue("userIdB"))
	start, _ := strconv.Atoi(c.Request.FormValue("start"))
	end, _ := strconv.Atoi(c.Request.FormValue("end"))
	isRev, _ := strconv.ParseBool(c.Request.FormValue("isRev"))
	res := models.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	utils.RespOKList(c.Writer, 0, res, len(res))
}

func SearchFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriends(uint(userId))

	//c.JSON(200, gin.H{
	//	"code":    0, // 0 成功 -1 失败
	//	"message": "查询好友列表成功！",
	//	"data":    users,
	//})
	utils.RespOKList(c.Writer, 0, users, len(users))

}

func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetId, _ := strconv.Atoi(c.Request.FormValue("targetId"))
	code, msg := models.AddFriend(uint(userId), uint(targetId))

	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	name := c.Request.FormValue("name")
	desc := c.Request.FormValue("desc")
	icon := c.Request.FormValue("icon")
	community := models.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Desc = desc
	community.Img = icon
	code, msg := models.CreateCommunity(&community)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	data, msg := models.LoadCommunity(uint(ownerId))
	if len(data) != 0 {
		utils.RespOKList(c.Writer, 0, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func JoinGroup(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	comId := c.Request.FormValue("comId")
	code, msg := models.JoinGroup(uint(userId), comId)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}

}
func FindByID(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))

	//	name := c.Request.FormValue("name")
	data := models.FindByID(uint(userId))
	utils.RespOK(c.Writer, data, "ok")
}
func JoinGroupChat(c *gin.Context) {
	groupId, _ := strconv.Atoi(c.Request.FormValue("groupId"))
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))

	models.JoinGroupChat(int64(groupId), int64(userId))

	utils.RespOK(c.Writer, nil, "ok")
}
