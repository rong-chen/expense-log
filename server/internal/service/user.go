package service

import (
	"errors"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"expense-log/pkg/utils"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req *model.RegisterRequest) (*model.TokenResponse, error)
	Login(req *model.LoginRequest) (*model.TokenResponse, error)
	RefreshToken(refreshToken string) (*model.TokenResponse, error)
	GetUserInfo(userID uuid.UUID) (*model.UserInfoResponse, error)
	UpdatePassword(userID uuid.UUID, oldPwd, newPwd string) error
}
type userService struct {
	repo          repository.UserRepository
	jwtSecret     []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewUserService(repo repository.UserRepository, jwtCfg model.JWTConfig) UserService {
	accessExpire, err := time.ParseDuration(jwtCfg.AccessTokenExpire)
	if err != nil {
		accessExpire = 15 * time.Minute
	}
	refreshExpire, err := time.ParseDuration(jwtCfg.RefreshTokenExpire)
	if err != nil {
		refreshExpire = 7 * 24 * time.Hour
	}
	return &userService{
		repo:          repo,
		jwtSecret:     []byte(jwtCfg.Secret),
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
	}
}

// Register 注册新用户
func (u *userService) Register(req *model.RegisterRequest) (*model.TokenResponse, error) {
	// 1. 检测手机号是否已注册
	exists, err := u.repo.PhoneExists(req.Phone)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("该手机号已注册")
	}

	// 2. bcrypt 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 3. 构建用户，设置默认昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = "用户" + req.Phone[len(req.Phone)-4:]
	}

	user := &model.User{
		Phone:     req.Phone,
		Password:  string(hashedPassword),
		Nickname:  nickname,
		LastLogin: time.Now().Unix(),
	}

	// 4. 创建用户
	if err := u.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// 5. 生成双 Token
	return u.generateTokenPair(user.ID)
}

// Login 用户登录
func (u *userService) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	// 1. 查找用户
	user, err := u.repo.GetUserByPhone(req.Phone)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("手机号未注册")
		}
		return nil, err
	}

	// 2. 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 3. 更新最后登录时间
	_ = u.repo.UpdateLastLogin(user.ID, time.Now().Unix())

	// 4. 生成双 Token
	return u.generateTokenPair(user.ID)
}

// RefreshToken 使用 Refresh Token 换取新的双 Token
func (u *userService) RefreshToken(refreshToken string) (*model.TokenResponse, error) {
	// 1. 解析 Token
	claims, err := utils.ParseToken(refreshToken, u.jwtSecret)
	if err != nil {
		return nil, errors.New("Refresh Token 无效或已过期")
	}

	// 2. 校验必须是 Refresh Token
	if claims.TokenType != utils.TokenTypeRefresh {
		return nil, errors.New("请提供 Refresh Token")
	}

	// 3. 确认用户仍然存在
	_, err = u.repo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 4. 生成新的双 Token
	return u.generateTokenPair(claims.UserID)
}

// GetUserInfo 获取用户信息
func (u *userService) GetUserInfo(userID uuid.UUID) (*model.UserInfoResponse, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &model.UserInfoResponse{
		UID:       user.UID,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Email:     user.Email,
		LastLogin: user.LastLogin,
	}, nil
}

// generateTokenPair 内部方法：生成双 Token
func (u *userService) generateTokenPair(userID uuid.UUID) (*model.TokenResponse, error) {
	accessToken, refreshToken, err := utils.CreateTokenPair(userID, u.jwtSecret, u.accessExpire, u.refreshExpire)
	if err != nil {
		return nil, errors.New("Token 生成失败")
	}
	return &model.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// UpdatePassword 修改密码
func (u *userService) UpdatePassword(userID uuid.UUID, oldPwd, newPwd string) error {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPwd)); err != nil {
		return errors.New("原密码错误")
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	
	return u.repo.UpdatePassword(userID, string(hashedPassword))
}
