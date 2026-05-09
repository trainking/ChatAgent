package handler

import (
	"github.com/chatagent/server/internal/service"
	"github.com/gin-gonic/gin"
)

func AuthClaims(c *gin.Context) *service.Claims {
	v, _ := c.Get("user_claims")
	claims, _ := v.(*service.Claims)
	return claims
}
