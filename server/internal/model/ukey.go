package model

import (
	"expense-log/global"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Ukey 自动化凭证表
type Ukey struct {
	global.Model
	UserID    uuid.UUID `gorm:"type:char(36);not null;index" json:"user_id"` // 关联 User ID
	SecretKey string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"secret_key"` // API Key 内容
	Name      string    `gorm:"type:varchar(100);not null" json:"name"` // 设备/应用识别名
}

// BeforeCreate 钩子，自动生成 UUID
func (u *Ukey) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID, err = uuid.NewV7()
	}
	return
}
