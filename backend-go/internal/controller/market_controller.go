package controller

import (
	"log"
	"net/http"
	"strconv"
	"strings"

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
		"1.000001",  // 上证指数
		"0.399001",  // 深证成指
		"0.399006",  // 创业板指
		"1.000300",  // 沪深300
		"1.000905",  // 中证500
		"1.000852",  // 中证1000
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
	daysStr := ctx.DefaultQuery("days", "120")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 120
	}

	log.Printf("[MarketController] GET /api/v1/market/history?code=%s&days=%d", code, days)

	data, err := c.svc.FetchIndexHistory(code, days)
	if err != nil {
		log.Printf("[MarketController] 获取历史K线失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50005, "获取历史K线失败")
		return
	}

	response.OK(ctx, data)
}

// GetFundDistribution GET /api/v1/market/fund-distribution
// 获取全市场基金当日涨跌分布
func (c *MarketController) GetFundDistribution(ctx *gin.Context) {
	log.Printf("[MarketController] GET /api/v1/market/fund-distribution")

	data, err := c.svc.FetchFundDistribution()
	if err != nil {
		log.Printf("[MarketController] 获取基金涨跌分布失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50006, "获取基金涨跌分布失败")
		return
	}

	response.OK(ctx, data)
}

// GetFundQuotes GET /api/v1/market/fund-quotes?codes=110011,161725
// 批量获取基金实时估值
func (c *MarketController) GetFundQuotes(ctx *gin.Context) {
	codesStr := ctx.Query("codes")
	if codesStr == "" {
		response.OK(ctx, []service.FundQuote{})
		return
	}

	codes := strings.Split(codesStr, ",")
	log.Printf("[MarketController] GET /api/v1/market/fund-quotes codes=%v", codes)

	data, err := c.svc.FetchFundQuotes(codes)
	if err != nil {
		log.Printf("[MarketController] 获取基金估值失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50007, "获取基金估值失败")
		return
	}

	response.OK(ctx, data)
}
