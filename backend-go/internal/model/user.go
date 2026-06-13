package model

import "time"

// User 用户模型
type User struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Username     string     `gorm:"type:varchar(64);uniqueIndex" json:"username"`
	Email        string     `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255)" json:"-"`
	Phone        *string    `gorm:"type:varchar(20)" json:"phone,omitempty"`
	AvatarURL    *string    `gorm:"type:varchar(512)" json:"avatar_url,omitempty"`
	Status       int16      `gorm:"default:1" json:"status"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (User) TableName() string { return "users" }
