package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/middleware"
	"room-api/internal/response"
	"room-api/internal/service"
)

type UploadHandler struct {
	uploads *service.UploadService
	users   *service.UserService
}

func NewUploadHandler(uploads *service.UploadService, users *service.UserService) *UploadHandler {
	return &UploadHandler{uploads: uploads, users: users}
}

func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 500, "invalid params")
		return
	}

	avatarURL, err := h.uploads.UploadAvatar(file)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"avatar_url": avatarURL})
}

func (h *UploadHandler) UpdateMyAvatar(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 500, "invalid params")
		return
	}

	avatarURL, err := h.uploads.UploadAvatar(file)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	user, err := h.users.UpdateAvatar(userID, avatarURL)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"user": user})
}
