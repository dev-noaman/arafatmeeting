package main

import (
	"fmt"
	"log"
	"time"

	"mini-meeting/internal/config"
	"mini-meeting/internal/database"
	"mini-meeting/internal/handlers"
	"mini-meeting/internal/repositories"
	"mini-meeting/internal/routes"
	"mini-meeting/internal/services"
	"mini-meeting/internal/workers"
	"mini-meeting/pkg/cache"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := database.Connect(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := cache.Connect(&cfg.Redis); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cache.Close()

	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	// Build CORS origins from env var + localhost fallbacks
	origins := cfg.Server.FrontendURL
	if origins == "" {
		origins = "http://localhost:5173,http://127.0.0.1:5173"
	} else {
		origins += ",http://localhost:5173,http://127.0.0.1:5173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length,Content-Type",
	}))

	// Repositories (only summarizer-related)
	sessionRepo := repositories.NewSummarizerSessionRepository(database.GetDB())
	chunkRepo := repositories.NewAudioChunkRepository(database.GetDB())
	transcriptRepo := repositories.NewTranscriptRepository(database.GetDB())

	// Services
	livekitService := services.NewLiveKitService(cfg)
	openRouterService := services.NewOpenRouterService(cfg)
	emailService := services.NewEmailService(cfg)

	summarizationService := services.NewSummarizationService(sessionRepo, openRouterService, emailService, cfg)
	normalizationService := services.NewNormalizationService(sessionRepo, transcriptRepo, summarizationService)
	transcriptionService := services.NewTranscriptionService(sessionRepo, chunkRepo, transcriptRepo, normalizationService, cfg)
	summarizerService := services.NewSummarizerService(sessionRepo, chunkRepo, transcriptRepo, livekitService, transcriptionService, cfg)

	// Handlers
	livekitHandler := handlers.NewLiveKitHandler(livekitService, summarizerService, cfg)
	lobbyHandler := handlers.NewLobbyHandler(livekitService, cfg)
	lobbyWSHandler := handlers.NewLobbyWSHandler(livekitService, cfg)
	summarizerHandler := handlers.NewSummarizerHandler(summarizerService)

	// Background workers
	transcriptionWorker := workers.NewTranscriptionWorker(sessionRepo, transcriptionService, 60*time.Minute, 15*time.Minute)
	go transcriptionWorker.Start()
	normalizationWorker := workers.NewNormalizationWorker(sessionRepo, normalizationService, 60*time.Minute, 15*time.Minute)
	go normalizationWorker.Start()
	summarizationWorker := workers.NewSummarizationWorker(sessionRepo, summarizationService, 60*time.Minute, 15*time.Minute)
	go summarizationWorker.Start()

	routes.SetupRoutes(app, livekitHandler, lobbyHandler, lobbyWSHandler, summarizerHandler, cfg)

	app.Get("/api/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on port %s...", cfg.Server.Port)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
