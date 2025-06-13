package http

import (
	"anastasia_gofman_backend/internal/delivery/http/handler"
	"anastasia_gofman_backend/internal/delivery/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(authorHandler *handler.AuthorHandler, artHandler *handler.ArtHandler, welcomeHandler *handler.WelcomeHandler, eventHandler *handler.EventHandler) *gin.Engine {
	router := gin.Default()
	router.RedirectTrailingSlash = false // Disable automatic redirect for trailing slashes

	router.Use(middleware.DetailedLogging())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true                                                       // Allows all origins
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}  // Specify methods
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"} // Specify headers
	config.AllowCredentials = true
	router.Use(cors.New(config))

	welcom := router.Group("/")
	welcom.GET("/", welcomeHandler.Welcome)
	api := router.Group("/api")
	{
		authors := api.Group("/authors")
		{
			authors.GET("", authorHandler.GetAllAuthors)
			authors.GET("/:id", authorHandler.GetAuthorByID)
			authors.POST("", authorHandler.CreateAuthor)
			authors.PUT("/:id", authorHandler.FullUpdateAuthor)
			authors.PATCH("/:id", authorHandler.PartialUpdateAuthor)
			authors.DELETE("/:id", authorHandler.DeleteAuthor)
			authors.POST("/with_photos", authorHandler.CreateAuthorWithPhotos)
			authors.POST("/:id/main_photo", authorHandler.AddMainPhotoToAuthor)
			// authors.POST("/:id/preview_photo", authorHandler.AddPreviewPhotoToAuthor)
			authors.POST("/:id/photos", authorHandler.AddPhotosToAuthor)
			authors.PATCH("/:id/photos", authorHandler.PatchAuthorPhotos)
			authors.GET("/:id/with_arts", authorHandler.GetAuthorWithArts)

		}
		arts := api.Group("/arts")
		{
			arts.GET("", artHandler.GetAllArts)
			arts.GET("/:id", artHandler.GetArtByID)
			arts.POST("", artHandler.CreateArt)
			arts.POST("/with_photos", artHandler.CreateArtWithPhotos)
			arts.PUT("/:id", artHandler.FullUpdateArt)
			arts.PATCH("/:id", artHandler.PartialUpdateArt)
			arts.DELETE("/:id", artHandler.DeleteArt)

			arts.GET("/:id/main_photo", artHandler.GetMainPhoto)
			arts.POST("/:id/main_photo", artHandler.AddMainPhotoToArt)

			arts.POST("/:id/photos", artHandler.AddPhotosToArt)
			arts.PATCH("/:id/photos", artHandler.PatchArtPhotos)

			arts.POST("/:id/author/:author_id", artHandler.AddAuthorToArt)
		}
		events := api.Group("/events")
		{
			events.GET("", eventHandler.GetAllEvents)
			events.GET("/:id", eventHandler.GetEventByID)
			events.POST("", eventHandler.CreateEvent)
			events.POST("/with_photos", eventHandler.CreateEventWithPhotos)
			events.PUT("/:id", eventHandler.FullUpdateEvent)
			events.PATCH("/:id", eventHandler.PartialUpdateEvent)
			events.DELETE("/:id", eventHandler.DeleteEvent)

			events.GET("/:id/main_photo", eventHandler.GetMainPhoto)
			events.POST("/:id/main_photo", eventHandler.AddMainPhotoToEvent)

			events.POST("/:id/photos", eventHandler.AddPhotosToEvent)
			events.PATCH("/:id/photos", eventHandler.PatchEventPhotos)
		}
	}

	return router
}

// dissabilyty
// financial planing
