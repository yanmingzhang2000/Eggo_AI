package repository

import (
	"github.com/jishengdan/backend-go/internal/model"
	"gorm.io/gorm"
)

// PortfolioRepository 虚拟盘数据访问
type PortfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) *PortfolioRepository {
	return &PortfolioRepository{db: db}
}

// ── 账户 ──────────────────────────────────────────────────────────────────

func (r *PortfolioRepository) ListAccounts(userID int64) ([]model.VirtualAccount, error) {
	var accounts []model.VirtualAccount
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&accounts).Error
	return accounts, err
}

func (r *PortfolioRepository) GetAccount(accountID int64) (*model.VirtualAccount, error) {
	var account model.VirtualAccount
	err := r.db.Where("id = ?", accountID).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *PortfolioRepository) CreateAccount(account *model.VirtualAccount) error {
	return r.db.Create(account).Error
}

func (r *PortfolioRepository) UpdateAccount(account *model.VirtualAccount) error {
	return r.db.Save(account).Error
}

func (r *PortfolioRepository) DeleteAccount(accountID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id = ?", accountID).Delete(&model.VirtualPosition{}).Error; err != nil {
			return err
		}
		if err := tx.Where("account_id = ?", accountID).Delete(&model.VirtualOrder{}).Error; err != nil {
			return err
		}
		if err := tx.Where("account_id = ?", accountID).Delete(&model.VirtualTransaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("account_id = ?", accountID).Delete(&model.VirtualDCAPlan{}).Error; err != nil {
			return err
		}
		return tx.Where("id = ?", accountID).Delete(&model.VirtualAccount{}).Error
	})
}

// ── 持仓 ──────────────────────────────────────────────────────────────────

func (r *PortfolioRepository) GetPositions(accountID int64) ([]model.VirtualPosition, error) {
	var positions []model.VirtualPosition
	err := r.db.Where("account_id = ? AND shares > 0", accountID).Order("created_at ASC").Find(&positions).Error
	return positions, err
}

func (r *PortfolioRepository) GetPosition(accountID int64, fundCode string) (*model.VirtualPosition, error) {
	var pos model.VirtualPosition
	err := r.db.Where("account_id = ? AND fund_code = ?", accountID, fundCode).First(&pos).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

func (r *PortfolioRepository) UpsertPosition(pos *model.VirtualPosition) error {
	return r.db.Save(pos).Error
}

// ── 订单 ──────────────────────────────────────────────────────────────────

func (r *PortfolioRepository) CreateOrder(order *model.VirtualOrder) error {
	return r.db.Create(order).Error
}

func (r *PortfolioRepository) GetPendingOrders(accountID int64) ([]model.VirtualOrder, error) {
	var orders []model.VirtualOrder
	err := r.db.Where("account_id = ? AND status = 'pending'", accountID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *PortfolioRepository) GetPendingOrdersByDate(orderDate string) ([]model.VirtualOrder, error) {
	var orders []model.VirtualOrder
	err := r.db.Where("status = 'pending' AND order_date = ?", orderDate).Find(&orders).Error
	return orders, err
}

func (r *PortfolioRepository) UpdateOrder(order *model.VirtualOrder) error {
	return r.db.Save(order).Error
}

// ── 交易流水 ──────────────────────────────────────────────────────────────

func (r *PortfolioRepository) CreateTransaction(tx *model.VirtualTransaction) error {
	return r.db.Create(tx).Error
}

func (r *PortfolioRepository) GetTransactions(accountID int64, limit int) ([]model.VirtualTransaction, error) {
	var txs []model.VirtualTransaction
	err := r.db.Where("account_id = ?", accountID).Order("created_at DESC").Limit(limit).Find(&txs).Error
	return txs, err
}

// ── 定投计划 ──────────────────────────────────────────────────────────────

func (r *PortfolioRepository) GetActiveDCAPLans() ([]model.VirtualDCAPlan, error) {
	var plans []model.VirtualDCAPlan
	err := r.db.Where("status = 'active' AND next_exec_date <= CURRENT_DATE").Find(&plans).Error
	return plans, err
}

func (r *PortfolioRepository) GetDCAPLansByAccount(accountID int64) ([]model.VirtualDCAPlan, error) {
	var plans []model.VirtualDCAPlan
	err := r.db.Where("account_id = ? AND status = 'active'", accountID).Order("next_exec_date ASC").Find(&plans).Error
	return plans, err
}
