package model

import (
	"crypto/sha256"
	"expense-log/global"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillSource 账单来源
type BillSource string

const (
	BillSourceEmail     BillSource = "email"     // 邮件自动拉取
	BillSourceUpload    BillSource = "upload"    // 手动上传文件
	BillSourceManual    BillSource = "manual"    // 手动输入
	BillSourceRecurring BillSource = "recurring" // 周期账单自动生成
)

// Bill 账单记录
type Bill struct {
	global.Model
	UserID   uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"` // 所属用户 (记录人)
	LedgerID *uuid.UUID `gorm:"type:uuid;index;default:null" json:"ledger_id"` // 所属账本 (为null或新兼容分配给个人账本)

	// --- 核心交易信息 ---
	TransactionNo   string     `gorm:"type:varchar(100);index" json:"transaction_no"`        // 交易单号 (支付宝/微信)
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"`             // 金额
	Merchant        string     `gorm:"type:varchar(200)" json:"merchant"`                     // 商户名称
	TransactionDate time.Time  `json:"transaction_date"`                                      // 交易日期
	Category        string     `gorm:"type:varchar(50)" json:"category"`                      // 分类(餐饮/交通/购物等)
	Remark          string     `gorm:"type:text" json:"remark"`                               // 备注
	Source          BillSource `gorm:"type:varchar(20);not null;default:'manual'" json:"source"` // 来源

	// --- 指纹去重 ---
	Fingerprint string `gorm:"type:varchar(64);uniqueIndex;not null" json:"fingerprint"` // SHA256 指纹

	// --- 关联信息 ---
	EmailID      *uuid.UUID `gorm:"type:uuid;index" json:"email_id,omitempty"`    // 关联邮件 (可为空)
	OriginalFile string     `gorm:"type:varchar(500)" json:"original_file"`       // 原始文件路径
	RawContent   string     `gorm:"type:text" json:"-"`                           // VLM 原始返回 (不传给前端)
	Tags         []Tag      `gorm:"many2many:bill_tags;" json:"tags,omitempty"`   // 账单标签
}

func (b *Bill) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	// 自动生成指纹
	if b.Fingerprint == "" {
		b.Fingerprint = GenerateFingerprint(b.UserID, b.TransactionNo, b.Amount, b.TransactionDate, b.Merchant)
	}
	return
}

// GenerateFingerprint 生成账单指纹
//   - 有交易单号: SHA256(userID + transaction_no) → 100% 精确去重
//   - 无交易单号: SHA256(userID + amount + date + merchant) → 模糊去重
func GenerateFingerprint(userID uuid.UUID, transactionNo string, amount float64, date time.Time, merchant string) string {
	var raw string
	if transactionNo != "" {
		raw = fmt.Sprintf("%s|%s", userID.String(), transactionNo)
	} else {
		raw = fmt.Sprintf("%s|%.2f|%s|%s", userID.String(), amount, date.Format("2006-01-02"), merchant)
	}
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", hash)
}
