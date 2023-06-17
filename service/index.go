package service

import (
	"fmt"
	"gin_chat/models"
	"github.com/gin-gonic/gin"
	"html/template"
	"strconv"
)

// GetIndex
// @Tags 首页
// @Success 200 {string} welcome
// @Router /index [get]
func GetIndex(ctx *gin.Context) {
	//ctx.JSON(200, gin.H{
	//	"message": "welcome",
	//})
	ind, err := template.ParseFiles("index.html", "views/chat/head.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(ctx.Writer, "index")
}

func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "index")
}

func ToChat(ctx *gin.Context) {
	//ctx.JSON(200, gin.H{
	//	"message": "welcome",
	//})
	ind, err := template.ParseFiles("views/chat/index.html",
		"views/chat/head.html",
		"views/chat/foot.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/main.html",
		"views/chat/createcom.html",
		"views/chat/userinfo.html",
	)
	if err != nil {
		panic(err)
	}

	userId, _ := strconv.Atoi(ctx.Query("userId"))
	token := ctx.Query("token")
	user := models.UserBasic{}
	user.ID = uint(userId)
	user.Identity = token

	fmt.Println("Tochat >>>>>>>>>", user)
	ind.Execute(ctx.Writer, user)
}

func Chat(c *gin.Context) {
	models.Chat1(c.Writer, c.Request)
}
