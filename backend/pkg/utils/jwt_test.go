package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateToken_ValidToken(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(t, secret, "user-1", "test@example.com", "user", time.Now().Add(time.Hour))
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected UserID 'user-1', got '%s'", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("expected Email 'test@example.com', got '%s'", claims.Email)
	}
	if claims.Role != "user" {
		t.Errorf("expected Role 'user', got '%s'", claims.Role)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	secret := "correct-secret"
	token := createTestToken(t, secret, "user-1", "test@example.com", "user", time.Now().Add(time.Hour))
	_, err := ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(t, secret, "user-1", "test@example.com", "user", time.Now().Add(-time.Hour))
	_, err := ValidateToken(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := ValidateToken("", "test-secret")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	_, err := ValidateToken("not-a-valid-jwt-token", "test-secret")
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}

func TestValidateToken_TamperedSignature(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(t, secret, "user-1", "test@example.com", "user", time.Now().Add(time.Hour))
	// Tamper with the signature (last part)
	parts := splitToken(token)
	if len(parts) != 3 {
		t.Fatal("token does not have 3 parts")
	}
	tampered := parts[0] + "." + parts[1] + ".tampered-signature"
	_, err := ValidateToken(tampered, secret)
	if err == nil {
		t.Fatal("expected error for tampered signature, got nil")
	}
}

func TestValidateToken_AdminRole(t *testing.T) {
	secret := "test-secret"
	token := createTestToken(t, secret, "admin-1", "admin@example.com", "admin", time.Now().Add(time.Hour))
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Role != "admin" {
		t.Errorf("expected Role 'admin', got '%s'", claims.Role)
	}
}

func TestValidateToken_UUIDUserID(t *testing.T) {
	secret := "test-secret"
	uuid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	token := createTestToken(t, secret, uuid, "user@example.com", "user", time.Now().Add(time.Hour))
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != uuid {
		t.Errorf("expected UserID '%s', got '%s'", uuid, claims.UserID)
	}
}

// Helper: create a valid JWT token for testing
func createTestToken(t *testing.T, secret, userID, email, role string, expiresAt time.Time) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})
	tkn, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}
	return tkn
}

// Helper: split JWT token by "."
func splitToken(token string) []string {
	result := make([]string, 0, 3)
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			result = append(result, token[start:i])
			start = i + 1
		}
	}
	result = append(result, token[start:])
	return result
}
