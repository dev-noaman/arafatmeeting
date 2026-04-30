# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mini-Meeting is a video conferencing platform with AI-powered meeting summaries. Built with React 19 + TypeScript (frontend), Go (backend — LiveKit/lobby/summarizer only), and LiveKit for real-time video/audio.

## Architecture

### Hybrid Architecture: InsForge + Go Backend
- **InsForge (BaaS)**: Auth, database (PostgreSQL + PostgREST), storage. Frontend talks to InsForge SDK directly for auth, meetings, attendees.
- **Go Backend (Fiber)**: Only handles LiveKit tokens, lobby WebSocket, and AI summarizer workers. Validates InsForge-issued JWTs.
- **Frontend**: Talks to both InsForge SDK (auth + data) and Go backend (LiveKit + lobby + summarizer) via axios with InsForge JWT in Authorization header.
- **Cloud Services**: LiveKit (video), Redis Cloud, OpenRouter (LLM), Brevo (email), Whisper (transcription)

### Data Flow
```
Frontend → InsForge SDK  → Auth (signup, login, OAuth, password reset)
Frontend → InsForge SDK  → Database (meetings, attendees CRUD via PostgREST)
Frontend → Go Backend    → LiveKit tokens, lobby WS, summarizer
Go Backend validates InsForge JWTs (HS256, sub=email, role field)
```

### Backend Structure (`/backend`)
```
cmd/
  server/main.go          # Application entry point
internal/
  config/                 # Configuration loading from env
  database/               # PostgreSQL connection (summarizer tables only)
  handlers/               # LiveKit, lobby, summarizer HTTP/WS handlers
  middleware/             # JWT auth middleware (validates InsForge tokens)
  models/                 # GORM models (SummarizerSession, AudioChunk, Transcript only)
  repositories/           # Database access layer (summarizer-related only)
  routes/                 # Route group setup (livekit, lobby, summarizer)
  services/               # Business logic (livekit, lobby, email, summarization pipeline)
  workers/                # Background workers (transcription, normalization, summarization)
pkg/
  cache/                  # Redis client + lobby hub
  utils/                  # JWT validation helpers
```

### Frontend Structure (`/frontend`)
```
src/
  components/
    auth/                 # Login, Register, ProtectedRoute
    common/               # Reusable UI (Button, Input, Modal, Toast)
    layout/               # Header, Layout
    meeting/              # LiveKit integration, lobby, controls
  hooks/                  # Custom React hooks (useAuth)
  pages/                  # Route pages (Login, Dashboard, Meeting, etc.)
  router/                 # React Router setup
  services/
    api/                  # Axios client (Go backend) + service modules
    insforge/             # InsForge SDK client (client.ts)
  store/                  # Zustand state management
  types/                  # TypeScript type definitions
  utils/                  # Utilities
```

## Development Commands

### Backend
```bash
# Run with hot-reload (development)
air -c .air.linux.toml

# Build
go build -o server.exe ./cmd/server

# Run directly (Go installed)
go run ./cmd/server/main.go
```

### Frontend
```bash
# Development server
npm run dev

# Build for production
npm run build

# Lint
npm run lint

# Type check
npx tsc --noEmit
```

### Docker
```bash
# Start all services (backend, frontend, postgres, redis, livekit, whisper)
docker-compose up --build

# Start specific service
docker-compose up backend

# View logs
docker-compose logs -f backend
```

## Environment Setup

### Backend (.env)
Copy `backend/.env.example` to `backend/.env`. Key variables:
- `DATABASE_URL` - PostgreSQL connection (for summarizer tables)
- `REDIS_HOST`, `REDIS_PORT` - Redis connection
- `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET` - LiveKit credentials
- `OPEN_ROUTER_API_KEY` - AI summarization API key
- `BRAVO_API_KEY` - Brevo email API key
- `JWT_SECRET` - Must match InsForge JWT secret (for token validation)

### Frontend (.env)
```
VITE_API_BASE_URL=http://localhost:3002/api/v1
VITE_LIVEKIT_URL=ws://localhost:7880
VITE_INSFORGE_URL=<insforge-instance-url>
VITE_INSFORGE_ANON_KEY=<insforge-anon-key>
```

## Key Patterns

### InsForge SDK Client (`services/insforge/client.ts`)
```ts
import { createClient } from '@insforge/sdk';
export const insforge = createClient({
  baseUrl: import.meta.env.VITE_INSFORGE_URL,
  anonKey: import.meta.env.VITE_INSFORGE_ANON_KEY,
});
```
- Auth: `insforge.auth.signUp()`, `signInWithPassword()`, `signOut()`, `signInWithOAuth()`, `getCurrentUser()`, `verifyEmail()`, `sendResetPasswordEmail()`, `resetPassword()`
- Database: `insforge.database.from('meetings').select().eq('creator_id', userId)`
- RPC: `insforge.database.rpc('generate_meeting_code')`
- Token access: `insforge.getHttpClient().getHeaders()` returns Authorization header for Go backend calls

### Go Backend JWT Validation
- InsForge issues JWTs with `{sub, email, role}` — signed HS256
- Go middleware extracts `sub` as `userID` (UUID string), `email`, and `role`
- No token generation in Go — only validation

### User IDs
- InsForge uses UUID strings (e.g., `"a1b2c3d4-..."`)
- All user ID fields are `string` type throughout the stack (frontend types, Go models, Redis cache)

### Meeting Code Format
Generated as `xxx-xxxx-xxx` (10 lowercase letters with dashes) via PostgreSQL `generate_meeting_code()` RPC

### Summarizer Pipeline
Background workers process in order: Audio capture → Transcription → Normalization → Summarization → Email notification

## API Routes

### Go Backend Routes (InsForge JWT required for protected routes)

#### LiveKit (`/api/v1/livekit`)
- `POST /token` - Generate room token (public, guests need this)
- `GET /participants/count` - Participant count (public)
- `GET /participants` - List participants (auth required)
- `POST /remove-participant` - Remove participant (auth required)
- `POST /mute-participant` - Mute participant track (auth required)
- `POST /end-meeting` - End meeting for all (auth required)

#### Lobby (`/api/v1/lobby`, WebSocket `/ws/lobby`)
- `POST /request` - Request to join meeting (public, guests)
- `DELETE /request` - Cancel join request (public)
- `WS /ws/lobby/visitor` - Visitor waits for approval
- `WS /ws/lobby/admin` - Admin approves/rejects requests (auth required)

#### Summarizer (`/api/v1/meetings/:id/summarizer`, `/api/v1/sessions`)
- `POST /meetings/:id/summarizer/start` - Start summarizer
- `POST /meetings/:id/summarizer/stop` - Stop summarizer
- `GET /meetings/:id/summarizer/sessions` - List sessions for meeting
- `GET /sessions` - List user's sessions (paginated)
- `GET /sessions/:id` - Get session detail
- `DELETE /sessions/:id` - Delete session

#### Health
- `GET /api/v1/health` - Health check

### InsForge SDK Routes (frontend → InsForge directly)
- Auth: signup, login, OAuth, email verification, password reset
- Database: meetings CRUD, attendees CRUD, user profile

## Database

### InsForge-Managed Tables
- `users` - InsForge auth users
- `meetings` - Meeting code (unique), creator_id (UUID), timestamps
- `meeting_attendees` - Meeting invitees with email/name
- `saved_attendees` - User's saved attendee contacts
- `meeting_cache` - Lightweight cache for Go backend ownership checks
- RLS policies enforce user-scoped access via `auth.uid()::text`

### Go-Managed Tables (local PostgreSQL)
- `summarizer_sessions` - AI summarization job tracking
- `audio_chunks` - Recording segments for transcription
- `transcripts` - Meeting transcripts

## Frontend State Management

### Zustand Stores
- `authStore` - User, isAuthenticated, isLoading (no token — InsForge manages session)
- `uiStore` - Toast notifications, modal state
- `meetingStore` - Active meeting state

### Routing
- `PublicRoutes` - Landing, Login, Register, ForgotPassword, ResetPassword
- `ProtectedRoutes` - Dashboard, Profile, Sessions, MeetingHistory
- `AdminRoutes` - Admin user management
- `MeetingRoutes` - Lobby, Room, Meeting detail

## Testing

No automated tests currently exist. Manual testing via:
1. Start backend + frontend
2. Open http://localhost:5173
3. Register/login flow (via InsForge)
4. Create meeting → copy code → open incognito → join meeting

## Common Issues

### Frontend blank page
Check browser console (F12). Common causes:
- Go backend API not running (LiveKit/lobby calls fail)
- InsForge instance unreachable
- IPv6/IPv4 binding issue (fix: add `host: '0.0.0.0'` to vite.config.ts)

### Port conflicts
- Backend: 3002 (dev), 3000 (Docker)
- Frontend: 5173 (Vite default)
- LiveKit: 7880, 7881, 7882/UDP
- PostgreSQL: 5432
- Redis: 6379
- Whisper: 9000

### InsForge JWT mismatch
If Go backend rejects valid InsForge tokens, ensure `JWT_SECRET` in backend `.env` matches the InsForge project's JWT secret.

### Database
- InsForge tables managed via InsForge dashboard or `npx @insforge/cli db import`
- Go summarizer tables managed via GORM AutoMigrate
