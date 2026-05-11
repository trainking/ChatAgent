package handler

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/internal/websocket"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type WidgetHandler struct {
	db                *sqlx.DB
	inboxRepo         *repository.InboxRepository
	contactRepo       *repository.ContactRepository
	contactInboxRepo  *repository.ContactInboxRepository
	convRepo          *repository.ConversationRepository
	msgRepo           *repository.MessageRepository
	hub               *websocket.Hub
}

func NewWidgetHandler(
	db *sqlx.DB,
	inboxRepo *repository.InboxRepository,
	contactRepo *repository.ContactRepository,
	contactInboxRepo *repository.ContactInboxRepository,
	convRepo *repository.ConversationRepository,
	msgRepo *repository.MessageRepository,
	hub *websocket.Hub,
) *WidgetHandler {
	return &WidgetHandler{
		db:               db,
		inboxRepo:        inboxRepo,
		contactRepo:      contactRepo,
		contactInboxRepo: contactInboxRepo,
		convRepo:         convRepo,
		msgRepo:          msgRepo,
		hub:              hub,
	}
}

func randomSuffix() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type widgetAuthReq struct {
	InboxID     string `json:"inbox_id" binding:"required"`
	Fingerprint string `json:"fingerprint"`
}

func (h *WidgetHandler) Auth(c *gin.Context) {
	var req widgetAuthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	inbox, err := h.inboxRepo.FindByID(req.InboxID)
	if err != nil {
		response.ErrorMsg(c, errcode.NotFound, "inbox not found")
		return
	}
	if inbox.InboxType != "website" {
		response.ErrorMsg(c, errcode.InvalidParam, "inbox is not a website type")
		return
	}
	if inbox.Status != "enabled" {
		response.ErrorMsg(c, errcode.InvalidParam, "inbox is disabled")
		return
	}

	var contact *model.Contact

	if req.Fingerprint != "" {
		contact, _ = h.contactRepo.FindByFingerprint(req.Fingerprint)
	}

	if contact == nil {
		fingerprints := "[]"
		if req.Fingerprint != "" {
			fingerprints = `["` + req.Fingerprint + `"]`
		}
		contact = &model.Contact{
			Name:                "V_" + randomSuffix(),
			ContactType:         "visitor",
			BrowserFingerprints: fingerprints,
			CustomAttrs:         "{}",
		}
		if err := h.contactRepo.Create(contact); err != nil {
			response.Error(c, errcode.DBError)
			return
		}
	}

	ci, err := h.contactInboxRepo.FindByContactInbox(contact.ID, req.InboxID)
	if err != nil || ci == nil {
		token := websocket.GeneratePubsubToken()
		if _, err := h.db.Exec(
			"INSERT INTO contact_inboxes (contact_id, inbox_id, source_id, pubsub_token) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING",
			contact.ID, req.InboxID, req.Fingerprint, token,
		); err != nil {
			response.Error(c, errcode.DBError)
			return
		}
		ci = &model.ContactInbox{
			ContactID:   contact.ID,
			InboxID:     req.InboxID,
			SourceID:    req.Fingerprint,
			PubsubToken: token,
		}
	}

	response.Success(c, gin.H{
		"pubsub_token":     ci.PubsubToken,
		"contact_id":       contact.ID,
		"contact_inbox_id": ci.ContactID,
		"welcome_title":    inbox.WelcomeTitle,
		"welcome_message":  inbox.WelcomeMessage,
	})
}

type widgetMessageReq struct {
	PubsubToken string `json:"pubsub_token" binding:"required"`
	Content     string `json:"content" binding:"required"`
	ContentType string `json:"content_type"`
	Name        string `json:"name"`
	Email       string `json:"email"`
}

func (h *WidgetHandler) SendMessage(c *gin.Context) {
	var req widgetMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if req.ContentType == "" {
		req.ContentType = "text/html"
	}

	// Find contact_inbox by pubsub_token
	ci, err := h.contactInboxRepo.FindByPubsubToken(req.PubsubToken)
	if err != nil {
		response.Error(c, errcode.Forbidden)
		return
	}

	// Find contact
	contact, err := h.contactRepo.FindByID(ci.ContactID)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	// Update contact name/email if provided
	if req.Name != "" && contact.Name == "" || (req.Name != "" && contact.Name[:2] == "V_") {
		h.contactRepo.Update(contact.ID, map[string]interface{}{"name": req.Name})
	}
	if req.Email != "" && contact.Email == "" {
		h.contactRepo.Update(contact.ID, map[string]interface{}{"email": req.Email})
	}

	// Find open conversation for this contact+inbox, or create a new one
	var conv *model.Conversation
	convs, _, err := h.convRepo.List(repository.ConversationFilter{
		InboxID:  ci.InboxID,
		Page:     1,
		PageSize: 1,
	})
	if err == nil {
		for _, c := range convs {
			if c.ContactID == ci.ContactID && c.Status != "resolved" && c.Status != "closed" {
				conv = &c
				break
			}
		}
	}

	if conv == nil {
		conv = &model.Conversation{
			InboxID:   ci.InboxID,
			ContactID: ci.ContactID,
			Status:    "open",
			Priority:  "medium",
			Subject:   truncateSubject(req.Content, 100),
		}
		if err := h.convRepo.Create(conv); err != nil {
			response.Error(c, errcode.DBError)
			return
		}
	}

	// Create message
	msg := &model.Message{
		ConversationID: conv.ID,
		SenderType:     "contact",
		SenderID:       contact.ID,
		Content:        req.Content,
		MessageType:    "incoming",
		ContentType:    req.ContentType,
		Status:         "sent",
		Metadata:       "{}",
	}
	if err := h.msgRepo.Create(msg); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	// Update conversation metadata
	h.convRepo.OnNewMessage(conv.ID, "contact", true)
	h.convRepo.IncrementUnread(conv.ID)

	// Broadcast to agents
	h.hub.Broadcast("inbox:"+ci.InboxID, "conversation.created", gin.H{
		"conversation_id": conv.ID,
		"contact_id":      contact.ID,
		"contact_name":    contact.Name,
	})
	h.hub.Broadcast("inbox:"+ci.InboxID, "message.created", msg)

	response.Success(c, gin.H{
		"conversation_id": conv.ID,
		"message":         msg,
	})
}

func truncateSubject(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
