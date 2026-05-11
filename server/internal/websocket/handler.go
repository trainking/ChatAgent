package websocket

import (
	"net/http"

	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type HandlerDeps struct {
	Hub           *Hub
	AuthSvc       *service.AuthService
	ContactInboxRepo *repository.ContactInboxRepository
	UserRepo         *repository.UserRepository
}

func HandleWebSocket(deps *HandlerDeps) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		clientType := c.Query("type")

		if tokenStr == "" || clientType == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token or type"})
			return
		}

		var client *Client

		switch clientType {
		case "agent":
			claims, err := deps.AuthSvc.ValidateToken(tokenStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}
			agentClaims := claims.(*service.Claims)

			user, err := deps.UserRepo.FindByID(agentClaims.UserID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
				return
			}

			client = &Client{
				ID:         agentClaims.UserID,
				ClientType: ClientTypeAgent,
				Hub:        deps.Hub,
				Send:       make(chan []byte, 256),
				Channels:   make(map[string]bool),
				UserID:     agentClaims.UserID,
				UserEmail:  agentClaims.Email,
				UserName:   user.Name,
				Role:       agentClaims.Role,
			}

		case "widget":
			ci, err := deps.ContactInboxRepo.FindByPubsubToken(tokenStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}

			client = &Client{
				ID:             ci.PubsubToken,
				ClientType:     ClientTypeWidget,
				Hub:            deps.Hub,
				Send:           make(chan []byte, 256),
				Channels:       make(map[string]bool),
				ContactInboxID: ci.ContactID,
			}

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client.Conn = conn
		deps.Hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
