package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type WelcomeHandler struct {
}

func (h *WelcomeHandler) Welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Welcome to the Art Gallery API"})
}

func NewWelcomeHandler() *WelcomeHandler {
	return &WelcomeHandler{}
}
