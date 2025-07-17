package handler

import (
	"anastasia_gofman_backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	service service.SettingsService
}

func NewSettingsHandler(service service.SettingsService) *SettingsHandler {
	return &SettingsHandler{service: service}
}

// @Summary Get settings
// @Description Получает настройки
// @Accept json
// @Produce json
// @Tags Settings
// @Success 200 {object} entity.Settings
// @Router /api/settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// @Summary Update settings
// @Description Обновляет настройки
// @Accept json
// @Produce json
// @Tags Settings
// @Param settings body entity.Settings true "Settings"
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.UpdateSettings(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// @Summary Overwrite settings
// @Description Перезаписывает настройки
// @Accept json
// @Produce json
// @Tags Settings
// @Param settings body entity.Settings true "Settings"
func (h *SettingsHandler) OverwriteSettings(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.service.OverwriteSettings(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}
