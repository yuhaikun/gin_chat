package models

import "gorm.io/gorm"

// Group_basic 群关系
type GroupBasic struct {
	gorm.Model
	Name    string
	OwnerId uint
	Icon    string
	Type    int // level
	Desc    string
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
