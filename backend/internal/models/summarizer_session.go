package models

import "time"

// SummarizerSession represents a meeting summarization session
// Status flow: STARTED → CAPTURED → TRANSCRIBED → NORMALIZED → SUMMARIZED
// On failure, status remains at the failed stage and the Error field is populated
type SummarizerSession struct {
	ID         uint                    `gorm:"primaryKey" json:"id"`
	MeetingID  uint                    `gorm:"not null" json:"meeting_id"`
	UserID     string                  `gorm:"not null;size:36" json:"user_id"`
	UserEmail  string                  `gorm:"size:255" json:"user_email"`
	Status     SummarizerSessionStatus `gorm:"not null;default:STARTED" json:"status"`
	Transcript *string                 `json:"transcript"`
	Summary    *string                 `json:"summary"`
	Error      *string                 `json:"error"`
	StartedAt  time.Time               `gorm:"not null" json:"started_at"`
	EndedAt    *time.Time              `json:"ended_at,omitempty"`
	CreatedAt  time.Time               `json:"-"`
	UpdatedAt  time.Time               `json:"-"`
}

// Status constants for state machine
// On failure, status remains at current stage and Error field is populated
type SummarizerSessionStatus string

const (
	StatusStarted     SummarizerSessionStatus = "STARTED"     // Bot joined, capturing audio
	StatusCaptured    SummarizerSessionStatus = "CAPTURED"    // Audio captured, ready for transcription
	StatusTranscribed SummarizerSessionStatus = "TRANSCRIBED" // Text extracted, ready for normalization
	StatusNormalized  SummarizerSessionStatus = "NORMALIZED"  // Text normalized, ready for summarization
	StatusSummarized  SummarizerSessionStatus = "SUMMARIZED"  // Final summary generated
)
