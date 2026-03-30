package controller

import (
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UkeyController interface {
	CreateUkey(c *gin.Context)
	ListUkeys(c *gin.Context)
	DeleteUkey(c *gin.Context)
}

type ukeyController struct {
	serv   service.UkeyService
	domain string
}

func NewUkeyController(serv service.UkeyService, domain string) UkeyController {
	return &ukeyController{serv: serv, domain: domain}
}

func (ctrl *ukeyController) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := userIDValue.(uuid.UUID)
	return userID, ok
}

type CreateUkeyRequest struct {
	Name string `json:"name" binding:"required"`
}

func (ctrl *ukeyController) CreateUkey(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	var req CreateUkeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误")
		return
	}

	secret, err := ctrl.serv.CreateUkey(userID, req.Name)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40002, err.Error())
		return
	}

	response.Success(c, gin.H{
		"host": ctrl.domain + "/api/v1/bill/image",
		"ukey": "Bearer " + secret,
	})
}

func (ctrl *ukeyController) ListUkeys(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	res, err := ctrl.serv.ListUkeys(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "获取失败")
		return
	}

	response.Success(c, res)
}

func (ctrl *ukeyController) DeleteUkey(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}

	ukeyIDStr := c.Param("id")
	ukeyID, err := uuid.Parse(ukeyIDStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的 ID")
		return
	}

	if err := ctrl.serv.DeleteUkey(ukeyID, userID); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "删除失败")
		return
	}

	response.Success(c, nil)
}
