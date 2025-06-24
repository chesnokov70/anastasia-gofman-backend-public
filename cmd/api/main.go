package main

import (
	// _ "anastasia_gofman_backend/docs"
	_ "anastasia_gofman_backend/cmd/docs"
	"anastasia_gofman_backend/internal/delivery/http"
	_ "anastasia_gofman_backend/internal/delivery/http/dto"
	"anastasia_gofman_backend/internal/delivery/http/handler"

	"anastasia_gofman_backend/internal/entity"

	"anastasia_gofman_backend/internal/repository/postgres"
	"anastasia_gofman_backend/internal/service"
	"anastasia_gofman_backend/pkg/config"
	"anastasia_gofman_backend/pkg/database"
	pkgService "anastasia_gofman_backend/pkg/service"
	"log"
	stdhttp "net/http"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Art Gallery API
// @version 1.0
// @description Backend для галереи anastasii_gofman
// @BasePath /
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// имиграция
	err = db.AutoMigrate(
		&entity.Author{},
		&entity.Photo{},
		&entity.Art{},
		&entity.Event{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := db.Exec(`
		UPDATE arts
		SET author_id = NULL
		WHERE author_id IS NOT NULL AND author_id NOT IN (SELECT id FROM authors)
	`).Error; err != nil {
		log.Printf("Warning: Failed to clean up orphan arts: %v", err)
	}

	photoRepository := postgres.NewPhotoRepository(db)
	artRepository := postgres.NewArtRepository(db)

	// Инициализация Stripe сервиса
	stripeService := pkgService.NewStripeService(
		cfg.Stripe.SecretKey, // test key
		cfg.Stripe.PublicKey, // prod key (пока используем тот же)
		false,                // isLive - по умолчанию тестовый режим
	)

	artService := service.NewArtService(artRepository, photoRepository, stripeService)
	artHandler := handler.NewArtHandler(artService)

	authorRepository := postgres.NewAuthorRepository(db)
	authorRepository.CreateDefaultAuthor()
	authorService := service.NewAuthorService(authorRepository, photoRepository, artRepository)
	authorHandler := handler.NewAuthorHandler(authorService)

	eventRepository := postgres.NewEventRepository(db)
	eventService := service.NewEventService(eventRepository, photoRepository)
	eventHandler := handler.NewEventHandler(eventService)

	welcomeHandler := handler.NewWelcomeHandler()

	paymentHandler := handler.NewPaymentHandler(stripeService)

	router := http.NewRouter(authorHandler, artHandler, welcomeHandler, eventHandler, paymentHandler)
	router.Static("/uploads", "./uploads")
	// / Swagger route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.DefaultModelsExpandDepth(2)))

	// ginSwagger.WrapHandler(swaggerfiles.Handler,
	// 	ginSwagger.URL("http://localhost:8080/swagger/doc.json"),
	// )

	// Redirect /docs to /swagger/index.html
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(stdhttp.StatusMovedPermanently, "/swagger/index.html")
	})

	log.Printf("Starting server on port %s in %s mode", cfg.Server.Port, cfg.Server.Mode)
	if cfg.Stripe.SecretKey == "sk_test_" {
		log.Printf("Warning: Using default Stripe test key. Set STRIPE_SECRET_KEY environment variable.")
	} else {
		log.Printf("Stripe service initialized in production mode")
	}

	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// package main

// import (
// 	"context"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/gin-gonic/gin"

// 	"anastasia_gofman_backend/internal/delivery/http/handler"
// 	"anastasia_gofman_backend/internal/delivery/http/router"
// 	"anastasia_gofman_backend/internal/repository/postgres"
// 	"anastasia_gofman_backend/internal/service"
// 	"anastasia_gofman_backend/pkg/config"
// 	"anastasia_gofman_backend/pkg/database"
// )

// func main() {
// 	// 1. Загрузка конфигурации
// 	cfg, err := config.Load("configs/config.yaml")
// 	if err != nil {
// 		log.Fatalf("Failed to load config: %v", err)
// 	}

// 	// 2. Подключение к базе данных
// 	db, err := database.NewPostgresConnection(cfg.Database)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to database: %v", err)
// 	}

// 	// 3. Инициализация репозиториев
// 	authorRepo := postgres.NewAuthorRepository(db)
// 	artRepo := postgres.NewArtRepository(db)

// 	// 4. Инициализация сервисов
// 	authorService := service.NewAuthorService(authorRepo)
// 	artService := service.NewArtService(artRepo)

// 	// 5. Инициализация обработчиков
// 	authorHandler := handler.NewAuthorHandler(authorService)
// 	artHandler := handler.NewArtHandler(artService)

// 	// 6. Настройка Gin
// 	gin.SetMode(cfg.Server.Mode)
// 	r := gin.Default()

// 	// 7. Настройка маршрутов
// 	router.SetupRoutes(r, authorHandler, artHandler)

// 	// 8. Создание HTTP-сервера
// 	srv := &http.Server{
// 		Addr:    ":" + cfg.Server.Port,
// 		Handler: r,
// 	}

// 	// 9. Запуск сервера в горутине
// 	go func() {
// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatalf("Failed to start server: %v", err)
// 		}
// 	}()

// 	// 10. Graceful shutdown
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
// 	<-quit

// 	log.Println("Shutting down server...")

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	if err := srv.Shutdown(ctx); err != nil {
// 		log.Fatalf("Server forced to shutdown: %v", err)
// 	}

// 	log.Println("Server exited properly")
// }
