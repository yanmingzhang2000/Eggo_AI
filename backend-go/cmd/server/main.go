package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jishengdan/backend-go/internal/config"
	"github.com/jishengdan/backend-go/internal/router"
)

func main() {
	// 1. 加载配置
	cfg := loadConfig()

	// 2. 初始化数据库
	db := initDB(cfg)

	// 3. 初始化路由
	r := router.Setup(db, cfg.JWT.Secret)

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
			Port: 8080,
			Mode: "debug",
		},
		Database: config.DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     3306,
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASS", "root123"),
			DBName:   getEnv("DB_NAME", "jishengdan"),
		},
		JWT: config.JWTConfig{
			Secret:     getEnv("JWT_SECRET", "eggo_jwt_secret_2024"),
			ExpireHour: 72,
		},
	}
}

// initDB 初始化数据库连接
func initDB(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
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
