package controller

import (
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type InvitationController interface {
	Generate(c *gin.Context)
	List(c *gin.Context)
}

type invitationController struct {
	serv service.InvitationService
}

func NewInvitationController(serv service.InvitationService) InvitationController {
	return &invitationController{
		serv: serv,
	}
}

type GenerateRequest struct {
	Count int    `json:"count" binding:"required,min=1,max=100"`
	Role  string `json:"role" binding:"omitempty,oneof=user admin"`
}

func (con *invitationController) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	codes, err := con.serv.Generate(req.Count, req.Role)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, "生成邀请码失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{"codes": codes})
}

func (con *invitationController) List(c *gin.Context) {
	invitations, err := con.serv.List()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50001, "获取邀请码列表失败: "+err.Error())
		return
	}

	response.Success(c, invitations)
}
