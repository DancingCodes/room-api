package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/middleware"
	"room-api/internal/response"
	"room-api/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

type registerRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	EmailCode string `json:"email_code"`
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type updateMeRequest struct {
	Nickname string `json:"nickname"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 500, "invalid params")
		return
	}
	if req.EmailCode == "" {
		response.Error(c, 500, "invalid email code")
		return
	}

	result, err := h.users.Register(req.Username, req.Email, req.Password, req.Nickname, req.AvatarURL)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 500, "invalid params")
		return
	}

	result, err := h.users.Login(req.Username, req.Password)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}

	user, err := h.users.Me(userID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"user": user})
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}

	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 500, "invalid params")
		return
	}

	user, err := h.users.UpdateNickname(userID, req.Nickname)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"user": user})
}
