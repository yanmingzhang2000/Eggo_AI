package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jishengdan/backend-go/internal/dto"
	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/internal/repository"
	"github.com/jishengdan/backend-go/internal/service"
	"github.com/jishengdan/backend-go/pkg/response"
)

// WatchlistController 自选基金控制器
type WatchlistController struct {
	repo    *repository.WatchlistRepository
	fundSvc *service.FundService
}

// NewWatchlistController 创建 WatchlistController 实例
func NewWatchlistController(repo *repository.WatchlistRepository, fundSvc *service.FundService) *WatchlistController {
	return &WatchlistController{repo: repo, fundSvc: fundSvc}
}

// List GET /api/v1/watchlist — 获取当前用户自选列表
func (c *WatchlistController) List(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		response.Fail(ctx, http.StatusUnauthorized, 40101, "未登录")
		return
	}

	items, err := c.repo.FindByUserID(userID)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取自选列表失败")
		return
	}

	result := make([]dto.WatchlistItem, 0, len(items))
	for _, item := range items {
		wi := dto.WatchlistItem{
			FundCode: item.Fund.FundCode,
			FundName: item.Fund.FundName,
			FundType: item.Fund.FundType,
			Remark:   item.Remark,
		}
		result = append(result, wi)
	}

	response.OK(ctx, result)
}

// Add POST /api/v1/watchlist — 添加自选
func (c *WatchlistController) Add(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		response.Fail(ctx, http.StatusUnauthorized, 40101, "未登录")
		return
	}

	var req dto.AddWatchlistRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "参数错误: "+err.Error())
		return
	}

	// 确保基金存在（自动播种）
	fund, err := c.fundSvc.EnsureFundPublic(req.FundCode)
	if err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40002, "基金不存在: "+err.Error())
		return
	}

	// 检查是否已收藏
	exists, _ := c.repo.Exists(userID, fund.ID)
	if exists {
		response.OK(ctx, gin.H{"already": true})
		return
	}

	var remark *string
	if req.Remark != "" {
		remark = &req.Remark
	}

	w := &model.Watchlist{
		UserID: userID,
		FundID: fund.ID,
		Remark: remark,
	}
	if err := c.repo.Create(w); err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50002, "添加自选失败")
		return
	}

	response.OK(ctx, gin.H{"added": true, "fundCode": fund.FundCode, "fundName": fund.FundName})
}

// Remove DELETE /api/v1/watchlist/:code — 取消自选
func (c *WatchlistController) Remove(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		response.Fail(ctx, http.StatusUnauthorized, 40101, "未登录")
		return
	}

	fundCode := ctx.Param("code")

	// 查基金 ID
	fund, err := c.fundSvc.EnsureFundPublic(fundCode)
	if err != nil {
		response.Fail(ctx, http.StatusNotFound, 40401, "基金不存在")
		return
	}

	if err := c.repo.Delete(userID, fund.ID); err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50003, "取消自选失败")
		return
	}

	response.OK(ctx, gin.H{"removed": true})
}

// Check GET /api/v1/watchlist/check/:code — 检查是否已收藏
func (c *WatchlistController) Check(ctx *gin.Context) {
	userID := ctx.GetString("userID")
	if userID == "" {
		response.OK(ctx, gin.H{"watched": false})
		return
	}

	fundCode := ctx.Param("code")
	fund, err := c.fundSvc.EnsureFundPublic(fundCode)
	if err != nil {
		response.OK(ctx, gin.H{"watched": false})
		return
	}

	exists, _ := c.repo.Exists(userID, fund.ID)
	response.OK(ctx, gin.H{"watched": exists})
}
