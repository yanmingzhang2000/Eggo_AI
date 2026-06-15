package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/controller"
	"github.com/jishengdan/backend-go/internal/middleware"
	"github.com/jishengdan/backend-go/internal/repository"
	"github.com/jishengdan/backend-go/internal/service"
)

// Setup 初始化路由
func Setup(db *gorm.DB, jwtSecret string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// ── 依赖注入 ──────────────────────────────────────────────
	// Repository 层
	userRepo := repository.NewUserRepository(db)
	eggRepo := repository.NewEggRepository(db)

	// Service 层
	authSvc := service.NewAuthService(userRepo, jwtSecret)
	eggSvc := service.NewEggService(eggRepo)
	marketSvc := service.NewMarketService()

	// Controller 层
	authCtrl := controller.NewAuthController(authSvc)
	eggCtrl := controller.NewEggController(eggSvc)
	marketCtrl := controller.NewMarketController(marketSvc)

	// ── 路由注册 ──────────────────────────────────────────────
	api := r.Group("/api/v1")
	{
		// 认证
		auth := api.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
			auth.POST("/guest", authCtrl.GuestLogin)
		}

		// 母鸡状态（核心接口，支持游客访问）
		egg := api.Group("/egg")
		{
			egg.GET("/status", eggCtrl.GetStatus)
		}

		// 大盘行情（公开接口）
		market := api.Group("/market")
		{
			market.GET("/indices", marketCtrl.GetIndices)
			market.GET("/sectors", marketCtrl.GetSectors)
			market.GET("/concepts", marketCtrl.GetConcepts)
			market.GET("/overview", marketCtrl.GetOverview)
			market.GET("/index-options", marketCtrl.GetIndexOptions)
			market.GET("/history", marketCtrl.GetIndexHistory)
			market.GET("/fund-distribution", marketCtrl.GetFundDistribution)
			market.GET("/fund-quotes", marketCtrl.GetFundQuotes)
			market.GET("/intraday", marketCtrl.GetIntraday)
		}
	}

	return r
}
