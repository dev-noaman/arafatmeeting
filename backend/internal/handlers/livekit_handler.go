package handlers

import (
	"fmt"
	"strings"

	"mini-meeting/internal/config"
	"mini-meeting/internal/handlers/dto"
	"mini-meeting/internal/services"
	"mini-meeting/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type LiveKitHandler struct {
	livekitService    services.LiveKitServiceInterface
	summarizerService services.SummarizerServiceInterface
	config            *config.Config
}

func NewLiveKitHandler(
	livekitService services.LiveKitServiceInterface,
	summarizerService services.SummarizerServiceInterface,
	cfg *config.Config,
) *LiveKitHandler {
	return &LiveKitHandler{
		livekitService:    livekitService,
		summarizerService: summarizerService,
		config:            cfg,
	}
}

// GenerateToken generates a LiveKit token for joining a meeting
func (h *LiveKitHandler) GenerateToken(c *fiber.Ctx) error {
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

	var req dto.GenerateTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var userName string
	var identity string
	var userRole string
	var metadata string

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

	token, err := h.livekitService.CreateJoinToken(
		req.MeetingCode,
		identity,
		userName,
		userRole,
		metadata,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	response := dto.GenerateTokenResponse{
		Token:    token,
		URL:      h.livekitService.GetURL(),
		RoomCode: req.MeetingCode,
		Identity: identity,
		UserName: userName,
	}

	return c.JSON(response)
}

// RemoveParticipant removes a participant from a meeting (admin only)
func (h *LiveKitHandler) RemoveParticipant(c *fiber.Ctx) error {
	_, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req dto.RemoveParticipantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.livekitService.RemoveParticipant(req.MeetingCode, req.ParticipantIdentity)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove participant",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListParticipants lists all participants in a meeting
func (h *LiveKitHandler) ListParticipants(c *fiber.Ctx) error {
	meetingCode := c.Query("meeting_code")
	if meetingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "meeting_code is required",
		})
	}

	participants, err := h.livekitService.ListParticipants(meetingCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list participants",
		})
	}

	participantInfos := make([]dto.ParticipantInfo, 0, len(participants))
	for _, p := range participants {
		participantInfos = append(participantInfos, dto.ParticipantInfo{
			Identity: p.Identity,
			Name:     p.Name,
			State:    p.State.String(),
			Metadata: p.Metadata,
			JoinedAt: p.JoinedAt,
		})
	}

	return c.JSON(dto.ListParticipantsResponse{
		Participants: participantInfos,
	})
}

// GetParticipantCount returns the number of participants in a meeting (public)
func (h *LiveKitHandler) GetParticipantCount(c *fiber.Ctx) error {
	meetingCode := c.Query("meeting_code")
	if meetingCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "meeting_code is required",
		})
	}

	participants, err := h.livekitService.ListParticipants(meetingCode)
	count := 0
	if err == nil {
		count = len(participants)
	}

	return c.JSON(fiber.Map{
		"count": count,
	})
}

// MuteParticipant mutes/unmutes a specific track for a participant (admin only)
func (h *LiveKitHandler) MuteParticipant(c *fiber.Ctx) error {
	_, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req dto.MuteParticipantRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.livekitService.MuteParticipantTrack(req.MeetingCode, req.ParticipantIdentity, req.TrackSid, req.Muted)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to mute participant",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// EndMeeting ends a meeting for all participants (admin only)
func (h *LiveKitHandler) EndMeeting(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req dto.EndMeetingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check for active summarizer session and stop it if running
	// Derive meeting ID from code — for now, skip the meeting lookup since
	// summarizer sessions track by meeting ID, not code
	// This will be handled via meeting_cache in a future iteration

	err := h.livekitService.DeleteRoom(req.MeetingCode)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to end meeting",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
