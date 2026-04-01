package model

import (
	"expense-log/global"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InvitationCode 邀请码模型
type InvitationCode struct {
	global.Model
	Code   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"`
	IsUsed bool      `gorm:"default:false" json:"is_used"`
	UsedBy uuid.UUID `gorm:"type:uuid" json:"used_by"` // 使用该邀请码的用户ID
}

func (i *InvitationCode) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}
