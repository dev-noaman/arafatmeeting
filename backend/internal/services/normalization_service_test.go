package services

import (
	"mini-meeting/internal/models"
	"testing"
)

func TestMergeConsecutiveSpeakers_EmptyInput(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.mergeConsecutiveSpeakers([]models.Transcript{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d segments", len(result))
	}
}

func TestMergeConsecutiveSpeakers_SingleTranscript(t *testing.T) {
	svc := &NormalizationService{}
	transcripts := []models.Transcript{
		{UserIdentity: "alice_1", Text: "  Hello world  ", StartTime: 0.0, EndTime: 5.0},
	}
	result := svc.mergeConsecutiveSpeakers(transcripts)
	if len(result) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result))
	}
	if result[0].UserIdentity != "alice_1" {
		t.Errorf("expected user alice_1, got %s", result[0].UserIdentity)
	}
	if result[0].Text != "Hello world" {
		t.Errorf("expected trimmed text 'Hello world', got '%s'", result[0].Text)
	}
	if result[0].StartTime != 0.0 {
		t.Errorf("expected start 0.0, got %f", result[0].StartTime)
	}
	if result[0].EndTime != 5.0 {
		t.Errorf("expected end 5.0, got %f", result[0].EndTime)
	}
}

func TestMergeConsecutiveSpeakers_MergesSameSpeaker(t *testing.T) {
	svc := &NormalizationService{}
	transcripts := []models.Transcript{
		{UserIdentity: "alice_1", Text: "Hello", StartTime: 0.0, EndTime: 5.0},
		{UserIdentity: "alice_1", Text: "World", StartTime: 5.0, EndTime: 10.0},
		{UserIdentity: "bob_2", Text: "Hi", StartTime: 10.0, EndTime: 15.0},
	}
	result := svc.mergeConsecutiveSpeakers(transcripts)
	if len(result) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(result))
	}
	if result[0].UserIdentity != "alice_1" {
		t.Errorf("expected first segment alice_1, got %s", result[0].UserIdentity)
	}
	if result[0].Text != "Hello World" {
		t.Errorf("expected merged text 'Hello World', got '%s'", result[0].Text)
	}
	if result[0].EndTime != 10.0 {
		t.Errorf("expected end time 10.0, got %f", result[0].EndTime)
	}
	if result[1].UserIdentity != "bob_2" {
		t.Errorf("expected second segment bob_2, got %s", result[1].UserIdentity)
	}
}

func TestMergeConsecutiveSpeakers_AlternatingSpeakers(t *testing.T) {
	svc := &NormalizationService{}
	transcripts := []models.Transcript{
		{UserIdentity: "alice_1", Text: "A", StartTime: 0.0, EndTime: 1.0},
		{UserIdentity: "bob_2", Text: "B", StartTime: 1.0, EndTime: 2.0},
		{UserIdentity: "alice_1", Text: "C", StartTime: 2.0, EndTime: 3.0},
	}
	result := svc.mergeConsecutiveSpeakers(transcripts)
	if len(result) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(result))
	}
}

func TestMergeConsecutiveSpeakers_SameSpeakerNotConsecutive(t *testing.T) {
	svc := &NormalizationService{}
	transcripts := []models.Transcript{
		{UserIdentity: "alice_1", Text: "A", StartTime: 0.0, EndTime: 1.0},
		{UserIdentity: "bob_2", Text: "B", StartTime: 1.0, EndTime: 2.0},
		{UserIdentity: "alice_1", Text: "C", StartTime: 2.0, EndTime: 3.0},
	}
	result := svc.mergeConsecutiveSpeakers(transcripts)
	// alice appears twice non-consecutively, should NOT be merged
	if result[0].Text != "A" {
		t.Errorf("expected first alice segment 'A', got '%s'", result[0].Text)
	}
	if result[2].Text != "C" {
		t.Errorf("expected second alice segment 'C', got '%s'", result[2].Text)
	}
}

func TestFormatMeetingDocument_EmptySegments(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.formatMeetingDocument([]MergedSegment{})
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestFormatMeetingDocument_SingleSegment(t *testing.T) {
	svc := &NormalizationService{}
	segments := []MergedSegment{
		{UserIdentity: "alice_1", Text: "Hello", StartTime: 0.0, EndTime: 5.0},
	}
	result := svc.formatMeetingDocument(segments)
	expected := "[00:00:00] alice: Hello"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestFormatMeetingDocument_MultipleSegments(t *testing.T) {
	svc := &NormalizationService{}
	segments := []MergedSegment{
		{UserIdentity: "alice_1", Text: "Hello", StartTime: 0.0, EndTime: 5.0},
		{UserIdentity: "bob_2", Text: "Hi", StartTime: 5.0, EndTime: 10.0},
	}
	result := svc.formatMeetingDocument(segments)
	expected := "[00:00:00] alice: Hello\n\n[00:00:05] bob: Hi"
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

func TestFormatTimestamp_Zero(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.formatTimestamp(0)
	if result != "00:00:00" {
		t.Errorf("expected '00:00:00', got '%s'", result)
	}
}

func TestFormatTimestamp_SixtySeconds(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.formatTimestamp(60)
	if result != "00:01:00" {
		t.Errorf("expected '00:01:00', got '%s'", result)
	}
}

func TestFormatTimestamp_Hour(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.formatTimestamp(3661)
	if result != "01:01:01" {
		t.Errorf("expected '01:01:01', got '%s'", result)
	}
}

func TestFormatTimestamp_MultiHour(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.formatTimestamp(7325)
	if result != "02:02:05" {
		t.Errorf("expected '02:02:05', got '%s'", result)
	}
}

func TestExtractDisplayName_SimpleName(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.extractDisplayName("alice")
	if result != "alice" {
		t.Errorf("expected 'alice', got '%s'", result)
	}
}

func TestExtractDisplayName_WithSuffix(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.extractDisplayName("alice_1")
	if result != "alice" {
		t.Errorf("expected 'alice', got '%s'", result)
	}
}

func TestExtractDisplayName_WithUnderscoreInName(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.extractDisplayName("alice_smith_123")
	if result != "alice_smith" {
		t.Errorf("expected 'alice_smith', got '%s'", result)
	}
}

func TestExtractDisplayName_EmailStyle(t *testing.T) {
	svc := &NormalizationService{}
	result := svc.extractDisplayName("user_domain.com_abc123")
	if result != "user_domain.com" {
		t.Errorf("expected 'user_domain.com', got '%s'", result)
	}
}
