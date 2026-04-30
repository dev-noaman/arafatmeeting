# InsForge Hybrid Migration Design

**Date:** 2026-04-28
**Status:** Approved
**Scope:** Migrate Mini-Meeting from Go backend + Neon PostgreSQL to InsForge for auth + DB, keeping Go for LiveKit/lobby/workers.

**Assumption:** No production data exists in Neon. This is a development-stage migration. All tables start fresh in InsForge.

---

## 1. Architecture

### Before

```
Frontend → Go Backend → Neon PostgreSQL
```

All auth, user management, meetings, attendees, LiveKit, lobby, and summarization go through the Go backend.

### After

```
Frontend
  ├─ InsForge SDK → InsForge (auth, meetings, attendees, profiles)
  └─ Axios client  → Go Backend → InsForge PostgreSQL (LiveKit, lobby, summarizer workers)
```

### Responsibility Split

| Responsibility | InsForge | Go Backend |
|---------------|----------|------------|
| User registration/login | Yes | No |
| Email verification | Yes | No |
| Password reset | Yes | No |
| OAuth (Google, GitHub) | Yes | No |
| User profiles / admin user mgmt | Yes | No |
| Meeting CRUD | Yes | No |
| Attendee management | Yes | No |
| Saved attendees | Yes | No |
| Meeting code generation | Yes (DB function or client-side) | No |
| LiveKit token generation | No | Yes |
| LiveKit room management | No | Yes |
| WebSocket lobby | No | Yes |
| Summarizer orchestration | No | Yes |
| Background workers (transcription, normalization, summarization) | No | Yes |
| JWT validation | Issues tokens | Validates tokens |

---

## 2. Frontend Changes

### 2.1 InsForge SDK Setup

Install `@insforge/sdk` and create a shared client:

```ts
import { createClient } from '@insforge/sdk';

export const insforge = createClient({
  baseUrl: 'https://7c85mvkc.us-east.insforge.app',
  anonKey: '<from .env.local NEXT_PUBLIC_INSFORGE_ANON_KEY>',
});
```

**Prerequisite:** Validate SDK API surface by running `npx @insforge/cli docs db-sdk-typescript` and `npx @insforge/cli docs auth-sdk-typescript` before implementing. The method signatures in this spec are provisional and must be confirmed against actual SDK docs.

### 2.2 Auth Service Rewrite

Replace `src/services/api/auth.service.ts` with InsForge SDK calls:

| Method | Current | After |
|--------|---------|-------|
| login | `POST /auth/login` | `insforge.auth.login(email, password)` |
| register | `POST /auth/register` | `insforge.auth.register(email, password, name)` |
| verifyEmail | `POST /auth/verify-email` | `insforge.auth.verifyEmail(code)` |
| forgotPassword | `POST /auth/forgot-password` | `insforge.auth.forgotPassword(email)` |
| resetPassword | `POST /auth/reset-password` | `insforge.auth.resetPassword(code, newPassword)` |
| initiateOAuthLogin | `window.location.href = backend/oauth` | InsForge OAuth redirect |

### 2.3 Auth Store Changes

- Remove `tokenStorage.ts` — InsForge SDK manages tokens
- `auth.store.ts` reads auth state from InsForge SDK session instead of manual localStorage
- `authActions.ts` calls InsForge SDK methods, updates Zustand from SDK state
- `userFetcher.ts` likely unnecessary — SDK handles session hydration
- `ProtectedRoute` and `AuthInitializer` stay structurally identical, consuming `useAuth()` as before

### 2.4 Database Operations → InsForge SDK

**Meetings** (`meeting.service.ts`):
- `createMeeting` → `insforge.database.from('meetings').insert({...}).select()`
- `getMyMeetings` → `insforge.database.from('meetings').select().eq('creator_id', userId)`
- `getMeeting` → `insforge.database.from('meetings').select().eq('id', id)`
- `getMeetingByCode` → `insforge.database.from('meetings').select().eq('meeting_code', code)`
- `deleteMeeting` → `insforge.database.from('meetings').delete().eq('id', id)`

**Attendees** (`attendee.service.ts`):
- Meeting attendees → `insforge.database.from('meeting_attendees')`
- Saved attendees → `insforge.database.from('saved_attendees')`

**User profiles** (`user/user-profile.service.ts`):
- `getMe` → InsForge SDK user session
- `updateMe` → InsForge SDK user update

**Admin user management** (`user/user-admin.service.ts`):
- InsForge built-in admin user management

### 2.5 Go Backend API Calls (Stay with Axios)

The existing axios client (`src/services/api/client.ts`) continues calling the Go backend for:

- All LiveKit endpoints (`livekit/`)
- All Lobby endpoints (`lobby/`)
- All Summarizer endpoints (`sessions/`)
- The client attaches the InsForge token in `Authorization: Bearer` header

### 2.6 OAuth Page Changes

- `OAuthSuccess.tsx`: Instead of reading `?token=` from URL, handle InsForge OAuth callback
- `OAuthError.tsx`: Keep as-is for error display

### 2.7 Type Alignment

Map existing types to InsForge SDK return shapes:

```ts
// User comes from InsForge auth, fields map as:
// id → sub claim (UUID string)
// email → email
// name → user_metadata.name
// role → user_metadata.role or app_metadata.role
// provider → app_metadata.provider
```

**Key type change:** `User.id` changes from `number` to `string` (UUID). All code referencing user IDs as numbers must be updated to use strings.

Meeting and attendee types map directly to InsForge collection rows. The `meeting_code` field uses `VARCHAR(12)` (format: `xxx-xxxx-xxx`).

---

## 3. Database Schema

### 3.1 InsForge-Managed Tables

Created via `npx @insforge/cli db query` or SQL import:

```sql
-- Meetings
CREATE TABLE meetings (
  id BIGSERIAL PRIMARY KEY,
  meeting_code VARCHAR(12) NOT NULL UNIQUE,
  creator_id TEXT NOT NULL REFERENCES auth.users(id),
  meeting_link TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Meeting Attendees (invited people)
CREATE TABLE meeting_attendees (
  id BIGSERIAL PRIMARY KEY,
  meeting_id BIGINT NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
  user_id TEXT REFERENCES auth.users(id),
  email TEXT NOT NULL,
  name TEXT,
  invited_at TIMESTAMPTZ DEFAULT NOW(),
  joined_at TIMESTAMPTZ
);

-- Saved Attendees (user's address book)
CREATE TABLE saved_attendees (
  id BIGSERIAL PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES auth.users(id),
  email TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, email)
);

-- Meeting code generation function
CREATE OR REPLACE FUNCTION generate_meeting_code()
RETURNS TEXT AS $$
DECLARE
  chars TEXT := 'abcdefghijklmnopqrstuvwxyz';
  code TEXT;
BEGIN
  code := substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) || '-' ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) || '-' ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1) ||
          substr(chars, floor(random()*26+1)::int, 1);
  RETURN code;
END;
$$ LANGUAGE plpgsql;

-- RLS Policies
ALTER TABLE meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_attendees ENABLE ROW LEVEL SECURITY;
ALTER TABLE saved_attendees ENABLE ROW LEVEL SECURITY;

-- Meetings: creator can CRUD their own
CREATE POLICY "Users can view own meetings" ON meetings
  FOR SELECT USING (creator_id = auth.uid()::text);
CREATE POLICY "Users can create meetings" ON meetings
  FOR INSERT WITH CHECK (creator_id = auth.uid()::text);
CREATE POLICY "Users can update own meetings" ON meetings
  FOR UPDATE USING (creator_id = auth.uid()::text);
CREATE POLICY "Users can delete own meetings" ON meetings
  FOR DELETE USING (creator_id = auth.uid()::text);

-- Meeting attendees: visible to meeting creator + invited user
CREATE POLICY "Meeting attendees visible to creator" ON meeting_attendees
  FOR SELECT USING (
    EXISTS (SELECT 1 FROM meetings WHERE meetings.id = meeting_attendees.meeting_id AND meetings.creator_id = auth.uid()::text)
    OR user_id = auth.uid()::text
  );
CREATE POLICY "Creator can manage attendees" ON meeting_attendees
  FOR ALL USING (
    EXISTS (SELECT 1 FROM meetings WHERE meetings.id = meeting_attendees.meeting_id AND meetings.creator_id = auth.uid()::text)
  );

-- Saved attendees: owner only
CREATE POLICY "Users can manage own saved attendees" ON saved_attendees
  FOR ALL USING (user_id = auth.uid()::text);

-- Indexes
CREATE INDEX idx_meetings_creator ON meetings(creator_id);
CREATE INDEX idx_meetings_code ON meetings(meeting_code);
CREATE INDEX idx_meeting_attendees_meeting ON meeting_attendees(meeting_id);
CREATE INDEX idx_saved_attendees_user ON saved_attendees(user_id);

-- Public read policy for meeting lookup by code (guests need to find meetings)
CREATE POLICY "Public meeting code lookup" ON meetings
  FOR SELECT USING (true);
```

### 3.2 Go-Managed Tables (Stay in Go's Database)

The Go backend still needs its own tables for summarizer-related data. **These models must be updated** to remove GORM foreign keys to removed models:

**Changes to `summarizer_session.go`:**
- `MeetingID uint` stays as a plain value field (no foreign key to Meeting model)
- `UserID uint` changes to `UserID string` (InsForge user IDs are UUID strings)
- Remove `Meeting Meeting` and `User User` GORM association fields
- Remove `gorm:"foreignKey:MeetingID"` and `gorm:"foreignKey:UserID"` tags

**Changes to `audio_chunk.go` and `transcript.go`:**
- Same pattern: remove foreign key associations, convert user IDs to strings

**New table — `meeting_cache`:**
```sql
CREATE TABLE meeting_cache (
  meeting_id BIGSERIAL PRIMARY KEY,
  meeting_code VARCHAR(12) NOT NULL UNIQUE,
  creator_id TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```
This lightweight table allows the lobby and summarizer handlers to verify meeting ownership without querying InsForge. The Go backend populates it when the summarizer starts (via the InsForge REST API or a webhook).

These tables are in the same InsForge PostgreSQL instance but managed by Go's GORM AutoMigrate (not exposed via InsForge SDK).

---

## 4. Go Backend Changes

### 4.1 Type Migration: `uint` → `string` for User IDs

InsForge uses UUID strings for user IDs. Every Go handler and model currently using `uint` for user IDs must change to `string`:

- `c.Locals("userID")` changes from `.(uint)` to `.(string)` throughout all handlers
- `SummarizerSession.UserID` changes from `uint` to `string`
- `meeting_cache.CreatorID` is `string`
- All repository queries filtering by `user_id` change from `uint` to `string`

### 4.2 JWT Middleware Rework

The middleware interface stays identical — extract Bearer token, validate, set context locals. The validation changes:

- **Option A (Shared Secret):** InsForge and Go share the same HS256 secret. Minimal code change — just ensure the `JWT_SECRET` env var matches InsForge's signing key.
- **Option B (JWKS):** Go fetches InsForge's public key and validates RS256 tokens. More secure but more complex.

**Recommendation:** Option A for simplicity. The three context locals (`userID`, `email`, `role`) must still be set for all downstream handlers to work.

Claims mapping:
- InsForge `sub` → Go `userID` (as string, not uint)
- InsForge `email` → Go `email`
- InsForge `role` (from user_metadata `role` field or app_metadata) → Go `role`

**Admin middleware:** The `AdminMiddleware` checks `role != "admin"`. After migration, the role comes from InsForge's `app_metadata.role`. Ensure admin users in InsForge have `role: "admin"` set in their app_metadata. This is configured in the InsForge dashboard or via the admin API.

### 4.3 SummarizationService Dependency Resolution

`SummarizationService` currently depends on `UserService` and `EmailService`:
- `s.userService.GetUserByID(session.UserID)` — Replace with InsForge REST API call to fetch user by ID, or store the user's email in the `SummarizerSession` at creation time to avoid the lookup.
- `s.emailService.SendSessionReadyEmail(...)` — Either keep `EmailService` for this single purpose (simplest), or replace with an InsForge edge function call. **Recommendation:** Keep `EmailService` since it's only used for session-ready notifications, not auth.

### 4.4 Files Removed (~25 files)

```
internal/handlers/auth_handler.go
internal/handlers/user_handler.go
internal/handlers/meeting_handler.go
internal/handlers/saved_attendee_handler.go
internal/handlers/dto/user.go
internal/handlers/dto/meeting.go
internal/services/user_service.go
internal/services/meeting_service.go
internal/services/saved_attendee_service.go
internal/repositories/user_repository.go
internal/repositories/meeting_repository.go
internal/repositories/saved_attendee_repository.go
internal/models/user.go
internal/models/meeting.go
internal/models/saved_attendee.go
internal/cache/oauth_state.go
internal/routes/auth.go
internal/routes/user.go
pkg/oauth/google.go
pkg/oauth/github.go
pkg/oauth/oauth.go
pkg/oauth/provider.go
pkg/utils/password.go
pkg/utils/jwt.go (GenerateToken removed; ValidateToken reworked)
```

**Note:** `email_service.go` stays because `SummarizationService` uses it for session-ready notifications.

### 4.5 Files That Stay

```
cmd/server/main.go                   (trimmed wiring)
internal/config/config.go            (trimmed: remove OAuth, DB host/port/user/pass)
internal/middleware/auth.go           (reworked: validate InsForge tokens, userID as string)
internal/handlers/livekit_handler.go (userID uint → string)
internal/handlers/lobby_handler.go   (userID uint → string)
internal/handlers/lobby_ws_handler.go(userID uint → string, meeting ownership via meeting_cache)
internal/handlers/summarizer_handler.go (userID uint → string, meeting_cache for ownership)
internal/services/livekit_service.go
internal/services/summarizer_service.go (meeting ownership via meeting_cache)
internal/services/transcription_service.go
internal/services/normalization_service.go
internal/services/summarization_service.go (replace UserService dep with InsForge API or cached email)
internal/services/email_service.go      (kept for session-ready notifications)
internal/services/openrouter_service.go
internal/workers/transcription_worker.go
internal/workers/normalization_worker.go
internal/workers/summarization_worker.go
internal/cache/lobby.go
internal/cache/lobby_hub.go
internal/repositories/summarizer_session_repository.go (UserID uint → string)
internal/repositories/audio_chunk_repository.go
internal/repositories/transcript_repository.go
internal/models/summarizer_session.go (remove FK to Meeting/User, UserID uint → string)
internal/models/audio_chunk.go
internal/models/transcript.go
internal/handlers/dto/livekit.go
internal/handlers/dto/lobby.go
internal/handlers/dto/summarizer.go
pkg/cache/redis.go
```

### 4.6 Remaining API Surface

```
PUBLIC (no auth):
  GET    /api/v1/health
  POST   /api/v1/livekit/token
  GET    /api/v1/livekit/participants/count
  POST   /api/v1/lobby/request
  DELETE /api/v1/lobby/request
  WS     /ws/lobby/visitor
  WS     /ws/lobby/admin

AUTHENTICATED (InsForge JWT):
  POST   /api/v1/livekit/remove-participant
  POST   /api/v1/livekit/mute-participant
  POST   /api/v1/livekit/end-meeting
  GET    /api/v1/livekit/participants
  POST   /api/v1/meetings/:id/summarizer/start
  POST   /api/v1/meetings/:id/summarizer/stop
  GET    /api/v1/meetings/:id/summarizer/sessions
  GET    /api/v1/sessions/
  GET    /api/v1/sessions/:id
  DELETE /api/v1/sessions/:id
```

~16 endpoints, down from ~35. No auth endpoints, no user management, no meeting CRUD, no saved attendees.

### 4.7 LiveKit and Lobby Handler Changes

The LiveKit and Lobby handlers currently do optional inline JWT validation on public endpoints (e.g., `/livekit/token` accepts both authenticated and guest requests). These inline validation calls must also use the InsForge JWT secret instead of the Go-issued secret.

The lobby admin WebSocket currently validates the JWT from the query param to verify the user is the meeting creator. Since meeting data moves to InsForge, the lobby handler verifies meeting ownership via the `meeting_cache` table instead of querying the `meetings` table directly.

---

## 5. Migration Steps (High-Level)

### Phase 1: Database Setup
1. Create InsForge tables (meetings, meeting_attendees, saved_attendees, meeting_cache) via SQL migration
2. Set up RLS policies
3. Create meeting code generation function
4. Run InsForge DB migration via `npx @insforge/cli db query`

### Phase 2: Frontend Auth Migration
1. Install `@insforge/sdk`, validate SDK API via docs
2. Create InsForge client
3. Rewrite auth service to use InsForge SDK
4. Update auth store to read from InsForge session
5. Remove tokenStorage.ts
6. Update OAuth callback pages
7. Update `User` type: `id` from `number` to `string`
8. Test full auth flow (register, login, verify email, password reset, OAuth)

### Phase 3: Frontend DB Migration
1. Rewrite meeting service to use InsForge SDK
2. Rewrite attendee service to use InsForge SDK
3. Rewrite user service to use InsForge SDK
4. Update all types to use string user IDs
5. Test meeting CRUD, attendee management

### Phase 4: Go Backend Cleanup
1. Update SummarizerSession model: remove FK associations, `UserID uint` → `UserID string`
2. Update meeting_cache model
3. Rework JWT middleware for InsForge token validation (`userID` as string)
4. Fix `SummarizationService`: replace `UserService` dependency with InsForge API or cached email
5. Update all handlers: `c.Locals("userID").(uint)` → `c.Locals("userID").(string)`
6. Remove auth, user, meeting, attendee handlers/services/repositories/models
7. Trim main.go wiring
8. Test LiveKit, lobby, summarizer flows

### Phase 5: Integration Testing
1. Register new user via InsForge
2. Login and get InsForge token
3. Create meeting via InsForge SDK
4. Join meeting via Go LiveKit endpoint
5. Test lobby flow (visitor request → admin approve → LiveKit token)
6. Test summarizer flow (start → transcription → normalization → summarization)
7. Verify RLS policies prevent unauthorized access

### Phase 6: Environment & CORS
1. Add `NEXT_PUBLIC_INSFORGE_URL` and `NEXT_PUBLIC_INSFORGE_ANON_KEY` to frontend `.env`
2. Remove `BRAVO_API_KEY` from Go backend `.env` (kept if email service stays)
3. Remove `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET` from Go backend `.env`
4. Configure InsForge CORS to allow frontend origin
5. Update Google/GitHub OAuth redirect URLs to InsForge callback URLs

---

## 6. Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| InsForge JWT claims don't match Go's expected format | Map claims in middleware; test with actual InsForge tokens before removing old auth |
| Meeting code collisions | `VARCHAR(12) UNIQUE` constraint + retry logic in generation function |
| Lobby can't verify meeting ownership after moving meetings out of Go DB | `meeting_cache` table populated when summarizer starts or via webhook |
| InsForge SDK API surface differs from documented examples | Fetch SDK docs via `npx @insforge/cli docs` before implementing each service; pin SDK version |
| SummarizerSession FK to removed models breaks AutoMigrate | Remove GORM foreign key associations before removing models; meeting_id/user_id become plain value fields |
| SummarizationService depends on removed UserService/EmailService | Replace UserService with InsForge REST API call; keep EmailService for notifications |
| User ID type mismatch (uint vs UUID string) | Convert all `uint` user ID fields to `string` across Go models, handlers, and frontend types |
| OAuth redirect URLs change | Update Google/GitHub OAuth provider settings to point to InsForge callback URLs |
| Admin role not propagated to InsForge JWT | Set `role: "admin"` in InsForge `app_metadata` for admin users; verify in claims mapping |
| Mid-migration rollback needed | Each phase is independently testable. Keep old auth code until Phase 5 passes. Revert by switching frontend API endpoints back to Go backend. |

---

## 7. Environment Variables

### Frontend `.env` — Add
```
VITE_INSFORGE_URL=https://7c85mvkc.us-east.insforge.app
VITE_INSFORGE_ANON_KEY=<anon key from .insforge/project.json or .env.local>
```

### Frontend `.env` — Keep
```
VITE_API_BASE_URL (still needed for Go backend calls to LiveKit/lobby/summarizer)
```

### Go Backend `.env` — Remove
```
GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET
GOOGLE_REDIRECT_URL
GITHUB_CLIENT_ID
GITHUB_CLIENT_SECRET
GITHUB_REDIRECT_URL
DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME (replaced by DATABASE_URL)
```

### Go Backend `.env` — Keep
```
DATABASE_URL (now pointing to InsForge PostgreSQL connection string)
JWT_SECRET (must match InsForge's signing key for shared-secret validation)
PORT, ENV, FRONTEND_URL
REDIS_HOST, REDIS_PORT, REDIS_USERNAME, REDIS_PASSWORD
LIVEKIT_API_KEY, LIVEKIT_API_SECRET, LIVEKIT_URL
OPEN_ROUTER_API_KEY, OPEN_ROUTER_BASE_URL, OPEN_ROUTER_MODEL, etc.
WHISPER_URL, WHISPER_MODEL, etc.
SUMMARIZER_* settings
BRAVO_API_KEY, BREVO_SENDER_EMAIL, BREVO_SENDER_NAME (kept for session-ready emails)
```
