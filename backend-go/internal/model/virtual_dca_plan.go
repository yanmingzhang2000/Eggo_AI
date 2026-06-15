package model

import "time"

// VirtualDCAPlan 定投计划
type VirtualDCAPlan struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"column:user_id;not null" json:"userId"`
	FundCode     string    `gorm:"column:fund_code;size:10;not null" json:"fundCode"`
	FundName     string    `gorm:"column:fund_name;size:100" json:"fundName"`
	Amount       float64   `gorm:"column:amount;not null" json:"amount"`          // 每次定投金额
	Frequency    string    `gorm:"column:frequency;size:10;not null" json:"frequency"` // weekly / biweekly / monthly
	ExecDay      int       `gorm:"column:exec_day;not null" json:"execDay"`       // 周几(1-7) 或 几号(1-31)
	NextExecDate time.Time `gorm:"column:next_exec_date;type:date;not null" json:"nextExecDate"`
	Status       string    `gorm:"column:status;size:10;default:active" json:"status"` // active / paused
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (VirtualDCAPlan) TableName() string {
	return "virtual_dca_plan"
}
