package handler

import (
	"math/rand"
	"strings"
	"time"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

type createUserReq struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type updateUserReq struct {
	Name               string `json:"name"`
	Role               string `json:"role"`
	Status             string `json:"status"`
	MustChangePassword *bool  `json:"must_change_password"`
}

func (h *UserHandler) List(c *gin.Context) {
	users, err := h.repo.List()
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	claims := AuthClaims(c)
	if claims == nil || claims.Role != "super_admin" {
		filtered := make([]model.User, 0)
		for _, u := range users {
			if u.Role != "super_admin" {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	response.Success(c, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if !strings.Contains(req.Email, "@") {
		response.ErrorMsg(c, errcode.InvalidParam, "invalid email format")
		return
	}

	existing, _ := h.repo.FindByEmail(req.Email)
	if existing != nil {
		response.Error(c, errcode.EmailExists)
		return
	}

	if req.Role == "" {
		req.Role = "agent"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	user := &model.User{
		Email:              req.Email,
		Name:               req.Name,
		PasswordHash:       string(hash),
		Role:               req.Role,
		Status:             "active",
		MustChangePassword: true,
	}

	if err := h.repo.Create(user); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")
	user, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	var req updateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Status != "" {
		user.Status = req.Status
	}
	if req.MustChangePassword != nil {
		user.MustChangePassword = *req.MustChangePassword
	}

	if err := h.repo.Update(user); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	claims := AuthClaims(c)
	if claims != nil && claims.UserID == id {
		response.ErrorMsg(c, errcode.InvalidParam, "cannot delete yourself")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := c.Param("id")
	claims := AuthClaims(c)
	if claims == nil || claims.Role != "super_admin" {
		response.Error(c, errcode.Forbidden)
		return
	}

	newPwd := randomPassword(6)
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	if err := h.repo.AdminResetPassword(id, string(hash)); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, gin.H{"password": newPwd})
}

func randomPassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
