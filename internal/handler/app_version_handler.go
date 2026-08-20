package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/response"
	"room-api/internal/service"
)

type AppVersionHandler struct {
	versions *service.AppVersionService
}

func NewAppVersionHandler(versions *service.AppVersionService) *AppVersionHandler {
	return &AppVersionHandler{versions: versions}
}

func (h *AppVersionHandler) Latest(c *gin.Context) {
	version, err := h.versions.Latest()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, version)
}
