package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"room-api/internal/auth"
	"room-api/internal/realtime"
	"room-api/internal/response"
	"room-api/internal/service"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
)

type WSHandler struct {
	tokens *auth.Service
	rooms  *service.RoomService
	hub    *realtime.Hub
}

func NewWSHandler(tokens *auth.Service, rooms *service.RoomService, hub *realtime.Hub) *WSHandler {
	return &WSHandler{tokens: tokens, rooms: rooms, hub: hub}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *WSHandler) ConnectRoom(c *gin.Context) {
	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return
	}

	claims, ok := h.parseClaims(c)
	if !ok {
		response.Error(c, 401, "未登录")
		return
	}

	if _, err := h.rooms.Detail(claims.UserID, roomID); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := realtime.NewClient(roomID, claims.UserID, conn)
	h.hub.Add(client)
	h.run(client)
}

func (h *WSHandler) run(client *realtime.Client) {
	done := make(chan struct{})

	client.Conn.SetReadLimit(512)
	if err := client.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return
	}
	client.Conn.SetPongHandler(func(string) error {
		return client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	go func() {
		defer close(done)
		for {
			if _, _, err := client.Conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer func() { _ = client.Conn.Close() }()
	defer h.hub.Remove(client)
	defer h.leaveAfterDisconnect(client)

	for {
		select {
		case <-done:
			return
		case payload, ok := <-client.Send:
			if !ok {
				return
			}
			if err := client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := client.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			if err := client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return
			}
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHandler) leaveAfterDisconnect(client *realtime.Client) {
	result, err := h.rooms.Leave(client.UserID, client.RoomID)
	if err != nil || result == nil || !result.Left {
		return
	}

	h.hub.Broadcast(client.RoomID, realtime.Event{
		Type:   "member.left",
		RoomID: client.RoomID,
		Data: gin.H{
			"user_id":         client.UserID,
			"current_members": result.CurrentMemberSize,
		},
	})

	if result.OwnerChanged {
		h.hub.Broadcast(client.RoomID, realtime.Event{
			Type:   "room.owner_changed",
			RoomID: client.RoomID,
			Data: gin.H{
				"owner_id": result.NewOwnerUserID,
				"members":  result.RemainingMembers,
			},
		})
	}
}

func (h *WSHandler) parseClaims(c *gin.Context) (*auth.Claims, bool) {
	header := c.GetHeader("Authorization")
	const prefix = "Bearer "
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return nil, false
	}

	claims, err := h.tokens.Parse(header[len(prefix):])
	if err != nil {
		return nil, false
	}
	return claims, true
}
