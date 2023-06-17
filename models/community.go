package models

import (
	"fmt"
	"gin_chat/utils"
	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	//Ch       *amqp.Channel
	//PubQueue *amqp.Queue
	Name    string
	OwnerId uint
	Img     string
	Desc    string
}

func (table *Community) TableName() string {
	return "community"
}

func CreateCommunity(community *Community) (int, string) {
	tx := utils.DB.Begin()
	// 事务一旦开始，不论什么异常最终都会rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if len(community.Name) == 0 {
		return -1, "群名称不能为空"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}

	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(err)
		tx.Rollback()
		return -1, "建群失败"
	}
	contact := Contact{}
	contact.OwnerId = community.OwnerId
	contact.TargetId = community.ID
	contact.Type = 2
	if err := utils.DB.Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "添加群关系失败"
	}
	tx.Commit()
	return 0, "建群成功"
}

func LoadCommunity(ownerId uint) ([]*Community, string) {
	//data := make([]*Community, 10)
	//utils.DB.Where("owner_id = ?", ownerId).Find(&data)
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)

	utils.DB.Where("owner_id = ? and type = 2", ownerId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.TargetId)
	}
	data := make([]*Community, 10)
	utils.DB.Where("id in ?", objIds).Find(&data)

	return data, "查询成功"
}

func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	//contact.TargetId = comId
	contact.Type = 2
	community := Community{}

	utils.DB.Where("id = ? or name = ?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type=2", userId, comId).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}

}
