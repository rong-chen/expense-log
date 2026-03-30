package repository

import (
	"expense-log/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BillRepository interface {
	// ExistsByFingerprint 通过指纹检查账单是否已存在（核心去重方法）
	ExistsByFingerprint(fingerprint string) (bool, error)
	// ExistsByTransactionNo 通过交易单号检查是否已存在
	ExistsByTransactionNo(transactionNo string) (bool, error)
	// Create 创建账单记录
	Create(bill *model.Bill) error
	// GetByUserID 获取用户的所有账单
	GetByUserID(userID uuid.UUID, page, pageSize int) ([]model.Bill, int64, error)
	// GetByID 获取单条账单
	GetByID(id uuid.UUID) (*model.Bill, error)
}

type billRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) BillRepository {
	return &billRepository{db: db}
}

func (r *billRepository) ExistsByFingerprint(fingerprint string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Bill{}).Where("fingerprint = ?", fingerprint).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *billRepository) ExistsByTransactionNo(transactionNo string) (bool, error) {
	if transactionNo == "" {
		return false, nil
	}
	var count int64
	err := r.db.Model(&model.Bill{}).Where("transaction_no = ?", transactionNo).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *billRepository) Create(bill *model.Bill) error {
	return r.db.Create(bill).Error
}

func (r *billRepository) GetByUserID(userID uuid.UUID, page, pageSize int) ([]model.Bill, int64, error) {
	var bills []model.Bill
	var total int64

	query := r.db.Model(&model.Bill{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("transaction_date DESC").Offset(offset).Limit(pageSize).Find(&bills).Error
	return bills, total, err
}

func (r *billRepository) GetByID(id uuid.UUID) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.Where("id = ?", id).First(&bill).Error
	if err != nil {
		return nil, err
	}
	return &bill, nil
}
