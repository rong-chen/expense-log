package service

import (
	"context"
	"encoding/json"
	"expense-log/internal/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type BillService interface {
	GetTrendStats(userID uuid.UUID) ([]TrendStatResponse, error)
	GetCategoryStats(userID uuid.UUID) ([]CategoryStatResponse, error)
	GetDashboardStats(userID uuid.UUID) (*DashboardStatResponse, error)
	GetBillDetail(userID uuid.UUID, billID uuid.UUID) (*model.Bill, error)
	GetBillList(userID uuid.UUID, page, pageSize int, keyword, category, date string) ([]model.Bill, int64, error)
	UpdateRemark(userID uuid.UUID, billID uuid.UUID, remark string) error
	UpdateBill(userID uuid.UUID, billID uuid.UUID, dto UpdateBillDTO) error
	DeleteBill(userID uuid.UUID, billID uuid.UUID) error
	InvalidateUserCache(userID uuid.UUID)
}

type UpdateBillDTO struct {
	Amount    float64
	Merchant  string
	Category  string
	Remark    string
	CreatedAt time.Time
}

type billService struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewBillService(db *gorm.DB, rdb *redis.Client) BillService {
	return &billService{db: db, rdb: rdb}
}

type TrendStatResponse struct {
	Month   string  `json:"month"`
	Expense float64 `json:"expense"` // 暂时不需要 income，按需扩展
}

type CategoryStatResponse struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type DashboardStatResponse struct {
	MonthExpense     float64 `json:"month_expense"`
	LastMonthExpense float64 `json:"last_month_expense"`
	MonthIncome      float64 `json:"month_income"`
	BillCount        int64   `json:"bill_count"`
	PendingEmail     int64   `json:"pending_email"`
}

func (s *billService) GetTrendStats(userID uuid.UUID) ([]TrendStatResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%s:stats:trend", userID.String())

	// 1. 读取缓存
	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var res []TrendStatResponse
		if json.Unmarshal([]byte(cached), &res) == nil {
			return res, nil
		}
	}

	now := time.Now()
	sixMonthsAgo := now.AddDate(0, -5, 0)
	startOfMonth := time.Date(sixMonthsAgo.Year(), sixMonthsAgo.Month(), 1, 0, 0, 0, 0, now.Location())

	var results []TrendStatResponse
	err := s.db.Model(&model.Bill{}).
		Select("to_char(created_at, 'YYYY-MM') as month, sum(amount) as expense").
		Where("user_id = ? AND created_at >= ? AND (category != '退款' OR category IS NULL)", userID, startOfMonth).
		Group("to_char(created_at, 'YYYY-MM')").
		Order("month asc").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 补齐 6 个月（即使某个月没有数据也要返回 0）
	resMap := make(map[string]float64)
	for _, r := range results {
		resMap[r.Month] = r.Expense
	}

	var finalRes []TrendStatResponse
	for i := -5; i <= 0; i++ {
		m := now.AddDate(0, i, 0).Format("2006-01")
		finalRes = append(finalRes, TrendStatResponse{
			Month:   m,
			Expense: resMap[m],
		})
	}

	// 2. 写入缓存 (5分钟)
	if resBytes, err := json.Marshal(finalRes); err == nil {
		s.rdb.Set(ctx, cacheKey, resBytes, 5*time.Minute)
	}

	return finalRes, nil
}

func (s *billService) GetCategoryStats(userID uuid.UUID) ([]CategoryStatResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%s:stats:category", userID.String())

	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var res []CategoryStatResponse
		if json.Unmarshal([]byte(cached), &res) == nil {
			return res, nil
		}
	}

	var results []CategoryStatResponse

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	err := s.db.Model(&model.Bill{}).
		Select("category as name, count(*) as value"). // 暂时以笔数占比为例，或者用 sum(amount)
		Where("user_id = ? AND created_at >= ? AND category != '' AND category IS NOT NULL", userID, startOfMonth).
		Group("category").
		Order("value desc").
		Scan(&results).Error

	if err == nil {
		if resBytes, e := json.Marshal(results); e == nil {
			s.rdb.Set(ctx, cacheKey, resBytes, 5*time.Minute)
		}
	}

	return results, err
}

func (s *billService) GetDashboardStats(userID uuid.UUID) (*DashboardStatResponse, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%s:stats:dashboard", userID.String())

	if cached, err := s.rdb.Get(ctx, cacheKey).Result(); err == nil {
		var res DashboardStatResponse
		if json.Unmarshal([]byte(cached), &res) == nil {
			return &res, nil
		}
	}

	var res DashboardStatResponse

	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 这个月总支出（排除退款）
	var monthExpense float64
	s.db.Model(&model.Bill{}).
		Where("user_id = ? AND created_at >= ? AND (category != '退款' OR category IS NULL)", userID, startOfMonth).
		Select("COALESCE(sum(amount), 0)").Scan(&monthExpense)

	// 本月账单数（排除退款）
	var billCount int64
	s.db.Model(&model.Bill{}).
		Where("user_id = ? AND created_at >= ? AND (category != '退款' OR category IS NULL)", userID, startOfMonth).
		Count(&billCount)

	// 上月总支出（排除退款）
	lastMonthStart := startOfMonth.AddDate(0, -1, 0)
	var lastMonthExpense float64
	s.db.Model(&model.Bill{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ? AND (category != '退款' OR category IS NULL)", userID, lastMonthStart, startOfMonth).
		Select("COALESCE(sum(amount), 0)").Scan(&lastMonthExpense)

	// 未处理邮件 (status = 0 或者 parsing_status)
	var pendingEmail int64
	s.db.Table("email_messages").
		Where("user_id = ? AND status = 0", userID).
		Count(&pendingEmail)

	res.MonthExpense = monthExpense
	res.LastMonthExpense = lastMonthExpense
	res.BillCount = billCount
	res.PendingEmail = pendingEmail
	// 收入可以暂时为0，因为账单模型里目前并未严格区分类型，假设全为支出
	res.MonthIncome = 0.0

	if resBytes, err := json.Marshal(res); err == nil {
		s.rdb.Set(ctx, cacheKey, resBytes, 5*time.Minute)
	}

	return &res, nil
}

func (s *billService) GetBillDetail(userID uuid.UUID, billID uuid.UUID) (*model.Bill, error) {
	var bill model.Bill
	err := s.db.Where("id = ? AND user_id = ?", billID, userID).First(&bill).Error
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

func (s *billService) GetBillList(userID uuid.UUID, page, pageSize int, keyword, category, date string) ([]model.Bill, int64, error) {
	// 懒得每次去重写 repo 实例，直接内敛使用。因为原来没有在这个 service 里注入 repo。
	// 为了最快实现，我们用 gorm db 直接查
	var bills []model.Bill
	var total int64

	query := s.db.Model(&model.Bill{}).Where("user_id = ?", userID)

	if keyword != "" {
		query = query.Where("merchant LIKE ? OR remark LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if date != "" {
		// 利用 to_char 兼容 YYYY-MM 或 YYYY-MM-DD 的前缀匹配
		query = query.Where("to_char(created_at, 'YYYY-MM-DD') LIKE ?", date+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&bills).Error
	return bills, total, err
}

func (s *billService) UpdateRemark(userID uuid.UUID, billID uuid.UUID, remark string) error {
	return s.db.Model(&model.Bill{}).
		Where("id = ? AND user_id = ?", billID, userID).
		Update("remark", remark).Error
}

func (s *billService) UpdateBill(userID uuid.UUID, billID uuid.UUID, dto UpdateBillDTO) error {
	updates := map[string]interface{}{
		"amount":     dto.Amount,
		"merchant":   dto.Merchant,
		"category":   dto.Category,
		"remark":     dto.Remark,
		"created_at": dto.CreatedAt,
	}
	return s.db.Model(&model.Bill{}).Where("id = ? AND user_id = ?", billID, userID).Updates(updates).Error
}

func (s *billService) DeleteBill(userID uuid.UUID, billID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", billID, userID).Delete(&model.Bill{}).Error
}

func (s *billService) InvalidateUserCache(userID uuid.UUID) {
	ctx := context.Background()
	s.rdb.Del(ctx,
		fmt.Sprintf("user:%s:stats:trend", userID.String()),
		fmt.Sprintf("user:%s:stats:category", userID.String()),
		fmt.Sprintf("user:%s:stats:dashboard", userID.String()),
	)
}
