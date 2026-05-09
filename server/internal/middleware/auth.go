package middleware

import (
	"strings"

	"github.com/chatagent/server/internal/service"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

const UserClaimsKey = "user_claims"

func Auth(validateFunc func(string) (interface{}, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(c, errcode.Unauthorized)
			c.Abort()
			return
		}

		claims, err := validateFunc(parts[1])
		if err != nil {
			response.Error(c, errcode.TokenExpired)
			c.Abort()
			return
		}

		c.Set(UserClaimsKey, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) *service.Claims {
	v, exists := c.Get(UserClaimsKey)
	if !exists {
		return nil
	}
	claims, ok := v.(*service.Claims)
	if !ok {
		return nil
	}
	return claims
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil || (claims.Role != "admin" && claims.Role != "super_admin") {
			response.Error(c, errcode.Forbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil || claims.Role != "super_admin" {
			response.Error(c, errcode.Forbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}
