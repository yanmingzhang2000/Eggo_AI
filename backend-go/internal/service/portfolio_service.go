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
	mkt   *MarketService
}

func NewPortfolioService(repo *repository.PortfolioRepository, mkt *MarketService) *PortfolioService {
	return &PortfolioService{repo: repo, mkt: mkt}
}

// ── 账户 ──────────────────────────────────────────────────────────────────

// AccountSummary 账户概览（含总资产计算）
type AccountSummary struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	UserID         string  `json:"userId"`
	InitialBalance float64 `json:"initialBalance"`
	CashBalance    float64 `json:"cashBalance"`
	FrozenCash     float64 `json:"frozenCash"`
	TotalAssets    float64 `json:"totalAssets"`
	TotalProfit    float64 `json:"totalProfit"`
	TotalReturn    float64 `json:"totalReturn"`
	PendingCount   int     `json:"pendingCount"`
	CreatedAt      string  `json:"createdAt"`
}

// ListAccounts 获取用户所有鸡笼
func (s *PortfolioService) ListAccounts(userID string) ([]AccountSummary, error) {
	accounts, err := s.repo.ListAccounts(userID)
	if err != nil {
		return nil, err
	}

	var summaries []AccountSummary
	for _, acc := range accounts {
		summary, err := s.buildSummary(&acc)
		if err != nil {
			continue
		}
		summaries = append(summaries, *summary)
	}
	return summaries, nil
}

// CreateAccount 创建虚拟账户（鸡笼）
func (s *PortfolioService) CreateAccount(userID string, name string, initialBalance float64) (*model.VirtualAccount, error) {
	if name == "" {
		name = "我的鸡笼"
	}

	account := &model.VirtualAccount{
		UserID:         userID,
		Name:           name,
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

// GetAccountSummary 获取单个账户概览
func (s *PortfolioService) GetAccountSummary(accountID int64) (*AccountSummary, error) {
	account, err := s.repo.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("账户不存在")
	}
	return s.buildSummary(account)
}

func (s *PortfolioService) buildSummary(account *model.VirtualAccount) (*AccountSummary, error) {
	positions, _ := s.repo.GetPositions(account.ID)
	pendingOrders, _ := s.repo.GetPendingOrders(account.ID)

	var positionValue float64
	for _, pos := range positions {
		quotes, _ := s.mkt.FetchFundQuotes([]string{pos.FundCode})
		if len(quotes) > 0 && quotes[0].EstNav > 0 {
			positionValue += pos.Shares * quotes[0].EstNav
		} else {
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
		ID:             account.ID,
		Name:           account.Name,
		UserID:         account.UserID,
		InitialBalance: account.InitialBalance,
		CashBalance:    account.CashBalance,
		FrozenCash:     account.FrozenCash,
		TotalAssets:    math.Round(totalAssets*100) / 100,
		TotalProfit:    account.TotalProfit,
		TotalReturn:    totalReturn,
		PendingCount:   len(pendingOrders),
		CreatedAt:      account.CreatedAt.Format("2006-01-02 15:04"),
	}, nil
}

// DeleteAccount 删除鸡笼（级联删除持仓/订单/流水/定投）
func (s *PortfolioService) DeleteAccount(accountID int64) error {
	account, err := s.repo.GetAccount(accountID)
	if err != nil {
		return fmt.Errorf("账户不存在")
	}
	if account.FrozenCash > 0 {
		return fmt.Errorf("该鸡笼有未结算订单，无法删除")
	}
	return s.repo.DeleteAccount(accountID)
}

// ── 持仓 ──────────────────────────────────────────────────────────────────

// PositionView 前端展示用的持仓
type PositionView struct {
	FundCode       string  `json:"fundCode"`
	FundName       string  `json:"fundName"`
	Shares         float64 `json:"shares"`
	AvgCost        float64 `json:"avgCost"`
	TotalCost      float64 `json:"totalCost"`
	CurrentNav     float64 `json:"currentNav"`
	MarketValue    float64 `json:"marketValue"`
	UnrealizedPnL  float64 `json:"unrealizedPnL"`
	ReturnPct      float64 `json:"returnPct"`
	DividendMethod string  `json:"dividendMethod"`
}

func (s *PortfolioService) GetPositions(accountID int64) ([]PositionView, error) {
	positions, err := s.repo.GetPositions(accountID)
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

type BuyRequest struct {
	FundCode string  `json:"fundCode" binding:"required"`
	FundName string  `json:"fundName"`
	Amount   float64 `json:"amount" binding:"required,min=1"`
}

func (s *PortfolioService) Buy(accountID int64, req *BuyRequest) (*model.VirtualOrder, error) {
	account, err := s.repo.GetAccount(accountID)
	if err != nil {
		return nil, fmt.Errorf("账户不存在")
	}

	if account.CashBalance < req.Amount {
		return nil, fmt.Errorf("可用余额不足: %.2f < %.2f", account.CashBalance, req.Amount)
	}

	today := time.Now()

	order := &model.VirtualOrder{
		AccountID: accountID,
		UserID:    account.UserID,
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

	account.CashBalance -= req.Amount
	account.FrozenCash += req.Amount
	if err := s.repo.UpdateAccount(account); err != nil {
		return nil, err
	}

	return order, nil
}

// ── 结算 ──────────────────────────────────────────────────────────────────

func (s *PortfolioService) SettleOrders(settleDate string) (int, error) {
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
	quotes, err := s.mkt.FetchFundQuotes([]string{order.FundCode})
	if err != nil || len(quotes) == 0 {
		return fmt.Errorf("无法获取基金 %s 净值", order.FundCode)
	}

	nav := quotes[0].EstNav
	if nav <= 0 {
		return fmt.Errorf("基金 %s 净值无效: %f", order.FundCode, nav)
	}

	account, err := s.repo.GetAccount(order.AccountID)
	if err != nil {
		return err
	}

	switch order.OrderType {
	case "buy":
		shares := math.Round((order.Amount/nav)*10000) / 10000
		order.SettleNav = &nav
		order.SettleShares = &shares
		order.SettleAmount = &order.Amount
		order.Status = "confirmed"

		pos, err := s.repo.GetPosition(order.AccountID, order.FundCode)
		if err != nil {
			pos = &model.VirtualPosition{
				AccountID:      order.AccountID,
				UserID:         order.UserID,
				FundCode:       order.FundCode,
				FundName:       order.FundName,
				Shares:         shares,
				AvgCost:        nav,
				TotalCost:      order.Amount,
				DividendMethod: "reinvest",
			}
		} else {
			newTotalCost := pos.TotalCost + order.Amount
			newShares := pos.Shares + shares
			pos.AvgCost = newTotalCost / newShares
			pos.Shares = newShares
			pos.TotalCost = newTotalCost
		}
		if err := s.repo.UpsertPosition(pos); err != nil {
			return err
		}

		account.FrozenCash -= order.Amount

	case "sell":
		sellShares := order.Amount
		pos, err := s.repo.GetPosition(order.AccountID, order.FundCode)
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

		pos.Shares -= sellShares
		pos.TotalCost = math.Round(pos.Shares*pos.AvgCost*100) / 100
		if err := s.repo.UpsertPosition(pos); err != nil {
			return err
		}

		account.CashBalance += sellAmount
		account.FrozenCash -= sellAmount
		account.TotalProfit += math.Round((sellAmount-(sellShares*pos.AvgCost))*100) / 100
	}

	if err := s.repo.UpdateOrder(order); err != nil {
		return err
	}
	if err := s.repo.UpdateAccount(account); err != nil {
		return err
	}

	tx := &model.VirtualTransaction{
		AccountID: order.AccountID,
		UserID:    order.UserID,
		FundCode:  order.FundCode,
		TxType:    order.OrderType,
		Amount:    *order.SettleAmount,
		Shares:    *order.SettleShares,
		Nav:       nav,
	}
	return s.repo.CreateTransaction(tx)
}

// GetPendingOrders 获取待结算订单
func (s *PortfolioService) GetPendingOrders(accountID int64) ([]model.VirtualOrder, error) {
	return s.repo.GetPendingOrders(accountID)
}

// GetTransactions 获取交易流水
func (s *PortfolioService) GetTransactions(accountID int64, limit int) ([]model.VirtualTransaction, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetTransactions(accountID, limit)
}
