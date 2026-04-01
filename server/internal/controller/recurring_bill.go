package controller

import (
	"expense-log/internal/model"
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"html"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecurringBillController interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	ToggleActive(c *gin.Context)
}

type recurringBillController struct {
	serv service.RecurringBillService
}

func NewRecurringBillController(serv service.RecurringBillService) RecurringBillController {
	return &recurringBillController{serv: serv}
}

func (ctrl *recurringBillController) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := userIDValue.(uuid.UUID)
	return userID, ok
}

type CreateRecurringBillRequest struct {
	Amount     float64 `json:"amount" binding:"required"`
	Merchant   string  `json:"merchant" binding:"required"`
	Category   string  `json:"category"`
	Remark     string  `json:"remark"`
	DayOfMonth int     `json:"day_of_month" binding:"required,min=1,max=31"`
	ExecuteNow bool    `json:"execute_now"`
}

func (ctrl *recurringBillController) Create(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	var req CreateRecurringBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	rb := model.RecurringBill{
		Amount:     req.Amount,
		Merchant:   html.EscapeString(req.Merchant),
		Category:   html.EscapeString(req.Category),
		Remark:     html.EscapeString(req.Remark),
		DayOfMonth: req.DayOfMonth,
	}
	if err := ctrl.serv.Create(userID, &rb, req.ExecuteNow); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "创建失败")
		return
	}

	response.Success(c, rb)
}

func (ctrl *recurringBillController) List(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	list, err := ctrl.serv.List(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "获取失败")
		return
	}

	response.Success(c, list)
}

type UpdateRecurringBillRequest struct {
	Amount     float64 `json:"amount" binding:"required"`
	Merchant   string  `json:"merchant" binding:"required"`
	Category   string  `json:"category"`
	Remark     string  `json:"remark"`
	DayOfMonth int     `json:"day_of_month" binding:"required,min=1,max=31"`
}

func (ctrl *recurringBillController) Update(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID")
		return
	}

	var req UpdateRecurringBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	rb := model.RecurringBill{
		Amount:     req.Amount,
		Merchant:   html.EscapeString(req.Merchant),
		Category:   html.EscapeString(req.Category),
		Remark:     html.EscapeString(req.Remark),
		DayOfMonth: req.DayOfMonth,
	}
	if err := ctrl.serv.Update(userID, id, &rb); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "更新失败")
		return
	}

	response.Success(c, nil)
}

func (ctrl *recurringBillController) Delete(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID")
		return
	}

	if err := ctrl.serv.Delete(userID, id); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "删除失败")
		return
	}

	response.Success(c, nil)
}

func (ctrl *recurringBillController) ToggleActive(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID")
		return
	}

	if err := ctrl.serv.ToggleActive(userID, id); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "切换状态失败")
		return
	}

	response.Success(c, nil)
}
