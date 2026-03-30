package controller

import (
	"expense-log/internal/model"
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmailController interface {
	BindEmail(*gin.Context)
	GetEmails(*gin.Context)
	DeleteEmail(*gin.Context)
}

type emailController struct {
	serv service.EmailService
}

func NewEmailController(serv service.EmailService) EmailController {
	return &emailController{serv: serv}
}

// BindEmail 绑定邮箱账户
func (e *emailController) BindEmail(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	var req model.BindEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}
	account, err := e.serv.BindAccount(userID, &req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40010, err.Error())
		return
	}
	response.Success(c, account)
}

// GetEmails 获取绑定的邮箱列表
func (e *emailController) GetEmails(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	accounts, err := e.serv.GetAccounts(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.Success(c, accounts)
}

// DeleteEmail 解绑邮箱
func (e *emailController) DeleteEmail(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "无效的邮箱ID")
		return
	}
	if err := e.serv.DeleteAccount(id, userID); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, err.Error())
		return
	}
	response.Success(c, nil)
}

// getUserID 从 JWT 中间件上下文获取 userID 的通用方法
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, 40100, "未获取到用户信息")
		return uuid.Nil, false
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Fail(c, http.StatusInternalServerError, 50001, "用户ID解析失败")
		return uuid.Nil, false
	}
	return userID, true
}
