package auth

import "time"

// JWTManager JWT 管理器
// type JWTManager struct {
//     secret     []byte
//     expireHour int
// }
//
// func NewJWTManager(secret string, expireHour int) *JWTManager {}
// func (m *JWTManager) Generate(userID string) (string, time.Time, error) {}
// func (m *JWTManager) Parse(tokenStr string) (string, error) {} // returns userID
