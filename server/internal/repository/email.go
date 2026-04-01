package repository

import (
	"expense-log/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EmailRepository interface {
	ExistsByMessageID(messageID string) (bool, error)
	CreateRecord(record *model.EmailRecord) error
	CreateAttachment(attachment *model.EmailAttachment) error
	GetUnprocessed() ([]model.EmailRecord, error)
	MarkProcessed(id uuid.UUID) error

	// --- 邮箱账户管理 ---
	CreateAccount(account *model.UserEmailAccount) error
	GetAccountsByUserID(userID uuid.UUID) ([]model.UserEmailAccount, error)
	GetAllEnabledAccounts() ([]model.UserEmailAccount, error)
	DeleteAccount(id uuid.UUID, userID uuid.UUID) error
	CountTotalAccounts() (int64, error)
}

type emailRepository struct {
	db *gorm.DB
}

func NewEmailRepository(db *gorm.DB) EmailRepository {
	return &emailRepository{db: db}
}

func (r *emailRepository) ExistsByMessageID(messageID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.EmailRecord{}).Where("message_id = ?", messageID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *emailRepository) CreateRecord(record *model.EmailRecord) error {
	return r.db.Create(record).Error
}

func (r *emailRepository) CreateAttachment(attachment *model.EmailAttachment) error {
	return r.db.Create(attachment).Error
}

func (r *emailRepository) GetUnprocessed() ([]model.EmailRecord, error) {
	var records []model.EmailRecord
	err := r.db.Where("processed = ?", false).Find(&records).Error
	return records, err
}

func (r *emailRepository) MarkProcessed(id uuid.UUID) error {
	return r.db.Model(&model.EmailRecord{}).Where("id = ?", id).Update("processed", true).Error
}

// --- 邮箱账户管理 ---

func (r *emailRepository) CreateAccount(account *model.UserEmailAccount) error {
	return r.db.Create(account).Error
}

func (r *emailRepository) GetAccountsByUserID(userID uuid.UUID) ([]model.UserEmailAccount, error) {
	var accounts []model.UserEmailAccount
	err := r.db.Where("user_id = ?", userID).Find(&accounts).Error
	return accounts, err
}

func (r *emailRepository) GetAllEnabledAccounts() ([]model.UserEmailAccount, error) {
	var accounts []model.UserEmailAccount
	err := r.db.Where("enabled = ?", true).Find(&accounts).Error
	return accounts, err
}

func (r *emailRepository) DeleteAccount(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.UserEmailAccount{}).Error
}

func (r *emailRepository) CountTotalAccounts() (int64, error) {
	var count int64
	err := r.db.Model(&model.UserEmailAccount{}).Count(&count).Error
	return count, err
}
