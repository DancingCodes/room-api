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

type registerRequest struct {
	Account   string `json:"account"`
	Email     string `json:"email"`
	EmailCode string `json:"email_code"`
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type updateMeRequest struct {
	Nickname string `json:"nickname"`
}

type emailCodeRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	EmailCode   string `json:"email_code"`
	NewPassword string `json:"new_password"`
}

func (h *UserHandler) SendRegisterCode(c *gin.Context) {
	var req emailCodeRequest
	if !bindUserJSON(c, &req) {
		return
	}

	if err := h.codes.SendRegisterCode(req.Email); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, nil)
}

func (h *UserHandler) Register(c *gin.Context) {
	var req registerRequest
	if !bindUserJSON(c, &req) {
		return
	}
	if req.EmailCode == "" {
		response.Error(c, 500, "验证码错误")
		return
	}

	result, err := h.users.Register(req.Account, req.Email, req.EmailCode, req.Password, req.Nickname, req.AvatarURL)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req loginRequest
	if !bindUserJSON(c, &req) {
		return
	}

	result, err := h.users.Login(req.Account, req.Password)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, result)
}

func (h *UserHandler) SendPasswordResetCode(c *gin.Context) {
	var req emailCodeRequest
	if !bindUserJSON(c, &req) {
		return
	}

	if err := h.codes.SendResetPasswordCode(req.Email); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, nil)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if !bindUserJSON(c, &req) {
		return
	}

	if err := h.users.ResetPassword(req.Email, req.EmailCode, req.NewPassword); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, nil)
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

	user, err := h.users.UpdateNickname(userID, req.Nickname)
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
