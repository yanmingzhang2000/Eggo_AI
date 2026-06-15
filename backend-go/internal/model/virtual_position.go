package model

import "time"

// VirtualPosition 基金持仓
type VirtualPosition struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         int64     `gorm:"column:user_id;not null" json:"userId"`
	FundCode       string    `gorm:"column:fund_code;size:10;not null" json:"fundCode"`
	FundName       string    `gorm:"column:fund_name;size:100" json:"fundName"`
	Shares         float64   `gorm:"column:shares;default:0" json:"shares"`
	AvgCost        float64   `gorm:"column:avg_cost;default:0" json:"avgCost"`
	TotalCost      float64   `gorm:"column:total_cost;default:0" json:"totalCost"`
	DividendMethod string    `gorm:"column:dividend_method;size:10;default:reinvest" json:"dividendMethod"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (VirtualPosition) TableName() string {
	return "virtual_position"
}
