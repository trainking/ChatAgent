package router

import (
	"github.com/chatagent/server/config"
	"github.com/chatagent/server/internal/handler"
	"github.com/chatagent/server/internal/middleware"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func Setup(cfg *config.Config, db *sqlx.DB) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	sysCfgRepo := repository.NewSystemConfigRepository(db)
	totpRepo := repository.NewUserTOTPRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg)
	twoFASvc := service.NewTwoFactorService(totpRepo, sysCfgRepo, cfg.JWT.Secret)

	authHandler := handler.NewAuthHandler(authSvc, twoFASvc)
	userHandler := handler.NewUserHandler(userRepo)
	permHandler := handler.NewPermissionHandler(permRepo)
	sysHandler := handler.NewSystemHandler(twoFASvc)
	twoFAHandler := handler.NewTwoFactorHandler(twoFASvc, authSvc)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		v1.GET("/auth/status", authHandler.Status)
		v1.POST("/auth/init", authHandler.InitRoot)
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/2fa/verify", twoFAHandler.VerifyLogin)

		protected := v1.Group("")
		protected.Use(middleware.Auth(authSvc.ValidateToken))
		{
			protected.GET("/me", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "authenticated"})
			})
			protected.POST("/auth/change-password", authHandler.ChangePassword)
			protected.GET("/auth/2fa/status", twoFAHandler.GetStatus)
			protected.POST("/auth/2fa/setup", twoFAHandler.Setup)
			protected.POST("/auth/2fa/verify-setup", twoFAHandler.VerifySetup)

			admin := protected.Group("/users")
			admin.Use(middleware.RequireAdmin())
			{
				admin.GET("", userHandler.List)
				admin.POST("", userHandler.Create)
				admin.PUT("/:id", userHandler.Update)
				admin.DELETE("/:id", userHandler.Delete)
				admin.PUT("/:id/reset-password", userHandler.ResetPassword)
			}

			sa := protected.Group("/roles")
			sa.Use(middleware.RequireSuperAdmin())
			{
				sa.GET("/:role/permissions", permHandler.GetRolePermissions)
				sa.PUT("/:role/permissions", permHandler.SetRolePermissions)
			}
			protected.GET("/permissions", permHandler.ListAll)

			system := protected.Group("/system")
			system.Use(middleware.RequireSuperAdmin())
			{
				system.GET("/2fa/config", sysHandler.Get2FAConfig)
				system.PUT("/2fa/config", sysHandler.Set2FAConfig)
			}
		}
	}

	return r
}
