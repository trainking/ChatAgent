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

type MessageHandler struct {
	repo    *repository.MessageRepository
	convRepo *repository.ConversationRepository
	hub     *websocket.Hub
}

func NewMessageHandler(repo *repository.MessageRepository, convRepo *repository.ConversationRepository, hub *websocket.Hub) *MessageHandler {
	return &MessageHandler{repo: repo, convRepo: convRepo, hub: hub}
}

func (h *MessageHandler) List(c *gin.Context) {
	convID := c.Param("id")
	beforeID := c.DefaultQuery("before_id", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	msgs, err := h.repo.FindByConversation(convID, beforeID, limit)
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	if msgs == nil {
		msgs = []model.Message{}
	}

	response.Success(c, msgs)
}

type createMessageReq struct {
	Content     string `json:"content" binding:"required"`
	ContentType string `json:"content_type"`
	FileURL     string `json:"file_url"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	Private     bool   `json:"private"`
}

func (h *MessageHandler) Create(c *gin.Context) {
	convID := c.Param("id")
	claims := AuthClaims(c)

	var req createMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if req.ContentType == "" {
		req.ContentType = "text/html"
	}

	conv, err := h.convRepo.FindByID(convID)
	if err != nil {
		response.Error(c, errcode.ConversationNotFound)
		return
	}

	msg := &model.Message{
		ConversationID: convID,
		SenderType:     "user",
		SenderID:       claims.UserID,
		Content:        req.Content,
		MessageType:    "outgoing",
		ContentType:    req.ContentType,
		FileURL:        req.FileURL,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		Private:        req.Private,
		Status:         "sent",
		Metadata:       "{}",
	}

	if err := h.repo.Create(msg); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	// Update conversation metadata
	h.convRepo.OnNewMessage(convID, "user", false)

	// Handle status transitions
	if conv.Status == "pending" {
		h.convRepo.UpdateStatus(convID, "open")
	}

	// Broadcast to subscribers
	event := "message.created"
	if msg.Private {
		event = "message.created"
	}

	h.hub.Broadcast("conversation:"+convID, event, msg)
	h.hub.Broadcast("inbox:"+conv.InboxID, event, msg)

	response.Success(c, msg)
}

func (h *MessageHandler) Retry(c *gin.Context) {
	convID := c.Param("id")
	msgID := c.Param("fid") // :fid from route /conversations/:id/messages/:fid/retry

	msg, err := h.repo.FindByID(msgID)
	if err != nil {
		response.Error(c, errcode.MessageNotFound)
		return
	}

	if msg.ConversationID != convID {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if msg.Status != "failed" {
		response.ErrorMsg(c, errcode.InvalidParam, "only failed messages can be retried")
		return
	}

	if err := h.repo.UpdateStatus(msgID, "sent"); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	msg.Status = "sent"

	h.hub.Broadcast("conversation:"+convID, "message.updated", msg)

	response.Success(c, msg)
}
