package handler

// import (
// 	"anastasia_gofman_backend/internal/delivery/http/dto"
// 	"anastasia_gofman_backend/internal/service"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type PressHandler struct {
// 	pressService *service.PressService
// }

// func NewPressHandler(pressService *service.PressService) *PressHandler {
// 	return &PressHandler{pressService: pressService}
// }

// func (h *PressHandler) GetAllPress(c *gin.Context) {
// 	press, err := h.pressService.GetAllPress()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, press)
// }

// func (h *PressHandler) GetPressByID(c *gin.Context) {
// 	id := c.Param("id")
// 	press, err := h.pressService.GetPressByID(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, press)
// }

// func (h *PressHandler) CreatePress(c *gin.Context) {
// 	var press dto.CreatePressDTO
// 	if err := c.ShouldBindJSON(&press); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	press, err := h.pressService.CreatePress(press)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, press)
// }
