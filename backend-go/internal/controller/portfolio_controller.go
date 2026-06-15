package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/internal/service"
	"github.com/jishengdan/backend-go/pkg/response"
)

// PortfolioController 虚拟盘控制器
type PortfolioController struct {
	svc *service.PortfolioService
}

func NewPortfolioController(svc *service.PortfolioService) *PortfolioController {
	return &PortfolioController{svc: svc}
}

// 从上下文获取 user_id（中间件设置）
func getUserID(ctx *gin.Context) int64 {
	uid, _ := ctx.Get("userID")
	if uid == nil {
		// 开发模式：默认用户 ID 为 1
		return 1
	}
	return uid.(int64)
}

// CreateAccount POST /api/v1/portfolio/account
func (c *PortfolioController) CreateAccount(ctx *gin.Context) {
	var req struct {
		InitialBalance float64 `json:"initialBalance" binding:"required,min=1000"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "初始资金至少 1000 元")
		return
	}

	userID := getUserID(ctx)
	log.Printf("[PortfolioController] CreateAccount userId=%d balance=%.0f", userID, req.InitialBalance)

	account, err := c.svc.CreateAccount(userID, req.InitialBalance)
	if err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40002, err.Error())
		return
	}

	response.OK(ctx, account)
}

// GetAccount GET /api/v1/portfolio/account
func (c *PortfolioController) GetAccount(ctx *gin.Context) {
	userID := getUserID(ctx)
	summary, err := c.svc.GetAccountSummary(userID)
	if err != nil {
		response.Fail(ctx, http.StatusNotFound, 40401, err.Error())
		return
	}
	response.OK(ctx, summary)
}

// GetPositions GET /api/v1/portfolio/positions
func (c *PortfolioController) GetPositions(ctx *gin.Context) {
	userID := getUserID(ctx)
	positions, err := c.svc.GetPositions(userID)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取持仓失败")
		return
	}
	if positions == nil {
		positions = []service.PositionView{}
	}
	response.OK(ctx, positions)
}

// Buy POST /api/v1/portfolio/buy
func (c *PortfolioController) Buy(ctx *gin.Context) {
	var req service.BuyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "参数错误: fundCode + amount 必填")
		return
	}

	userID := getUserID(ctx)
	log.Printf("[PortfolioController] Buy userId=%d fund=%s amount=%.0f", userID, req.FundCode, req.Amount)

	order, err := c.svc.Buy(userID, &req)
	if err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40003, err.Error())
		return
	}

	response.OK(ctx, order)
}

// GetPendingOrders GET /api/v1/portfolio/orders/pending
func (c *PortfolioController) GetPendingOrders(ctx *gin.Context) {
	userID := getUserID(ctx)
	orders, err := c.svc.GetPendingOrders(userID)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取订单失败")
		return
	}
	if orders == nil {
		orders = []model.VirtualOrder{}
	}
	response.OK(ctx, orders)
}

// GetTransactions GET /api/v1/portfolio/transactions
func (c *PortfolioController) GetTransactions(ctx *gin.Context) {
	userID := getUserID(ctx)
	txs, err := c.svc.GetTransactions(userID, 50)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取流水失败")
		return
	}
	if txs == nil {
		txs = []model.VirtualTransaction{}
	}
	response.OK(ctx, txs)
}
