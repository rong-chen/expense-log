package model

import (
	"expense-log/global"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag 用户自定义标签
type Tag struct {
	global.Model
	UserID uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Name   string    `gorm:"type:varchar(50);not null" json:"name"`
	Color  string    `gorm:"type:varchar(20);default:'#3498db'" json:"color"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID, err = uuid.NewV7()
	}
	return
}

// BillTag 账单-标签 多对多关联表
type BillTag struct {
	BillID uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"bill_id"`
	TagID  uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"tag_id"`
}
