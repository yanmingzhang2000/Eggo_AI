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
	portfolioRepo := repository.NewPortfolioRepository(db)
	portfolioSvc := service.NewPortfolioService(portfolioRepo, marketSvc)

	// Controller 层
	authCtrl := controller.NewAuthController(authSvc)
	eggCtrl := controller.NewEggController(eggSvc)
	marketCtrl := controller.NewMarketController(marketSvc)
	portfolioCtrl := controller.NewPortfolioController(portfolioSvc)

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

		// 虚拟养鸡（基）— 模拟盘（多鸡笼）
		portfolio := api.Group("/portfolio")
		{
			// 鸡笼管理
			portfolio.GET("/accounts", portfolioCtrl.ListAccounts)
			portfolio.POST("/accounts", portfolioCtrl.CreateAccount)
			portfolio.GET("/accounts/:id", portfolioCtrl.GetAccount)
			portfolio.DELETE("/accounts/:id", portfolioCtrl.DeleteAccount)

			// 鸡笼内操作
			portfolio.GET("/accounts/:id/positions", portfolioCtrl.GetPositions)
			portfolio.POST("/accounts/:id/buy", portfolioCtrl.Buy)
			portfolio.GET("/accounts/:id/orders/pending", portfolioCtrl.GetPendingOrders)
			portfolio.GET("/accounts/:id/transactions", portfolioCtrl.GetTransactions)
		}
	}

	return r
}
