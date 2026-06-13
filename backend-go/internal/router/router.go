package router

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jishengdan/backend-go/internal/controller"
	"github.com/jishengdan/backend-go/internal/repository"
	"github.com/jishengdan/backend-go/internal/service"
)

// Setup 初始化路由
func Setup(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// ── 依赖注入 ──────────────────────────────────────────────
	// Repository 层
	eggRepo := repository.NewEggRepository(db)

	// Service 层
	eggSvc := service.NewEggService(eggRepo)

	// Controller 层
	eggCtrl := controller.NewEggController(eggSvc)

	// ── 路由注册 ──────────────────────────────────────────────
	api := r.Group("/api/v1")
	{
		// 母鸡状态（核心接口）
		egg := api.Group("/egg")
		{
			egg.GET("/status", eggCtrl.GetStatus)
		}

		// 预留路由（后续扩展）
		// auth := api.Group("/auth")
		// {
		//     auth.POST("/register", userCtrl.Register)
		//     auth.POST("/login",    userCtrl.Login)
		// }
		//
		// funds := api.Group("/funds")
		// {
		//     funds.GET("",          fundCtrl.List)
		//     funds.GET("/:code",    fundCtrl.Detail)
		//     funds.GET("/:code/nav", fundCtrl.NavHistory)
		// }
	}

	return r
}
