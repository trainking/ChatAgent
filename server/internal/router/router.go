package router

import (
	"github.com/chatagent/server/config"
	"github.com/chatagent/server/internal/handler"
	"github.com/chatagent/server/internal/middleware"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/service"
	"github.com/chatagent/server/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

func Setup(cfg *config.Config, db *sqlx.DB, rdb *redis.Client) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	r.Static("/uploads", "./uploads")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepo := repository.NewUserRepository(db)
	permRepo := repository.NewPermissionRepository(db)
	sysCfgRepo := repository.NewSystemConfigRepository(db)
	totpRepo := repository.NewUserTOTPRepository(db)
	actRepo := repository.NewActivityRepository(db)
	inboxRepo := repository.NewInboxRepository(db)
	contactRepo := repository.NewContactRepository(db)
	contactInboxRepo := repository.NewContactInboxRepository(db)
	convRepo := repository.NewConversationRepository(db, rdb)
	msgRepo := repository.NewMessageRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg)
	twoFASvc := service.NewTwoFactorService(totpRepo, sysCfgRepo, cfg.JWT.Secret)

	hub := websocket.NewHub(convRepo)
	go hub.Run()

	authHandler := handler.NewAuthHandler(authSvc, twoFASvc, userRepo, actRepo)
	userHandler := handler.NewUserHandler(userRepo)
	permHandler := handler.NewPermissionHandler(permRepo)
	sysHandler := handler.NewSystemHandler(twoFASvc)
	twoFAHandler := handler.NewTwoFactorHandler(twoFASvc, authSvc, userRepo, actRepo)
	profileHandler := handler.NewProfileHandler(userRepo, actRepo)
	inboxHandler := handler.NewInboxHandler(inboxRepo, userRepo)
	conversationH := handler.NewConversationHandler(convRepo, msgRepo, contactRepo, userRepo, inboxRepo, hub)
	messageH := handler.NewMessageHandler(msgRepo, convRepo, hub)
	contactH := handler.NewContactHandler(contactRepo, convRepo, db)
	widgetH := handler.NewWidgetHandler(db, inboxRepo, contactRepo, contactInboxRepo, convRepo, msgRepo, hub)
	uploadH := handler.NewUploadHandler()

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
			protected.POST("/auth/logout", authHandler.Logout)
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

			inboxes := protected.Group("/inboxes")
			{
				inboxes.GET("", inboxHandler.List)
				inboxes.POST("", inboxHandler.Create)
				inboxes.GET("/:id", inboxHandler.Get)
				inboxes.PUT("/:id", inboxHandler.Update)
				inboxes.DELETE("/:id", inboxHandler.Delete)
			}

			profile := protected.Group("/profile")
			{
				profile.GET("", profileHandler.GetProfile)
				profile.PUT("", profileHandler.UpdateProfile)
				profile.POST("/avatar", profileHandler.UploadAvatar)
				profile.GET("/activities", profileHandler.GetActivities)
				profile.PUT("/status", profileHandler.UpdateStatus)
			}

			system := protected.Group("/system")
			system.Use(middleware.RequireSuperAdmin())
			{
				system.GET("/2fa/config", sysHandler.Get2FAConfig)
				system.PUT("/2fa/config", sysHandler.Set2FAConfig)
			}

			conversations := protected.Group("/conversations")
			{
				conversations.GET("", conversationH.List)
				conversations.GET("/unread-count", conversationH.UnreadCount)
				conversations.GET("/:id", conversationH.Get)
				conversations.PUT("/:id/assign", conversationH.Assign)
				conversations.PUT("/:id/status", conversationH.ChangeStatus)
				conversations.PUT("/:id/priority", conversationH.ChangePriority)
			}

			protected.GET("/conversations/:id/messages", messageH.List)
			protected.POST("/conversations/:id/messages", messageH.Create)
			protected.POST("/conversations/:id/messages/:fid/retry", messageH.Retry)

			protected.POST("/upload", uploadH.Upload)

			contacts := protected.Group("/contacts")
			{
				contacts.GET("", contactH.List)
				contacts.POST("", contactH.Create)
				contacts.GET("/:id", contactH.Get)
				contacts.PUT("/:id", contactH.Update)
				contacts.POST("/:id/merge", contactH.Merge)
				contacts.GET("/:id/conversations", contactH.GetConversations)
			}
		}

		v1.POST("/widget/auth", widgetH.Auth)
		v1.POST("/widget/messages", widgetH.SendMessage)
	}

	wsDeps := &websocket.HandlerDeps{
		Hub:              hub,
		AuthSvc:          authSvc,
		ContactInboxRepo: contactInboxRepo,
		UserRepo:         userRepo,
		InboxRepo:        inboxRepo,
	}
	r.GET("/ws", websocket.HandleWebSocket(wsDeps))

	return r
}
