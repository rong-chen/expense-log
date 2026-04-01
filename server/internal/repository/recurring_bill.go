package repository

import (
	"expense-log/internal/model"
	"gorm.io/gorm"
)

type RecurringBillRepository interface {
	CountTotal() (int64, error)
}

type recurringBillRepository struct {
	db *gorm.DB
}

func NewRecurringBillRepository(db *gorm.DB) RecurringBillRepository {
	return &recurringBillRepository{db: db}
}

func (r *recurringBillRepository) CountTotal() (int64, error) {
	var count int64
	err := r.db.Model(&model.RecurringBill{}).Count(&count).Error
	return count, err
}
