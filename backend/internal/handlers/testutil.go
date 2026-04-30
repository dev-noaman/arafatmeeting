package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"mini-meeting/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func createTestApp(
	livekitHandler *LiveKitHandler,
	lobbyHandler *LobbyHandler,
	lobbyWSHandler *LobbyWSHandler,
	summarizerHandler *SummarizerHandler,
) *fiber.App {
	app := fiber.New()

	api := app.Group("/api/v1")

	if livekitHandler != nil {
		publicLK := api.Group("/livekit")
		publicLK.Post("/token", livekitHandler.GenerateToken)
		publicLK.Get("/participants/count", livekitHandler.GetParticipantCount)
	}

	if lobbyHandler != nil {
		publicLobby := api.Group("/lobby")
		publicLobby.Post("/request", lobbyHandler.RequestToJoin)
		publicLobby.Delete("/request", lobbyHandler.CancelRequest)
	}

	if livekitHandler != nil {
		cfg := livekitHandler.config
		protectedLK := api.Group("/livekit", middleware.AuthMiddleware(cfg))
		protectedLK.Get("/participants", livekitHandler.ListParticipants)
		protectedLK.Post("/remove-participant", livekitHandler.RemoveParticipant)
		protectedLK.Post("/mute-participant", livekitHandler.MuteParticipant)
		protectedLK.Post("/end-meeting", livekitHandler.EndMeeting)
	}

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return app
}

func createTestJWT(secret, userID, email, role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	tkn, _ := token.SignedString([]byte(secret))
	return tkn
}

// testRequestBuilder helps build requests for Fiber's app.Test()
type testRequest struct {
	method  string
	path    string
	body    io.Reader
	headers map[string]string
}

func (r *testRequest) toFiberRequest() *http.Request {
	req, _ := http.NewRequest(r.method, "http://test"+r.path, r.body)
	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	return req
}

func postJSON(method, path string, body interface{}, authHeader ...string) *testRequest {
	data, _ := json.Marshal(body)
	req := &testRequest{
		method:  "POST",
		path:    path,
		body:    bytes.NewReader(data),
		headers: map[string]string{"Content-Type": "application/json"},
	}
	if len(authHeader) > 0 {
		req.headers["Authorization"] = authHeader[0]
	}
	return req
}

func getJSON(path string, authHeader ...string) *testRequest {
	req := &testRequest{
		method:  "GET",
		path:    path,
		headers: map[string]string{},
	}
	if len(authHeader) > 0 {
		req.headers["Authorization"] = authHeader[0]
	}
	return req
}

func execRequest(app *fiber.App, req *testRequest) (*http.Response, error) {
	return app.Test(req.toFiberRequest())
}

func decodeJSON(resp *http.Response, v interface{}) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected status %d, got %d: %s", expected, resp.StatusCode, string(body))
	}
}
