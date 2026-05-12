package handler

import (
	"strconv"

	"github.com/chatagent/server/internal/model"
	"github.com/chatagent/server/internal/repository"
	"github.com/chatagent/server/pkg/errcode"
	"github.com/chatagent/server/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type ContactHandler struct {
	repo     *repository.ContactRepository
	convRepo *repository.ConversationRepository
	db       *sqlx.DB
}

func NewContactHandler(repo *repository.ContactRepository, convRepo *repository.ConversationRepository, db *sqlx.DB) *ContactHandler {
	return &ContactHandler{repo: repo, convRepo: convRepo, db: db}
}

func (h *ContactHandler) List(c *gin.Context) {
	inboxID := c.DefaultQuery("inbox_id", "")
	contactType := c.DefaultQuery("type", "")
	search := c.DefaultQuery("search", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := repository.ContactFilter{
		InboxID:  inboxID,
		Type:     contactType,
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	}

	contacts, total, err := h.repo.List(filter)
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	if contacts == nil {
		contacts = []model.Contact{}
	}

	response.Page(c, total, contacts)
}

type createContactReq struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	ContactType string `json:"contact_type"`
	CustomAttrs string `json:"custom_attrs"`
}

func (h *ContactHandler) Create(c *gin.Context) {
	var req createContactReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if req.Name == "" {
		response.ErrorMsg(c, errcode.InvalidParam, "name is required")
		return
	}
	if req.ContactType == "" {
		req.ContactType = "visitor"
	}
	if req.CustomAttrs == "" {
		req.CustomAttrs = "{}"
	}

	if req.Email != "" {
		existing, _ := h.repo.FindByEmail(req.Email)
		if existing != nil {
			response.Error(c, errcode.ContactExists)
			return
		}
	}

	contact := &model.Contact{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		ContactType: req.ContactType,
		CustomAttrs: req.CustomAttrs,
	}

	if err := h.repo.Create(contact); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	response.Success(c, contact)
}

func (h *ContactHandler) Get(c *gin.Context) {
	id := c.Param("id")

	contact, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	response.Success(c, contact)
}

type updateContactReq struct {
	Name        *string `json:"name"`
	Email       *string `json:"email"`
	Phone       *string `json:"phone"`
	ContactType *string `json:"contact_type"`
	CustomAttrs *string `json:"custom_attrs"`
}

func (h *ContactHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req updateContactReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	_, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.ContactType != nil {
		updates["contact_type"] = *req.ContactType
	}
	if req.CustomAttrs != nil {
		updates["custom_attrs"] = *req.CustomAttrs
	}

	if len(updates) == 0 {
		response.ErrorMsg(c, errcode.InvalidParam, "no fields to update")
		return
	}

	if err := h.repo.Update(id, updates); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	response.Success(c, nil)
}

type mergeReq struct {
	TargetID string `json:"target_id" binding:"required"`
}

func (h *ContactHandler) Merge(c *gin.Context) {
	id := c.Param("id")

	var req mergeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.InvalidParam)
		return
	}

	if id == req.TargetID {
		response.ErrorMsg(c, errcode.InvalidParam, "cannot merge into self")
		return
	}

	source, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	target, err := h.repo.FindByID(req.TargetID)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	// Transfer conversations to target
	tx, err := h.db.Beginx()
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec("UPDATE conversations SET contact_id = $1 WHERE contact_id = $2", target.ID, source.ID); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	// Drop source contact_inboxes that would conflict with existing target rows,
	// then transfer the remaining rows to the target contact.
	if _, err := tx.Exec(
		`DELETE FROM contact_inboxes s
		 WHERE s.contact_id = $1
		   AND EXISTS (
		     SELECT 1 FROM contact_inboxes t
		     WHERE t.contact_id = $2 AND t.inbox_id = s.inbox_id
		   )`,
		source.ID, target.ID,
	); err != nil {
		response.Error(c, errcode.DBError)
		return
	}
	if _, err := tx.Exec(
		"UPDATE contact_inboxes SET contact_id = $1 WHERE contact_id = $2",
		target.ID, source.ID,
	); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	// Merge fingerprints
	var sourceFPs string
	if err := tx.Get(&sourceFPs, "SELECT browser_fingerprints FROM contacts WHERE id = $1", source.ID); err != nil {
		response.Error(c, errcode.DBError)
		return
	}
	if sourceFPs != "" && sourceFPs != "[]" {
		var targetFPs string
		if err := tx.Get(&targetFPs, "SELECT browser_fingerprints FROM contacts WHERE id = $1", target.ID); err != nil {
			response.Error(c, errcode.DBError)
			return
		}
		// Simple merge — append source FPs to target
		// In production, we'd parse and deduplicate JSONB arrays
		if _, err := tx.Exec(
			"UPDATE contacts SET browser_fingerprints = browser_fingerprints::jsonb || ($1)::jsonb WHERE id = $2",
			sourceFPs, target.ID,
		); err != nil {
			response.Error(c, errcode.DBError)
			return
		}
	}

	// Soft-delete source contact
	if _, err := tx.Exec("UPDATE contacts SET email = '', phone = '', blocked = true, updated_at = NOW() WHERE id = $1", source.ID); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	if err := tx.Commit(); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	response.Success(c, nil)
}

func (h *ContactHandler) GetConversations(c *gin.Context) {
	id := c.Param("id")

	_, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	convs, err := h.convRepo.FindByContact(id)
	if err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	if convs == nil {
		convs = []model.Conversation{}
	}

	response.Success(c, convs)
}

func (h *ContactHandler) Block(c *gin.Context) {
	id := c.Param("id")

	_, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	if err := h.repo.Block(id); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	response.Success(c, nil)
}

func (h *ContactHandler) Unblock(c *gin.Context) {
	id := c.Param("id")

	_, err := h.repo.FindByID(id)
	if err != nil {
		response.Error(c, errcode.ContactNotFound)
		return
	}

	if err := h.repo.Unblock(id); err != nil {
		response.Error(c, errcode.DBError)
		return
	}

	response.Success(c, nil)
}
