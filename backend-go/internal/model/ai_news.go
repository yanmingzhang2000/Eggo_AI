package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// StringArray 自定义字符串数组类型，兼容 MySQL JSON 字段
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("StringArray.Scan: 无法转换 %T 到 []byte", value)
	}
	return json.Unmarshal(bytes, a)
}

// AiNews AI 过滤新闻模型
type AiNews struct {
	ID            string      `gorm:"type:char(36);primaryKey" json:"id"`
	Source        string      `gorm:"type:varchar(64)" json:"source"`
	SourceURL     *string     `gorm:"type:varchar(1024)" json:"source_url,omitempty"`
	Title         string      `gorm:"type:varchar(512)" json:"title"`
	Summary       *string     `gorm:"type:text" json:"summary,omitempty"`
	Content       *string     `gorm:"type:text" json:"content,omitempty"`
	RelatedFunds  StringArray `gorm:"type:json;default:'[]'" json:"related_funds"`
	Sentiment     *int16      `json:"sentiment,omitempty"`
	Importance    *int16      `json:"importance,omitempty"`
	Tags          StringArray `gorm:"type:json;default:'[]'" json:"tags"`
	PublishedAt   *time.Time  `json:"published_at,omitempty"`
	ProcessedAt   time.Time   `json:"processed_at"`
	CreatedAt     time.Time   `json:"created_at"`
}

func (AiNews) TableName() string { return "ai_news" }
