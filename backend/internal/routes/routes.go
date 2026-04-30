package routes

import (
	"mini-meeting/internal/config"
	"mini-meeting/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	livekitHandler *handlers.LiveKitHandler,
	lobbyHandler *handlers.LobbyHandler,
	lobbyWSHandler *handlers.LobbyWSHandler,
	summarizerHandler *handlers.SummarizerHandler,
	cfg *config.Config,
) {
	api := app.Group("/api/v1")

	setupSummarizerRoutes(api, summarizerHandler, cfg)
	setupLiveKitRoutes(api, livekitHandler, cfg)
	setupLobbyRoutes(app, api, lobbyHandler, lobbyWSHandler)
}
