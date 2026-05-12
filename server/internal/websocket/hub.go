package websocket

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/chatagent/server/internal/repository"
)

type Hub struct {
	clients    map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
	agentConns map[string]*Client
	convRepo   *repository.ConversationRepository
}

func NewHub(convRepo *repository.ConversationRepository) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		agentConns: make(map[string]*Client),
		convRepo:   convRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if client.ClientType == ClientTypeAgent {
				h.agentConns[client.UserID] = client
			}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if client.ClientType == ClientTypeAgent {
				if h.agentConns[client.UserID] == client {
					delete(h.agentConns, client.UserID)
				}
			}
			for ch := range client.Channels {
				if clients, ok := h.clients[ch]; ok {
					delete(clients, client)
					if len(clients) == 0 {
						delete(h.clients, ch)
					}
				}
			}
			close(client.Send)
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client.Channels[channel] = true
	if _, ok := h.clients[channel]; !ok {
		h.clients[channel] = make(map[*Client]bool)
	}
	h.clients[channel][client] = true
}

func (h *Hub) Unsubscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(client.Channels, channel)
	if clients, ok := h.clients[channel]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, channel)
		}
	}
}

func (h *Hub) Broadcast(channel string, event string, data interface{}) {
	msg, err := json.Marshal(WSEvent{Event: event, Data: data})
	if err != nil {
		log.Printf("broadcast marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[channel]; ok {
		for client := range clients {
			select {
			case client.Send <- msg:
			default:
				go func(c *Client) {
					h.Unregister <- c
				}(client)
			}
		}
	}
}

func (h *Hub) SendToUser(userID string, event string, data interface{}) {
	msg, err := json.Marshal(WSEvent{Event: event, Data: data})
	if err != nil {
		log.Printf("send to user marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.agentConns[userID]; ok {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

func (h *Hub) ProcessMessage(client *Client, message []byte) {
	var req WSRequest
	if err := json.Unmarshal(message, &req); err != nil {
		log.Printf("ws unmarshal error: %v", err)
		return
	}

	switch req.Action {
	case "subscribe":
		if !h.CanSubscribe(client, req.Channel) {
			log.Printf("ws subscribe denied: client=%s channel=%s", client.ID, req.Channel)
			return
		}
		h.Subscribe(client, req.Channel)
	case "unsubscribe":
		h.Unsubscribe(client, req.Channel)
	default:
		log.Printf("ws unknown action: %s", req.Action)
	}
}

func (h *Hub) CanSubscribe(client *Client, channel string) bool {
	if channel == "" {
		return false
	}

	if strings.HasPrefix(channel, "inbox:") {
		if client.ClientType != ClientTypeAgent {
			return false
		}
		inboxID := strings.TrimPrefix(channel, "inbox:")
		return stringInSlice(inboxID, client.InboxIDs)
	}

	if strings.HasPrefix(channel, "conversation:") {
		if h.convRepo == nil {
			return false
		}
		convID := strings.TrimPrefix(channel, "conversation:")
		conv, err := h.convRepo.FindByID(convID)
		if err != nil {
			return false
		}
		if client.ClientType == ClientTypeWidget {
			return conv.ContactID == client.ContactID && conv.InboxID == client.InboxID
		}
		return stringInSlice(conv.InboxID, client.InboxIDs)
	}

	return false
}

func stringInSlice(value string, values []string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.agentConns[userID]
	return ok
}
