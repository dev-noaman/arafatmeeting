package services

import (
	"mini-meeting/internal/handlers/dto"
	"mini-meeting/internal/models"
	"github.com/livekit/protocol/livekit"
)

// LiveKitServiceInterface defines the contract for LiveKit operations used by handlers.
// This interface enables testing without a real LiveKit server.
type LiveKitServiceInterface interface {
	CreateJoinToken(RoomCode, identity, userName, userRole, metadata string) (string, error)
	CreateBotToken(RoomCode string, sessionID uint) (string, error)
	GetURL() string
	RemoveParticipant(RoomCode, participantIdentity string) error
	ListParticipants(RoomCode string) ([]*livekit.ParticipantInfo, error)
	DeleteRoom(RoomCode string) error
	MuteParticipantTrack(RoomCode, participantIdentity, trackSid string, muted bool) error
	UpdateParticipantMetadata(RoomCode, participantIdentity, metadata string) error
	UpdateRoomMetadata(RoomCode, metadata string) error
}

// SummarizerServiceInterface defines the contract for summarizer operations used by handlers.
type SummarizerServiceInterface interface {
	StartSummarizer(meetingID uint, userID, userEmail string) (*models.SummarizerSession, error)
	StopSummarizer(sessionID uint, userID string) (int64, error)
	GetActiveSession(meetingID uint) (*models.SummarizerSession, error)
	GetSessionByID(sessionID uint) (*models.SummarizerSession, error)
	GetSessions(userID string, page, pageSize int) (*dto.PaginatedSessionsResponse, error)
	GetSession(sessionID uint, userID string) (*dto.SessionResponse, error)
	DeleteSession(sessionID uint, userID string) error
	GetSessionsByMeetingID(meetingID uint, userID string) ([]models.SummarizerSession, error)
}

// Ensure LiveKitService implements the interface at compile time
var _ LiveKitServiceInterface = (*LiveKitService)(nil)

// Ensure SummarizerService implements the interface at compile time
var _ SummarizerServiceInterface = (*SummarizerService)(nil)

// --- Mock implementations for testing ---

// MockLiveKitService is a mock implementation of LiveKitServiceInterface
type MockLiveKitService struct {
	CreateJoinTokenFunc  func(RoomCode, identity, userName, userRole, metadata string) (string, error)
	CreateBotTokenFunc   func(RoomCode string, sessionID uint) (string, error)
	GetURLFunc           func() string
	RemoveParticipantFunc func(RoomCode, participantIdentity string) error
	ListParticipantsFunc func(RoomCode string) ([]*livekit.ParticipantInfo, error)
	DeleteRoomFunc       func(RoomCode string) error
	MuteParticipantTrackFunc func(RoomCode, participantIdentity, trackSid string, muted bool) error
}

func (m *MockLiveKitService) CreateJoinToken(RoomCode, identity, userName, userRole, metadata string) (string, error) {
	if m.CreateJoinTokenFunc != nil {
		return m.CreateJoinTokenFunc(RoomCode, identity, userName, userRole, metadata)
	}
	return "mock-token", nil
}

func (m *MockLiveKitService) CreateBotToken(RoomCode string, sessionID uint) (string, error) {
	if m.CreateBotTokenFunc != nil {
		return m.CreateBotTokenFunc(RoomCode, sessionID)
	}
	return "mock-bot-token", nil
}

func (m *MockLiveKitService) GetURL() string {
	if m.GetURLFunc != nil {
		return m.GetURLFunc()
	}
	return "ws://localhost:7880"
}

func (m *MockLiveKitService) RemoveParticipant(RoomCode, participantIdentity string) error {
	if m.RemoveParticipantFunc != nil {
		return m.RemoveParticipantFunc(RoomCode, participantIdentity)
	}
	return nil
}

func (m *MockLiveKitService) ListParticipants(RoomCode string) ([]*livekit.ParticipantInfo, error) {
	if m.ListParticipantsFunc != nil {
		return m.ListParticipantsFunc(RoomCode)
	}
	return []*livekit.ParticipantInfo{}, nil
}

func (m *MockLiveKitService) DeleteRoom(RoomCode string) error {
	if m.DeleteRoomFunc != nil {
		return m.DeleteRoomFunc(RoomCode)
	}
	return nil
}

func (m *MockLiveKitService) MuteParticipantTrack(RoomCode, participantIdentity, trackSid string, muted bool) error {
	if m.MuteParticipantTrackFunc != nil {
		return m.MuteParticipantTrackFunc(RoomCode, participantIdentity, trackSid, muted)
	}
	return nil
}

func (m *MockLiveKitService) UpdateParticipantMetadata(RoomCode, participantIdentity, metadata string) error {
	return nil
}

func (m *MockLiveKitService) UpdateRoomMetadata(RoomCode, metadata string) error {
	return nil
}

// MockSummarizerService is a mock implementation of SummarizerServiceInterface
type MockSummarizerService struct {
	StartSummarizerFunc       func(meetingID uint, userID, userEmail string) (*models.SummarizerSession, error)
	StopSummarizerFunc        func(sessionID uint, userID string) (int64, error)
	GetActiveSessionFunc      func(meetingID uint) (*models.SummarizerSession, error)
	GetSessionByIDFunc        func(sessionID uint) (*models.SummarizerSession, error)
	GetSessionsFunc           func(userID string, page, pageSize int) (*dto.PaginatedSessionsResponse, error)
	GetSessionFunc            func(sessionID uint, userID string) (*dto.SessionResponse, error)
	DeleteSessionFunc         func(sessionID uint, userID string) error
	GetSessionsByMeetingIDFunc func(meetingID uint, userID string) ([]models.SummarizerSession, error)
}

func (m *MockSummarizerService) StartSummarizer(meetingID uint, userID, userEmail string) (*models.SummarizerSession, error) {
	if m.StartSummarizerFunc != nil {
		return m.StartSummarizerFunc(meetingID, userID, userEmail)
	}
	return nil, nil
}

func (m *MockSummarizerService) StopSummarizer(sessionID uint, userID string) (int64, error) {
	if m.StopSummarizerFunc != nil {
		return m.StopSummarizerFunc(sessionID, userID)
	}
	return 0, nil
}

func (m *MockSummarizerService) GetActiveSession(meetingID uint) (*models.SummarizerSession, error) {
	if m.GetActiveSessionFunc != nil {
		return m.GetActiveSessionFunc(meetingID)
	}
	return nil, nil
}

func (m *MockSummarizerService) GetSessionByID(sessionID uint) (*models.SummarizerSession, error) {
	if m.GetSessionByIDFunc != nil {
		return m.GetSessionByIDFunc(sessionID)
	}
	return nil, nil
}

func (m *MockSummarizerService) GetSessions(userID string, page, pageSize int) (*dto.PaginatedSessionsResponse, error) {
	if m.GetSessionsFunc != nil {
		return m.GetSessionsFunc(userID, page, pageSize)
	}
	return &dto.PaginatedSessionsResponse{
		Data:       []dto.SessionsList{},
		Total:      0,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 0,
	}, nil
}

func (m *MockSummarizerService) GetSession(sessionID uint, userID string) (*dto.SessionResponse, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(sessionID, userID)
	}
	return nil, nil
}

func (m *MockSummarizerService) DeleteSession(sessionID uint, userID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(sessionID, userID)
	}
	return nil
}

func (m *MockSummarizerService) GetSessionsByMeetingID(meetingID uint, userID string) ([]models.SummarizerSession, error) {
	if m.GetSessionsByMeetingIDFunc != nil {
		return m.GetSessionsByMeetingIDFunc(meetingID, userID)
	}
	return nil, nil
}

// Ensure mocks implement interfaces at compile time
var _ LiveKitServiceInterface = (*MockLiveKitService)(nil)
var _ SummarizerServiceInterface = (*MockSummarizerService)(nil)
