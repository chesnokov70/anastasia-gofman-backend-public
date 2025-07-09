package handler

import (
	"anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"net/http"
	"strconv"

	"anastasia_gofman_backend/internal/entity"

	"github.com/gin-gonic/gin"
)

type CollectionHandler struct {
	collectionService service.ArtCollectionService
	artService        service.ArtService
}

func NewCollectionHandler(collectionService service.ArtCollectionService, artService service.ArtService) *CollectionHandler {
	return &CollectionHandler{collectionService: collectionService, artService: artService}
}

// @Summary Get all collections
// @Description Получает все коллекции. Сортировка: NEW - сначала новые мальчишки, OLD - старые мужчины, BIG - где больше всего картинок, SMALL - где меньше всего, AVALIBLE - где есть картинки на продажу, при этом новое - раньше
// @Param with_arts query bool false "With arts" default(true)
// @Param sorting query string false "Sorting type" Enums(NEW, OLD, BIG, SMALL, AVALIBLE) default(NEW)
// @Accept json
// @Produce json
// @Tags Collections
// @Success 200 {object} []dto.CollectionResponseDTO
// @Router /api/collections [get]
func (h *CollectionHandler) GetAllCollections(c *gin.Context) {
	sorting := c.DefaultQuery("sorting", "NEW")
	with_arts := c.DefaultQuery("with_arts", "true")
	with_arts_bool, err := strconv.ParseBool(with_arts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid with_arts parameter"})
		return
	}

	collections, err := h.collectionService.GetAllCollections(sorting, with_arts_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	response := gin.H{
		"data": dto.ToCollectionResponseDTOs(collections, base_url),
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Get collection by ID
// @Description Получает коллекцию по ID
// @Tags Collections
// @Param id path int true "Collection ID"
// @Param with_arts query bool false "With arts" default(true)
// @Accept json
// @Produce json
// @Success 200 {object} dto.CollectionResponseDTO
// @Router /api/collections/{id} [get]
func (h *CollectionHandler) GetCollectionByID(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}
	with_arts := c.DefaultQuery("with_arts", "true")
	with_arts_bool, err := strconv.ParseBool(with_arts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid with_arts parameter"})
		return
	}
	collection, err := h.collectionService.GetCollectionByID(uint(id_uint), with_arts_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToCollectionResponseDTO(collection, base_url))
}

// @Summary Create collection
// @Description Создает коллекцию
// @Tags Collections
// @Accept json
// @Produce json
// @Param collection body dto.CreateCollectionDTO true "Collection"
// @Success 201 {object} dto.CollectionResponseDTO
// @Router /api/collections [post]
func (h *CollectionHandler) CreateCollection(c *gin.Context) {
	var collection dto.CreateCollectionDTO
	if err := c.ShouldBindJSON(&collection); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	collection_entity, err := h.collectionService.CreateCollection(collection.ToEntity())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusCreated, dto.ToCollectionResponseDTO(collection_entity, base_url))
}

// @Summary Full update collection
// @Description Полное обновление коллекции
// @Tags Collections
// @Accept json
// @Produce json
// @Param id path int true "Collection ID"
// @Param collection body dto.UpdateCollectionDTO true "Collection"
// @Success 200 {object} dto.CollectionResponseDTO
// @Router /api/collections/{id} [put]
func (h *CollectionHandler) FullUpdateCollection(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}
	var collection dto.UpdateCollectionDTO
	if err := c.ShouldBindJSON(&collection); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	collection_entity, err := h.collectionService.FullUpdateCollection(collection.ToEntity(uint(id_uint)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if collection.ArtsIds != nil {
		collection_entity, err = h.collectionService.AddArtsToCollection(collection_entity.ID, collection.ArtsIds, collection.RemoveNotInIds || false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToCollectionResponseDTO(collection_entity, base_url))
}

// @Summary Partial update collection
// @Description Частичное обновление коллекции
// @Tags Collections
// @Param id path int true "Collection ID"
// @Param collection body dto.UpdateCollectionDTO true "Collection"
// @Success 200 {object} dto.CollectionResponseDTO
// @Router /api/collections/{id} [patch]
func (h *CollectionHandler) PartialUpdateCollection(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	processedData := make(map[string]interface{})

	if nameData, exists := updateData["name"]; exists {
		if nameMap, ok := nameData.(map[string]interface{}); ok {
			translatedName := entity.TranslatedText{}
			if en, exists := nameMap["en"]; exists {
				if enStr, ok := en.(string); ok {
					translatedName.EN = enStr
				}
			}
			if ru, exists := nameMap["ru"]; exists {
				if ruStr, ok := ru.(string); ok {
					translatedName.RU = ruStr
				}
			}
			if es, exists := nameMap["es"]; exists {
				if esStr, ok := es.(string); ok {
					translatedName.ES = esStr
				}
			}
			processedData["name"] = translatedName
		}
	}

	if descData, exists := updateData["description"]; exists {
		if descMap, ok := descData.(map[string]interface{}); ok {
			translatedDesc := entity.TranslatedText{}
			if en, exists := descMap["en"]; exists {
				if enStr, ok := en.(string); ok {
					translatedDesc.EN = enStr
				}
			}
			if ru, exists := descMap["ru"]; exists {
				if ruStr, ok := ru.(string); ok {
					translatedDesc.RU = ruStr
				}
			}
			if es, exists := descMap["es"]; exists {
				if esStr, ok := es.(string); ok {
					translatedDesc.ES = esStr
				}
			}
			processedData["description"] = translatedDesc
		}
	}

	collection_entity, err := h.collectionService.PartialUpdateCollection(uint(id_uint), processedData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if artsIdsData, exists := updateData["arts_ids"]; exists {
		if artsIdsSlice, ok := artsIdsData.([]interface{}); ok {
			var arts_ids []uint
			for _, artId := range artsIdsSlice {
				if artIdFloat, ok := artId.(float64); ok {
					arts_ids = append(arts_ids, uint(artIdFloat))
				}
			}

			remove_not_in_ids := false
			if removeFlag, exists := updateData["remove_not_in_ids"]; exists {
				if removeBool, ok := removeFlag.(bool); ok {
					remove_not_in_ids = removeBool
				}
			}

			collection_entity, err = h.collectionService.AddArtsToCollection(uint(id_uint), arts_ids, remove_not_in_ids)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToCollectionResponseDTO(collection_entity, base_url))
}

// @Summary Delete collection
// @Description Удаляет коллекцию, если delete_action = DELETE_ART, то удаляет все арты и картинки каскадом(причем в отдельном потоке - не ждет пока удалятся все кртинки и страйп продукты - иншаллах сработает, впервые потрогал тут асинхронность - прикольно очень), если delete_action = SAVE_ART, то сохраняет все картинки просто очищает им поле коллекция и теперь они - обычный арт
// @Tags Collections
// @Param delete_action query string false "Delete action" Enums(DELETE_ART, SAVE_ART) default(DELETE_ART)
// @Param id path int true "Collection ID"
// @Success 200 {object} map[string]string
// @Router /api/collections/{id} [delete]
func (h *CollectionHandler) DeleteCollection(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}
	delete_action := c.DefaultQuery("delete_action", "DELETE_ART")
	if delete_action != "DELETE_ART" && delete_action != "SAVE_ART" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delete_action parameter"})
		return
	}
	err = h.collectionService.DeleteCollection(uint(id_uint), delete_action, h.artService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Collection deleted successfully"})
}

// @Summary Add arts to collection
// @Description Добавляет арты в коллекцию
// @Tags Collections
// @Param id path int true "Collection ID"
// @Param arts body []uint true "Arts"
// @Param remove_not_in_ids query bool false "Remove not in ids" default(false)
// @Success 200 {object} map[string]string
// @Router /api/collections/{id}/arts [post]
func (h *CollectionHandler) AddArtsToCollection(c *gin.Context) {
	id := c.Param("id")
	id_uint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id parameter"})
		return
	}
	var arts []uint
	if err := c.ShouldBindJSON(&arts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	remove_not_in_ids := c.DefaultQuery("remove_not_in_ids", "false")
	remove_not_in_ids_bool, err := strconv.ParseBool(remove_not_in_ids)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid remove_not_in_ids parameter"})
		return
	}
	collection_result, err := h.collectionService.AddArtsToCollection(uint(id_uint), arts, remove_not_in_ids_bool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	base_url := config.GetBaseURL()
	c.JSON(http.StatusOK, dto.ToCollectionResponseDTO(collection_result, base_url))
}
