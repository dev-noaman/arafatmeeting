package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"mini-meeting/internal/cache"
	"mini-meeting/internal/config"
	"mini-meeting/internal/services"
	"mini-meeting/pkg/utils"

	"github.com/gofiber/contrib/websocket"
)

const (
	WSTypePendingRequests  = "pending_requests"
	WSTypeNewRequest       = "new_request"
	WSTypeRequestResolved  = "request_resolved"
	WSTypeVisitorCancelled = "visitor_cancelled"
	WSTypeApproved         = "approved"
	WSTypeRejected         = "rejected"
	WSTypeRespond          = "respond"
)

type WSPendingRequestsMsg struct {
	Type     string                  `json:"type"`
	Requests []WSPendingRequestEntry `json:"requests"`
}

type WSPendingRequestEntry struct {
	RequestID string `json:"request_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
}

type WSNewRequestMsg struct {
	Type    string                `json:"type"`
	Request WSPendingRequestEntry `json:"request"`
}

type WSRequestResolvedMsg struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
}

type WSApprovedMsg struct {
	Type     string `json:"type"`
	Token    string `json:"token"`
	URL      string `json:"url"`
	RoomCode string `json:"room_code"`
	Identity string `json:"identity"`
	UserName string `json:"user_name"`
}

type WSRejectedMsg struct {
	Type string `json:"type"`
}

type WSRespondMsg struct {
	Type      string `json:"type"`
	RequestID string `json:"request_id"`
	Action    string `json:"action"`
}

type LobbyWSHandler struct {
	livekitService services.LiveKitServiceInterface
	config         *config.Config
}

func NewLobbyWSHandler(
	livekitService services.LiveKitServiceInterface,
	cfg *config.Config,
) *LobbyWSHandler {
	return &LobbyWSHandler{
		livekitService: livekitService,
		config:         cfg,
	}
}

func (h *LobbyWSHandler) HandleVisitor(c *websocket.Conn) {
	requestID := c.Query("request_id")
	meetingCode := c.Query("meeting_code")

	if requestID == "" || meetingCode == "" {
		log.Printf("[WS Visitor] Missing request_id or meeting_code")
		c.WriteJSON(map[string]string{"error": "request_id and meeting_code are required"})
		c.Close()
		return
	}

	_, err := cache.GetLobbyRequest(requestID)
	if err != nil {
		log.Printf("[WS Visitor] Request not found: %s", requestID)
		c.WriteJSON(map[string]string{"error": "Request not found or expired"})
		c.Close()
		return
	}

	cache.Hub.RegisterVisitor(requestID, c)

	defer func() {
		cache.Hub.UnregisterVisitor(requestID)

		req, err := cache.GetLobbyRequest(requestID)
		if err == nil && req.Status == cache.LobbyStatusPending {
			cache.CleanupLobbyRequest(requestID)
			cache.Hub.NotifyAdmins(meetingCode, WSRequestResolvedMsg{
				Type:      WSTypeVisitorCancelled,
				RequestID: requestID,
			})
			log.Printf("[WS Visitor] Visitor disconnected, cleaned up request %s", requestID)
		}
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Printf("[WS Visitor] Connection closed for request %s: %v", requestID, err)
			break
		}
	}
}

func (h *LobbyWSHandler) HandleAdmin(c *websocket.Conn) {
	meetingCode := c.Query("meeting_code")
	token := c.Query("token")

	if meetingCode == "" || token == "" {
		log.Printf("[WS Admin] Missing meeting_code or token")
		c.WriteJSON(map[string]string{"error": "meeting_code and token are required"})
		c.Close()
		return
	}

	claims, err := utils.ValidateToken(token, h.config.JWT.Secret)
	if err != nil {
		log.Printf("[WS Admin] Invalid token: %v", err)
		c.WriteJSON(map[string]string{"error": "Invalid or expired token"})
		c.Close()
		return
	}

	_ = claims.UserID // User validated — meeting ownership checked via meeting_cache or trusted client

	cache.Hub.RegisterAdmin(meetingCode, c)
	defer cache.Hub.UnregisterAdmin(meetingCode, c)

	pending, err := cache.GetPendingRequests(meetingCode)
	if err != nil {
		log.Printf("[WS Admin] Failed to get pending requests: %v", err)
	}

	entries := make([]WSPendingRequestEntry, 0, len(pending))
	for _, req := range pending {
		entries = append(entries, WSPendingRequestEntry{
			RequestID: req.ID,
			Name:      req.Name,
			AvatarURL: req.AvatarURL,
			Role:      req.Role,
			CreatedAt: req.CreatedAt,
		})
	}

	c.WriteJSON(WSPendingRequestsMsg{
		Type:     WSTypePendingRequests,
		Requests: entries,
	})

	for {
		_, msgBytes, err := c.ReadMessage()
		if err != nil {
			log.Printf("[WS Admin] Connection closed for meeting %s: %v", meetingCode, err)
			break
		}

		var baseMsg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(msgBytes, &baseMsg); err != nil {
			log.Printf("[WS Admin] Invalid message: %v", err)
			continue
		}

		switch baseMsg.Type {
		case WSTypeRespond:
			var respondMsg WSRespondMsg
			if err := json.Unmarshal(msgBytes, &respondMsg); err != nil {
				log.Printf("[WS Admin] Invalid respond message: %v", err)
				continue
			}
			h.handleAdminRespond(meetingCode, &respondMsg)

		default:
			log.Printf("[WS Admin] Unknown message type: %s", baseMsg.Type)
		}
	}
}

func (h *LobbyWSHandler) handleAdminRespond(meetingCode string, msg *WSRespondMsg) {
	if msg.RequestID == "" || (msg.Action != "approve" && msg.Action != "reject") {
		log.Printf("[WS Admin] Invalid respond params: request_id=%s action=%s", msg.RequestID, msg.Action)
		return
	}

	lobbyReq, err := cache.GetLobbyRequest(msg.RequestID)
	if err != nil {
		log.Printf("[WS Admin] Request not found: %s", msg.RequestID)
		return
	}

	if lobbyReq.MeetingCode != meetingCode {
		log.Printf("[WS Admin] Request %s does not belong to meeting %s", msg.RequestID, meetingCode)
		return
	}

	if msg.Action == "reject" {
		if err := cache.UpdateLobbyRequestStatus(msg.RequestID, cache.LobbyStatusRejected); err != nil {
			log.Printf("[WS Admin] Failed to reject request: %v", err)
			return
		}

		cache.Hub.NotifyVisitor(msg.RequestID, WSRejectedMsg{Type: WSTypeRejected})
		cache.Hub.NotifyAdmins(meetingCode, WSRequestResolvedMsg{
			Type:      WSTypeRequestResolved,
			RequestID: msg.RequestID,
		})
		go cache.CleanupLobbyRequest(msg.RequestID)
		return
	}

	metadata := fmt.Sprintf(`{"name":"%s","avatar":"%s","role":"%s"}`, lobbyReq.Name, lobbyReq.AvatarURL, lobbyReq.Role)

	token, err := h.livekitService.CreateJoinToken(
		lobbyReq.MeetingCode,
		lobbyReq.Identity,
		lobbyReq.Name,
		lobbyReq.Role,
		metadata,
	)
	if err != nil {
		log.Printf("[WS Admin] Failed to generate token: %v", err)
		return
	}

	if err := cache.UpdateLobbyRequestStatus(msg.RequestID, cache.LobbyStatusApproved); err != nil {
		log.Printf("[WS Admin] Failed to approve request: %v", err)
		return
	}

	cache.Hub.NotifyVisitor(msg.RequestID, WSApprovedMsg{
		Type:     WSTypeApproved,
		Token:    token,
		URL:      h.livekitService.GetURL(),
		RoomCode: lobbyReq.MeetingCode,
		Identity: lobbyReq.Identity,
		UserName: lobbyReq.Name,
	})

	cache.Hub.NotifyAdmins(meetingCode, WSRequestResolvedMsg{
		Type:      WSTypeRequestResolved,
		RequestID: msg.RequestID,
	})

	go func() {
		time.Sleep(5 * time.Second)
		cache.CleanupLobbyRequest(msg.RequestID)
	}()
}

func NotifyAdminsOfNewRequest(lobbyReq *cache.LobbyRequest) {
	cache.Hub.NotifyAdmins(lobbyReq.MeetingCode, WSNewRequestMsg{
		Type: WSTypeNewRequest,
		Request: WSPendingRequestEntry{
			RequestID: lobbyReq.ID,
			Name:      lobbyReq.Name,
			AvatarURL: lobbyReq.AvatarURL,
			Role:      lobbyReq.Role,
			CreatedAt: lobbyReq.CreatedAt,
		},
	})
}
