package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jishengdan/backend-go/internal/service"
	"github.com/jishengdan/backend-go/pkg/response"
)

// EggController 母鸡状态控制器
type EggController struct {
	svc *service.EggService
}

// NewEggController 创建 EggController 实例
func NewEggController(svc *service.EggService) *EggController {
	return &EggController{svc: svc}
}

// GetStatus GET /api/v1/egg/status
// 获取母鸡状态总接口：今日指标 + AI新闻线索 + 三铁律决策
func (c *EggController) GetStatus(ctx *gin.Context) {
	// TODO: 从 JWT 中间件获取 userID
	userID := ctx.Query("user_id")
	if userID == "" {
		userID = "default-user" // 临时默认值，生产环境应从 JWT 获取
	}

	log.Printf("[EggController] GET /api/v1/egg/status, userID=%s", userID)

	result, err := c.svc.GetEggStatus(userID)
	if err != nil {
		log.Printf("[EggController] 获取母鸡状态失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取母鸡状态失败")
		return
	}

	response.OK(ctx, result)
}
