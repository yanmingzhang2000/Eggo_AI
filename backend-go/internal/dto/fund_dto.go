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

// ── 基金详情页面 DTO ─────────────────────────────────────────────

// FundDetailResponse 基金详情页完整响应
type FundDetailResponse struct {
	FundCode    string            `json:"fundCode"`
	FundName    string            `json:"fundName"`
	UnitNav     float64           `json:"unitNav"`
	AccNav      *float64          `json:"accNav,omitempty"`
	DailyReturn float64           `json:"dailyReturn"`
	NavDate     string            `json:"navDate"`
	// 多周期收益率（从净值计算）
	WeekReturn    float64  `json:"weekReturn"`
	MonthReturn   *float64 `json:"monthReturn,omitempty"`
	QuarterReturn *float64 `json:"quarterReturn,omitempty"`
	YearReturn    *float64 `json:"yearReturn,omitempty"`
	// 趋势指标
	ConsecutiveUp  int    `json:"consecutiveUp"`
	ConsecutiveDown int   `json:"consecutiveDown"`
	// 信号分析
	HasNegNews    bool `json:"hasNegNews"`
	HasPosPolicy  bool `json:"hasPosPolicy"`
	SentimentCool bool `json:"sentimentCool"`
	// 评级
	MaxDrawdown *float64 `json:"maxDrawdown,omitempty"`
	// 基本信息
	FundType      *string `json:"fundType,omitempty"`
	Manager       *string `json:"manager,omitempty"`
	Custodian     *string `json:"custodian,omitempty"`
	InceptionDate *string `json:"inceptionDate,omitempty"`
	Benchmark     *string `json:"benchmark,omitempty"`
}

// NavPoint 净值历史点（前端图表用）
type NavPoint struct {
	Date        string  `json:"date"`
	UnitNav     float64 `json:"unitNav"`
	AccNav      float64 `json:"accNav,omitempty"`
	DailyReturn float64 `json:"dailyReturn,omitempty"`
}

// NavHistoryResponse 净值历史响应
type NavHistoryResponse struct {
	FundCode string     `json:"fundCode"`
	FundName string     `json:"fundName"`
	Points   []NavPoint `json:"points"`
}

// FundAnalysis 单只基金三铁律分析结果
type FundAnalysis struct {
	FundCode     string  `json:"fundCode"`
	FundName     string  `json:"fundName"`
	SignalType   string  `json:"signalType"`   // "buy" | "sell" | "hold" | "observe"
	SignalText   string  `json:"signalText"`   // "建议买入" | "建议卖出" | "建议持有" | "继续观察"
	SignalEmoji  string  `json:"signalEmoji"`  // 对应 emoji
	Confidence   float64 `json:"confidence"`   // 置信度 0~1
	Reason       string  `json:"reason"`       // 分析依据
	TrendLabel   string  `json:"trendLabel"`   // "上升期" | "下跌期" | "震荡期" | "平稳期"
	TrendColor   string  `json:"trendColor"`   // 颜色 #ff4d4f / #00d68f / #f7ba1e / #787878
	Suggestions  []string `json:"suggestions"` // 操作建议列表
}

// FundAnalysisResponse 基金分析响应
type FundAnalysisResponse struct {
	FundCode    string       `json:"fundCode"`
	FundName    string       `json:"fundName"`
	Analysis    FundAnalysis `json:"analysis"`
	GeneratedAt string       `json:"generatedAt"`
}
