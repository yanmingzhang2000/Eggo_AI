package dto

import "github.com/shopspring/decimal"

// FundQuery 基金列表查询参数
type FundQuery struct {
	Keyword  string `form:"keyword"`
	FundType string `form:"fund_type"`
	Page     int    `form:"page,default=1"  binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
}

// FundItem 基金列表项
type FundItem struct {
	FundCode string  `json:"fund_code"`
	FundName string  `json:"fund_name"`
	FundType *string `json:"fund_type,omitempty"`
	Manager  *string `json:"manager,omitempty"`
}

// FundDetail 基金详情
type FundDetail struct {
	FundItem
	Custodian     *string `json:"custodian,omitempty"`
	InceptionDate *string `json:"inception_date,omitempty"`
	Benchmark     *string `json:"benchmark,omitempty"`
}

// NavItem 净值数据项
type NavItem struct {
	NavDate     string           `json:"nav_date"`
	UnitNav     decimal.Decimal  `json:"unit_nav"`
	AccNav      *decimal.Decimal `json:"acc_nav,omitempty"`
	DailyReturn *decimal.Decimal `json:"daily_return,omitempty"`
}
