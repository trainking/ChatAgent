package handler

import (
	"strings"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type InboxHandler struct {
	repo     *repository.InboxRepository
	userRepo *repository.UserRepository
}

func NewInboxHandler(repo *repository.InboxRepository, userRepo *repository.UserRepository) *InboxHandler {
	return &InboxHandler{repo: repo, userRepo: userRepo}
}

type createInboxReq struct {
	Name           string   `json:"name" binding:"required"`
	Description    string   `json:"description"`
	WelcomeTitle   string   `json:"welcome_title"`
	WelcomeMessage string   `json:"welcome_message"`
	InboxType      string   `json:"inbox_type" binding:"required"`
	Status         string   `json:"status"`
	Collaborators  []string `json:"collaborators"`
}

type updateInboxReq struct {
	Description    string   `json:"description"`
	WelcomeTitle   string   `json:"welcome_title"`
	WelcomeMessage string   `json:"welcome_message"`
	Status         string   `json:"status"`
	Collaborators  []string `json:"collaborators"`
}

func (h *InboxHandler) List(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	isAdmin := claims.Role == "admin" || claims.Role == "super_admin"
	inboxes, err := h.repo.ListByUser(claims.UserID, isAdmin)
	if err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, inboxes)
}

func (h *InboxHandler) Get(c *gin.Context) {
	id := c.Param("id")
	inbox, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	inbox.Creator, _ = h.repo.GetCreator(id)
	inbox.Collaborators, _ = h.repo.GetCollaborators(id)
	if inbox.Collaborators == nil {
		inbox.Collaborators = []model.User{}
	}

	response.Success(c, inbox)
}

func (h *InboxHandler) Create(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	var req createInboxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.ErrorMsg(c, errcode.InvalidParam, "name is required")
		return
	}

	if req.InboxType != "website" && req.InboxType != "api" {
		response.ErrorMsg(c, errcode.InvalidParam, "inbox_type must be website or api")
		return
	}

	if len(req.WelcomeTitle) > 32 {
		response.ErrorMsg(c, errcode.InvalidParam, "welcome_title max 32 characters")
		return
	}
	if len(req.WelcomeMessage) > 255 {
		response.ErrorMsg(c, errcode.InvalidParam, "welcome_message max 255 characters")
		return
	}

	existing, _ := h.repo.FindByName(req.Name)
	if existing != nil {
		response.ErrorMsg(c, errcode.InvalidParam, "inbox name already exists")
		return
	}

	status := req.Status
	if status == "" {
		status = "enabled"
	}
	if status != "enabled" && status != "disabled" {
		response.ErrorMsg(c, errcode.InvalidParam, "status must be enabled or disabled")
		return
	}

	inbox := &model.Inbox{
		Name:           req.Name,
		Description:    req.Description,
		WelcomeTitle:   req.WelcomeTitle,
		WelcomeMessage: req.WelcomeMessage,
		InboxType:      req.InboxType,
		Status:         status,
		CreatedBy:      claims.UserID,
	}

	if err := h.repo.Create(inbox); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	for _, uid := range req.Collaborators {
		if err := h.repo.AddCollaborator(inbox.ID, uid); err != nil {
			continue
		}
	}

	inbox.Creator, _ = h.repo.GetCreator(inbox.ID)
	inbox.Collaborators, _ = h.repo.GetCollaborators(inbox.ID)
	if inbox.Collaborators == nil {
		inbox.Collaborators = []model.User{}
	}

	response.Success(c, inbox)
}

func (h *InboxHandler) Update(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	id := c.Param("id")
	inbox, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	isAdmin := claims.Role == "admin" || claims.Role == "super_admin"
	isCollaborator, _ := h.isCollaborator(id, claims.UserID)
	if !isAdmin && inbox.CreatedBy != claims.UserID && !isCollaborator {
		response.Error(c, errcode.Forbidden)
		return
	}

	var req updateInboxReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if len(req.WelcomeTitle) > 32 {
		response.ErrorMsg(c, errcode.InvalidParam, "welcome_title max 32 characters")
		return
	}
	if len(req.WelcomeMessage) > 255 {
		response.ErrorMsg(c, errcode.InvalidParam, "welcome_message max 255 characters")
		return
	}

	inbox.Description = req.Description
	inbox.WelcomeTitle = req.WelcomeTitle
	inbox.WelcomeMessage = req.WelcomeMessage
	if req.Status != "" {
		if req.Status != "enabled" && req.Status != "disabled" {
			response.ErrorMsg(c, errcode.InvalidParam, "status must be enabled or disabled")
			return
		}
		inbox.Status = req.Status
	}

	if err := h.repo.Update(inbox); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	if req.Collaborators != nil {
		if err := h.repo.RemoveAllCollaborators(id); err != nil {
			response.Error(c, errcode.ServerError)
			return
		}
		for _, uid := range req.Collaborators {
			if err := h.repo.AddCollaborator(id, uid); err != nil {
				continue
			}
		}
	}

	inbox.Creator, _ = h.repo.GetCreator(id)
	inbox.Collaborators, _ = h.repo.GetCollaborators(id)
	if inbox.Collaborators == nil {
		inbox.Collaborators = []model.User{}
	}

	response.Success(c, inbox)
}

func (h *InboxHandler) Delete(c *gin.Context) {
	claims := AuthClaims(c)
	if claims == nil {
		response.Error(c, errcode.Unauthorized)
		return
	}

	id := c.Param("id")
	inbox, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.NotFound)
		return
	}

	isSuperAdmin := claims.Role == "super_admin"
	if inbox.CreatedBy != claims.UserID && !isSuperAdmin {
		response.Error(c, errcode.Forbidden)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		response.Error(c, errcode.ServerError)
		return
	}

	response.Success(c, nil)
}

func (h *InboxHandler) isCollaborator(inboxID, userID string) (bool, error) {
	collabs, err := h.repo.GetCollaborators(inboxID)
	if err != nil {
		return false, err
	}
	for _, u := range collabs {
		if u.ID == userID {
			return true, nil
		}
	}
	return false, nil
}
