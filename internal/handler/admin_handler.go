package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/auth"
	"room-api/internal/response"
	"room-api/internal/service"
)

type AdminHandler struct {
	admins   *auth.AdminService
	rooms    *service.RoomService
	versions *service.AppVersionService
}

type adminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAdminHandler(admins *auth.AdminService, rooms *service.RoomService, versions *service.AppVersionService) *AdminHandler {
	return &AdminHandler{admins: admins, rooms: rooms, versions: versions}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req adminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 500, "参数错误")
		return
	}

	token, err := h.admins.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, 401, "账号或密码错误")
		return
	}
	response.OK(c, gin.H{"token": token})
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.rooms.Stats()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	recentRooms, err := h.rooms.AdminList(1, 5)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	version, err := h.versions.LatestAdmin()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{
		"active_rooms":    stats.ActiveRooms,
		"current_members": stats.CurrentMembers,
		"current_version": version,
		"recent_rooms":    recentRooms.List,
	})
}

func (h *AdminHandler) Rooms(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)
	result, err := h.rooms.AdminList(page, pageSize)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *AdminHandler) RoomDetail(c *gin.Context) {
	roomID, ok := parseUintParam(c, "room_id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return
	}
	result, err := h.rooms.AdminDetail(roomID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, result)
}
