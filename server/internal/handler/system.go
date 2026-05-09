package handler

import (
	"github.com/chatagent/server/internal/service"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	twoFactorSvc *service.TwoFactorService
}

func NewSystemHandler(twoFactorSvc *service.TwoFactorService) *SystemHandler {
	return &SystemHandler{twoFactorSvc: twoFactorSvc}
}

func (h *SystemHandler) Get2FAConfig(c *gin.Context) {
	config, err := h.twoFactorSvc.GetConfig()
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	response.Success(c, config)
}

type set2FAReq struct {
	Enabled bool   `json:"enabled"`
	Issuer  string `json:"issuer"`
}

func (h *SystemHandler) Set2FAConfig(c *gin.Context) {
	var req set2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}
	if err := h.twoFactorSvc.SetConfig(req.Enabled, req.Issuer); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	response.Success(c, nil)
}
