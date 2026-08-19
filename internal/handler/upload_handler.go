package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/response"
	"room-api/internal/service"
)

type UploadHandler struct {
	uploads *service.UploadService
}

func NewUploadHandler(uploads *service.UploadService) *UploadHandler {
	return &UploadHandler{uploads: uploads}
}

func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 500, "参数错误")
		return
	}

	imageURL, err := h.uploads.UploadAvatar(file)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.OK(c, gin.H{"url": imageURL})
}
