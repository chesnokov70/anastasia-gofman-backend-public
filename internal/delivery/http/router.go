package http

import (
	"anastasia_gofman_backend/internal/delivery/http/handler"
	"anastasia_gofman_backend/internal/delivery/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/webstradev/gin-pagination/v2/pkg/pagination"
)

func NewRouter(authorHandler *handler.AuthorHandler, artHandler *handler.ArtHandler, welcomeHandler *handler.WelcomeHandler, eventHandler *handler.EventHandler, paymentHandler *handler.PaymentHandler, collectionHandler *handler.CollectionHandler, pressHandler *handler.PressHandler, articleHandler *handler.ArticleHandler, translationHandler *handler.TranslationHandler) *gin.Engine {
	// func NewRouter(authorHandler *handler.AuthorHandler, artHandler *handler.ArtHandler, welcomeHandler *handler.WelcomeHandler, eventHandler *handler.EventHandler, paymentHandler *handler.PaymentHandler, collectionHandler *handler.CollectionHandler) *gin.Engine {
	router := gin.Default()
	router.RedirectTrailingSlash = false // Disable automatic redirect for trailing slashes

	router.Use(middleware.DetailedLogging())

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true                                                       // Allows all origins
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}  // Specify methods
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"} // Specify headers
	config.AllowCredentials = true
	router.Use(cors.New(config))

	paginationMiddleware := pagination.New(
		pagination.WithMinPageSize(2),
		pagination.WithMaxPageSize(100),
	)

	welcom := router.Group("/")
	welcom.GET("/", welcomeHandler.Welcome)
	api := router.Group("/api")
	{
		authors := api.Group("/authors")
		{
			authors.GET("", paginationMiddleware, authorHandler.GetAllAuthors)
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
			authors.DELETE("/:id/main_photo", authorHandler.DeleteMainPhotoFromAuthor)

		}
		arts := api.Group("/arts")
		{
			arts.GET("", paginationMiddleware, artHandler.GetAllArts)
			arts.GET("/:id", artHandler.GetArtByID)

			arts.POST("", artHandler.CreateArt)
			arts.POST("/with_photos", artHandler.CreateArtWithPhotos)
			arts.PUT("/:id", artHandler.FullUpdateArt)
			arts.PATCH("/:id", artHandler.PartialUpdateArt)

			arts.DELETE("/:id", artHandler.DeleteArt)

			arts.GET("/:id/main_photo", artHandler.GetMainPhoto)
			arts.POST("/:id/main_photo", artHandler.AddMainPhotoToArt)
			arts.DELETE("/:id/main_photo", artHandler.DeleteMainPhotoFromArt)

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
			events.DELETE("/:id/main_photo", eventHandler.DeleteMainPhotoFromEvent)
			events.POST("/:id/photos", eventHandler.AddPhotosToEvent)
			events.PATCH("/:id/photos", eventHandler.PatchEventPhotos)
		}
		payments := api.Group("/payments")
		{
			payments.POST("/products", paymentHandler.CreateProduct)
			payments.GET("/products/:id", paymentHandler.GetProduct)
			payments.PUT("/products/:id", paymentHandler.UpdateProduct)
			payments.DELETE("/products/:id", paymentHandler.DeleteProduct)

			payments.GET("/customers", paymentHandler.GetCustomers)

			payments.GET("/payment-intents", paymentHandler.GetPayments)
			payments.POST("/payment-links", paymentHandler.CreatePaymentLink)
			payments.POST("/refunds", paymentHandler.CreateRefund)

			payments.GET("/balance", paymentHandler.GetBalance)
		}
		collections := api.Group("/collections")
		{
			collections.GET("", collectionHandler.GetAllCollections)
			collections.GET("/:id", collectionHandler.GetCollectionByID)
			collections.POST("", collectionHandler.CreateCollection)
			collections.PUT("/:id", collectionHandler.FullUpdateCollection)
			collections.PATCH("/:id", collectionHandler.PartialUpdateCollection)
			collections.DELETE("/:id", collectionHandler.DeleteCollection)
		}
		press := api.Group("/press")
		{
			press.GET("", pressHandler.GetAllPress)
			press.GET("/:id", pressHandler.GetPressByID)
			press.POST("", pressHandler.CreatePress)
			press.POST("/with_photos", pressHandler.CreatePressWithPhotos)
			press.DELETE("/:id", pressHandler.DeletePress)
			press.PUT("/:id", pressHandler.FullUpdatePress)
			press.PATCH("/:id", pressHandler.PartialUpdatePress)
			press.POST("/:id/main_photo", pressHandler.AddMainPhotoToPress)
			press.DELETE("/:id/main_photo", pressHandler.DeleteMainPhotoFromPress)
			press.POST("/:id/photos", pressHandler.AddPhotosToPress)
			press.PATCH("/:id/photos", pressHandler.PatchPressPhotos)
			press.GET("/:id/main_photo", pressHandler.GetMainPhoto)
		}
		articles := api.Group("/articles")
		{
			articles.GET("", articleHandler.GetAllArticles)
			articles.GET("/:id", articleHandler.GetArticleByID)
			articles.POST("", articleHandler.CreateArticle)
			articles.POST("/with_photos", articleHandler.CreateArticleWithPhotos)
			articles.DELETE("/:id", articleHandler.DeleteArticle)
			articles.PUT("/:id", articleHandler.FullUpdateArticle)
			articles.PATCH("/:id", articleHandler.PartialUpdateArticle)
			articles.POST("/:id/main_photo", articleHandler.AddMainPhotoToArticle)
			articles.DELETE("/:id/main_photo", articleHandler.DeleteMainPhotoFromArticle)
			articles.POST("/:id/photos", articleHandler.AddPhotosToArticle)
			articles.PATCH("/:id/photos", articleHandler.PatchArticlePhotos)
			articles.GET("/:id/main_photo", articleHandler.GetMainPhoto)

		}

		api.POST("/translate", translationHandler.TranslateText)
	}

	return router
}

// dissabilyty
// financial planing
