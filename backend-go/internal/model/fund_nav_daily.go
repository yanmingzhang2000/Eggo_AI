package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// FundNavDaily 每日净值模型
type FundNavDaily struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	FundID      string          `gorm:"type:uuid;index" json:"fund_id"`
	NavDate     time.Time       `gorm:"type:date" json:"nav_date"`
	UnitNav     decimal.Decimal `gorm:"type:numeric(12,4)" json:"unit_nav"`
	AccNav      *decimal.Decimal `gorm:"type:numeric(12,4)" json:"acc_nav,omitempty"`
	DailyReturn *decimal.Decimal `gorm:"type:numeric(8,4)" json:"daily_return,omitempty"`
	TotalReturn *decimal.Decimal `gorm:"type:numeric(8,4)" json:"total_return,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`

	Fund Fund `gorm:"foreignKey:FundID" json:"fund,omitempty"`
}

func (FundNavDaily) TableName() string { return "fund_nav_daily" }
