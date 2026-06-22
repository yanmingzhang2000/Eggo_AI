package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jishengdan/backend-go/internal/config"
	"github.com/jishengdan/backend-go/internal/router"
	"github.com/jishengdan/backend-go/pkg/tushare"
)

func main() {
	// 1. 加载配置
	cfg := loadConfig()

	// 2. 初始化数据库
	db := initDB(cfg)

	// 3. 初始化 Tushare 客户端
	tsClient := tushare.NewClient(cfg.Tushare.Token)

	// 4. 初始化路由
	r := router.Setup(db, cfg.JWT.Secret, tsClient)

	// 4. 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("🐔 鸡生蛋 API 服务启动中... %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// loadConfig 加载配置
func loadConfig() *config.Config {
	// TODO: 从环境变量或配置文件加载
	return &config.Config{
		Server: config.ServerConfig{
			Port: getEnvAsInt("PORT", 8080),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: config.DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     5432,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASS", ""),
			DBName:   getEnv("DB_NAME", "jishengdan"),
		},
		JWT: config.JWTConfig{
			Secret:     getEnv("JWT_SECRET", "eggo_jwt_secret_2024"),
			ExpireHour: 72,
		},
		Tushare: config.TushareConfig{
			Token: getEnv("TUSHARE_TOKEN", ""),
		},
	}
}

// initDB 初始化数据库连接
func initDB(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error

	// 优先使用 DATABASE_URL（Render 等平台提供）
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// 使用独立环境变量构建 DSN
		sslMode := getEnv("DB_SSLMODE", "disable")
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.DBName,
			cfg.Database.Port,
			sslMode,
		)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	log.Println("数据库连接成功")
	return db
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvAsInt 获取环境变量（整数），带默认值
func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
	}
	return defaultVal
}
