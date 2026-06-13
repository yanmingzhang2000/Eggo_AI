package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jishengdan/backend-go/internal/dto"
	"github.com/jishengdan/backend-go/internal/service"
	"github.com/jishengdan/backend-go/pkg/response"
)

// AuthController 认证控制器
type AuthController struct {
	svc *service.AuthService
}

// NewAuthController 创建 AuthController 实例
func NewAuthController(svc *service.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

// Register POST /api/v1/auth/register
func (c *AuthController) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	log.Printf("[AuthController] Register: username=%s", req.Username)

	result, err := c.svc.Register(req)
	if err != nil {
		log.Printf("[AuthController] Register failed: %v", err)
		response.Fail(ctx, http.StatusBadRequest, 40002, err.Error())
		return
	}

	response.OK(ctx, result)
}

// Login POST /api/v1/auth/login
func (c *AuthController) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	log.Printf("[AuthController] Login: username=%s", req.Username)

	result, err := c.svc.Login(req)
	if err != nil {
		log.Printf("[AuthController] Login failed: %v", err)
		response.Fail(ctx, http.StatusUnauthorized, 40101, err.Error())
		return
	}

	response.OK(ctx, result)
}

// GuestLogin POST /api/v1/auth/guest
func (c *AuthController) GuestLogin(ctx *gin.Context) {
	log.Printf("[AuthController] GuestLogin")

	result, err := c.svc.GuestLogin()
	if err != nil {
		log.Printf("[AuthController] GuestLogin failed: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50001, err.Error())
		return
	}

	response.OK(ctx, result)
}
