package controller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jishengdan/backend-go/internal/service"
	"github.com/jishengdan/backend-go/pkg/response"
)

// FundController 基金控制器
type FundController struct {
	svc *service.FundService
}

// NewFundController 创建 FundController 实例
func NewFundController(svc *service.FundService) *FundController {
	return &FundController{svc: svc}
}

// GetDetail GET /api/v1/funds/:code/detail
// 获取单只基金详情
func (c *FundController) GetDetail(ctx *gin.Context) {
	code := ctx.Param("code")
	log.Printf("[FundController] GET /api/v1/funds/%s/detail", code)

	data, err := c.svc.GetFundDetail(code)
	if err != nil {
		log.Printf("[FundController] 获取基金详情失败: %v", err)
		response.Fail(ctx, http.StatusNotFound, 40401, "获取基金详情失败: "+err.Error())
		return
	}

	response.OK(ctx, data)
}

// GetNavHistory GET /api/v1/funds/:code/nav-history
// 获取基金净值历史
func (c *FundController) GetNavHistory(ctx *gin.Context) {
	code := ctx.Param("code")
	daysStr := ctx.DefaultQuery("days", "120")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 120
	}

	log.Printf("[FundController] GET /api/v1/funds/%s/nav-history?days=%d", code, days)

	data, err := c.svc.GetNavHistory(code, days)
	if err != nil {
		log.Printf("[FundController] 获取净值历史失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取净值历史失败")
		return
	}

	response.OK(ctx, data)
}

// GetAnalysis GET /api/v1/funds/:code/analysis
// 获取单只基金三铁律分析建议
func (c *FundController) GetAnalysis(ctx *gin.Context) {
	code := ctx.Param("code")
	log.Printf("[FundController] GET /api/v1/funds/%s/analysis", code)

	data, err := c.svc.AnalyzeFund(code)
	if err != nil {
		log.Printf("[FundController] 获取基金分析失败: %v", err)
		response.Fail(ctx, http.StatusInternalServerError, 50002, "获取基金分析失败")
		return
	}

	response.OK(ctx, data)
}
