package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/middleware"
	"room-api/internal/realtime"
	"room-api/internal/response"
	"room-api/internal/service"
)

type MessageHandler struct {
	messages *service.MessageService
	hub      *realtime.Hub
}

func NewMessageHandler(messages *service.MessageService, hub *realtime.Hub) *MessageHandler {
	return &MessageHandler{messages: messages, hub: hub}
}

type createMessageRequest struct {
	Content string `json:"content"`
}

func (h *MessageHandler) List(c *gin.Context) {
	userID, roomID, ok := currentUserAndRoomID(c)
	if !ok {
		return
	}

	limit := parseIntDefault(c.Query("limit"), 20)
	beforeID := parseUintDefault(c.Query("before_id"), 0)

	result, err := h.messages.List(userID, roomID, limit, beforeID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *MessageHandler) Create(c *gin.Context) {
	userID, roomID, ok := currentUserAndRoomID(c)
	if !ok {
		return
	}

	var req createMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 500, "参数错误")
		return
	}

	message, err := h.messages.Create(userID, roomID, req.Content)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	h.hub.Broadcast(roomID, realtime.Event{
		Type:   "message.created",
		RoomID: roomID,
		Data:   gin.H{"message": message},
	})

	response.OK(c, gin.H{"message": message})
}

func currentUserAndRoomID(c *gin.Context) (uint64, uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "未登录")
		return 0, 0, false
	}

	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return 0, 0, false
	}

	return userID, roomID, true
}
