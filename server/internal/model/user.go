package model

import (
	"expense-log/global"

	"github.com/btcsuite/btcutil/base58"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User 内部使用
type User struct {
	global.Model
	// --- 核心登录信息 ---
	Phone    string `gorm:"type:varchar(20);uniqueIndex;not null" json:"phone"` // 唯一索引，用于登录
	Password string `gorm:"type:varchar(255);not null" json:"-"`                // 存储加密后的哈希，json排除

	// --- 基础个人资料 ---
	Nickname string `gorm:"type:varchar(50)" json:"nickname"`     // 昵称
	Avatar   string `gorm:"type:varchar(255)" json:"avatar"`      // 头像URL
	Email    string `gorm:"type:varchar(100);index" json:"email"` // 备用联系方式
	UID      string `gorm:"type:varchar(50)" json:"uid"`

	// --- 业务相关 ---
	LastLogin int64 `json:"last_login"` // 最后登录时间戳
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID, err = uuid.NewV7()
		if err != nil {
			return err // 处理可能的系统时钟错误
		}
	}
	if u.UID == "" {
		//比特币，创建uid，用于识别用户
		u.UID = base58.Encode(u.ID[:])
	}
	return
}

// --- 请求 DTO ---

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}



// --- 响应 DTO ---

// TokenResponse 双Token响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	UID       string `json:"uid"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	LastLogin int64  `json:"last_login"`
}
