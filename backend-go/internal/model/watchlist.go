package model

import "time"

// Watchlist 自选基金模型
type Watchlist struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    string    `gorm:"type:uuid;index" json:"user_id"`
	FundID    string    `gorm:"type:uuid;index" json:"fund_id"`
	Remark    *string   `gorm:"type:varchar(256)" json:"remark,omitempty"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Fund Fund `gorm:"foreignKey:FundID" json:"fund,omitempty"`
}

func (Watchlist) TableName() string { return "watchlist" }
