package model

import "time"

// VirtualTransaction 交易流水
type VirtualTransaction struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID int64     `gorm:"column:account_id;not null" json:"accountId"`
	UserID    int64     `gorm:"column:user_id;not null" json:"userId"`
	FundCode  string    `gorm:"column:fund_code;size:10;not null" json:"fundCode"`
	TxType    string    `gorm:"column:tx_type;size:12;not null" json:"txType"` // buy / sell / dividend / dca_buy
	Amount    float64   `gorm:"column:amount" json:"amount"`
	Shares    float64   `gorm:"column:shares" json:"shares"`
	Nav       float64   `gorm:"column:nav" json:"nav"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (VirtualTransaction) TableName() string {
	return "virtual_transaction"
}
