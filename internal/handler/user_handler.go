package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/middleware"
	"room-api/internal/response"
	"room-api/internal/service"
)

type UserHandler struct {
	users *service.UserService
	codes *service.EmailCodeService
}

func NewUserHandler(users *service.UserService, codes *service.EmailCodeService) *UserHandler {
	return &UserHandler{users: users, codes: codes}
}

type emailLoginRequest struct {
	Email     string `json:"email"`
	EmailCode string `json:"email_code"`
}

type updateMeRequest struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type emailCodeRequest struct {
	Email string `json:"email"`
}

func (h *UserHandler) SendEmailCode(c *gin.Context) {
	var req emailCodeRequest
	if !bindUserJSON(c, &req) {
		return
	}

	if err := h.codes.SendEmailCode(req.Email); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, nil)
}

func (h *UserHandler) EmailLogin(c *gin.Context) {
	var req emailLoginRequest
	if !bindUserJSON(c, &req) {
		return
	}

	result, err := h.users.EmailLogin(req.Email, req.EmailCode)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := userCurrentUser(c)
	if !ok {
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
	userID, ok := userCurrentUser(c)
	if !ok {
		return
	}

	var req updateMeRequest
	if !bindUserJSON(c, &req) {
		return
	}

	user, err := h.users.UpdateProfile(userID, req.Nickname, req.AvatarURL)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"user": user})
}

func userCurrentUser(c *gin.Context) (uint64, bool) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "未登录")
		return 0, false
	}
	return userID, true
}

func bindUserJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, 500, "参数错误")
		return false
	}
	return true
}
