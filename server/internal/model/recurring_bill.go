package model

import (
	"expense-log/global"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RecurringBill 周期性账单（订阅/固定支出）
type RecurringBill struct {
	global.Model
	UserID     uuid.UUID `gorm:"type:uuid;index;not null" json:"user_id"`
	Amount     float64   `gorm:"type:decimal(12,2);not null" json:"amount"`       // 每期金额
	Merchant   string    `gorm:"type:varchar(200);not null" json:"merchant"`      // 商户/项目名称
	Category   string    `gorm:"type:varchar(50)" json:"category"`                // 分类
	Remark     string    `gorm:"type:text" json:"remark"`                         // 备注
	DayOfMonth int       `gorm:"not null" json:"day_of_month"`                    // 每月几号扣款 (1-31)
	IsActive   bool      `gorm:"not null;default:true" json:"is_active"`          // 是否启用
	LastExecAt *string   `gorm:"type:varchar(10)" json:"last_exec_at,omitempty"`  // 上次执行日期 YYYY-MM-DD
}

func (r *RecurringBill) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID, err = uuid.NewV7()
	}
	return
}
