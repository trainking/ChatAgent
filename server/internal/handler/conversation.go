package handler

import (
	"strconv"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/websocket"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	repo        *repository.ConversationRepository
	msgRepo     *repository.MessageRepository
	contactRepo *repository.ContactRepository
	userRepo    *repository.UserRepository
	inboxRepo   *repository.InboxRepository
	hub         *websocket.Hub
}

func NewConversationHandler(
	repo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	contactRepo *repository.ContactRepository,
	userRepo *repository.UserRepository,
	inboxRepo *repository.InboxRepository,
	hub *websocket.Hub,
) *ConversationHandler {
	return &ConversationHandler{
		repo: repo, msgRepo: msgRepo, contactRepo: contactRepo,
		userRepo: userRepo, inboxRepo: inboxRepo, hub: hub,
	}
}

type conversationListReq struct {
	InboxID    string `json:"inbox_id"`
	Status     string `json:"status"`
	Priority   string `json:"priority"`
	AssigneeID string `json:"assignee_id"`
	Search     string `json:"search"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

func (h *ConversationHandler) List(c *gin.Context) {
	claims := AuthClaims(c)

	var req conversationListReq
	req.InboxID = c.DefaultQuery("inbox_id", "")
	req.Status = c.DefaultQuery("status", "")
	req.Priority = c.DefaultQuery("priority", "")
	req.AssigneeID = c.DefaultQuery("assignee_id", "")
	req.Search = c.DefaultQuery("search", "")
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := repository.ConversationFilter{
		InboxID:    req.InboxID,
		Status:     req.Status,
		Priority:   req.Priority,
		AssigneeID: req.AssigneeID,
		Search:     req.Search,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	if claims.Role == "agent" {
		if filter.AssigneeID == "me" {
			// agents can only see their own
		} else {
			// agents see only their inboxes' conversations
			filter.AssigneeID = "" // agents can't filter by other agents
		}
	}

	convs, total, err := h.repo.List(filter)
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	// Enrich with contact and last message
	type enrichedConv struct {
		model.Conversation
		ContactName  string `json:"contact_name"`
		ContactEmail string `json:"contact_email"`
		LastMessage  *model.Message `json:"last_message"`
	}

	result := make([]enrichedConv, 0, len(convs))
	for _, conv := range convs {
		ec := enrichedConv{Conversation: conv}
		if contact, err := h.contactRepo.FindByID(conv.ContactID); err == nil {
			ec.ContactName = contact.Name
			ec.ContactEmail = contact.Email
		}
		if msg, err := h.msgRepo.GetLastMessage(conv.ID); err == nil {
			ec.LastMessage = msg
		}
		result = append(result, ec)
	}

	response.Page(c, total, result)
}

func (h *ConversationHandler) Get(c *gin.Context) {
	id := c.Param("id")
	conv, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ConversationNotFound)
		return
	}

	type detailResp struct {
		model.Conversation
		Contact  *model.Contact `json:"contact"`
		Assignee *model.User    `json:"assignee"`
	}

	resp := detailResp{Conversation: *conv}

	if contact, err := h.contactRepo.FindByID(conv.ContactID); err == nil {
		resp.Contact = contact
	}
	if conv.AssigneeID != nil {
		if user, err := h.userRepo.FindByID(*conv.AssigneeID); err == nil {
			resp.Assignee = &model.User{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				Role:      user.Role,
				AvatarURL: user.AvatarURL,
			}
		}
	}

	// Mark messages as read
	h.msgRepo.MarkConversationRead(id, "")

	response.Success(c, resp)
}

type assignReq struct {
	AssigneeID string `json:"assignee_id"`
}

func (h *ConversationHandler) Assign(c *gin.Context) {
	id := c.Param("id")
	claims := AuthClaims(c)

	var req assignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	conv, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ConversationNotFound)
		return
	}

	var assigneeID *string
	if req.AssigneeID != "" {
		assigneeID = &req.AssigneeID
	}

	if err := h.repo.UpdateAssignee(id, assigneeID); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	h.hub.Broadcast("inbox:"+conv.InboxID, "conversation.assignee_changed", gin.H{
		"conversation_id": id,
		"assignee_id":     assigneeID,
		"changed_by":      claims.UserID,
	})

	response.Success(c, nil)
}

type statusReq struct {
	Status      string `json:"status"`
	SnoozedUntil *int64 `json:"snoozed_until"`
}

func (h *ConversationHandler) ChangeStatus(c *gin.Context) {
	id := c.Param("id")

	var req statusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if req.Status != "open" && req.Status != "pending" && req.Status != "resolved" && req.Status != "snoozed" {
		response.ErrorMsg(c, errcode.InvalidParam, "invalid status")
		return
	}

	conv, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ConversationNotFound)
		return
	}

	oldStatus := conv.Status

	if err := h.repo.UpdateStatus(id, req.Status); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	h.hub.Broadcast("inbox:"+conv.InboxID, "conversation.status_changed", gin.H{
		"conversation_id": id,
		"old_status":      oldStatus,
		"new_status":      req.Status,
	})

	response.Success(c, nil)
}

type priorityReq struct {
	Priority string `json:"priority"`
}

func (h *ConversationHandler) ChangePriority(c *gin.Context) {
	id := c.Param("id")

	var req priorityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" && req.Priority != "urgent" {
		response.ErrorMsg(c, errcode.InvalidParam, "invalid priority")
		return
	}

	conv, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ConversationNotFound)
		return
	}

	if err := h.repo.UpdatePriority(id, req.Priority); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	h.hub.Broadcast("inbox:"+conv.InboxID, "conversation.priority_changed", gin.H{
		"conversation_id": id,
		"priority":        req.Priority,
	})

	response.Success(c, nil)
}

func (h *ConversationHandler) UnreadCount(c *gin.Context) {
	claims := AuthClaims(c)

	// For admins, get unread across all inboxes
	// For agents, get unread across their inboxes
	isAdmin := claims.Role == "admin" || claims.Role == "super_admin"
	inboxes, err := h.inboxRepo.ListByUser(claims.UserID, isAdmin)
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	inboxIDs := make([]string, len(inboxes))
	for i, in := range inboxes {
		inboxIDs[i] = in.ID
	}

	total, err := h.repo.GetUnreadCount(claims.UserID, inboxIDs)
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	response.Success(c, gin.H{"count": total})
}
