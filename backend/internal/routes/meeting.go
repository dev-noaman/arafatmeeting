package routes

import (
	"mini-meeting/internal/config"
	"mini-meeting/internal/handlers"
	"mini-meeting/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func setupSummarizerRoutes(
	api fiber.Router,
	summarizerHandler *handlers.SummarizerHandler,
	cfg *config.Config,
) {
	meetings := api.Group("/meetings", middleware.AuthMiddleware(cfg))
	meetings.Post("/:id/summarizer/start", summarizerHandler.StartSummarizer)
	meetings.Post("/:id/summarizer/stop", summarizerHandler.StopSummarizer)
	meetings.Get("/:id/summarizer/sessions", summarizerHandler.GetSessionsByMeetingID)

	sessions := api.Group("/sessions", middleware.AuthMiddleware(cfg))
	sessions.Get("/", summarizerHandler.GetSessions)
	sessions.Get("/:id", summarizerHandler.GetSession)
	sessions.Delete("/:id", summarizerHandler.DeleteSession)
}
