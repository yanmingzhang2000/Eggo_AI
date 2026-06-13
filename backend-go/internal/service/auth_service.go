package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/jishengdan/backend-go/internal/dto"
	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/internal/repository"
)

// AuthService 认证服务
type AuthService struct {
	userRepo *repository.UserRepository
	jwtSecret []byte
}

// NewAuthService 创建 AuthService 实例
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

// JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsGuest  bool   `json:"is_guest"`
	jwt.RegisteredClaims
}

// Register 用户注册
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.FindByUsername(req.Username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	existingEmail, _ := s.userRepo.FindByEmail(req.Email)
	if existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Status:       1,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, errors.New("创建用户失败")
	}

	// 生成 JWT Token
	token, err := s.generateToken(user.ID, user.Username, false)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
		IsGuest: false,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.Update(user)

	// 生成 JWT Token
	token, err := s.generateToken(user.ID, user.Username, false)
	if err != nil {
		return nil, errors.New("生成 Token 失败")
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
		IsGuest: false,
	}, nil
}

// GuestLogin 游客登录
func (s *AuthService) GuestLogin() (*dto.GuestResponse, error) {
	// 生成游客 Token（有效期 24 小时）
	token, err := s.generateGuestToken()
	if err != nil {
		return nil, errors.New("生成游客 Token 失败")
	}

	return &dto.GuestResponse{
		Token:   token,
		IsGuest: true,
		User: dto.UserInfo{
			ID:       "guest",
			Username: "游客用户",
			Email:    "",
		},
	}, nil
}

// generateToken 生成 JWT Token
func (s *AuthService) generateToken(userID, username string, isGuest bool) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		IsGuest:  isGuest,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// generateGuestToken 生成游客 Token
func (s *AuthService) generateGuestToken() (string, error) {
	claims := &Claims{
		UserID:   "guest",
		Username: "游客用户",
		IsGuest:  true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ParseToken 解析 JWT Token
func (s *AuthService) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的 Token")
}
