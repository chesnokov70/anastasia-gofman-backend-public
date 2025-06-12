package handler

import (
	"net/http"

	// "anastasia_gofman_backend/internal/repository/postgres"

	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/entity"
	"anastasia_gofman_backend/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuthorHandler struct {
	authorService service.AuthorService
}

// @Summary Get all authors
// @Description Получаем всех авторов
// @Tags Authors
// @Accept json
// @Produce json
// @Success 200 {array} []entity.Author
// @Router /api/authors [get]
func (h *AuthorHandler) GetAllAuthors(c *gin.Context) {
	authors, err := h.authorService.GetAllAuthors()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, authors)
}

// @Summary Get author by ID
// @Description Получаем автора по ID
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Router /api/authors/{id} [get]
func (h *AuthorHandler) GetAuthorByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	author, err := h.authorService.GetAuthorByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, author)
}

// @Summary Create author
// @Description С кайфом создаем автора, все поля необязательные кроме имени, ну и в mail пихать нужно mail нормальный, но можно ничего не пихать
// @Tags Authors
// @Accept json
// @Produce json
// @Param data body dto.CreateAuthorDTO true "Author data"
// @Success 201 {object} entity.Author
// @Failure 400 {object} map[string]string
// @Router /api/authors [post]
func (h *AuthorHandler) CreateAuthor(c *gin.Context) {
	var authorDTO dto.CreateAuthorDTO
	if err := c.ShouldBindJSON(&authorDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	author := entity.Author{
		Name: authorDTO.Name.ToEntity(),
		Bio:  authorDTO.Bio.ToEntity(),
		Contact: entity.ContactInfo{
			Email: authorDTO.Contact.Email,
			Phone: authorDTO.Contact.Phone,
			Links: entity.SocialLink{
				Instagram: authorDTO.Contact.Links.Instagram,
				Telegram:  authorDTO.Contact.Links.Telegram,
				Vkontakte: authorDTO.Contact.Links.Vkontakte,
				Facebook:  authorDTO.Contact.Links.Facebook,
				Twitter:   authorDTO.Contact.Links.Twitter,
				Youtube:   authorDTO.Contact.Links.Youtube,
				Linkedin:  authorDTO.Contact.Links.Linkedin,
				Whatsapp:  authorDTO.Contact.Links.Whatsapp,
				Pinterest: authorDTO.Contact.Links.Pinterest,
				Behance:   authorDTO.Contact.Links.Behance,
			},
		},
	}

	author, err := h.authorService.CreateAuthor(author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, author)
}

// @Summary Delete author
// @Description Удаляем автора
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Success 200 {object} map[string]string
func (h *AuthorHandler) DeleteAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	err = h.authorService.DeleteAuthor(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "author deleted"})
}

// @Summary Partial update author
// @Description Обновляем автора частично, все поля необязательные, не передал = осталось прежним
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Param data body dto.UpdateAuthorDTO true "Update data"
// @Success 200 {object} entity.Author
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/authors/:id [patch]
func (h *AuthorHandler) PartialUpdateAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	author, err := h.authorService.PartialUpdateAuthor(uint(id), updateData)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, author)
}

// @Summary Update author
// @Description Обновляем автора полностью, все поля необязательные, но сотрется вся инфа, которая была раньше и которая не передана
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path int true "Author ID"
// @Param data body dto.UpdateAuthorDTO true "Author data"
// @Success 200 {object} entity.Author
// @Failure 400 {object} map[string]string
// @Router /api/authors/{id} [put]
func (h *AuthorHandler) FullUpdateAuthor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var authorDTO dto.UpdateAuthorDTO
	if err := c.ShouldBindJSON(&authorDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	author := entity.Author{
		ID:   uint(id),
		Name: authorDTO.Name.ToEntity(),
		Bio:  authorDTO.Bio.ToEntity(),
		Contact: entity.ContactInfo{
			Email: authorDTO.Contact.Email,
			Phone: authorDTO.Contact.Phone,
			Links: entity.SocialLink{
				Instagram: authorDTO.Contact.Links.Instagram,
				Telegram:  authorDTO.Contact.Links.Telegram,
				Vkontakte: authorDTO.Contact.Links.Vkontakte,
				Facebook:  authorDTO.Contact.Links.Facebook,
				Twitter:   authorDTO.Contact.Links.Twitter,
				Youtube:   authorDTO.Contact.Links.Youtube,
				Linkedin:  authorDTO.Contact.Links.Linkedin,
				Whatsapp:  authorDTO.Contact.Links.Whatsapp,
				Pinterest: authorDTO.Contact.Links.Pinterest,
				Behance:   authorDTO.Contact.Links.Behance,
			},
		},
	}
	author, err = h.authorService.FullUpdateAuthor(author)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, author)
}

func NewAuthorHandler(authorService service.AuthorService) *AuthorHandler {
	return &AuthorHandler{authorService: authorService}
}
