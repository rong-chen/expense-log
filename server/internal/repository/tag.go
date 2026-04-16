package repository

import (
	"expense-log/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TagRepository interface {
	Create(tag *model.Tag) error
	GetByUserID(userID uuid.UUID) ([]model.Tag, error)
	GetByID(id uuid.UUID) (*model.Tag, error)
	Delete(id uuid.UUID, userID uuid.UUID) error
	ExistsByName(userID uuid.UUID, name string) (bool, error)

	// 账单-标签关联
	SetBillTags(billID uuid.UUID, tagIDs []uuid.UUID) error
	GetBillTags(billID uuid.UUID) ([]model.Tag, error)
	GetBillIDsByTagID(tagID uuid.UUID) ([]uuid.UUID, error)
}

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(tag *model.Tag) error {
	return r.db.Create(tag).Error
}

func (r *tagRepository) GetByUserID(userID uuid.UUID) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&tags).Error
	return tags, err
}

func (r *tagRepository) GetByID(id uuid.UUID) (*model.Tag, error) {
	var tag model.Tag
	err := r.db.Where("id = ?", id).First(&tag).Error
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) Delete(id uuid.UUID, userID uuid.UUID) error {
	// 先删除关联关系
	r.db.Where("tag_id = ?", id).Delete(&model.BillTag{})
	// 再删除标签
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Tag{}).Error
}

func (r *tagRepository) ExistsByName(userID uuid.UUID, name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Tag{}).Where("user_id = ? AND name = ?", userID, name).Count(&count).Error
	return count > 0, err
}

// SetBillTags 设置账单的标签（先清后写）
func (r *tagRepository) SetBillTags(billID uuid.UUID, tagIDs []uuid.UUID) error {
	// 删除旧的关联
	if err := r.db.Where("bill_id = ?", billID).Delete(&model.BillTag{}).Error; err != nil {
		return err
	}
	// 批量插入新关联
	if len(tagIDs) > 0 {
		var billTags []model.BillTag
		for _, tid := range tagIDs {
			billTags = append(billTags, model.BillTag{BillID: billID, TagID: tid})
		}
		return r.db.Create(&billTags).Error
	}
	return nil
}

// GetBillTags 获取账单关联的所有标签
func (r *tagRepository) GetBillTags(billID uuid.UUID) ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.
		Joins("JOIN bill_tags ON bill_tags.tag_id = tags.id").
		Where("bill_tags.bill_id = ?", billID).
		Find(&tags).Error
	return tags, err
}

// GetBillIDsByTagID 获取某标签下的所有账单ID
func (r *tagRepository) GetBillIDsByTagID(tagID uuid.UUID) ([]uuid.UUID, error) {
	var billIDs []uuid.UUID
	err := r.db.Model(&model.BillTag{}).Where("tag_id = ?", tagID).Pluck("bill_id", &billIDs).Error
	return billIDs, err
}
