package service

import (
	"context"
	"errors"
	"expense-log/internal/model"
	"expense-log/internal/repository"
	"expense-log/pkg/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req *model.RegisterRequest) (*model.TokenResponse, error)
	Login(req *model.LoginRequest) (*model.TokenResponse, error)
	RefreshToken(refreshToken string) (*model.TokenResponse, error)
	GetUserInfo(userID uuid.UUID) (*model.UserInfoResponse, error)
	UpdatePassword(userID uuid.UUID, oldPwd, newPwd string) error
	Logout(userID uuid.UUID) error
}
type userService struct {
	repo          repository.UserRepository
	invitationRepo repository.InvitationRepository
	rdb           *redis.Client
	jwtSecret     []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
}

func NewUserService(repo repository.UserRepository, invitationRepo repository.InvitationRepository, rdb *redis.Client, jwtCfg model.JWTConfig) UserService {
	accessExpire, err := time.ParseDuration(jwtCfg.AccessTokenExpire)
	if err != nil {
		accessExpire = 15 * time.Minute
	}
	refreshExpire, err := time.ParseDuration(jwtCfg.RefreshTokenExpire)
	if err != nil {
		refreshExpire = 7 * 24 * time.Hour
	}
	return &userService{
		repo:           repo,
		invitationRepo: invitationRepo,
		rdb:            rdb,
		jwtSecret:      []byte(jwtCfg.Secret),
		accessExpire:   accessExpire,
		refreshExpire:  refreshExpire,
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

	// 1.1. 校验邀请码
	invitation, err := u.invitationRepo.GetByCode(req.InvitationCode)
	if err != nil {
		return nil, errors.New("邀请码无效或已被使用")
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
		Role:      invitation.Role, // 设置为邀请码预设的角色
		LastLogin: time.Now().Unix(),
	}

	// 4. 创建用户
	if err := u.repo.CreateUser(user); err != nil {
		return nil, err
	}

	// 4.1. 标记邀请码为已使用
	_ = u.invitationRepo.MarkAsUsed(invitation.Code, user.ID)

	// 5. 生成双 Token
	return u.generateTokenPair(user.ID)
}

// Login 用户登录
func (u *userService) Login(req *model.LoginRequest) (*model.TokenResponse, error) {
	ctx := context.Background()
	lockKey := "login_lock:" + req.Phone
	failKey := "login_fail:" + req.Phone

	// 0. 检查是否已被锁定
	if u.rdb.Exists(ctx, lockKey).Val() > 0 {
		ttl, _ := u.rdb.TTL(ctx, lockKey).Result()
		return nil, fmt.Errorf("登录失败次数过多，请 %d 分钟后再试", int(ttl.Minutes())+1)
	}

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
		// 密码错误，记录失败次数
		failCount, _ := u.rdb.Incr(ctx, failKey).Result()
		if failCount == 1 {
			u.rdb.Expire(ctx, failKey, 15*time.Minute)
		}
		if failCount >= 5 {
			// 达到5次，锁定15分钟
			u.rdb.Set(ctx, lockKey, "locked", 15*time.Minute)
			u.rdb.Del(ctx, failKey)
			return nil, errors.New("登录失败次数过多，账号已锁定15分钟")
		}
		remaining := 5 - failCount
		return nil, fmt.Errorf("密码错误，还可尝试 %d 次", remaining)
	}

	// 3. 登录成功，清除失败记录
	u.rdb.Del(ctx, failKey, lockKey)

	// 4. 更新最后登录时间
	_ = u.repo.UpdateLastLogin(user.ID, time.Now().Unix())

	// 5. 生成双 Token
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
		Role:      user.Role,
	}, nil
}

// generateTokenPair 内部方法：生成双 Token (带 SSO 会话逻辑)
func (u *userService) generateTokenPair(userID uuid.UUID) (*model.TokenResponse, error) {
	// 生成此次登录/刷新的专属 SessionID
	sessionID, _ := uuid.NewV7()
	sessionIDStr := sessionID.String()

	// 存入 Redis，有效期与 Refresh Token 保持一致。这会覆盖该用户之前的旧 session，实现单设备登录踢人机制。
	ctx := context.Background()
	err := u.rdb.Set(ctx, "session:"+userID.String(), sessionIDStr, u.refreshExpire).Err()
	if err != nil {
		return nil, errors.New("会话初始化失败")
	}

	accessToken, refreshToken, err := utils.CreateTokenPair(userID, sessionIDStr, u.jwtSecret, u.accessExpire, u.refreshExpire)
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

// Logout 登出：废止 Redis 中的 Session，使所有旧 Token 立即失效
func (u *userService) Logout(userID uuid.UUID) error {
	ctx := context.Background()
	return u.rdb.Del(ctx, "session:"+userID.String()).Err()
}
