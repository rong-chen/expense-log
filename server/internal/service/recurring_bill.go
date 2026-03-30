package service

import (
	"expense-log/internal/model"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecurringBillService interface {
	Create(userID uuid.UUID, rb *model.RecurringBill) error
	List(userID uuid.UUID) ([]model.RecurringBill, error)
	Update(userID uuid.UUID, id uuid.UUID, rb *model.RecurringBill) error
	Delete(userID uuid.UUID, id uuid.UUID) error
	ToggleActive(userID uuid.UUID, id uuid.UUID) error
	ExecuteDailyTask()
	StartScheduler()
}

type recurringBillService struct {
	db *gorm.DB
}

func NewRecurringBillService(db *gorm.DB) RecurringBillService {
	return &recurringBillService{db: db}
}

func (s *recurringBillService) Create(userID uuid.UUID, rb *model.RecurringBill) error {
	rb.UserID = userID
	return s.db.Create(rb).Error
}

func (s *recurringBillService) List(userID uuid.UUID) ([]model.RecurringBill, error) {
	var list []model.RecurringBill
	err := s.db.Where("user_id = ?", userID).Order("day_of_month ASC").Find(&list).Error
	return list, err
}

func (s *recurringBillService) Update(userID uuid.UUID, id uuid.UUID, rb *model.RecurringBill) error {
	updates := map[string]interface{}{
		"amount":       rb.Amount,
		"merchant":     rb.Merchant,
		"category":     rb.Category,
		"remark":       rb.Remark,
		"day_of_month": rb.DayOfMonth,
	}
	return s.db.Model(&model.RecurringBill{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}

func (s *recurringBillService) Delete(userID uuid.UUID, id uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.RecurringBill{}).Error
}

func (s *recurringBillService) ToggleActive(userID uuid.UUID, id uuid.UUID) error {
	var rb model.RecurringBill
	if err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&rb).Error; err != nil {
		return err
	}
	return s.db.Model(&rb).Update("is_active", !rb.IsActive).Error
}

// ExecuteDailyTask 每天执行一次，检查当天应该自动记账的订阅项目
func (s *recurringBillService) ExecuteDailyTask() {
	now := time.Now()
	today := now.Format("2006-01-02")
	dayOfMonth := now.Day()

	// 判断本月最后一天（用于处理 31 号边界）
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()

	var bills []model.RecurringBill
	query := s.db.Where("is_active = ? AND (last_exec_at IS NULL OR last_exec_at != ?)", true, today)

	if dayOfMonth == lastDay {
		// 今天是本月最后一天：除了匹配今天的号数，还要匹配所有大于本月最大天数的设置
		// 比如 2月28日，要同时触发设置为 28、29、30、31 号的配置
		query = query.Where("day_of_month <= ? OR day_of_month = ?", lastDay, dayOfMonth)
		// 更精确：day_of_month >= dayOfMonth (当天或更大的都要触发)
		query = s.db.Where("is_active = ? AND (last_exec_at IS NULL OR last_exec_at != ?) AND day_of_month >= ?", true, today, dayOfMonth)
	} else {
		query = query.Where("day_of_month = ?", dayOfMonth)
	}

	if err := query.Find(&bills).Error; err != nil {
		log.Printf("[RecurringBill] 查询待执行任务失败: %v", err)
		return
	}

	if len(bills) == 0 {
		log.Printf("[RecurringBill] %s 无需执行的周期账单", today)
		return
	}

	log.Printf("[RecurringBill] %s 发现 %d 条待执行的周期账单", today, len(bills))

	for _, rb := range bills {
		// 生成指纹用于去重
		fingerprint := model.GenerateFingerprint("", rb.Amount, now, rb.Merchant)

		bill := model.Bill{
			UserID:          rb.UserID,
			Amount:          rb.Amount,
			Merchant:        rb.Merchant,
			Category:        rb.Category,
			Remark:          fmt.Sprintf("[周期账单] %s", rb.Remark),
			TransactionDate: now,
			Source:          "recurring",
			Fingerprint:     fingerprint,
		}

		if err := s.db.Create(&bill).Error; err != nil {
			log.Printf("[RecurringBill] 创建账单失败 (recurring_id=%s): %v", rb.ID, err)
			continue
		}

		// 更新 LastExecAt 防止重复执行
		s.db.Model(&rb).Update("last_exec_at", today)

		log.Printf("[RecurringBill] 成功: %s %.2f元 (%s)", rb.Merchant, rb.Amount, rb.Category)
	}
}

// StartScheduler 启动定时调度器
// 策略：启动时立即补偿检查 + 每天凌晨 00:05 定时执行
// 这样即使服务器在凌晨宕机或重启，也不会漏掉当天的周期账单
func (s *recurringBillService) StartScheduler() {
	go func() {
		log.Println("[RecurringBill] 周期账单调度器已启动")

		// 1. 启动时立即执行一次补偿检查（处理服务器宕机期间遗漏的账单）
		log.Println("[RecurringBill] 执行启动补偿检查...")
		s.ExecuteDailyTask()

		// 2. 然后进入每日定时循环
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 5, 0, 0, now.Location())
			waitDuration := next.Sub(now)

			log.Printf("[RecurringBill] 下次执行时间: %s (等待 %s)", next.Format("2006-01-02 15:04:05"), waitDuration)

			time.Sleep(waitDuration)

			s.ExecuteDailyTask()
		}
	}()
}
