package model

import (
	"crypto/rand"
	"expense-log/global"
	"math/big"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LedgerType 账本类型
type LedgerType string

const (
	LedgerTypePersonal LedgerType = "personal" // 个人账本
	LedgerTypeShared   LedgerType = "shared"   // 共享账本
)

// Ledger 账本模型
type Ledger struct {
	global.Model
	Name        string     `gorm:"type:varchar(100)" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	OwnerID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_id"`
	Type        LedgerType `gorm:"type:varchar(20);default:'shared'" json:"type"`
	InviteCode  string     `gorm:"type:varchar(10);uniqueIndex" json:"invite_code"` // 6位邀请码
	
	// 关联
	Members []LedgerMember `gorm:"foreignKey:LedgerID" json:"members,omitempty"`
}

func (l *Ledger) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	// 自动生成首个邀请码
	if l.InviteCode == "" {
		l.InviteCode = GenerateInviteCode()
	}
	return
}

// GenerateInviteCode 生成6位随机大小写字母+数字邀请码
func GenerateInviteCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
}

// LedgerRole 队伍内角色
type LedgerRole string

const (
	LedgerRoleOwner  LedgerRole = "owner"  // 创建者
	LedgerRoleMember LedgerRole = "member" // 普通成员
)

// LedgerMember 账本成员多对多或一对多结构
type LedgerMember struct {
	LedgerID uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"ledger_id"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null;primaryKey" json:"user_id"`
	Role     LedgerRole `gorm:"type:varchar(20);default:'member'" json:"role"`
	// 可选预加载用户信息
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
