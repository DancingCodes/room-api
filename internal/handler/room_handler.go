package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"room-api/internal/middleware"
	"room-api/internal/realtime"
	"room-api/internal/response"
	"room-api/internal/service"
)

type RoomHandler struct {
	rooms *service.RoomService
	hub   *realtime.Hub
}

func NewRoomHandler(rooms *service.RoomService, hub *realtime.Hub) *RoomHandler {
	return &RoomHandler{rooms: rooms, hub: hub}
}

type createRoomRequest struct {
	MaxMembers uint8 `json:"max_members"`
}

type updateMicRequest struct {
	MicStatus string `json:"mic_status"`
}

func (h *RoomHandler) List(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)

	result, err := h.rooms.List(page, pageSize)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *RoomHandler) Create(c *gin.Context) {
	userID, ok := roomCurrentUser(c)
	if !ok {
		return
	}

	var req createRoomRequest
	if !bindRoomJSON(c, &req) {
		return
	}

	result, err := h.rooms.Create(userID, req.MaxMembers)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *RoomHandler) Detail(c *gin.Context) {
	userID, roomID, ok := roomCurrentUserAndRoomID(c)
	if !ok {
		return
	}

	result, err := h.rooms.Detail(userID, roomID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	if len(result.Members) > 0 {
		joined := result.Members[len(result.Members)-1]
		h.hub.Broadcast(roomID, realtime.Event{
			Type:   "member.joined",
			RoomID: roomID,
			Data: gin.H{
				"member":          joined,
				"current_members": result.Room.CurrentMembers,
			},
		})
	}

	response.OK(c, result)
}

func (h *RoomHandler) Join(c *gin.Context) {
	userID, roomID, ok := roomCurrentUserAndRoomID(c)
	if !ok {
		return
	}

	result, err := h.rooms.Join(userID, roomID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *RoomHandler) Leave(c *gin.Context) {
	userID, roomID, ok := roomCurrentUserAndRoomID(c)
	if !ok {
		return
	}

	result, err := h.rooms.Leave(userID, roomID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	if result != nil && result.Left {
		h.hub.Broadcast(roomID, realtime.Event{
			Type:   "member.left",
			RoomID: roomID,
			Data: gin.H{
				"user_id":         userID,
				"current_members": result.CurrentMemberSize,
			},
		})

		if result.OwnerChanged {
			h.hub.Broadcast(roomID, realtime.Event{
				Type:   "room.owner_changed",
				RoomID: roomID,
				Data: gin.H{
					"owner_id": result.NewOwnerUserID,
					"members":  result.RemainingMembers,
				},
			})
		}
	}

	response.OK(c, nil)
}

func (h *RoomHandler) UpdateMicStatus(c *gin.Context) {
	userID, roomID, ok := roomCurrentUserAndRoomID(c)
	if !ok {
		return
	}

	var req updateMicRequest
	if !bindRoomJSON(c, &req) {
		return
	}

	member, err := h.rooms.UpdateMicStatus(userID, roomID, req.MicStatus)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	h.hub.Broadcast(roomID, realtime.Event{
		Type:   "member.mic_updated",
		RoomID: roomID,
		Data:   gin.H{"member": member},
	})

	response.OK(c, gin.H{"member": member})
}

func roomCurrentUser(c *gin.Context) (uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "未登录")
		return 0, false
	}
	return userID, true
}

func roomCurrentUserAndRoomID(c *gin.Context) (uint64, uint64, bool) {
	userID, ok := roomCurrentUser(c)
	if !ok {
		return 0, 0, false
	}

	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return 0, 0, false
	}

	return userID, roomID, true
}

func bindRoomJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, 500, "参数错误")
		return false
	}
	return true
}

func parseUintParam(c *gin.Context, name string) (uint64, bool) {
	value, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || value == 0 {
		return 0, false
	}
	return value, true
}

func parseIntDefault(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseUintDefault(value string, fallback uint64) uint64 {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
