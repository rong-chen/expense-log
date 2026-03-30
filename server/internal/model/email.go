package model

import (
	"expense-log/global"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailRecord 邮件记录（用于去重追踪）
type EmailRecord struct {
	global.Model
	UserID        uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`                  // 所属用户
	AccountID     uuid.UUID `gorm:"type:uuid;index;not null" json:"account_id"`               // 关联邮箱账户
	MessageID     string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"message_id"` // 邮件 Message-ID
	Subject       string    `gorm:"type:text" json:"subject"`
	FromAddress   string    `gorm:"type:varchar(255)" json:"from_address"`
	Date          time.Time `json:"date"`
	HasAttachment bool      `json:"has_attachment"`
	Processed     bool      `gorm:"default:false" json:"processed"`
}

func (e *EmailRecord) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}

// EmailAttachment 邮件附件记录
type EmailAttachment struct {
	global.Model
	EmailID  uuid.UUID `gorm:"type:uuid;index;not null" json:"email_id"`
	Filename string    `gorm:"type:varchar(255)" json:"filename"`
	FilePath string    `gorm:"type:varchar(500)" json:"file_path"`
	MimeType string    `gorm:"type:varchar(100)" json:"mime_type"`
	Size     int64     `json:"size"`
}

func (a *EmailAttachment) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}

// UserEmailAccount 用户绑定的邮箱账户（IMAP 凭证存数据库）
type UserEmailAccount struct {
	global.Model
	UserID   uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Host     string    `gorm:"type:varchar(100);not null" json:"host"`     // IMAP 服务器地址
	Port     int       `gorm:"default:993" json:"port"`                   // IMAP 端口
	Username string    `gorm:"type:varchar(100);not null" json:"username"` // 邮箱地址
	Password string    `gorm:"type:varchar(255);not null" json:"-"`        // 授权码 (json排除)
	TLS      bool      `gorm:"default:true" json:"tls"`                   // 是否使用 TLS
	Folder   string    `gorm:"type:varchar(50);default:'INBOX'" json:"folder"`
	Enabled  bool      `gorm:"default:true" json:"enabled"` // 是否启用
}

func (a *UserEmailAccount) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return
}

// BindEmailRequest 绑定邮箱请求
type BindEmailRequest struct {
	Host     string `json:"host" binding:"required"`     // 如 imap.qq.com
	Port     int    `json:"port"`                        // 默认 993
	Username string `json:"username" binding:"required"` // 邮箱地址
	Password string `json:"password" binding:"required"` // 授权码
	TLS      *bool  `json:"tls"`                         // 默认 true
	Folder   string `json:"folder"`                      // 默认 INBOX
}

// EmailAccountResponse 邮箱账户响应（不包含密码）
type EmailAccountResponse struct {
	ID       uuid.UUID `json:"id"`
	Host     string    `json:"host"`
	Username string    `json:"username"`
	Port     int       `json:"port"`
	TLS      bool      `json:"tls"`
	Folder   string    `json:"folder"`
	Enabled  bool      `json:"enabled"`
}
