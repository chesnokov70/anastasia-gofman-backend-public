package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TranslationHandler struct {
	translationService service.TranslationService
}

func NewTranslationHandler(translationService service.TranslationService) *TranslationHandler {
	return &TranslationHandler{
		translationService: translationService,
	}
}

// @Summary Перевести текст на указанные языки, используя жптиху. Кушает текст и массив языков в формате ISO 639-1, но наверное в другом формате языки тоже съест
// @Description Переводит текст на указанные языки с помощью OpenAI API
// @Tags Translation
// @Accept json
// @Produce json
// @Param request body dto.TranslationRequest true "Запрос на перевод"
// @Success 200 {object} dto.TranslationResponse "Переводы текста"
// @Failure 400 {object} map[string]interface{} "Неверный запрос"
// @Failure 500 {object} map[string]interface{} "Внутренняя ошибка сервера"
// @Router /api/translate [post]
func (h *TranslationHandler) TranslateText(c *gin.Context) {
	var request dto.TranslationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Валидация входных данных
	if request.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Text field is required and cannot be empty",
		})
		return
	}

	if len(request.Languages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Languages array is required and cannot be empty",
		})
		return
	}

	// Проверяем, что языки не пустые
	for i, lang := range request.Languages {
		if lang == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Language code cannot be empty",
				"index": i,
			})
			return
		}
	}

	// Вызываем сервис перевода
	translations, err := h.translationService.TranslateText(request.Text, request.Languages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to translate text",
			"details": err.Error(),
		})
		return
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, dto.TranslationResponse(translations))
}
