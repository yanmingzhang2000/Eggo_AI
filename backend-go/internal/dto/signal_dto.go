package dto

import "github.com/shopspring/decimal"

// SignalQuery 信号查询参数
type SignalQuery struct {
	Status     *int16  `form:"status"`
	SignalType string  `form:"signal_type"`
	Strategy   string  `form:"strategy"`
	Page       int     `form:"page,default=1"`
	PageSize   int     `form:"page_size,default=20"`
}

// SignalItem 信号列表项
type SignalItem struct {
	ID           string           `json:"id"`
	FundCode     string           `json:"fund_code"`
	FundName     string           `json:"fund_name"`
	SignalType   string           `json:"signal_type"`
	Strategy     string           `json:"strategy"`
	Confidence   decimal.Decimal  `json:"confidence"`
	Reason       *string          `json:"reason,omitempty"`
	TargetAmount *decimal.Decimal `json:"target_amount,omitempty"`
	Status       int16            `json:"status"`
	CreatedAt    string           `json:"created_at"`
}
