package handlers

import (
	"fmt"
	"strings"
	"time"

	"mini-meeting/internal/cache"
	"mini-meeting/internal/config"
	"mini-meeting/internal/handlers/dto"
	"mini-meeting/internal/services"
	"mini-meeting/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LobbyHandler struct {
	livekitService services.LiveKitServiceInterface
	config         *config.Config
}

func NewLobbyHandler(
	livekitService services.LiveKitServiceInterface,
	cfg *config.Config,
) *LobbyHandler {
	return &LobbyHandler{
		livekitService: livekitService,
		config:         cfg,
	}
}

// RequestToJoin handles a user's request to join a meeting.
func (h *LobbyHandler) RequestToJoin(c *fiber.Ctx) error {
	var req dto.LobbyJoinRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.MeetingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "meeting_code is required",
		})
	}

	var userID string
	var userEmail string
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			claims, err := utils.ValidateToken(parts[1], h.config.JWT.Secret)
			if err == nil {
				userID = claims.UserID
				userEmail = claims.Email
			}
		}
	}

	var userName, identity, userRole, metadata string

	if userID != "" {
		userRole = "user"
		userName = userEmail
		if req.UserName != "" {
			userName = req.UserName
		}

		identity = fmt.Sprintf("%s_%s", userName, userID)
		metadata = fmt.Sprintf(`{"name":"%s","avatar":"","role":"%s"}`, userName, userRole)
	} else {
		if req.UserName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "User name is required for guest users",
			})
		}

		userName = req.UserName
		userRole = "guest"
		identity = fmt.Sprintf("%s_%d", userName, c.Context().ConnID())
		metadata = fmt.Sprintf(`{"name":"%s","avatar":"","role":"%s"}`, userName, userRole)
	}

	// If admin (meeting creator), auto-approve and return token immediately
	if userRole == "admin" {
		token, err := h.livekitService.CreateJoinToken(
			req.MeetingCode, identity, userName, userRole, metadata,
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to generate token",
			})
		}

		return c.JSON(dto.LobbyJoinResponse{
			RequestID: "",
			Status:    "auto_approved",
			Token:     token,
			URL:       h.livekitService.GetURL(),
			RoomCode:  req.MeetingCode,
			Identity:  identity,
			UserName:  userName,
		})
	}

	// Non-admin: create a pending lobby request
	requestID := uuid.New().String()

	lobbyReq := &cache.LobbyRequest{
		ID:          requestID,
		MeetingCode: req.MeetingCode,
		UserID:      userID,
		Name:        userName,
		Identity:    identity,
		Role:        userRole,
		Status:      cache.LobbyStatusPending,
		CreatedAt:   time.Now().Unix(),
	}

	if err := cache.StoreLobbyRequest(lobbyReq); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create lobby request",
		})
	}

	NotifyAdminsOfNewRequest(lobbyReq)

	return c.JSON(dto.LobbyJoinResponse{
		RequestID: requestID,
		Status:    "pending",
	})
}

// CancelRequest lets a waiting user cancel their own join request.
func (h *LobbyHandler) CancelRequest(c *fiber.Ctx) error {
	requestID := c.Query("request_id")
	if requestID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "request_id is required",
		})
	}

	req, _ := cache.GetLobbyRequest(requestID)

	if err := cache.CleanupLobbyRequest(requestID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel request",
		})
	}

	if req != nil {
		cache.Hub.NotifyAdmins(req.MeetingCode, map[string]string{
			"type":       "visitor_cancelled",
			"request_id": requestID,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Request cancelled",
	})
}
