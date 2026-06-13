package model

import "time"

// Fund 基金基础信息模型
type Fund struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	FundCode      string     `gorm:"type:varchar(16);uniqueIndex" json:"fund_code"`
	FundName      string     `gorm:"type:varchar(128)" json:"fund_name"`
	FundType      *string    `gorm:"type:varchar(32)" json:"fund_type,omitempty"`
	Manager       *string    `gorm:"type:varchar(64)" json:"manager,omitempty"`
	Custodian     *string    `gorm:"type:varchar(64)" json:"custodian,omitempty"`
	InceptionDate *time.Time `gorm:"type:date" json:"inception_date,omitempty"`
	Benchmark     *string    `gorm:"type:varchar(128)" json:"benchmark,omitempty"`
	Status        int16      `gorm:"default:1" json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Fund) TableName() string { return "funds" }
