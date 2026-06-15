package model

import (
	"time"
)

// VirtualOrder 买卖订单（T日提交，T+1净值确认）
type VirtualOrder struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID    int64      `gorm:"column:account_id;not null" json:"accountId"`
	UserID       int64      `gorm:"column:user_id;not null" json:"userId"`
	FundCode     string     `gorm:"column:fund_code;size:10;not null" json:"fundCode"`
	FundName     string     `gorm:"column:fund_name;size:100" json:"fundName"`
	OrderType    string     `gorm:"column:order_type;size:4;not null" json:"orderType"` // buy / sell
	Amount       float64    `gorm:"column:amount;not null" json:"amount"`               // 买入金额 / 卖出份额
	Status       string     `gorm:"column:status;size:12;default:pending" json:"status"` // pending / confirmed / cancelled
	OrderDate    time.Time  `gorm:"column:order_date;type:date;not null" json:"orderDate"`
	SettleNav    *float64   `gorm:"column:settle_nav" json:"settleNav"`
	SettleShares *float64   `gorm:"column:settle_shares" json:"settleShares"`
	SettleAmount *float64   `gorm:"column:settle_amount" json:"settleAmount"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (VirtualOrder) TableName() string {
	return "virtual_order"
}
