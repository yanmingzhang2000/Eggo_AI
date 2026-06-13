package dto

// AddWatchlistRequest 添加自选请求
type AddWatchlistRequest struct {
	FundCode string `json:"fund_code" binding:"required"`
	Remark   string `json:"remark"`
}

// WatchlistItem 自选列表项
type WatchlistItem struct {
	FundCode  string  `json:"fund_code"`
	FundName  string  `json:"fund_name"`
	FundType  *string `json:"fund_type,omitempty"`
	Remark    *string `json:"remark,omitempty"`
	UnitNav   string  `json:"unit_nav,omitempty"`  // 最新净值
	DailyChg  string  `json:"daily_chg,omitempty"` // 日涨跌
}
