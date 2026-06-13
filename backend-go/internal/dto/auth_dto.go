package dto

// AuthResponse 认证响应
type AuthResponse struct {
	Token   string   `json:"token"`
	User    UserInfo `json:"user"`
	IsGuest bool     `json:"isGuest"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GuestResponse 游客响应
type GuestResponse struct {
	Token   string   `json:"token"`
	IsGuest bool     `json:"isGuest"`
	User    UserInfo `json:"user"`
}
