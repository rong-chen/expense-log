package repository

import (
	"expense-log/internal/model"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LedgerRepository interface {
	CreateLedger(ledger *model.Ledger) error
	GetUserLedgers(userID uuid.UUID) ([]model.Ledger, error)
	GetLedgerByID(id uuid.UUID) (*model.Ledger, error)
	GetLedgerByInviteCode(code string) (*model.Ledger, error)
	AddMember(ledgerID, userID uuid.UUID, role model.LedgerRole) error
	RemoveMember(ledgerID, userID uuid.UUID) error
	IsMember(ledgerID, userID uuid.UUID) bool
	GetLedgerRole(ledgerID, userID uuid.UUID) (model.LedgerRole, error)
}

type ledgerRepository struct {
	db *gorm.DB
}

func NewLedgerRepository(db *gorm.DB) LedgerRepository {
	return &ledgerRepository{db: db}
}

func (r *ledgerRepository) CreateLedger(ledger *model.Ledger) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ledger).Error; err != nil {
			return err
		}
		// 自动把自己加入为 Owner
		member := model.LedgerMember{
			LedgerID: ledger.ID,
			UserID:   ledger.OwnerID,
			Role:     model.LedgerRoleOwner,
		}
		return tx.Create(&member).Error
	})
}

func (r *ledgerRepository) GetUserLedgers(userID uuid.UUID) ([]model.Ledger, error) {
	var ledgers []model.Ledger
	// 查找用户加入的账本，按类型和创建时间排序（个人的排前面）
	err := r.db.
		Joins("JOIN ledger_members ON ledger_members.ledger_id = ledgers.id").
		Where("ledger_members.user_id = ?", userID).
		Order("ledgers.type ASC, ledgers.created_at DESC").
		Find(&ledgers).Error
	return ledgers, err
}

func (r *ledgerRepository) GetLedgerByID(id uuid.UUID) (*model.Ledger, error) {
	var ledger model.Ledger
	err := r.db.Preload("Members").Preload("Members.User").Where("id = ?", id).First(&ledger).Error
	return &ledger, err
}

func (r *ledgerRepository) GetLedgerByInviteCode(code string) (*model.Ledger, error) {
	var ledger model.Ledger
	err := r.db.Where("invite_code = ?", code).First(&ledger).Error
	return &ledger, err
}

func (r *ledgerRepository) AddMember(ledgerID, userID uuid.UUID, role model.LedgerRole) error {
	member := model.LedgerMember{
		LedgerID: ledgerID,
		UserID:   userID,
		Role:     role,
	}
	return r.db.Create(&member).Error
}

func (r *ledgerRepository) RemoveMember(ledgerID, userID uuid.UUID) error {
	return r.db.Where("ledger_id = ? AND user_id = ?", ledgerID, userID).Delete(&model.LedgerMember{}).Error
}

func (r *ledgerRepository) IsMember(ledgerID, userID uuid.UUID) bool {
	var count int64
	r.db.Model(&model.LedgerMember{}).Where("ledger_id = ? AND user_id = ?", ledgerID, userID).Count(&count)
	return count > 0
}

func (r *ledgerRepository) GetLedgerRole(ledgerID, userID uuid.UUID) (model.LedgerRole, error) {
	var member model.LedgerMember
	err := r.db.Where("ledger_id = ? AND user_id = ?", ledgerID, userID).First(&member).Error
	if err != nil {
		return "", fmt.Errorf("user not member of ledger")
	}
	return member.Role, nil
}
