package repository

import (
	"expense-log/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByPhone(phone string) (*model.User, error)
	GetUserByID(id uuid.UUID) (*model.User, error)
	CreateUser(user *model.User) error
	PhoneExists(phone string) (bool, error)
	UpdateLastLogin(id uuid.UUID, timestamp int64) error
	UpdatePassword(id uuid.UUID, password string) error
}
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetUserByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) PhoneExists(phone string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("phone = ?", phone).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) UpdateLastLogin(id uuid.UUID, timestamp int64) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("last_login", timestamp).Error
}

func (r *userRepository) UpdatePassword(id uuid.UUID, password string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password", password).Error
}
