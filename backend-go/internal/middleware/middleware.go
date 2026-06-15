package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CORSMiddleware 跨域处理
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowedOrigins := []string{
			"https://yanmingzhang2000.github.io",
			"http://localhost:5173",
			"http://localhost:3000",
		}

		allowed := false
		for _, o := range allowedOrigins {
			if strings.EqualFold(origin, o) {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	c.Next()
	}
}

// jwtClaims 与 auth_service.go 中的 Claims 保持一致
type jwtClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsGuest  bool   `json:"is_guest"`
	jwt.RegisteredClaims
}

// JWTAuthMiddleware 验证 JWT，拒绝游客和未登录请求
// 验证通过后将 userID (int64) 写入 ctx，供 controller 使用
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	secret := []byte(jwtSecret)
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token 无效或已过期，请重新登录",
			})
			return
		}

		claims, ok := token.Claims.(*jwtClaims)
		if !ok || claims.IsGuest {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "游客无权访问鸡笼，请注册或登录",
			})
			return
		}

		userID, err := strconv.ParseInt(claims.UserID, 10, 64)
		if err != nil || userID <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Token 用户信息异常",
			})
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
