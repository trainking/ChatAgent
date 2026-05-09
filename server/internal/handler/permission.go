package handler

import (
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	repo *repository.PermissionRepository
}

func NewPermissionHandler(repo *repository.PermissionRepository) *PermissionHandler {
	return &PermissionHandler{repo: repo}
}

func (h *PermissionHandler) ListAll(c *gin.Context) {
	perms, err := h.repo.ListAll()
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	response.Success(c, perms)
}

func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if role != "admin" && role != "agent" {
		response.ErrorMsg(c, errcode.InvalidParam, "invalid role")
		return
	}

	codes, err := h.repo.GetRolePermissions(role)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}
	if codes == nil {
		codes = []string{}
	}

	response.Success(c, gin.H{"role": role, "permissions": codes})
}

type setRolePermissionsReq struct {
	Permissions []string `json:"permissions" binding:"required"`
}

func (h *PermissionHandler) SetRolePermissions(c *gin.Context) {
	role := c.Param("role")
	if role != "admin" && role != "agent" {
		response.ErrorMsg(c, errcode.InvalidParam, "invalid role")
		return
	}

	var req setRolePermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if err := h.repo.SetRolePermissions(role, req.Permissions); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, nil)
}
