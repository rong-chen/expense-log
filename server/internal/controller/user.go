package controller

import (
	"expense-log/internal/model"
	"expense-log/internal/service"
	"expense-log/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetUserInfo(c *gin.Context)
	Logout(c *gin.Context)
	UpdatePassword(c *gin.Context)
}
type userController struct {
	serv          service.UserService
	refreshExpire time.Duration
}

func NewUserController(serv service.UserService, jwtCfg model.JWTConfig) UserController {
	refreshExpire, err := time.ParseDuration(jwtCfg.RefreshTokenExpire)
	if err != nil {
		refreshExpire = 7 * 24 * time.Hour
	}
	return &userController{
		serv:          serv,
		refreshExpire: refreshExpire,
	}
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// Register 用户注册
func (u *userController) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}
	tokens, err := u.serv.Register(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40002, err.Error())
		return
	}
	u.setRefreshTokenCookie(c, tokens.RefreshToken)
	response.Success(c, gin.H{"access_token": tokens.AccessToken})
}

// Login 用户登录
func (u *userController) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}
	tokens, err := u.serv.Login(&req)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, 40003, err.Error())
		return
	}
	u.setRefreshTokenCookie(c, tokens.RefreshToken)
	response.Success(c, gin.H{"access_token": tokens.AccessToken})
}

// RefreshToken 刷新 Token（从 HttpOnly Cookie 中读取 Refresh Token）
func (u *userController) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		response.Fail(c, http.StatusUnauthorized, 40104, "缺少 Refresh Token")
		return
	}
	tokens, err := u.serv.RefreshToken(refreshToken)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, 40105, err.Error())
		return
	}
	u.setRefreshTokenCookie(c, tokens.RefreshToken)
	response.Success(c, gin.H{"access_token": tokens.AccessToken})
}

// GetUserInfo 获取当前登录用户信息
func (u *userController) GetUserInfo(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, 40100, "未获取到用户信息")
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		response.Fail(c, http.StatusInternalServerError, 50001, "用户ID解析失败")
		return
	}
	userInfo, err := u.serv.GetUserInfo(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 50002, err.Error())
		return
	}
	response.Success(c, userInfo)
}

// setRefreshTokenCookie 将 Refresh Token 写入 HttpOnly Cookie
func (u *userController) setRefreshTokenCookie(c *gin.Context, token string) {
	c.SetCookie(
		"refresh_token",                // name
		token,                          // value
		int(u.refreshExpire.Seconds()), // maxAge (秒)
		"/api/v1/user/refresh",         // path: 仅刷新接口携带
		"",                             // domain: 自动匹配
		false,                          // secure: 允许 HTTP 传输
		true,                           // httpOnly: JS 无法读取
	)
}

// Logout 退出登录
func (u *userController) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/api/v1/user/refresh", "", false, true)
	response.Success(c, nil)
}

// UpdatePassword 修改密码
func (u *userController) UpdatePassword(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		response.Fail(c, http.StatusUnauthorized, 40100, "未授权")
		return
	}
	userID := userIDValue.(uuid.UUID)

	var req UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 40001, "参数不完整")
		return
	}

	if err := u.serv.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		response.Fail(c, http.StatusBadRequest, 40002, err.Error())
		return
	}
	response.Success(c, nil)
}
