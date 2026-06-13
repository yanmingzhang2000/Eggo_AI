package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// TradeSignal 交易信号模型
type TradeSignal struct {
	ID             string           `gorm:"type:char(36);primaryKey" json:"id"`
	UserID         string           `gorm:"type:char(36);index" json:"user_id"`
	FundID         string           `gorm:"type:char(36);index" json:"fund_id"`
	SignalType     string           `gorm:"type:varchar(8)" json:"signal_type"`
	Strategy       string           `gorm:"type:varchar(64)" json:"strategy"`
	Confidence     decimal.Decimal  `gorm:"type:decimal(5,2)" json:"confidence"`
	Reason         *string          `gorm:"type:text" json:"reason,omitempty"`
	RelatedNewsIDs StringArray      `gorm:"type:json;default:'[]'" json:"related_news_ids"`
	TargetAmount   *decimal.Decimal `gorm:"type:decimal(14,2)" json:"target_amount,omitempty"`
	ExecuteBefore  *time.Time       `json:"execute_before,omitempty"`
	Status         int16            `gorm:"default:0" json:"status"`
	ExecutedAt     *time.Time       `json:"executed_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Fund Fund `gorm:"foreignKey:FundID" json:"fund,omitempty"`
}

func (TradeSignal) TableName() string { return "trade_signals" }
