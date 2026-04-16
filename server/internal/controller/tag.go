package controller

import (
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"expense-log/pkg/response"
	"fmt"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TagController interface {
	ListTags(c *gin.Context)
	CreateTag(c *gin.Context)
	DeleteTag(c *gin.Context)
	SetBillTags(c *gin.Context)
	GetBillTags(c *gin.Context)
}

type tagController struct {
	repo repository.TagRepository
}

func NewTagController(repo repository.TagRepository) TagController {
	return &tagController{repo: repo}
}

func (ctrl *tagController) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := userIDValue.(uuid.UUID)
	return userID, ok
}

// ListTags 获取当前用户的所有标签
func (ctrl *tagController) ListTags(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	tags, err := ctrl.repo.GetByUserID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "获取标签失败")
		return
	}
	if tags == nil {
		tags = []model.Tag{}
	}
	response.Success(c, tags)
}

type createTagRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
}

// CreateTag 创建标签
func (ctrl *tagController) CreateTag(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "数据格式错误")
		return
	}

	name := html.EscapeString(req.Name)
	if len(name) == 0 || len(name) > 50 {
		response.Fail(c, http.StatusBadRequest, 40000, "标签名称长度应在1-50字符之间")
		return
	}

	// 检查重复
	exists, _ := ctrl.repo.ExistsByName(userID, name)
	if exists {
		response.Fail(c, http.StatusConflict, 40900, "标签名称已存在")
		return
	}

	color := req.Color
	if color == "" {
		color = "#3498db"
	}

	tag := &model.Tag{
		UserID: userID,
		Name:   name,
		Color:  color,
	}

	if err := ctrl.repo.Create(tag); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "创建标签失败: "+err.Error())
		return
	}

	response.Success(c, tag)
}

// DeleteTag 删除标签
func (ctrl *tagController) DeleteTag(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的标签ID")
		return
	}

	if err := ctrl.repo.Delete(tagID, userID); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "删除标签失败")
		return
	}

	response.Success(c, nil)
}

type setBillTagsRequest struct {
	TagIDs []string `json:"tag_ids"`
}

// SetBillTags 设置账单标签
func (ctrl *tagController) SetBillTags(c *gin.Context) {
	_, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	billID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的账单ID")
		return
	}

	var req setBillTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "数据格式错误")
		return
	}

	var tagIDs []uuid.UUID
	for _, idStr := range req.TagIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			fmt.Printf("[SetBillTags] 跳过无效 tagID: %s, err: %v\n", idStr, err)
			continue
		}
		tagIDs = append(tagIDs, id)
	}

	fmt.Printf("[SetBillTags] billID=%s, 收到 %d 个 tagID, 解析成功 %d 个\n", billID, len(req.TagIDs), len(tagIDs))

	if err := ctrl.repo.SetBillTags(billID, tagIDs); err != nil {
		fmt.Printf("[SetBillTags] 保存失败: %v\n", err)
		response.Fail(c, http.StatusInternalServerError, 50000, "设置标签失败")
		return
	}

	fmt.Printf("[SetBillTags] ✅ 保存成功\n")
	response.Success(c, nil)
}

// GetBillTags 获取账单的标签
func (ctrl *tagController) GetBillTags(c *gin.Context) {
	_, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	billID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "无效的账单ID")
		return
	}

	tags, err := ctrl.repo.GetBillTags(billID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "获取账单标签失败")
		return
	}
	if tags == nil {
		tags = []model.Tag{}
	}

	response.Success(c, tags)
}
