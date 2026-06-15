package model

import "time"

// VirtualAccount 虚拟养鸡账户
type VirtualAccount struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         int64     `gorm:"column:user_id;uniqueIndex;not null" json:"userId"`
	InitialBalance float64   `gorm:"column:initial_balance;not null" json:"initialBalance"`
	CashBalance    float64   `gorm:"column:cash_balance;not null" json:"cashBalance"`
	FrozenCash     float64   `gorm:"column:frozen_cash;default:0" json:"frozenCash"`
	TotalProfit    float64   `gorm:"column:total_profit;default:0" json:"totalProfit"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (VirtualAccount) TableName() string {
	return "virtual_account"
}
