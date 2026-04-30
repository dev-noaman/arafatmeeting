package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mini-meeting/internal/config"
	"net/http"
	"time"
)

const brevoAPIURL = "https://api.brevo.com/v3/smtp/email"

// EmailService handles sending transactional emails via Brevo REST API.
type EmailService struct {
	cfg        *config.BrevoConfig
	httpClient *http.Client
}

// NewEmailService creates a new EmailService instance.
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		cfg: &cfg.Brevo,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// brevoEmailRequest mirrors the Brevo transactional email API payload.
type brevoEmailRequest struct {
	Sender      brevoContact   `json:"sender"`
	To          []brevoContact `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
	TextContent string         `json:"textContent"`
}

type brevoContact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// SendVerificationEmail sends email verification code to user
func (s *EmailService) SendVerificationEmail(toEmail, toName, verificationCode string) error {
	if s.cfg.APIKey == "" {
		return fmt.Errorf("Brevo API key is not configured")
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 32px;">
  <div style="max-width: 560px; margin: 0 auto; background: #ffffff; border-radius: 8px; padding: 32px;">
    <h2 style="color: #1a1a2e; margin-top: 0;">Verify your email 📧</h2>
    <p style="color: #444; line-height: 1.6;">
      Hi %s,<br><br>
      Welcome to Mini Meeting! Please use the following verification code to activate your account:
    </p>
    <div style="text-align: center; margin: 24px 0;">
      <span style="display: inline-block; padding: 16px 32px; background: #6c63ff; color: #ffffff;
                   font-size: 24px; font-weight: bold; border-radius: 6px; letter-spacing: 4px;">
        %s
      </span>
    </div>
    <p style="color: #888; font-size: 14px;">
      This code will expire in 24 hours.<br>
      If you didn't request this, please ignore this email.
    </p>
  </div>
</body>
</html>`, toName, verificationCode)

	textBody := fmt.Sprintf("Hi %s,\n\nYour verification code is: %s\n\nThis code will expire in 24 hours.\n\nMini Meeting", toName, verificationCode)

	payload := brevoEmailRequest{
		Sender: brevoContact{
			Email: s.cfg.SenderEmail,
			Name:  s.cfg.SenderName,
		},
		To: []brevoContact{
			{Email: toEmail, Name: toName},
		},
		Subject:     "Verify your email - Mini Meeting",
		HTMLContent: htmlBody,
		TextContent: textBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Brevo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Brevo API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("EmailService: verification email sent to %s\n", toEmail)
	return nil
}

// SendPasswordResetEmail sends password reset code to user
func (s *EmailService) SendPasswordResetEmail(toEmail, toName, resetCode string, frontendURL string) error {
	if s.cfg.APIKey == "" {
		return fmt.Errorf("Brevo API key is not configured")
	}

	resetURL := fmt.Sprintf("%s/reset-password?code=%s", frontendURL, resetCode)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 32px;">
  <div style="max-width: 560px; margin: 0 auto; background: #ffffff; border-radius: 8px; padding: 32px;">
    <h2 style="color: #1a1a2e; margin-top: 0;">Reset your password 🔒</h2>
    <p style="color: #444; line-height: 1.6;">
      Hi %s,<br><br>
      You requested to reset your password. Use the code below or click the button:
    </p>
    <div style="text-align: center; margin: 24px 0;">
      <span style="display: inline-block; padding: 16px 32px; background: #e74c3c; color: #ffffff;
                   font-size: 24px; font-weight: bold; border-radius: 6px; letter-spacing: 4px;">
        %s
      </span>
    </div>
    <div style="text-align: center; margin: 24px 0;">
      <a href="%s" style="display: inline-block; padding: 12px 24px; background: #6c63ff; color: #ffffff;
                          text-decoration: none; border-radius: 6px; font-weight: bold;">
        Reset Password
      </a>
    </div>
    <p style="color: #888; font-size: 14px;">
      This code will expire in 1 hour.<br>
      If you didn't request this, please ignore this email.
    </p>
  </div>
</body>
</html>`, toName, resetCode, resetURL)

	textBody := fmt.Sprintf("Hi %s,\n\nYour password reset code is: %s\n\nOr visit: %s\n\nThis code will expire in 1 hour.\n\nMini Meeting", toName, resetCode, resetURL)

	payload := brevoEmailRequest{
		Sender: brevoContact{
			Email: s.cfg.SenderEmail,
			Name:  s.cfg.SenderName,
		},
		To: []brevoContact{
			{Email: toEmail, Name: toName},
		},
		Subject:     "Password reset - Mini Meeting",
		HTMLContent: htmlBody,
		TextContent: textBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Brevo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Brevo API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("EmailService: password reset email sent to %s\n", toEmail)
	return nil
}

// SendMeetingInvitationEmail sends meeting invitation to attendees
func (s *EmailService) SendMeetingInvitationEmail(attendeeEmail, attendeeName, inviterName, meetingCode, meetingTitle string, frontendURL string) error {
	if s.cfg.APIKey == "" {
		return fmt.Errorf("Brevo API key is not configured")
	}

	meetingURL := fmt.Sprintf("%s/%s", frontendURL, meetingCode)

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 32px;">
  <div style="max-width: 600px; margin: 0 auto; background: #ffffff; border-radius: 8px; padding: 32px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);">
    <div style="text-align: center; margin-bottom: 24px;">
      <h1 style="color: #6c63ff; margin: 0; font-size: 28px;">📅 Meeting Invitation</h1>
    </div>

    <p style="color: #444; line-height: 1.6; font-size: 16px;">
      Hi %s,
    </p>

    <p style="color: #444; line-height: 1.6; font-size: 16px;">
      <strong>%s</strong> has invited you to join a meeting.
    </p>

    <div style="background: #f8f9fa; border-left: 4px solid #6c63ff; padding: 16px; margin: 24px 0; border-radius: 4px;">
      <p style="margin: 0; color: #666; font-size: 14px; font-weight: bold;">MEETING DETAILS</p>
      <p style="margin: 8px 0 0 0; color: #1a1a2e; font-size: 18px; font-weight: bold;">%s</p>
      <p style="margin: 8px 0 0 0; color: #888; font-size: 14px;">Meeting Code: <strong style="color: #1a1a2e;">%s</strong></p>
    </div>

    <div style="text-align: center; margin: 32px 0;">
      <a href="%s" style="display: inline-block; padding: 14px 32px; background: #6c63ff; color: #ffffff;
                          text-decoration: none; border-radius: 8px; font-weight: bold; font-size: 16px;">
        🎥 Join Meeting
      </a>
    </div>

    <p style="color: #888; font-size: 14px; line-height: 1.6;">
      Or copy this link: <a href="%s" style="color: #6c63ff;">%s</a>
    </p>

    <hr style="border: none; border-top: 1px solid #eee; margin: 32px 0;">

    <p style="color: #aaa; font-size: 12px; text-align: center;">
      You received this invitation from %s via Mini Meeting.
    </p>
  </div>
</body>
</html>`, attendeeName, inviterName, meetingTitle, meetingCode, meetingURL, meetingURL, meetingURL, inviterName)

	textBody := fmt.Sprintf(`Hi %s,

%s has invited you to join a meeting.

MEETING DETAILS
Meeting Code: %s

Join here: %s

You received this invitation from %s via Mini Meeting.`, attendeeName, inviterName, meetingCode, meetingURL, inviterName)

	payload := brevoEmailRequest{
		Sender: brevoContact{
			Email: s.cfg.SenderEmail,
			Name:  s.cfg.SenderName,
		},
		To: []brevoContact{
			{Email: attendeeEmail, Name: attendeeName},
		},
		Subject:     fmt.Sprintf("📅 %s invited you to a meeting", inviterName),
		HTMLContent: htmlBody,
		TextContent: textBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Brevo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Brevo API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("EmailService: meeting invitation sent to %s for meeting %s\n", attendeeEmail, meetingCode)
	return nil
}

// SendSessionReadyEmail sends notification when meeting summary is ready
func (s *EmailService) SendSessionReadyEmail(toEmail, toName string, sessionID uint) error {
	if s.cfg.APIKey == "" {
		return fmt.Errorf("Brevo API key is not configured")
	}

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background: #f4f4f4; padding: 32px;">
  <div style="max-width: 560px; margin: 0 auto; background: #ffffff; border-radius: 8px; padding: 32px;">
    <h2 style="color: #1a1a2e; margin-top: 0;">Your meeting summary is ready! ✨</h2>
    <p style="color: #444; line-height: 1.6;">
      Hi %s,<br><br>
      Great news! The AI-powered summary and transcript for your meeting are now available.
    </p>
    <div style="text-align: center; margin: 24px 0;">
      <a href="#" style="display: inline-block; padding: 14px 32px; background: #6c63ff; color: #ffffff;
                          text-decoration: none; border-radius: 8px; font-weight: bold; font-size: 16px;">
        📄 View Summary
      </a>
    </div>
    <p style="color: #888; font-size: 14px; line-height: 1.6;">
      Log in to your Mini Meeting account to access the full transcript, key takeaways, and action items.
    </p>
  </div>
</body>
</html>`, toName)

	textBody := fmt.Sprintf("Hi %s,\n\nYour meeting summary is ready! Log in to your Mini Meeting account to view the full transcript and AI-powered summary.\n\nMini Meeting", toName)

	payload := brevoEmailRequest{
		Sender: brevoContact{
			Email: s.cfg.SenderEmail,
			Name:  s.cfg.SenderName,
		},
		To: []brevoContact{
			{Email: toEmail, Name: toName},
		},
		Subject:     "Your meeting summary is ready! ✨",
		HTMLContent: htmlBody,
		TextContent: textBody,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoAPIURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", s.cfg.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to Brevo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Brevo API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("EmailService: session ready notification sent to %s for session %d\n", toEmail, sessionID)
	return nil
}
