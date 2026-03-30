package repository

import (
	"expense-log/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UkeyRepository interface {
	Create(ukey *model.Ukey) error
	GetByID(id uuid.UUID) (*model.Ukey, error)
	GetByUserID(userID uuid.UUID) ([]*model.Ukey, error)
	GetBySecret(secret string) (*model.Ukey, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
}

type ukeyRepository struct {
	db *gorm.DB
}

func NewUkeyRepository(db *gorm.DB) UkeyRepository {
	return &ukeyRepository{db: db}
}

func (r *ukeyRepository) Create(ukey *model.Ukey) error {
	return r.db.Create(ukey).Error
}

func (r *ukeyRepository) GetByID(id uuid.UUID) (*model.Ukey, error) {
	var ukey model.Ukey
	err := r.db.Where("id = ?", id).Take(&ukey).Error
	return &ukey, err
}

func (r *ukeyRepository) GetByUserID(userID uuid.UUID) ([]*model.Ukey, error) {
	var ukeys []*model.Ukey
	err := r.db.Where("user_id = ?", userID).Find(&ukeys).Error
	return ukeys, err
}

func (r *ukeyRepository) GetBySecret(secret string) (*model.Ukey, error) {
	var ukey model.Ukey
	err := r.db.Where("secret_key = ?", secret).Take(&ukey).Error
	return &ukey, err
}

func (r *ukeyRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Ukey{}).Error
}
