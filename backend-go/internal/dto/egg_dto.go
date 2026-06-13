package dto

import "github.com/shopspring/decimal"

// EggStatusResponse 母鸡状态总响应（TypeScript 友好命名）
type EggStatusResponse struct {
	ChickenStatus ChickenStatus   `json:"chickenStatus"`
	TodayMetrics  []FundMetric    `json:"todayMetrics"`
	NewsClues     []NewsClue      `json:"newsClues"`
	Decision      Decision        `json:"decision"`
	GeneratedAt   string          `json:"generatedAt"`
}

// ChickenStatus 母鸡状态概览
type ChickenStatus struct {
	OverallState  string `json:"overallState"`  // "laying" | "resting" | "molting" | "soaking"
	StateEmoji    string `json:"stateEmoji"`    // 🐔🥚🌧️🛁
	StateDesc     string `json:"stateDesc"`     // 状态中文描述
	AlertLevel    string `json:"alertLevel"`    // "none" | "info" | "warning" | "critical"
	AlertMessage  string `json:"alertMessage"`  // 告警信息
}

// FundMetric 单只基金今日指标
type FundMetric struct {
	FundCode       string  `json:"fundCode"`
	FundName       string  `json:"fundName"`
	UnitNav        float64 `json:"unitNav"`
	DailyReturn    float64 `json:"dailyReturn"`    // 日涨跌幅 %
	ConsecutiveUp  int     `json:"consecutiveUp"`  // 连续上涨天数
	WeekReturn     float64 `json:"weekReturn"`     // 近5日累计涨幅 %
	HasNegNews     bool    `json:"hasNegNews"`
	HasPosPolicy   bool    `json:"hasPosPolicy"`
	SentimentCool  bool    `json:"sentimentCool"`
}

// NewsClue AI 过滤的新闻线索
type NewsClue struct {
	NewsID          string  `json:"newsId"`
	Title           string  `json:"title"`
	Source          string  `json:"source"`
	Sentiment       int     `json:"sentiment"`       // -1 / 0 / 1
	Importance      int     `json:"importance"`      // 1~5
	RelevanceReason string  `json:"relevanceReason"` // 关联性说明
	RelatedFunds    []string `json:"relatedFunds"`
	Tags            []string `json:"tags"`
	PublishedAt     string  `json:"publishedAt"`
}

// Decision 最终决策建议
type Decision struct {
	TriggeredRule string  `json:"triggeredRule"`   // "rule1_dca" | "rule2_hold" | "rule3_take_profit" | "none"
	RuleName      string  `json:"ruleName"`        // 规则中文名
	Action        string  `json:"action"`          // "dca" | "hold" | "take_profit" | "observe"
	ActionEmoji   string  `json:"actionEmoji"`     // 对应 emoji
	Suggestion    string  `json:"suggestion"`      // 最终建议文案
	Confidence    float64 `json:"confidence"`      // 置信度 0~1
	Reason        string  `json:"reason"`          // 决策依据
	TargetFunds   []string `json:"targetFunds"`    // 涉及的基金代码
}
