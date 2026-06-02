package model

import (
	"expense-log/global"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MCPKey MCP 访问密钥表
type MCPKey struct {
	global.Model
	Key        string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"key"`          // 生成的随机 API Key 令牌
	UserID     uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`                    // 绑定并代表的系统用户ID
	Applicant  string     `gorm:"type:varchar(100);not null" json:"applicant"`                // 申请人名称 / 应用名称
	Purpose    string     `gorm:"type:text" json:"purpose"`                                  // 申请用途说明
	Status     string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`  // pending | approved | rejected | revoked
	ApprovedBy *uuid.UUID `gorm:"type:uuid;default:null" json:"approved_by,omitempty"`        // 审批的管理员ID
	ApprovedAt *time.Time `json:"approved_at,omitempty"`
}

// BeforeCreate 钩子，自动生成 UUID
func (m *MCPKey) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}
