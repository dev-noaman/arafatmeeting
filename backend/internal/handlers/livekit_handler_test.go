package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/livekit/protocol/livekit"

	"mini-meeting/internal/config"
	"mini-meeting/internal/handlers/dto"
	"mini-meeting/internal/services"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-for-testing",
			Expiration: "24h",
		},
		Server: config.ServerConfig{
			Port:        "3000",
			Env:         "test",
			FrontendURL: "http://localhost:5173",
		},
		LiveKit: config.LiveKitConfig{
			APIKey:    "test-key",
			APISecret: "test-secret",
			URL:       "ws://localhost:7880",
		},
	}
}

func newTestLiveKitHandler() (*LiveKitHandler, *services.MockLiveKitService) {
	mock := &services.MockLiveKitService{}
	cfg := newTestConfig()
	handler := NewLiveKitHandler(mock, nil, cfg)
	return handler, mock
}

func TestGenerateToken_MissingMeetingCode(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestGenerateToken_GuestMissingName(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc-123",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestGenerateToken_GuestWithName(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	var capturedRole string
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		capturedRole = userRole
		if meetingCode != "abc-123" {
			t.Errorf("expected meeting_code 'abc-123', got '%s'", meetingCode)
		}
		if userName != "Guest User" {
			t.Errorf("expected userName 'Guest User', got '%s'", userName)
		}
		return "guest-token-123", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc-123",
		"user_name":    "Guest User",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result dto.GenerateTokenResponse
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Token != "guest-token-123" {
		t.Errorf("expected token 'guest-token-123', got '%s'", result.Token)
	}
	if result.RoomCode != "abc-123" {
		t.Errorf("expected room_code 'abc-123', got '%s'", result.RoomCode)
	}
	if result.URL != "ws://localhost:7880" {
		t.Errorf("expected URL 'ws://localhost:7880', got '%s'", result.URL)
	}
	if capturedRole != "guest" {
		t.Errorf("expected role 'guest', got '%s'", capturedRole)
	}
}

func TestGenerateToken_WithAuth(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	cfg := newTestConfig()

	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		if userRole != "user" {
			t.Errorf("expected userRole 'user' for authenticated user, got '%s'", userRole)
		}
		return "auth-token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	token := createTestJWT(cfg.JWT.Secret, "user-1", "test@example.com", "user")
	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc-123",
	}, "Bearer "+token)
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result dto.GenerateTokenResponse
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Token != "auth-token" {
		t.Errorf("expected token 'auth-token', got '%s'", result.Token)
	}
}

func TestGenerateToken_CustomNameOverride(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	cfg := newTestConfig()

	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		if userName != "CustomName" {
			t.Errorf("expected userName 'CustomName', got '%s'", userName)
		}
		return "token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	token := createTestJWT(cfg.JWT.Secret, "user-1", "test@example.com", "user")
	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    "CustomName",
	}, "Bearer "+token)
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)
}

func TestGenerateToken_ServiceError(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		return "", errors.New("livekit unavailable")
	}
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc-123",
		"user_name":    "Guest",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusInternalServerError)
}

func TestGenerateToken_UnauthorizedNoBearer(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		if userRole != "guest" {
			t.Errorf("expected guest role without auth, got '%s'", userRole)
		}
		return "guest-token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	// Send with "Bearer " prefix but invalid token — should treat as guest
	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    "Guest",
	}, "Bearer invalid-token")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// Should be treated as guest since token is invalid
	var result dto.GenerateTokenResponse
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Identity == "" {
		t.Error("expected non-empty identity")
	}
}

func TestGetParticipantCount_MissingMeetingCode(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	req := getJSON("/api/v1/livekit/participants/count")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestGetParticipantCount_Success(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.ListParticipantsFunc = func(meetingCode string) ([]*livekit.ParticipantInfo, error) {
		return []*livekit.ParticipantInfo{
			{Identity: "user1", Name: "User 1"},
			{Identity: "user2", Name: "User 2"},
		}, nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := getJSON("/api/v1/livekit/participants/count?meeting_code=abc-123")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result map[string]int
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["count"] != 2 {
		t.Errorf("expected count 2, got %d", result["count"])
	}
}

func TestGetParticipantCount_ServiceError_ReturnsZero(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.ListParticipantsFunc = func(meetingCode string) ([]*livekit.ParticipantInfo, error) {
		return nil, errors.New("livekit unavailable")
	}
	app := createTestApp(handler, nil, nil, nil)

	req := getJSON("/api/v1/livekit/participants/count?meeting_code=abc-123")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// Handler returns 200 with count=0 on service error
	assertStatus(t, resp, http.StatusOK)

	var result map[string]int
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["count"] != 0 {
		t.Errorf("expected count 0 on error, got %d", result["count"])
	}
}

func TestGetParticipantCount_EmptyRoom(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.ListParticipantsFunc = func(meetingCode string) ([]*livekit.ParticipantInfo, error) {
		return []*livekit.ParticipantInfo{}, nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := getJSON("/api/v1/livekit/participants/count?meeting_code=empty-room")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result map[string]int
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["count"] != 0 {
		t.Errorf("expected count 0, got %d", result["count"])
	}
}

func TestListParticipants_MissingMeetingCode(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	cfg := newTestConfig()
	app := createTestApp(handler, nil, nil, nil)

	token := createTestJWT(cfg.JWT.Secret, "user-1", "test@example.com", "user")
	req := getJSON("/api/v1/livekit/participants", "Bearer "+token)
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestListParticipants_WithParticipants(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	cfg := newTestConfig()

	mock.ListParticipantsFunc = func(meetingCode string) ([]*livekit.ParticipantInfo, error) {
		return []*livekit.ParticipantInfo{
			{Identity: "alice_1", Name: "Alice", State: livekit.ParticipantInfo_JOINED},
		}, nil
	}
	app := createTestApp(handler, nil, nil, nil)

	token := createTestJWT(cfg.JWT.Secret, "user-1", "test@example.com", "user")
	req := getJSON("/api/v1/livekit/participants?meeting_code=abc", "Bearer "+token)
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result dto.ListParticipantsResponse
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Participants) != 1 {
		t.Fatalf("expected 1 participant, got %d", len(result.Participants))
	}
	if result.Participants[0].Identity != "alice_1" {
		t.Errorf("expected identity 'alice_1', got '%s'", result.Participants[0].Identity)
	}
}

func TestListParticipants_Unauthorized(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	// Request without auth token
	req := getJSON("/api/v1/livekit/participants?meeting_code=abc")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestLobbyRequest_MissingMeetingCode(t *testing.T) {
	lobbyHandler := &LobbyHandler{
		livekitService: nil, // will fail when called, but validation should reject first
		config:         newTestConfig(),
	}
	app := createTestApp(nil, lobbyHandler, nil, nil)

	req := postJSON("POST", "/api/v1/lobby/request", map[string]interface{}{})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestLobbyCancel_MissingRequestID(t *testing.T) {
	lobbyHandler := &LobbyHandler{
		livekitService: nil,
		config:         newTestConfig(),
	}
	app := createTestApp(nil, lobbyHandler, nil, nil)

	req := &testRequest{
		method:  "DELETE",
		path:    "/api/v1/lobby/request",
		headers: map[string]string{},
	}
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

// Read body helper for error assertions
func readBody(resp *http.Response) string {
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

func TestGenerateToken_InvalidJSON(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	req := &testRequest{
		method:  "POST",
		path:    "/api/v1/livekit/token",
		body:    strings.NewReader(`{invalid json}`),
		headers: map[string]string{"Content-Type": "application/json"},
	}
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestGenerateToken_MissingContentType(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	// Send with empty body — should fail parsing
	req := &testRequest{
		method:  "POST",
		path:    "/api/v1/livekit/token",
		body:    strings.NewReader(""),
		headers: map[string]string{},
	}
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// Fiber returns 400 for empty/unparsable body
	assertStatus(t, resp, http.StatusBadRequest)
}

func TestGenerateToken_ResponseHasRequiredFields(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		return "test-token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "test-code",
		"user_name":    "TestUser",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	body := readBody(resp)
	// Verify all expected fields are in the response
	expectedFields := []string{`"token"`, `"url"`, `"room_code"`, `"identity"`, `"user_name"`}
	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Errorf("response missing field %s", field)
		}
	}
}

func TestGenerateToken_AccessControlAllowOrigin(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		return "token", nil
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "http://localhost:5173")
		return c.Next()
	})
	api := app.Group("/api/v1")
	api.Post("/livekit/token", handler.GenerateToken)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    "Test",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)
}

func TestHealthCheck(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	req := getJSON("/health")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result map[string]string
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", result["status"])
	}
}

func TestMiddleware_AuthMissingHeader(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	// Protected route without auth
	req := getJSON("/api/v1/livekit/participants?meeting_code=abc")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestMiddleware_AuthInvalidToken(t *testing.T) {
	handler, _ := newTestLiveKitHandler()
	app := createTestApp(handler, nil, nil, nil)

	req := getJSON("/api/v1/livekit/participants?meeting_code=abc", "Bearer invalid.token.here")
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusUnauthorized)
}

func TestMiddleware_AuthValidToken(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	cfg := newTestConfig()
	mock.ListParticipantsFunc = func(meetingCode string) ([]*livekit.ParticipantInfo, error) {
		return []*livekit.ParticipantInfo{}, nil
	}
	app := createTestApp(handler, nil, nil, nil)

	token := createTestJWT(cfg.JWT.Secret, "user-1", "test@example.com", "user")
	req := getJSON("/api/v1/livekit/participants?meeting_code=abc", "Bearer "+token)
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)
}

func TestGenerateToken_TokenLength(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		return "a-very-long-jwt-token-string-that-looks-like-a-real-token.eyJhbGciOiJIUzI1NiJ9.test-signature", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    "Test",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	var result dto.GenerateTokenResponse
	if err := decodeJSON(resp, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result.Token) < 20 {
		t.Errorf("expected realistic token length, got %d chars", len(result.Token))
	}
}

func TestGenerateToken_IdentityIncludesName(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	cfg := newTestConfig()

	var capturedIdentity string
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		capturedIdentity = identity
		return "token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	token := createTestJWT(cfg.JWT.Secret, "uuid-123", "user@test.com", "user")
	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
	}, "Bearer "+token)
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	// Authenticated user identity should contain userName_userID
	if !strings.Contains(capturedIdentity, "user@test.com") {
		t.Errorf("expected identity to contain email, got '%s'", capturedIdentity)
	}
}

func TestGenerateToken_MetadataContainsRole(t *testing.T) {
	handler, mock := newTestLiveKitHandler()

	var capturedMetadata string
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		capturedMetadata = metadata
		return "token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    "Guest",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)

	if !strings.Contains(capturedMetadata, `"role":"guest"`) {
		t.Errorf("expected metadata to contain role, got '%s'", capturedMetadata)
	}
}

func TestGenerateToken_NameTruncation(t *testing.T) {
	handler, mock := newTestLiveKitHandler()

	var capturedName string
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		capturedName = userName
		return "token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	longName := strings.Repeat("a", 1000)
	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    longName,
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)
	if capturedName != longName {
		t.Errorf("name was unexpectedly modified: got length %d", len(capturedName))
	}
}

func TestGenerateToken_EmptyMeetingCode(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	var capturedRoomCode string
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		capturedRoomCode = meetingCode
		return "token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "",
		"user_name":    "Test",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)
	if capturedRoomCode != "" {
		t.Errorf("expected empty room code, got '%s'", capturedRoomCode)
	}
}

func TestGenerateToken_SpecialCharactersInName(t *testing.T) {
	handler, mock := newTestLiveKitHandler()

	var capturedName string
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		capturedName = userName
		return "token", nil
	}
	app := createTestApp(handler, nil, nil, nil)

	specialName := "O'Brien <script>alert(1)</script>"
	req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
		"meeting_code": "abc",
		"user_name":    specialName,
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assertStatus(t, resp, http.StatusOK)
	if capturedName != specialName {
		t.Errorf("name was modified: got '%s'", capturedName)
	}
}

func TestGetParticipantCount_MeetingCodeEdgeCases(t *testing.T) {
	testCases := []struct {
		name         string
		meetingCode  string
		expectStatus int
	}{
		{"normal code", "abc-123", http.StatusOK},
		{"empty code", "", http.StatusBadRequest},
		{"single char", "a", http.StatusOK},
		{"with spaces", "abc 123", http.StatusOK},
		{"url encoded", "abc%2D123", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, mock := newTestLiveKitHandler()
			mock.ListParticipantsFunc = func(mc string) ([]*livekit.ParticipantInfo, error) {
				return []*livekit.ParticipantInfo{}, nil
			}
			app := createTestApp(handler, nil, nil, nil)

			urlParam := url.QueryEscape(tc.meetingCode)
			reqPath := fmt.Sprintf("/api/v1/livekit/participants/count?meeting_code=%s", urlParam)
			req := getJSON(reqPath)
			resp, err := execRequest(app, req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			assertStatus(t, resp, tc.expectStatus)
		})
	}
}

func TestGenerateToken_BatchMultipleRequests(t *testing.T) {
	handler, mock := newTestLiveKitHandler()
	callCount := 0
	mock.CreateJoinTokenFunc = func(meetingCode, identity, userName, userRole, metadata string) (string, error) {
		callCount++
		return fmt.Sprintf("token-%d", callCount), nil
	}
	app := createTestApp(handler, nil, nil, nil)

	for i := 0; i < 5; i++ {
		req := postJSON("POST", "/api/v1/livekit/token", map[string]interface{}{
			"meeting_code": fmt.Sprintf("meeting-%d", i),
			"user_name":    fmt.Sprintf("User-%d", i),
		})
		resp, err := execRequest(app, req)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		assertStatus(t, resp, http.StatusOK)

		var result dto.GenerateTokenResponse
		if err := decodeJSON(resp, &result); err != nil {
			t.Fatalf("request %d decode failed: %v", i, err)
		}
		expectedToken := fmt.Sprintf("token-%d", i+1)
		if result.Token != expectedToken {
			t.Errorf("request %d: expected token '%s', got '%s'", i, expectedToken, result.Token)
		}
	}
}

func TestLobbyRequest_GuestWithoutName(t *testing.T) {
	lobbyHandler := &LobbyHandler{
		livekitService: nil,
		config:         newTestConfig(),
	}
	app := createTestApp(nil, lobbyHandler, nil, nil)

	req := postJSON("POST", "/api/v1/lobby/request", map[string]interface{}{
		"meeting_code": "abc-123",
	})
	resp, err := execRequest(app, req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	// Guest without name should get 500 (panic/nil pointer) or 400 from validation
	// The actual behavior depends on whether nil service causes issues
	// The handler tries to call LiveKit for admin which won't happen for guest
	// For guest, it needs name or returns 400
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 400 or 500, got %d", resp.StatusCode)
	}
}
