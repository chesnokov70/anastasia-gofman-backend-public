package handler

// import (
// 	"anastasia_gofman_backend/internal/delivery/http/dto"
// 	"anastasia_gofman_backend/internal/service"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type ArticleHandler struct {
// 	articleService *service.ArticleService
// }

// func NewArticleHandler(articleService *service.ArticleService) *ArticleHandler {
// 	return &ArticleHandler{articleService: articleService}
// }

// func (h *ArticleHandler) GetAllArticles(c *gin.Context) {
// 	articles, err := h.articleService.GetAllArticles()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, articles)
// }

// func (h *ArticleHandler) GetArticleByID(c *gin.Context) {
// 	id := c.Param("id")
// 	article, err := h.articleService.GetArticleByID(id)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	c.JSON(http.StatusOK, article)
// }

// func (h *ArticleHandler) CreateArticle(c *gin.Context) {
// 	var article dto.CreateArticleDTO
// 	if err := c.ShouldBindJSON(&article); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// }
