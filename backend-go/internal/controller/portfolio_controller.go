package controller

import (
	"log"
	"net/http"
	"strconv"

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

func getUserID(ctx *gin.Context) int64 {
	uid, _ := ctx.Get("userID")
	if uid == nil {
		return 1
	}
	return uid.(int64)
}

func getAccountID(ctx *gin.Context) (int64, bool) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		response.Fail(ctx, http.StatusBadRequest, 40001, "无效的账户 ID")
		return 0, false
	}
	return id, true
}

// ── 鸡笼管理 ──────────────────────────────────────────────────────────────

// ListAccounts GET /api/v1/portfolio/accounts
func (c *PortfolioController) ListAccounts(ctx *gin.Context) {
	userID := getUserID(ctx)
	summaries, err := c.svc.ListAccounts(userID)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取鸡笼列表失败")
		return
	}
	if summaries == nil {
		summaries = []service.AccountSummary{}
	}
	response.OK(ctx, summaries)
}

// CreateAccount POST /api/v1/portfolio/accounts
func (c *PortfolioController) CreateAccount(ctx *gin.Context) {
	var req struct {
		Name           string  `json:"name"`
		InitialBalance float64 `json:"initialBalance" binding:"required,min=1000"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "初始资金至少 1000 元")
		return
	}

	userID := getUserID(ctx)
	log.Printf("[PortfolioController] CreateAccount userId=%d name=%s balance=%.0f", userID, req.Name, req.InitialBalance)

	account, err := c.svc.CreateAccount(userID, req.Name, req.InitialBalance)
	if err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40002, err.Error())
		return
	}

	response.OK(ctx, account)
}

// DeleteAccount DELETE /api/v1/portfolio/accounts/:id
func (c *PortfolioController) DeleteAccount(ctx *gin.Context) {
	accountID, ok := getAccountID(ctx)
	if !ok {
		return
	}

	if err := c.svc.DeleteAccount(accountID); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40003, err.Error())
		return
	}

	response.OK(ctx, gin.H{"message": "鸡笼已删除"})
}

// ── 账户详情 ──────────────────────────────────────────────────────────────

// GetAccount GET /api/v1/portfolio/accounts/:id
func (c *PortfolioController) GetAccount(ctx *gin.Context) {
	accountID, ok := getAccountID(ctx)
	if !ok {
		return
	}

	summary, err := c.svc.GetAccountSummary(accountID)
	if err != nil {
		response.Fail(ctx, http.StatusNotFound, 40401, err.Error())
		return
	}
	response.OK(ctx, summary)
}

// ── 持仓 ──────────────────────────────────────────────────────────────────

// GetPositions GET /api/v1/portfolio/accounts/:id/positions
func (c *PortfolioController) GetPositions(ctx *gin.Context) {
	accountID, ok := getAccountID(ctx)
	if !ok {
		return
	}

	positions, err := c.svc.GetPositions(accountID)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取持仓失败")
		return
	}
	if positions == nil {
		positions = []service.PositionView{}
	}
	response.OK(ctx, positions)
}

// ── 买入 ──────────────────────────────────────────────────────────────────

// Buy POST /api/v1/portfolio/accounts/:id/buy
func (c *PortfolioController) Buy(ctx *gin.Context) {
	accountID, ok := getAccountID(ctx)
	if !ok {
		return
	}

	var req service.BuyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40001, "参数错误: fundCode + amount 必填")
		return
	}

	log.Printf("[PortfolioController] Buy accountId=%d fund=%s amount=%.0f", accountID, req.FundCode, req.Amount)

	order, err := c.svc.Buy(accountID, &req)
	if err != nil {
		response.Fail(ctx, http.StatusBadRequest, 40004, err.Error())
		return
	}

	response.OK(ctx, order)
}

// ── 待结算订单 ──────────────────────────────────────────────────────────────

// GetPendingOrders GET /api/v1/portfolio/accounts/:id/orders/pending
func (c *PortfolioController) GetPendingOrders(ctx *gin.Context) {
	accountID, ok := getAccountID(ctx)
	if !ok {
		return
	}

	orders, err := c.svc.GetPendingOrders(accountID)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取订单失败")
		return
	}
	if orders == nil {
		orders = []model.VirtualOrder{}
	}
	response.OK(ctx, orders)
}

// ── 交易流水 ──────────────────────────────────────────────────────────────

// GetTransactions GET /api/v1/portfolio/accounts/:id/transactions
func (c *PortfolioController) GetTransactions(ctx *gin.Context) {
	accountID, ok := getAccountID(ctx)
	if !ok {
		return
	}

	txs, err := c.svc.GetTransactions(accountID, 50)
	if err != nil {
		response.Fail(ctx, http.StatusInternalServerError, 50001, "获取流水失败")
		return
	}
	if txs == nil {
		txs = []model.VirtualTransaction{}
	}
	response.OK(ctx, txs)
}
