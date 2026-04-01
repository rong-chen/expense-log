package repository

import (
	"expense-log/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvitationRepository interface {
	GetByCode(code string) (*model.InvitationCode, error)
	Create(invitation *model.InvitationCode) error
	MarkAsUsed(code string, userID uuid.UUID) error
	ListAll() ([]model.InvitationCode, error)
}

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) InvitationRepository {
	return &invitationRepository{
		db: db,
	}
}

func (r *invitationRepository) GetByCode(code string) (*model.InvitationCode, error) {
	var invitation model.InvitationCode
	err := r.db.Where("code = ? AND is_used = ?", code, false).First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) Create(invitation *model.InvitationCode) error {
	return r.db.Create(invitation).Error
}

func (r *invitationRepository) MarkAsUsed(code string, userID uuid.UUID) error {
	return r.db.Model(&model.InvitationCode{}).
		Where("code = ?", code).
		Updates(map[string]interface{}{
			"is_used": true,
			"used_by": userID,
		}).Error
}

func (r *invitationRepository) ListAll() ([]model.InvitationCode, error) {
	var invitations []model.InvitationCode
	err := r.db.Order("created_at desc").Find(&invitations).Error
	if err != nil {
		return nil, err
	}
	return invitations, nil
}
