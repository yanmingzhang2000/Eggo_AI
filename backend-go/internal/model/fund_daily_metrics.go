package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// FundDailyMetrics 基金每日指标汇总表
// 用于三铁律决策引擎的快速查询
type FundDailyMetrics struct {
	ID            string          `gorm:"type:char(36);primaryKey" json:"id"`
	FundID        string          `gorm:"type:char(36);index" json:"fund_id"`
	FundCode      string          `gorm:"type:varchar(16);index" json:"fund_code"`
	FundName      string          `gorm:"type:varchar(128)" json:"fund_name"`
	MetricsDate   time.Time       `gorm:"type:date;index" json:"metrics_date"`
	UnitNav       decimal.Decimal `gorm:"type:decimal(12,4)" json:"unit_nav"`
	DailyReturn   decimal.Decimal `gorm:"type:decimal(8,4)" json:"daily_return"`    // 日涨跌幅 %
	AccNav        *decimal.Decimal `gorm:"type:decimal(12,4)" json:"acc_nav,omitempty"`
	ConsecutiveUp int             `gorm:"default:0" json:"consecutive_up"`          // 连续上涨天数
	WeekReturn    decimal.Decimal `gorm:"type:decimal(8,4)" json:"week_return"`     // 近5日累计涨幅 %
	HasNegNews    bool            `gorm:"default:false" json:"has_neg_news"`        // 是否有负面新闻
	HasPosPolicy  bool            `gorm:"default:false" json:"has_pos_policy"`      // 是否有政策利好
	SentimentCool bool            `gorm:"default:false" json:"sentiment_cool"`      // 舆情是否降温
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	Fund Fund `gorm:"foreignKey:FundID" json:"fund,omitempty"`
}

func (FundDailyMetrics) TableName() string { return "fund_daily_metrics" }
