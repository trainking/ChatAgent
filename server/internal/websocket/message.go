package websocket

import "encoding/json"

type WSRequest struct {
	Action  string          `json:"action"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type WSEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type TypingData struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Typing   bool   `json:"typing"`
}
