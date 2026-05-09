package router

import (
	"github.com/chatagent/server/config"
	"github.com/chatagent/server/internal/middleware"
	"github.com/gin-gonic/gin"
)

func Setup(cfg *config.Config) *gin.Engine {
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

	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		// Auth routes will be added in M2
		// auth := v1.Group("/auth")
		// {
		//     auth.POST("/register", ...)
		//     auth.POST("/login", ...)
		// }
		//
		// Protected routes will be added in later modules
		// protected := v1.Group("")
		// protected.Use(middleware.Auth())
	}

	return r
}
