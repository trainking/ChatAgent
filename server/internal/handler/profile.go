package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	userRepo *repository.UserRepository
	actRepo  *repository.ActivityRepository
}

func NewProfileHandler(userRepo *repository.UserRepository, actRepo *repository.ActivityRepository) *ProfileHandler {
	return &ProfileHandler{userRepo: userRepo, actRepo: actRepo}
}

type updateProfileReq struct {
	Name string `json:"name" binding:"required"`
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	response.Success(c, user)
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.ErrorMsg(c, errcode.InvalidParam, "name is required")
		return
	}

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	if err := h.userRepo.UpdateProfile(claims.UserID, req.Name, user.AvatarURL); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	h.actRepo.Create(&model.ActivityLog{
		UserID:  claims.UserID,
		Event:   "update_profile",
		Module:  "profile",
		Content: fmt.Sprintf("Updated profile name from '%s' to '%s'", user.Name, req.Name),
	})

	user.Name = req.Name
	response.Success(c, user)
}

func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		response.ErrorMsg(c, errcode.InvalidParam, "avatar file is required")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		response.ErrorMsg(c, errcode.InvalidParam, "only jpg, png, gif are allowed")
		return
	}

	if file.Size > 2*1024*1024 {
		response.ErrorMsg(c, errcode.InvalidParam, "avatar file size must be less than 2MB")
		return
	}

	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	filename := fmt.Sprintf("%s_%d%s", claims.UserID, time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, filename)

	src, err := file.Open()
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	avatarURL := "/" + filepath.ToSlash(savePath)

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	if err := h.userRepo.UpdateProfile(claims.UserID, user.Name, avatarURL); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	h.actRepo.Create(&model.ActivityLog{
		UserID:  claims.UserID,
		Event:   "upload_avatar",
		Module:  "profile",
		Content: "Uploaded new avatar",
	})

	response.Success(c, gin.H{"avatar_url": avatarURL})
}

func (h *ProfileHandler) GetActivities(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	logs, total, err := h.actRepo.FindByUserID(claims.UserID, pageSize, offset)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Page(c, total, logs)
}
