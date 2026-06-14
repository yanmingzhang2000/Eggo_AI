package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jishengdan/backend-go/internal/service"
	"github.com/jishengdan/backend-go/pkg/response"
)

// MarketController 大盘行情控制器
type MarketController struct {
	svc *service.MarketService
}

// NewMarketController 创建 MarketController 实例
func NewMarketController(svc *service.MarketService) *MarketController {
	return &MarketController{svc: svc}
}

// GetIndices GET /api/v1/market/indices
// 获取大盘指数行情
func (c *MarketController) GetIndices(ctx *gin.Context) {
	log.Printf("[MarketController] GET /api/v1/market/indices")

	codes := []string{
		"sh000001",  // 上证指数
		"sz399001",  // 深证成指
		"sz399006",  // 创业板指
		"sh000300",  // 沪深300
		"sh000905",  // 中证500
		"sh000852",  // 中证1000
	}

	data, err := c.svc.FetchSinaIndex(codes)
	if err != nil {
		log.Printf("[MarketController] 获取指数行情失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取指数行情失败")
		return
	}

	response.OK(ctx, data)
}

// GetSectors GET /api/v1/market/sectors
// 获取板块行情
func (c *MarketController) GetSectors(ctx *gin.Context) {
	log.Printf("[MarketController] GET /api/v1/market/sectors")

	sectors, err := c.svc.FetchEastmoneySectors()
	if err != nil {
		log.Printf("[MarketController] 获取板块行情失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50002, "获取板块行情失败")
		return
	}

	response.OK(ctx, sectors)
}

// GetConcepts GET /api/v1/market/concepts
// 获取概念板块行情
func (c *MarketController) GetConcepts(ctx *gin.Context) {
	log.Printf("[MarketController] GET /api/v1/market/concepts")

	concepts, err := c.svc.FetchEastmoneyConcepts()
	if err != nil {
		log.Printf("[MarketController] 获取概念板块失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50003, "获取概念板块失败")
		return
	}

	response.OK(ctx, concepts)
}

// GetOverview GET /api/v1/market/overview
// 获取市场概览（涨跌家数、涨停跌停、北向资金等）
func (c *MarketController) GetOverview(ctx *gin.Context) {
	log.Printf("[MarketController] GET /api/v1/market/overview")

	stats, err := c.svc.FetchMarketOverview()
	if err != nil {
		log.Printf("[MarketController] 获取市场概览失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50004, "获取市场概览失败")
		return
	}

	// 获取北向资金
	stats.NorthFlow = c.svc.FetchNorthFlow()

	response.OK(ctx, stats)
}

// GetIndexOptions GET /api/v1/market/index-options
// 获取可选指数列表
func (c *MarketController) GetIndexOptions(ctx *gin.Context) {
	response.OK(ctx, service.GetIndexOptions())
}

// GetIndexHistory GET /api/v1/market/history
// 获取指数历史K线
func (c *MarketController) GetIndexHistory(ctx *gin.Context) {
	code := ctx.DefaultQuery("code", "1.000300")
	days := 120

	log.Printf("[MarketController] GET /api/v1/market/history?code=%s&days=%d", code, days)

	data, err := c.svc.FetchIndexHistory(code, days)
	if err != nil {
		log.Printf("[MarketController] 获取历史K线失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50005, "获取历史K线失败")
		return
	}

	response.OK(ctx, data)
}
