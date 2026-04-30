package handlers

import (
	"mini-meeting/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SummarizerHandler struct {
	service services.SummarizerServiceInterface
}

func NewSummarizerHandler(service services.SummarizerServiceInterface) *SummarizerHandler {
	return &SummarizerHandler{service: service}
}

// StartSummarizer starts the summarizer for a meeting
// POST /api/v1/meetings/:id/summarizer/start
func (h *SummarizerHandler) StartSummarizer(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	meetingID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid meeting ID",
		})
	}

	userEmail, _ := c.Locals("email").(string)

	session, err := h.service.StartSummarizer(uint(meetingID), userID, userEmail)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "meeting not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized: only meeting creator can start summarizer" {
			statusCode = fiber.StatusForbidden
		} else if err.Error() == "summarizer already running for this meeting" {
			statusCode = fiber.StatusConflict
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Summarizer started successfully",
		"data": fiber.Map{
			"session_id": session.ID,
			"meeting_id": session.MeetingID,
			"status":     session.Status,
			"started_at": session.StartedAt,
		},
	})
}

// StopSummarizer stops the summarizer for a meeting
// POST /api/v1/meetings/:id/summarizer/stop
func (h *SummarizerHandler) StopSummarizer(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	meetingID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid meeting ID",
		})
	}

	session, err := h.service.GetActiveSession(uint(meetingID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	totalChunks, err := h.service.StopSummarizer(session.ID, userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		errMsg := err.Error()
		if errMsg == "session not found" || errMsg == "meeting not found" {
			statusCode = fiber.StatusNotFound
		} else if errMsg == "unauthorized: only meeting creator can stop summarizer" {
			statusCode = fiber.StatusForbidden
		} else if len(errMsg) >= 28 && errMsg[:28] == "session is not active" {
			statusCode = fiber.StatusConflict
		}

		return c.Status(statusCode).JSON(fiber.Map{
			"error": errMsg,
		})
	}

	session, _ = h.service.GetSessionByID(session.ID)

	return c.JSON(fiber.Map{
		"message": "Summarizer stopped successfully",
		"data": fiber.Map{
			"session_id":   session.ID,
			"meeting_id":   session.MeetingID,
			"status":       session.Status,
			"started_at":   session.StartedAt,
			"ended_at":     session.EndedAt,
			"total_chunks": totalChunks,
		},
	})
}

// GetSessions retrieves a paginated list of sessions for the authenticated user
// GET /api/v1/sessions?page=1&page_size=10
func (h *SummarizerHandler) GetSessions(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 10)

	response, err := h.service.GetSessions(userID, page, pageSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// GetSession retrieves a specific session by ID
// GET /api/v1/sessions/:id
func (h *SummarizerHandler) GetSession(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	session, err := h.service.GetSession(uint(id), userID)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if err.Error() == "session not found" {
			statusCode = fiber.StatusNotFound
		} else if err.Error() == "unauthorized: session does not belong to user" {
			statusCode = fiber.StatusForbidden
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": session,
	})
}

// DeleteSession deletes a session by ID
// DELETE /api/v1/sessions/:id
func (h *SummarizerHandler) DeleteSession(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid session ID",
		})
	}

	if err := h.service.DeleteSession(uint(id), userID); err != nil {
		errMsg := err.Error()
		statusCode := fiber.StatusInternalServerError
		if len(errMsg) >= 17 && errMsg[:17] == "session not found" {
			statusCode = fiber.StatusNotFound
		} else if errMsg == "unauthorized: session does not belong to user" {
			statusCode = fiber.StatusForbidden
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"error": errMsg,
		})
	}

	return c.JSON(fiber.Map{
		"message": "Session deleted successfully",
	})
}

// GetSessionsByMeetingID retrieves all sessions for a specific meeting
// GET /api/v1/meetings/:id/summarizer/sessions
func (h *SummarizerHandler) GetSessionsByMeetingID(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	meetingID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid meeting ID",
		})
	}

	sessions, err := h.service.GetSessionsByMeetingID(uint(meetingID), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": sessions,
	})
}
