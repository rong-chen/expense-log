package model

import (
	"crypto/sha256"
	"database/sql/driver"
	"errors"
	"expense-log/global"
	"fmt"
	"strconv"
	"strings"
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
	Tags         []Tag      `gorm:"many2many:bill_tags;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"tags,omitempty"`   // 账单标签
	Embedding    Vector     `gorm:"type:vector(1536)" json:"-"`                  // 🌟 语义向量，类型为 pgvector vector(1536)
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

// Vector 自定义数据库向量映射类型 (支持 pgvector)
type Vector []float32

// Value 实现 driver.Valuer 接口，将 []float32 序列化为 PostgreSQL 识别的 '[x,y,z...]' 字符串格式
func (v Vector) Value() (driver.Value, error) {
	if len(v) == 0 {
		return nil, nil
	}
	var sb strings.Builder
	sb.WriteByte('[')
	for idx, val := range v {
		if idx > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatFloat(float64(val), 'f', -1, 32))
	}
	sb.WriteByte(']')
	return sb.String(), nil
}

// Scan 实现 sql.Scanner 接口，从数据库反序列化 pgvector 格式回到 Vector 类型
func (v *Vector) Scan(src interface{}) error {
	if src == nil {
		*v = nil
		return nil
	}

	var str string
	switch val := src.(type) {
	case string:
		str = val
	case []byte:
		str = string(val)
	default:
		return errors.New("failed to scan vector: unsupported type")
	}

	str = strings.Trim(str, "[]")
	if str == "" {
		*v = make(Vector, 0)
		return nil
	}

	parts := strings.Split(str, ",")
	res := make(Vector, len(parts))
	for idx, p := range parts {
		fVal, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return fmt.Errorf("failed to parse vector element: %w", err)
		}
		res[idx] = float32(fVal)
	}
	*v = res
	return nil
}
