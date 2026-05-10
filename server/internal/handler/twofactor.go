package handler

import (
	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/service"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type TwoFactorHandler struct {
	svc      *service.TwoFactorService
	authSvc  *service.AuthService
	userRepo *repository.UserRepository
	actRepo  *repository.ActivityRepository
}

func NewTwoFactorHandler(svc *service.TwoFactorService, authSvc *service.AuthService, userRepo *repository.UserRepository, actRepo *repository.ActivityRepository) *TwoFactorHandler {
	return &TwoFactorHandler{svc: svc, authSvc: authSvc, userRepo: userRepo, actRepo: actRepo}
}

type setup2FAResp struct {
	Secret  string `json:"secret"`
	QrURL   string `json:"qr_url"`
}

func (h *TwoFactorHandler) Setup(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	secret, qrURL, err := h.svc.GenerateSecret(claims.UserID, claims.Email)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, setup2FAResp{Secret: secret, QrURL: qrURL})
}

type verifySetupReq struct {
	Code string `json:"code" binding:"required"`
}

func (h *TwoFactorHandler) VerifySetup(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	var req verifySetupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if err := h.svc.VerifySetup(claims.UserID, req.Code); err != nil {
		response.ErrorMsg(c, errcode.InvalidParam, err.Error())
		return
	}

	response.Success(c, nil)
}

type verify2FAReq struct {
	TempToken string `json:"temp_token" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

func (h *TwoFactorHandler) VerifyLogin(c *gin.Context) {
	var req verify2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	user, err := h.svc.ValidateTempToken(req.TempToken)
	if err != nil {
		response.Error(c, errcode.TokenExpired)
		return
	}

	if err := h.svc.VerifyCode(user.ID, req.Code); err != nil {
		response.ErrorMsg(c, errcode.InvalidParam, err.Error())
		return
	}

	token, err := h.authSvc.GenerateToken(user)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	h.userRepo.UpdateLastLogin(user.ID)
	h.actRepo.Create(&model.ActivityLog{
		UserID:  user.ID,
		Event:   "login",
		Module:  "auth",
		Content: "User logged in (2FA verified)",
	})

	response.Success(c, gin.H{"token": token, "user": user})
}

func (h *TwoFactorHandler) GetStatus(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	has2FA := h.svc.HasEnabled2FA(claims.UserID)
	response.Success(c, gin.H{"enabled": has2FA})
}
