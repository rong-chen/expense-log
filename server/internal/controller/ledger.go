package controller

import (
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LedgerController interface {
	CreateSharedLedger(c *gin.Context)
	GetUserLedgers(c *gin.Context)
	JoinLedger(c *gin.Context)
}

type ledgerController struct {
	serv service.LedgerService
}

func NewLedgerController(serv service.LedgerService) LedgerController {
	return &ledgerController{serv: serv}
}

func (ctrl *ledgerController) getUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权的访问")
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		// 容错处理：部分中间件中可能会把UUID存为string
		if strID, ok := val.(string); ok {
			pID, err := uuid.Parse(strID)
			if err == nil {
				return pID, true
			}
		}
		response.Fail(c, http.StatusInternalServerError, 50000, "用户ID格式错误")
		return uuid.Nil, false
	}
	return id, true
}

type createLedgerRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (ctrl *ledgerController) CreateSharedLedger(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		return
	}

	var req createLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "参数错误")
		return
	}

	ledger, err := ctrl.serv.CreateSharedLedger(userID, req.Name, req.Description)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, err.Error())
		return
	}

	response.Success(c, ledger)
}

func (ctrl *ledgerController) GetUserLedgers(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		return
	}

	ledgers, err := ctrl.serv.GetUserLedgers(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50000, "获取账本列表失败")
		return
	}

	response.Success(c, ledgers)
}

type joinLedgerRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

func (ctrl *ledgerController) JoinLedger(c *gin.Context) {
	userID, ok := ctrl.getUserID(c)
	if !ok {
		return
	}

	var req joinLedgerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, "参数错误")
		return
	}

	ledger, err := ctrl.serv.JoinLedgerByCode(userID, req.InviteCode)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40000, err.Error())
		return
	}

	response.Success(c, ledger)
}
