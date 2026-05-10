package handler

import (
	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/service"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc      *service.AuthService
	twoFA    *service.TwoFactorService
	userRepo *repository.UserRepository
	actRepo  *repository.ActivityRepository
}

func NewAuthHandler(svc *service.AuthService, twoFA *service.TwoFactorService, userRepo *repository.UserRepository, actRepo *repository.ActivityRepository) *AuthHandler {
	return &AuthHandler{svc: svc, twoFA: twoFA, userRepo: userRepo, actRepo: actRepo}
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type initRootReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type changePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) Status(c *gin.Context) {
	initialized, err := h.svc.Status()
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	response.Success(c, gin.H{"initialized": initialized})
}

func (h *AuthHandler) InitRoot(c *gin.Context) {
	var req initRootReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	user, token, err := h.svc.InitRoot(req.Email, req.Password, req.Name)
	if err != nil {
		if err == service.ErrSystemInitialized {
			response.Error(c, errcode.Forbidden)
		} else {
			response.ErrorMsg(c, errcode.UnknownError, err.Error())
		}
		return
	}

	response.Success(c, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	user, _, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			response.Error(c, errcode.UserNotFound)
		case service.ErrWrongPassword:
			response.Error(c, errcode.WrongPassword)
		case service.ErrUserDisabled:
			response.Error(c, errcode.UserDisabled)
		default:
			response.ErrorMsg(c, errcode.UnknownError, err.Error())
		}
		return
	}

	twoFAEnabled, _ := h.twoFA.IsEnabled()
	if twoFAEnabled {
		has2FA := h.twoFA.HasEnabled2FA(user.ID)
		tempToken, err := h.twoFA.GenerateTempToken(user)
		if err != nil {
			response.Error(c, errcode.ServerError)
			return
		}
		if has2FA {
			response.Success(c, gin.H{
				"require_2fa": true,
				"temp_token":  tempToken,
			})
			return
		}
		response.Success(c, gin.H{
			"require_2fa_setup": true,
			"temp_token":        tempToken,
		})
		return
	}

	token, err := h.svc.GenerateToken(user)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	h.userRepo.UpdateLastLogin(user.ID)
	user.OnlineStatus = "online"
	h.actRepo.Create(&model.ActivityLog{
		UserID:  user.ID,
		Event:   "login",
		Module:  "auth",
		Content: "User logged in",
	})

	response.Success(c, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	var req changePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if err := h.svc.ChangePassword(claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		switch err {
		case service.ErrWrongPassword:
			response.ErrorMsg(c, errcode.WrongPassword, "old password is incorrect")
		case service.ErrSamePassword:
			response.ErrorMsg(c, errcode.InvalidParam, "new password must be different from old password")
		case service.ErrUserNotFound:
			response.Error(c, errcode.UserNotFound)
		default:
			response.ErrorMsg(c, errcode.UnknownError, err.Error())
		}
		return
	}

	h.actRepo.Create(&model.ActivityLog{
		UserID:  claims.UserID,
		Event:   "change_password",
		Module:  "auth",
		Content: "Password changed",
	})

	response.Success(c, nil)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	h.userRepo.UpdateOnlineStatus(claims.UserID, "offline")
	h.actRepo.Create(&model.ActivityLog{
		UserID:  claims.UserID,
		Event:   "logout",
		Module:  "auth",
		Content: "User logged out",
	})

	response.Success(c, nil)
}
