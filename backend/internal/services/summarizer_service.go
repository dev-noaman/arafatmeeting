package services

import (
	"errors"
	"fmt"
	"io"
	"mini-meeting/internal/config"
	"mini-meeting/internal/handlers/dto"
	"mini-meeting/internal/models"
	"mini-meeting/internal/repositories"
	"os"
	"path/filepath"
	"sync"
	"time"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media/oggwriter"
)

type SummarizerService struct {
	sessionRepo          *repositories.SummarizerSessionRepository
	chunkRepo            *repositories.AudioChunkRepository
	transcriptRepo       *repositories.TranscriptRepository
	livekitService       *LiveKitService
	transcriptionService *TranscriptionService
	cfg                  *config.Config

	activeRooms map[uint]*lksdk.Room
	roomMutex   sync.RWMutex
	wg          sync.WaitGroup
}

func NewSummarizerService(
	sessionRepo *repositories.SummarizerSessionRepository,
	chunkRepo *repositories.AudioChunkRepository,
	transcriptRepo *repositories.TranscriptRepository,
	livekitService *LiveKitService,
	transcriptionService *TranscriptionService,
	cfg *config.Config,
) *SummarizerService {
	return &SummarizerService{
		sessionRepo:          sessionRepo,
		chunkRepo:            chunkRepo,
		transcriptRepo:       transcriptRepo,
		livekitService:       livekitService,
		transcriptionService: transcriptionService,
		cfg:                  cfg,
		activeRooms:          make(map[uint]*lksdk.Room),
	}
}

// StartSummarizer starts a new summarizer session for a meeting
func (s *SummarizerService) StartSummarizer(meetingID uint, userID string, userEmail string) (*models.SummarizerSession, error) {
	// Check if there's already an active session
	existingSession, err := s.sessionRepo.FindActiveByMeetingID(meetingID)
	if err == nil && existingSession != nil {
		return nil, fmt.Errorf("summarizer already running for this meeting")
	}

	session := &models.SummarizerSession{
		MeetingID: meetingID,
		UserID:    userID,
		UserEmail: userEmail,
		Status:    models.StatusStarted,
		StartedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Get meeting code from the meeting_cache or derive from meeting ID
	// For now, use the meeting ID to generate a room name
	meetingCode := fmt.Sprintf("meeting-%d", meetingID)

	// Join LiveKit room as bot in background
	go func() {
		if err := s.joinLiveKitRoom(meetingCode, session.ID); err != nil {
			now := time.Now()
			errMsg := fmt.Sprintf("Failed to join LiveKit room: %v", err)
			s.sessionRepo.UpdateStatus(session.ID, models.StatusStarted, &errMsg, &now)
			fmt.Printf("%s\n", errMsg)
		}
	}()

	return session, nil
}

// StopSummarizer stops an active summarizer session
func (s *SummarizerService) StopSummarizer(sessionID uint, userID string) (int64, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return 0, fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != userID {
		return 0, fmt.Errorf("unauthorized: only meeting creator can stop summarizer")
	}

	if session.Status != models.StatusStarted {
		return 0, fmt.Errorf("session is not active (current status: %s)", session.Status)
	}

	s.roomMutex.Lock()
	if room, exists := s.activeRooms[sessionID]; exists {
		room.Disconnect()
		delete(s.activeRooms, sessionID)
	}
	s.roomMutex.Unlock()

	s.wg.Wait()

	now := time.Now()
	if err := s.sessionRepo.UpdateStatus(sessionID, models.StatusCaptured, nil, &now); err != nil {
		return 0, fmt.Errorf("failed to update session status: %w", err)
	}

	totalChunks, err := s.chunkRepo.CountBySessionID(sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to count chunks: %w", err)
	}

	go func() {
		fmt.Printf("Triggering background transcription for session %d\n", sessionID)
		if err := s.transcriptionService.ProcessSession(sessionID); err != nil {
			fmt.Printf("Failed to process session %d: %v\n", sessionID, err)
		}
	}()

	return totalChunks, nil
}

func (s *SummarizerService) joinLiveKitRoom(meetingCode string, sessionID uint) error {
	fmt.Printf("Attempting to join room %s for session %d\n", meetingCode, sessionID)

	token, err := s.livekitService.CreateBotToken(meetingCode, sessionID)
	if err != nil {
		return fmt.Errorf("failed to create bot token: %w", err)
	}

	fmt.Printf("Connecting bot to room %s using token (with Hidden=true permission)\n", meetingCode)
	room, err := lksdk.ConnectToRoomWithToken(s.livekitService.GetURL(), token, &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: func(track *webrtc.TrackRemote, publication *lksdk.RemoteTrackPublication, participant *lksdk.RemoteParticipant) {
				fmt.Printf("Track subscribed: %s from %s (kind: %s)\n", publication.SID(), participant.Identity(), track.Kind())
				if track.Kind() == webrtc.RTPCodecTypeAudio {
					fmt.Printf("Starting audio capture for user %s\n", participant.Identity())
					s.wg.Add(1)
					go s.handleAudioTrack(track, participant, sessionID)
				}
			},
			OnTrackPublished: func(publication *lksdk.RemoteTrackPublication, participant *lksdk.RemoteParticipant) {
				fmt.Printf("Track published: %s from %s\n", publication.SID(), participant.Identity())
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to room: %w", err)
	}

	s.roomMutex.Lock()
	s.activeRooms[sessionID] = room
	s.roomMutex.Unlock()

	fmt.Printf("Bot successfully joined room %s for session %d\n", meetingCode, sessionID)
	fmt.Printf("Waiting for participants to publish audio tracks...\n")

	return nil
}

func (s *SummarizerService) handleAudioTrack(track *webrtc.TrackRemote, participant *lksdk.RemoteParticipant, sessionID uint) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in audio handler: %v\n", r)
		}
	}()

	userIdentity := participant.Identity()
	chunkDuration := time.Duration(s.cfg.Summarizer.ChunkDurationSeconds) * time.Second

	fmt.Printf("Started processing audio for user %s in session %d (chunk duration: %v)\n", userIdentity, sessionID, chunkDuration)

	userDir, err := s.setupUserDirectory(sessionID, userIdentity)
	if err != nil {
		fmt.Printf("Failed to setup directory: %v\n", err)
		return
	}

	var (
		writer           *oggwriter.OggWriter
		currentChunkPath string
		chunkStartTime   time.Time
		chunkIndex       = 0
		packetCount      = 0
	)

	closeChunk := func() {
		if writer != nil {
			if err := writer.Close(); err != nil {
				fmt.Printf("Failed to close ogg writer: %v\n", err)
			}
			if packetCount > 0 {
				if err := s.createChunkMetadata(sessionID, userIdentity, chunkIndex, currentChunkPath, chunkStartTime); err != nil {
					fmt.Printf("Failed to create chunk metadata: %v\n", err)
				}
			}
		}
	}

	openNewChunk := func() error {
		chunkStartTime = time.Now()
		filename := fmt.Sprintf("chunk_%04d.ogg", chunkIndex)
		currentChunkPath = filepath.Join(userDir, filename)

		var err error
		writer, err = oggwriter.New(currentChunkPath, 48000, 2)
		if err != nil {
			return fmt.Errorf("failed to create ogg writer: %w", err)
		}
		return nil
	}

	defer closeChunk()

	if err := openNewChunk(); err != nil {
		fmt.Printf("Failed to create initial chunk: %v\n", err)
		return
	}

	for {
		rtp, _, err := track.ReadRTP()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("EOF reached for user %s after %d packets\n", userIdentity, packetCount)
				break
			}
			fmt.Printf("Error reading RTP: %v\n", err)
			continue
		}

		packetCount++
		if packetCount%100 == 0 {
			fmt.Printf("Received %d packets from %s\n", packetCount, userIdentity)
		}

		if err := writer.WriteRTP(rtp); err != nil {
			fmt.Printf("Failed to write RTP packet: %v\n", err)
			continue
		}

		if time.Since(chunkStartTime) >= chunkDuration {
			closeChunk()
			chunkIndex++
			if err := openNewChunk(); err != nil {
				fmt.Printf("Failed to create new chunk: %v\n", err)
				return
			}
			fmt.Printf("Rotated to chunk %d for user %s\n", chunkIndex, userIdentity)
		}
	}

	fmt.Printf("Finished processing audio for user %s in session %d (total packets: %d)\n", userIdentity, sessionID, packetCount)
}

func (s *SummarizerService) setupUserDirectory(sessionID uint, userIdentity string) (string, error) {
	userDir := filepath.Join(s.cfg.Summarizer.TempDir, fmt.Sprintf("%d", sessionID), fmt.Sprintf("user_%s", userIdentity))
	fmt.Printf("Creating directory: %s\n", userDir)

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for user %s: %w", userIdentity, err)
	}
	return userDir, nil
}

func (s *SummarizerService) createChunkMetadata(sessionID uint, userIdentity string, chunkIndex int, filePath string, startTime time.Time) error {
	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds()

	chunk := &models.AudioChunk{
		SessionID:       sessionID,
		UserIdentity:    userIdentity,
		ChunkIndex:      chunkIndex,
		FilePath:        filePath,
		StartTimestamp:  startTime,
		EndTimestamp:    endTime,
		DurationSeconds: duration,
	}

	if err := s.chunkRepo.Create(chunk); err != nil {
		return fmt.Errorf("failed to create audio chunk metadata: %w", err)
	}

	fmt.Printf("Saved metadata for chunk %d, user %s (%.2f seconds)\n", chunkIndex, userIdentity, duration)
	return nil
}

func (s *SummarizerService) GetActiveSession(meetingID uint) (*models.SummarizerSession, error) {
	session, err := s.sessionRepo.FindActiveByMeetingID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("no active summarizer session found")
	}
	return session, nil
}

func (s *SummarizerService) GetSessionByID(sessionID uint) (*models.SummarizerSession, error) {
	return s.sessionRepo.FindByID(sessionID)
}

func (s *SummarizerService) GetSessions(userID string, page, pageSize int) (*dto.PaginatedSessionsResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	sessions, total, err := s.sessionRepo.FindAllByUserIDPaginated(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	sessionList := make([]dto.SessionsList, len(sessions))
	for i, session := range sessions {
		sessionList[i] = dto.SessionsList{
			ID:        session.ID,
			Status:    session.Status,
			Error:     session.Error,
			StartedAt: session.StartedAt,
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.PaginatedSessionsResponse{
		Data:       sessionList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *SummarizerService) GetSession(sessionID uint, userID string) (*dto.SessionResponse, error) {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized: session does not belong to user")
	}

	return &dto.SessionResponse{
		ID:         session.ID,
		Status:     session.Status,
		Error:      session.Error,
		Transcript: session.Transcript,
		Summary:    session.Summary,
		StartedAt:  session.StartedAt,
		EndedAt:    session.EndedAt,
	}, nil
}

func (s *SummarizerService) DeleteSession(sessionID uint, userID string) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	if session.UserID != userID {
		return errors.New("unauthorized: session does not belong to user")
	}

	sessionDir := filepath.Join(s.cfg.Summarizer.TempDir, fmt.Sprintf("%d", sessionID))
	if err := os.RemoveAll(sessionDir); err != nil {
		fmt.Printf("Warning: failed to delete audio files for session %d: %v\n", sessionID, err)
	}

	if err := s.chunkRepo.DeleteBySessionID(sessionID); err != nil {
		return fmt.Errorf("failed to delete chunks: %w", err)
	}

	if err := s.transcriptRepo.DeleteBySessionID(sessionID); err != nil {
		return fmt.Errorf("failed to delete transcripts: %w", err)
	}

	if err := s.sessionRepo.Delete(sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (s *SummarizerService) GetSessionsByMeetingID(meetingID uint, userID string) ([]models.SummarizerSession, error) {
	sessions, err := s.sessionRepo.FindByMeetingID(meetingID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sessions: %w", err)
	}

	// Verify the user owns at least one session for this meeting (is the creator)
	owned := false
	for _, session := range sessions {
		if session.UserID == userID {
			owned = true
			break
		}
	}
	if !owned {
		return nil, errors.New("unauthorized: only meeting creator can view sessions")
	}

	return sessions, nil
}
