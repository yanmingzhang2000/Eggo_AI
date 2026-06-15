package service

import (
	"fmt"
	"math"
	"time"

	"github.com/jishengdan/backend-go/internal/model"
	"github.com/jishengdan/backend-go/internal/repository"
)

// PortfolioService 虚拟盘业务逻辑
type PortfolioService struct {
	repo  *repository.PortfolioRepository
	mkt   *MarketService // 用于获取基金净值
}

func NewPortfolioService(repo *repository.PortfolioRepository, mkt *MarketService) *PortfolioService {
	return &PortfolioService{repo: repo, mkt: mkt}
}

// ── 账户 ──────────────────────────────────────────────────────────────────

// AccountSummary 账户概览（含总资产计算）
type AccountSummary struct {
	UserID         int64   `json:"userId"`
	InitialBalance float64 `json:"initialBalance"`
	CashBalance    float64 `json:"cashBalance"`
	FrozenCash     float64 `json:"frozenCash"`
	TotalAssets    float64 `json:"totalAssets"` // 现金 + 持仓市值 + 冻结
	TotalProfit    float64 `json:"totalProfit"`
	TotalReturn    float64 `json:"totalReturn"` // 收益率 %
	PendingCount   int     `json:"pendingCount"`
}

// CreateAccount 创建虚拟账户
func (s *PortfolioService) CreateAccount(userID int64, initialBalance float64) (*model.VirtualAccount, error) {
	existing, _ := s.repo.GetAccount(userID)
	if existing != nil {
		return nil, fmt.Errorf("账户已存在")
	}

	account := &model.VirtualAccount{
		UserID:         userID,
		InitialBalance: initialBalance,
		CashBalance:    initialBalance,
		FrozenCash:     0,
		TotalProfit:    0,
	}
	if err := s.repo.CreateAccount(account); err != nil {
		return nil, err
	}
	return account, nil
}

// GetAccountSummary 获取账户概览（实时计算持仓市值）
func (s *PortfolioService) GetAccountSummary(userID int64) (*AccountSummary, error) {
	account, err := s.repo.GetAccount(userID)
	if err != nil {
		return nil, fmt.Errorf("账户不存在，请先创建")
	}

	positions, _ := s.repo.GetPositions(userID)
	pendingOrders, _ := s.repo.GetPendingOrders(userID)

	// 计算持仓总市值
	var positionValue float64
	for _, pos := range positions {
		// 获取基金最新估值
		quotes, _ := s.mkt.FetchFundQuotes([]string{pos.FundCode})
		if len(quotes) > 0 && quotes[0].EstNav > 0 {
			positionValue += pos.Shares * quotes[0].EstNav
		} else {
			// 无估值时用成本价近似
			positionValue += pos.TotalCost
		}
	}

	totalAssets := account.CashBalance + account.FrozenCash + positionValue
	totalReturn := 0.0
	if account.InitialBalance > 0 {
		totalReturn = ((totalAssets - account.InitialBalance) / account.InitialBalance) * 100
		totalReturn = math.Round(totalReturn*100) / 100
	}

	return &AccountSummary{
		UserID:         account.UserID,
		InitialBalance: account.InitialBalance,
		CashBalance:    account.CashBalance,
		FrozenCash:     account.FrozenCash,
		TotalAssets:    math.Round(totalAssets*100) / 100,
		TotalProfit:    account.TotalProfit,
		TotalReturn:    totalReturn,
		PendingCount:   len(pendingOrders),
	}, nil
}

// ── 持仓 ──────────────────────────────────────────────────────────────────

// PositionView 前端展示用的持仓
type PositionView struct {
	FundCode      string  `json:"fundCode"`
	FundName      string  `json:"fundName"`
	Shares        float64 `json:"shares"`
	AvgCost       float64 `json:"avgCost"`
	TotalCost     float64 `json:"totalCost"`
	CurrentNav    float64 `json:"currentNav"`    // 最新估值
	MarketValue   float64 `json:"marketValue"`   // 持仓市值
	UnrealizedPnL float64 `json:"unrealizedPnL"` // 浮动盈亏
	ReturnPct     float64 `json:"returnPct"`     // 收益率 %
	DividendMethod string `json:"dividendMethod"`
}

func (s *PortfolioService) GetPositions(userID int64) ([]PositionView, error) {
	positions, err := s.repo.GetPositions(userID)
	if err != nil {
		return nil, err
	}

	var views []PositionView
	for _, pos := range positions {
		view := PositionView{
			FundCode:       pos.FundCode,
			FundName:       pos.FundName,
			Shares:         pos.Shares,
			AvgCost:        pos.AvgCost,
			TotalCost:      pos.TotalCost,
			DividendMethod: pos.DividendMethod,
		}

		// 获取最新估值
		quotes, _ := s.mkt.FetchFundQuotes([]string{pos.FundCode})
		if len(quotes) > 0 && quotes[0].EstNav > 0 {
			view.CurrentNav = quotes[0].EstNav
		}

		if view.CurrentNav > 0 {
			view.MarketValue = math.Round(pos.Shares*view.CurrentNav*100) / 100
			view.UnrealizedPnL = math.Round((view.MarketValue-pos.TotalCost)*100) / 100
			if pos.TotalCost > 0 {
				view.ReturnPct = math.Round((view.UnrealizedPnL/pos.TotalCost)*10000) / 100
			}
		} else {
			view.MarketValue = pos.TotalCost
		}

		views = append(views, view)
	}
	return views, nil
}

// ── 买入 ──────────────────────────────────────────────────────────────────

// BuyRequest 买入请求
type BuyRequest struct {
	FundCode string  `json:"fundCode" binding:"required"`
	FundName string  `json:"fundName"`
	Amount   float64 `json:"amount" binding:"required,min=1"`
}

func (s *PortfolioService) Buy(userID int64, req *BuyRequest) (*model.VirtualOrder, error) {
	account, err := s.repo.GetAccount(userID)
	if err != nil {
		return nil, fmt.Errorf("账户不存在")
	}

	// 检查可用余额
	if account.CashBalance < req.Amount {
		return nil, fmt.Errorf("可用余额不足: %.2f < %.2f", account.CashBalance, req.Amount)
	}

	today := time.Now()

	// 创建待结算订单
	order := &model.VirtualOrder{
		UserID:    userID,
		FundCode:  req.FundCode,
		FundName:  req.FundName,
		OrderType: "buy",
		Amount:    req.Amount,
		Status:    "pending",
		OrderDate: today,
	}
	if err := s.repo.CreateOrder(order); err != nil {
		return nil, err
	}

	// 冻结资金
	account.CashBalance -= req.Amount
	account.FrozenCash += req.Amount
	if err := s.repo.UpdateAccount(account); err != nil {
		return nil, err
	}

	return order, nil
}

// ── 结算（T+1 日执行）─────────────────────────────────────────────────────

// SettleOrders 结算所有 pending 订单（由定时任务调用）
// settleDate 为结算日（T+1），用该日的 NAV 来结算 T 日的订单
func (s *PortfolioService) SettleOrders(settleDate string) (int, error) {
	// 查找 T-1 日的 pending 订单（即 order_date = settleDate 的前一天）
	// 简化处理：查找所有 pending 订单
	orders, err := s.repo.GetPendingOrdersByDate(settleDate)
	if err != nil {
		return 0, err
	}

	settled := 0
	for _, order := range orders {
		if err := s.settleOneOrder(&order); err != nil {
			fmt.Printf("[Settle] 订单 %d 结算失败: %v\n", order.ID, err)
			continue
		}
		settled++
	}
	return settled, nil
}

func (s *PortfolioService) settleOneOrder(order *model.VirtualOrder) error {
	// 获取基金当晚公布的净值
	quotes, err := s.mkt.FetchFundQuotes([]string{order.FundCode})
	if err != nil || len(quotes) == 0 {
		return fmt.Errorf("无法获取基金 %s 净值", order.FundCode)
	}

	nav := quotes[0].EstNav // 用最新估值作为结算净值（实际应取 T 日正式净值）
	if nav <= 0 {
		return fmt.Errorf("基金 %s 净值无效: %f", order.FundCode, nav)
	}

	account, err := s.repo.GetAccount(order.UserID)
	if err != nil {
		return err
	}

	switch order.OrderType {
	case "buy":
		// 计算份额（金额 / 净值，保留4位小数）
		shares := math.Round((order.Amount/nav)*10000) / 10000
		order.SettleNav = &nav
		order.SettleShares = &shares
		order.SettleAmount = &order.Amount
		order.Status = "confirmed"

		// 更新持仓
		pos, err := s.repo.GetPosition(order.UserID, order.FundCode)
		if err != nil {
			// 新建持仓
			pos = &model.VirtualPosition{
				UserID:         order.UserID,
				FundCode:       order.FundCode,
				FundName:       order.FundName,
				Shares:         shares,
				AvgCost:        nav,
				TotalCost:      order.Amount,
				DividendMethod: "reinvest",
			}
		} else {
			// 加仓：更新均价和总成本
			newTotalCost := pos.TotalCost + order.Amount
			newShares := pos.Shares + shares
			pos.AvgCost = newTotalCost / newShares
			pos.Shares = newShares
			pos.TotalCost = newTotalCost
		}
		if err := s.repo.UpsertPosition(pos); err != nil {
			return err
		}

		// 解冻资金 → 扣除（已冻结，现在确认扣除）
		account.FrozenCash -= order.Amount

	case "sell":
		// 卖出：order.Amount 是卖出份额
		sellShares := order.Amount
		pos, err := s.repo.GetPosition(order.UserID, order.FundCode)
		if err != nil {
			return fmt.Errorf("无持仓，无法卖出")
		}
		if pos.Shares < sellShares {
			return fmt.Errorf("持仓不足: %.4f < %.4f", pos.Shares, sellShares)
		}

		sellAmount := math.Round(sellShares*nav*100) / 100
		order.SettleNav = &nav
		order.SettleShares = &sellShares
		order.SettleAmount = &sellAmount
		order.Status = "confirmed"

		// 更新持仓
		pos.Shares -= sellShares
		pos.TotalCost = math.Round(pos.Shares*pos.AvgCost*100) / 100
		if err := s.repo.UpsertPosition(pos); err != nil {
			return err
		}

		// 卖出资金回到可用余额
		account.CashBalance += sellAmount
		account.FrozenCash -= sellAmount // 之前冻结的是份额对应的成本
		account.TotalProfit += math.Round((sellAmount-(sellShares*pos.AvgCost))*100) / 100
	}

	// 保存订单状态
	if err := s.repo.UpdateOrder(order); err != nil {
		return err
	}

	// 保存账户状态
	if err := s.repo.UpdateAccount(account); err != nil {
		return err
	}

	// 写入交易流水
	tx := &model.VirtualTransaction{
		UserID:   order.UserID,
		FundCode: order.FundCode,
		TxType:   order.OrderType,
		Amount:   *order.SettleAmount,
		Shares:   *order.SettleShares,
		Nav:      nav,
	}
	return s.repo.CreateTransaction(tx)
}

// GetPendingOrders 获取待结算订单
func (s *PortfolioService) GetPendingOrders(userID int64) ([]model.VirtualOrder, error) {
	return s.repo.GetPendingOrders(userID)
}

// GetTransactions 获取交易流水
func (s *PortfolioService) GetTransactions(userID int64, limit int) ([]model.VirtualTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetTransactions(userID, limit)
}
