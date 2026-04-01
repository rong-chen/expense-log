package controller

import (
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminController interface {
	ListUsers(c *gin.Context)
	UpdateUserRole(c *gin.Context)
}

type adminController struct {
	serv service.AdminService
}

func NewAdminController(serv service.AdminService) AdminController {
	return &adminController{
		serv: serv,
	}
}

func (con *adminController) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	users, total, err := con.serv.ListUsers(page, pageSize)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, "获取用户列表失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"users": users,
		"total": total,
	})
}

func (con *adminController) UpdateUserRole(c *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Role   string    `json:"role" binding:"required,oneof=user admin"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	if err := con.serv.UpdateUserRole(req.UserID, req.Role); err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, "更新角色失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

