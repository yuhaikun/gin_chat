package main

import (
	"gin_chat/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(mysql.Open("root:8871527yhk@tcp(127.0.0.1:3306)/ginchat?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		panic("failed to connect the database")
	}

	//db.AutoMigrate(&models.UserBasic{})

	//db.AutoMigrate(&models.Message{})
	//
	//db.AutoMigrate(&models.GroupBasic{})
	//
	//db.AutoMigrate(&models.Contact{})

	db.AutoMigrate(&models.Community{})

	////	create
	//user := models.UserBasic{}
	//user.Name = "张三"
	//db.Create(&user)
	//
	////	read
	//fmt.Println(db.First(&user, 1)) // 根据整形主键查找
	//
	//// update
	//db.Model(&user).Update("PassWord", 1234)
}
