package handler

import (
	"github.com/gin-gonic/gin"

	"room-api/internal/response"
	"room-api/internal/service"
)

type AdminAppVersionHandler struct {
	versions *service.AppVersionService
}

func NewAdminAppVersionHandler(versions *service.AppVersionService) *AdminAppVersionHandler {
	return &AdminAppVersionHandler{versions: versions}
}

func (h *AdminAppVersionHandler) List(c *gin.Context) {
	versions, err := h.versions.ListAdmin()
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, gin.H{"list": versions})
}

func (h *AdminAppVersionHandler) Create(c *gin.Context) {
	var input service.SaveAppVersionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 500, "参数错误")
		return
	}
	version, err := h.versions.CreateAdmin(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, version)
}

func (h *AdminAppVersionHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return
	}
	var input service.SaveAppVersionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, 500, "参数错误")
		return
	}
	version, err := h.versions.UpdateAdmin(id, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, version)
}

func (h *AdminAppVersionHandler) Publish(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return
	}
	version, err := h.versions.PublishAdmin(id)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, version)
}

func (h *AdminAppVersionHandler) Unpublish(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		response.Error(c, 500, "参数错误")
		return
	}
	if err := h.versions.UnpublishAdmin(id); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}
