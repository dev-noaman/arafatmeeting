# InsForge Hybrid Migration Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate Mini-Meeting to use InsForge for auth + database while keeping Go backend for LiveKit, lobby, and summarizer workers.

**Architecture:** Frontend talks to InsForge SDK for auth + meeting CRUD + attendees. Frontend talks to Go backend via axios for LiveKit, lobby WebSocket, and summarizer operations. Go backend validates InsForge-issued JWTs.

**Tech Stack:** InsForge SDK (`@insforge/sdk`), InsForge CLI (`npx @insforge/cli`), Go/Gin/Fiber, React 19, Zustand, TypeScript

**Spec:** `docs/superpowers/specs/2026-04-28-insforge-hybrid-migration-design.md`

---

## Phase 1: Database Setup

### Task 1: Create InsForge database schema

**Files:**
- Create: `migrations/insforge-schema.sql` (local migration file for reference)

- [ ] **Step 1: Create the SQL migration file**

```sql
-- meetings
CREATE TABLE meetings (
  id BIGSERIAL PRIMARY KEY,
  meeting_code VARCHAR(12) NOT NULL UNIQUE,
  creator_id TEXT NOT NULL,
  meeting_link TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- meeting attendees
CREATE TABLE meeting_attendees (
  id BIGSERIAL PRIMARY KEY,
  meeting_id BIGINT NOT NULL REFERENCES meetings(id) ON DELETE CASCADE,
  user_id TEXT,
  email TEXT NOT NULL,
  name TEXT,
  invited_at TIMESTAMPTZ DEFAULT NOW(),
  joined_at TIMESTAMPTZ
);

-- saved attendees
CREATE TABLE saved_attendees (
  id BIGSERIAL PRIMARY KEY,
  user_id TEXT NOT NULL,
  email TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(user_id, email)
);

-- meeting cache (for Go backend ownership checks)
CREATE TABLE meeting_cache (
  meeting_id BIGSERIAL PRIMARY KEY,
  meeting_code VARCHAR(12) NOT NULL UNIQUE,
  creator_id TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- meeting code generation function
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

-- RLS
ALTER TABLE meetings ENABLE ROW LEVEL SECURITY;
ALTER TABLE meeting_attendees ENABLE ROW LEVEL SECURITY;
ALTER TABLE saved_attendees ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view own meetings" ON meetings FOR SELECT USING (creator_id = auth.uid()::text);
CREATE POLICY "Users can create meetings" ON meetings FOR INSERT WITH CHECK (creator_id = auth.uid()::text);
CREATE POLICY "Users can update own meetings" ON meetings FOR UPDATE USING (creator_id = auth.uid()::text);
CREATE POLICY "Users can delete own meetings" ON meetings FOR DELETE USING (creator_id = auth.uid()::text);
CREATE POLICY "Public meeting code lookup" ON meetings FOR SELECT USING (true);

CREATE POLICY "Meeting attendees visible to creator" ON meeting_attendees FOR SELECT USING (
  EXISTS (SELECT 1 FROM meetings WHERE meetings.id = meeting_attendees.meeting_id AND meetings.creator_id = auth.uid()::text)
  OR user_id = auth.uid()::text
);
CREATE POLICY "Creator can manage attendees" ON meeting_attendees FOR ALL USING (
  EXISTS (SELECT 1 FROM meetings WHERE meetings.id = meeting_attendees.meeting_id AND meetings.creator_id = auth.uid()::text)
);

CREATE POLICY "Users can manage own saved attendees" ON saved_attendees FOR ALL USING (user_id = auth.uid()::text);

-- Indexes
CREATE INDEX idx_meetings_creator ON meetings(creator_id);
CREATE INDEX idx_meetings_code ON meetings(meeting_code);
CREATE INDEX idx_meeting_attendees_meeting ON meeting_attendees(meeting_id);
CREATE INDEX idx_saved_attendees_user ON saved_attendees(user_id);
```

- [ ] **Step 2: Run migration against InsForge**

Run each CREATE TABLE and CREATE INDEX via:
```bash
npx @insforge/cli db query "<SQL statement>"
```

Or save to a file and import:
```bash
npx @insforge/cli db import migrations/insforge-schema.sql
```

- [ ] **Step 3: Verify tables were created**

```bash
npx @insforge/cli db tables
npx @insforge/cli db indexes
```

Expected: `meetings`, `meeting_attendees`, `saved_attendees`, `meeting_cache` listed.

- [ ] **Step 4: Commit**

```bash
git add migrations/insforge-schema.sql
git commit -m "feat: add InsForge database schema migration"
```

---

## Phase 2: Frontend SDK + Auth Migration

### Task 2: Install InsForge SDK and create client

**Files:**
- Modify: `frontend/package.json`
- Create: `frontend/src/services/insforge/client.ts`
- Modify: `frontend/.env` or `.env.local`

- [ ] **Step 1: Fetch InsForge SDK docs to confirm API surface**

```bash
npx @insforge/cli docs db-sdk-typescript
npx @insforge/cli docs auth-sdk-typescript
```

Read output carefully. Adjust method names in subsequent tasks if the actual SDK API differs from assumed signatures.

- [ ] **Step 2: Install SDK**

```bash
cd frontend && npm install @insforge/sdk@latest
```

- [ ] **Step 3: Add env variables to frontend `.env`**

```
VITE_INSFORGE_URL=https://7c85mvkc.us-east.insforge.app
VITE_INSFORGE_ANON_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3OC0xMjM0LTU2NzgtOTBhYi1jZGVmMTIzNDU2NzgiLCJlbWFpbCI6ImFub25AaW5zZm9yZ2UuY29tIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NzczNzE2Njh9.r0tz2slApSpGCkRFTNokob2sot1XJO4LrqH42zGbY1U
```

- [ ] **Step 4: Create InsForge client**

`frontend/src/services/insforge/client.ts`:
```ts
import { createClient } from '@insforge/sdk';

export const insforge = createClient({
  baseUrl: import.meta.env.VITE_INSFORGE_URL,
  anonKey: import.meta.env.VITE_INSFORGE_ANON_KEY,
});
```

- [ ] **Step 5: Commit**

```bash
git add frontend/package.json frontend/package-lock.json frontend/src/services/insforge/client.ts
git commit -m "feat: install InsForge SDK and create client"
```

---

### Task 3: Update User types for UUID string IDs

**NOTE:** The `User` interface is defined in `frontend/src/types/user.types.ts`. The `Meeting` interface is in `frontend/src/types/meeting.types.ts`. Update both files.

**Files:**
- Modify: `frontend/src/types/user.types.ts` — `User.id: number` → `string`
- Modify: `frontend/src/types/meeting.types.ts` — `Meeting.creator_id: number` → `string`, `SummarizerSession.meeting_id` stays `number`

- [ ] **Step 1: Update `User` type — `id` from `number` to `string`**

In `frontend/src/types/user.types.ts`, find the `User` interface and change `id`:
```ts
export interface User {
  id: string;  // was number — InsForge uses UUID strings
  // ... rest unchanged
}
```

- [ ] **Step 2: Update `Meeting` type — `creator_id` from `number` to `string`**

In `frontend/src/types/meeting.types.ts`, find the `Meeting` interface:
```ts
export interface Meeting {
  id: number;
  creator_id: string;  // was number — InsForge user IDs are UUID strings
  // ... rest unchanged
}
```

- [ ] **Step 3: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

Expected: Type errors in files that compare `id` as number. Fix each by searching for `user.id` or `meeting.creator_id` used as number.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/
git commit -m "feat: update User and Meeting types for InsForge UUID string IDs"
```

---

### Task 4: Rewrite auth service with InsForge SDK

**Files:**
- Modify: `frontend/src/services/api/auth.service.ts`

- [ ] **Step 1: Rewrite auth.service.ts to use InsForge SDK**

```ts
import { insforge } from '../insforge/client';
import type { User } from '../../types/user.types';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export const authService = {
  login: async (request: LoginRequest): Promise<{ token: string; user: User }> => {
    const { data, error } = await insforge.auth.login(request.email, request.password);
    if (error) throw new Error(error.message || 'Login failed');

    const user: User = {
      id: data.user.id,
      email: data.user.email,
      name: data.user.user_metadata?.name || '',
      role: data.user.app_metadata?.role || 'user',
      provider: data.user.app_metadata?.provider || 'local',
      created_at: data.user.created_at,
    };
    return { token: data.session.access_token, user };
  },

  register: async (request: RegisterRequest): Promise<{ message: string; user: { id: string; email: string; name: string } }> => {
    const { data, error } = await insforge.auth.register(request.email, request.password, {
      data: { name: request.name },
    });
    if (error) throw new Error(error.message || 'Registration failed');
    return {
      message: 'User registered successfully',
      user: { id: data.user.id, email: data.user.email, name: request.name },
    };
  },

  initiateOAuthLogin: (provider: 'google' | 'github'): void => {
    insforge.auth.loginWithOAuth(provider, { redirectTo: `${window.location.origin}/auth/oauth-success` });
  },

  verifyEmail: async (code: string): Promise<{ message: string }> => {
    const { error } = await insforge.auth.verifyEmail(code);
    if (error) throw new Error(error.message || 'Verification failed');
    return { message: 'Email verified successfully' };
  },

  forgotPassword: async (email: string): Promise<{ message: string }> => {
    const { error } = await insforge.auth.forgotPassword(email);
    if (error) throw new Error(error.message || 'Failed to send reset email');
    return { message: 'Reset email sent' };
  },

  resetPassword: async (code: string, password: string): Promise<{ message: string }> => {
    const { error } = await insforge.auth.resetPassword(code, password);
    if (error) throw new Error(error.message || 'Password reset failed');
    return { message: 'Password reset successfully' };
  },
};
```

**IMPORTANT:** The method names above (e.g., `insforge.auth.login`, `insforge.auth.register`) are assumed based on the SDK docs overview. Check the actual SDK API from Task 2 Step 1 and adjust accordingly.

- [ ] **Step 2: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/services/api/auth.service.ts
git commit -m "feat: rewrite auth service to use InsForge SDK"
```

---

### Task 5: Update auth store — remove tokenStorage, use InsForge session

**Files:**
- Modify: `frontend/src/store/auth/auth.store.ts`
- Modify: `frontend/src/store/auth/authActions.ts`
- Modify: `frontend/src/store/auth/auth.types.ts`
- Delete: `frontend/src/store/auth/tokenStorage.ts`
- Modify: `frontend/src/store/auth/userFetcher.ts`

- [ ] **Step 1: Update auth.types.ts — remove `token` from state**

```ts
import type { User } from '../../types/user.types';

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface AuthActions {
  setUser: (user: User | null) => void;
  login: (user: User) => void;
  logout: () => Promise<void>;
  initialize: () => Promise<void>;
}

export type AuthStore = AuthState & AuthActions;
```

- [ ] **Step 2: Rewrite authActions.ts — use InsForge SDK for session**

```ts
import type { AuthActions } from './auth.types';
import type { User } from '../../types/user.types';
import { insforge } from '../../services/insforge/client';

export const createAuthActions = (
  set: (partial: Partial<{ user: User | null; isAuthenticated: boolean; isLoading: boolean }>) => void,
  get: () => { user: User | null; isAuthenticated: boolean; isLoading: boolean },
): AuthActions => ({
  setUser: (user: User | null) => {
    set({ user, isAuthenticated: !!user });
  },

  login: (user: User) => {
    set({ user, isAuthenticated: true, isLoading: false });
  },

  logout: async () => {
    await insforge.auth.logout();
    set({ user: null, isAuthenticated: false });
  },

  initialize: async () => {
    try {
      const { data } = await insforge.auth.getSession();
      if (data.session) {
        const u = data.session.user;
        const user: User = {
          id: u.id,
          email: u.email,
          name: u.user_metadata?.name || '',
          role: u.app_metadata?.role || 'user',
          provider: u.app_metadata?.provider || 'local',
          created_at: u.created_at,
        };
        set({ user, isAuthenticated: true, isLoading: false });
      } else {
        set({ user: null, isAuthenticated: false, isLoading: false });
      }
    } catch {
      set({ user: null, isAuthenticated: false, isLoading: false });
    }
  },
});
```

- [ ] **Step 3: Update auth.store.ts — remove token from state and persist**

```ts
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type { AuthStore } from './auth.types';
import { createAuthActions } from './authActions';

export const useAuthStore = create<AuthStore>()(
  devtools((set, get) => ({
    user: null,
    isAuthenticated: false,
    isLoading: true,
    ...createAuthActions(set, get),
  })),
);
```

- [ ] **Step 4: Delete `frontend/src/store/auth/tokenStorage.ts`**

Remove the file. InsForge SDK manages token persistence.

- [ ] **Step 5: Delete `frontend/src/store/auth/userFetcher.ts`**

Remove the file. No longer needed — InsForge SDK handles session fetching.

- [ ] **Step 6: Update `useAuth` hook — remove token/setAuthData**

`frontend/src/hooks/useAuth.ts`:
```ts
import { useAuthStore } from '../store';

export const useAuth = () => {
  const user = useAuthStore((state) => state.user);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isLoading = useAuthStore((state) => state.isLoading);
  const logout = useAuthStore((state) => state.logout);
  const login = useAuthStore((state) => state.login);
  const setUser = useAuthStore((state) => state.setUser);

  return { user, isAuthenticated, isLoading, logout, login, setUser };
};
```

- [ ] **Step 7: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 8: Commit**

```bash
git add -A frontend/src/store/auth/ frontend/src/hooks/useAuth.ts
git commit -m "feat: update auth store to use InsForge session, remove tokenStorage"
```

---

### Task 6: Update auth-consuming components

**Files:**
- Modify: `frontend/src/components/auth/LoginForm.tsx`
- Modify: `frontend/src/components/auth/RegisterForm.tsx`
- Modify: `frontend/src/components/auth/ProtectedRoute.tsx`
- Modify: `frontend/src/pages/OAuthSuccess.tsx`
- Modify: `frontend/src/router/components/AuthInitializer.tsx`
- Modify: `frontend/src/services/api/client.ts` (update 401 handler)

- [ ] **Step 1: Update LoginForm.tsx — `login` now takes only `user` (no token)**

In `handleLocalLogin`, change:
```ts
const { user } = await authService.login({ email: formData.email, password: formData.password });
login(user);
```

The `login` action no longer takes a token parameter.

- [ ] **Step 2: Update RegisterForm.tsx — same pattern**

In `handleVerificationSuccess`:
```ts
authService.login({ email: formData.email, password: formData.password }).then(({ user }) => {
  login(user);
  navigate('/dashboard');
});
```

- [ ] **Step 3: Update OAuthSuccess.tsx — use InsForge OAuth callback**

Replace the token-from-URL logic with InsForge session check:
```tsx
useEffect(() => {
  const processOAuth = async () => {
    try {
      const { data } = await insforge.auth.getSession();
      if (data.session) {
        const u = data.session.user;
        const user: User = {
          id: u.id, email: u.email, name: u.user_metadata?.name || '',
          role: u.app_metadata?.role || 'user', provider: u.app_metadata?.provider || 'local',
          created_at: u.created_at,
        };
        login(user);
        navigate('/dashboard', { replace: true });
      } else {
        setError('No session found');
        setTimeout(() => navigate('/login'), 3000);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'OAuth failed');
      setTimeout(() => navigate('/login'), 3000);
    }
  };
  processOAuth();
}, [navigate, login]);
```

- [ ] **Step 4: Update client.ts — get token from InsForge for Go backend calls**

Replace the request interceptor. InsForge SDK caches sessions in memory, so `getSession()` is synchronous/same-tick after the initial load:
```ts
import { insforge } from '../insforge/client';

apiClient.interceptors.request.use(async (config) => {
  const { data } = await insforge.auth.getSession();
  if (data.session?.access_token) {
    config.headers.Authorization = `Bearer ${data.session.access_token}`;
  }
  return config;
});
```

Also clean up the 401 response interceptor — remove the dead `localStorage.removeItem("token")` calls. Replace with InsForge SDK session cleanup:
```ts
if (error.response?.status === 401) {
  const currentPath = window.location.pathname;
  const isAuthPage = currentPath === "/login" || currentPath === "/register" ||
    currentPath === "/verify-email" || currentPath === "/forgot-password" ||
    currentPath === "/reset-password" || currentPath.startsWith("/auth/oauth");

  if (!isAuthPage) {
    console.warn("401 Unauthorized — redirecting to login");
    await insforge.auth.logout();
    window.location.href = "/login";
  }
}
```

- [ ] **Step 5: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/auth/ frontend/src/pages/ frontend/src/router/ frontend/src/services/api/client.ts
git commit -m "feat: update auth-consuming components for InsForge SDK"
```

---

## Phase 3: Frontend DB Migration

### Task 7: Rewrite meeting service with InsForge SDK

**Files:**
- Modify: `frontend/src/services/api/meeting.service.ts`

- [ ] **Step 1: Rewrite meeting.service.ts**

```ts
import { insforge } from '../insforge/client';
import type {
  Meeting, MeetingResponse, MeetingsResponse, MeetingDeleteResponse,
  SummarizerSession, SummarizerStartResponse, SummarizerStopResponse,
} from '../../types/meeting.types';
import apiClient from './client';

// Meeting CRUD goes to InsForge SDK
export const meetingService = {
  createMeeting: async (attendees?: string[]): Promise<Meeting> => {
    const { data: session } = await insforge.auth.getSession();
    const userId = session.session?.user.id;
    if (!userId) throw new Error('Not authenticated');

    const { data: codeResult } = await insforge.database.rpc('generate_meeting_code');
    const meetingCode = codeResult as string;

    const { data, error } = await insforge.database
      .from('meetings')
      .insert({ meeting_code: meetingCode, creator_id: userId })
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as Meeting;
  },

  getMyMeetings: async (): Promise<Meeting[]> => {
    const { data: session } = await insforge.auth.getSession();
    const userId = session.session?.user.id;
    const { data, error } = await insforge.database
      .from('meetings')
      .select('*')
      .eq('creator_id', userId)
      .order('created_at', { ascending: false });
    if (error) throw new Error(error.message);
    return (data as Meeting[]) || [];
  },

  getMeetingById: async (id: number): Promise<Meeting> => {
    const { data, error } = await insforge.database
      .from('meetings')
      .select('*')
      .eq('id', id)
      .single();
    if (error) throw new Error(error.message);
    return data as Meeting;
  },

  getMeetingByCode: async (code: string): Promise<Meeting> => {
    const { data, error } = await insforge.database
      .from('meetings')
      .select('*')
      .eq('meeting_code', code)
      .single();
    if (error) throw new Error(error.message);
    return data as Meeting;
  },

  deleteMeeting: async (id: number): Promise<void> => {
    const { error } = await insforge.database.from('meetings').delete().eq('id', id);
    if (error) throw new Error(error.message);
  },

  getAllMeetings: async (): Promise<Meeting[]> => {
    const { data, error } = await insforge.database
      .from('meetings')
      .select('*')
      .order('created_at', { ascending: false });
    if (error) throw new Error(error.message);
    return (data as Meeting[]) || [];
  },

  // Summarizer operations stay on Go backend
  startSummarizer: async (meetingId: number): Promise<SummarizerSession> => {
    const response = await apiClient.post<SummarizerStartResponse>(`/meetings/${meetingId}/summarizer/start`);
    return response.data.data;
  },

  stopSummarizer: async (meetingId: number): Promise<SummarizerSession> => {
    const response = await apiClient.post<SummarizerStopResponse>(`/meetings/${meetingId}/summarizer/stop`);
    return response.data.data;
  },

  getSummarizerSessions: async (meetingId: number): Promise<SummarizerSession[]> => {
    const response = await apiClient.get(`/meetings/${meetingId}/summarizer/sessions`);
    return response.data.data;
  },
};
```

- [ ] **Step 2: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/services/api/meeting.service.ts
git commit -m "feat: rewrite meeting service to use InsForge SDK for CRUD"
```

---

### Task 8: Rewrite attendee services with InsForge SDK

**Files:**
- Modify: `frontend/src/services/api/attendee.service.ts`

- [ ] **Step 1: Rewrite attendee.service.ts**

```ts
import { insforge } from '../insforge/client';
import type { AttendeeResponse, SavedAttendee } from '../../types/meeting.types';

export const attendeeService = {
  addAttendee: async (meetingId: number, email: string, name?: string): Promise<AttendeeResponse> => {
    const { data, error } = await insforge.database
      .from('meeting_attendees')
      .insert({ meeting_id: meetingId, email, name })
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as AttendeeResponse;
  },

  removeAttendee: async (meetingId: number, email: string): Promise<void> => {
    const { error } = await insforge.database
      .from('meeting_attendees')
      .delete()
      .eq('meeting_id', meetingId)
      .eq('email', email);
    if (error) throw new Error(error.message);
  },

  getSavedAttendees: async (): Promise<SavedAttendee[]> => {
    const { data, error } = await insforge.database
      .from('saved_attendees')
      .select('*')
      .order('created_at', { ascending: false });
    if (error) throw new Error(error.message);
    return (data as SavedAttendee[]) || [];
  },

  saveAttendee: async (email: string, name: string): Promise<SavedAttendee> => {
    const { data: session } = await insforge.auth.getSession();
    const userId = session.session?.user.id;
    const { data, error } = await insforge.database
      .from('saved_attendees')
      .insert({ user_id: userId, email, name })
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as SavedAttendee;
  },

  updateAttendee: async (id: number, email: string, name: string): Promise<SavedAttendee> => {
    const { data, error } = await insforge.database
      .from('saved_attendees')
      .update({ email, name })
      .eq('id', id)
      .select()
      .single();
    if (error) throw new Error(error.message);
    return data as SavedAttendee;
  },

  deleteAttendee: async (id: number): Promise<void> => {
    const { error } = await insforge.database.from('saved_attendees').delete().eq('id', id);
    if (error) throw new Error(error.message);
  },
};
```

- [ ] **Step 2: Verify build**

```bash
cd frontend && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/services/api/attendee.service.ts
git commit -m "feat: rewrite attendee service to use InsForge SDK"
```

---

### Task 9: Rewrite user services with InsForge SDK

**Files:**
- Modify: `frontend/src/services/api/user/user-profile.service.ts`
- Modify: `frontend/src/services/api/user/user-admin.service.ts`
- Modify: `frontend/src/services/api/user/index.ts` (if exists)

- [ ] **Step 1: Rewrite user-profile.service.ts**

```ts
import { insforge } from '../../insforge/client';
import type { User, UpdateUserRequest } from '../../../types/user.types';

export const userProfileService = {
  getCurrentUser: async (): Promise<User> => {
    const { data, error } = await insforge.auth.getSession();
    if (error || !data.session) throw new Error('Not authenticated');
    const u = data.session.user;
    return {
      id: u.id,
      email: u.email,
      name: u.user_metadata?.name || '',
      role: u.app_metadata?.role || 'user',
      provider: u.app_metadata?.provider || 'local',
      created_at: u.created_at,
    };
  },

  updateCurrentUser: async (updateData: UpdateUserRequest): Promise<User> => {
    const { data, error } = await insforge.auth.updateUser({ data: updateData });
    if (error) throw new Error(error.message);
    const u = data.user;
    return {
      id: u.id,
      email: u.email,
      name: u.user_metadata?.name || '',
      role: u.app_metadata?.role || 'user',
      provider: u.app_metadata?.provider || 'local',
      created_at: u.created_at,
    };
  },
};
```

- [ ] **Step 2: Rewrite user-admin.service.ts**

```ts
import { insforge } from '../../insforge/client';
import type { User, PaginatedUsersResponse } from '../../../types/user.types';

export const userAdminService = {
  getAllUsers: async (page: number = 1, pageSize: number = 10, search: string = ''): Promise<PaginatedUsersResponse> => {
    // InsForge admin user listing — check SDK docs for exact method
    // Fallback: use InsForge REST API directly
    const { data, error } = await insforge.auth.adminListUsers({ page, perPage: pageSize });
    if (error) throw new Error(error.message);
    const users: User[] = (data.users || []).map((u: any) => ({
      id: u.id,
      email: u.email,
      name: u.user_metadata?.name || '',
      role: u.app_metadata?.role || 'user',
      provider: u.app_metadata?.provider || 'local',
      created_at: u.created_at,
    }));
    return {
      data: search ? users.filter(u => u.email.includes(search) || u.name.includes(search)) : users,
      total: data.total || users.length,
      page,
      page_size: pageSize,
      total_pages: Math.ceil((data.total || users.length) / pageSize),
    };
  },

  getUserById: async (id: string): Promise<User> => {
    const { data, error } = await insforge.auth.adminGetUserById(id);
    if (error) throw new Error(error.message);
    return {
      id: data.user.id,
      email: data.user.email,
      name: data.user.user_metadata?.name || '',
      role: data.user.app_metadata?.role || 'user',
      provider: data.user.app_metadata?.provider || 'local',
      created_at: data.user.created_at,
    };
  },

  deleteUser: async (id: string): Promise<void> => {
    const { error } = await insforge.auth.adminDeleteUser(id);
    if (error) throw new Error(error.message);
  },
};
```

**IMPORTANT:** The admin method names (`adminListUsers`, `adminGetUserById`, `adminDeleteUser`) are assumed. Verify against SDK docs from Task 2 Step 1. The `getUserById` parameter also changes from `number` to `string` since IDs are UUIDs now.

- [ ] **Step 3: Verify build and commit**

```bash
cd frontend && npx tsc --noEmit
git add frontend/src/services/api/user/
git commit -m "feat: rewrite user services to use InsForge SDK"
```

---

## Phase 4: Go Backend Cleanup

**IMPORTANT: Tasks 10-12 MUST be done atomically in a single commit.** After changing `JWTClaims.UserID` to `string` (Task 11), the middleware sets `c.Locals("userID", someString)`. But handlers still do `c.Locals("userID").(uint)` — a runtime panic. Do all three tasks together, then verify build + commit once.

### Task 10-12: Atomic uint→string user ID migration (model + middleware + handlers)

**Files:**
- Modify: `backend/internal/models/summarizer_session.go`
- Modify: `backend/pkg/utils/jwt.go`
- Modify: `backend/internal/middleware/auth.go`
- Modify: `backend/internal/handlers/livekit_handler.go`
- Modify: `backend/internal/handlers/lobby_handler.go`
- Modify: `backend/internal/handlers/lobby_ws_handler.go`
- Modify: `backend/internal/handlers/summarizer_handler.go`
- Modify: `backend/internal/repositories/summarizer_session_repository.go`
- Modify: `backend/internal/models/audio_chunk.go`
- Modify: `backend/internal/models/transcript.go`

- [ ] **Step 1: Update SummarizerSession model**

In `backend/internal/models/summarizer_session.go`, change:
```go
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
```

Key changes:
- `UserID` changed from `uint` to `string` (UUID)
- Added `UserEmail` field (needed for email notifications after UserService removal)
- `MeetingID` stays `uint` but removed `Meeting` relation field
- Removed `Meeting Meeting` and `User User` association fields
- Exposed `meeting_id` and `user_id` in JSON (removed `json:"-"`)

- [ ] **Step 2: Update JWTClaims — UserID from uint to string**

In `backend/pkg/utils/jwt.go`:
```go
type JWTClaims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
```

Remove the `GenerateToken` function entirely (InsForge issues tokens). Keep `ValidateToken` as-is.

- [ ] **Step 3: Update middleware — userID as string**

In `backend/internal/middleware/auth.go`, the `c.Locals("userID", claims.UserID)` line stays the same — it now stores a `string` instead of `uint` because `claims.UserID` is now `string`.

- [ ] **Step 4: Replace all `c.Locals("userID").(uint)` with `c.Locals("userID").(string)`**

In every handler file, search for `c.Locals("userID").(uint)` and replace with `c.Locals("userID").(string)`. The variable `userID` becomes a `string` everywhere.

In `livekit_handler.go`, update:
- `GenerateToken`: `userID` comparison changes from `if userID > 0` to `if userID != ""`
- `RemoveParticipant`, `MuteParticipant`, `EndMeeting`: ownership checks against `meeting_cache` table instead of `meeting.CreatorID == userID`
- User name/avatar lookup: use data from JWT claims (email) or store in request metadata

In `lobby_handler.go` and `lobby_ws_handler.go`, same pattern:
- `userID` is now `string`
- Meeting ownership via `meeting_cache` table query instead of `meetingService.GetMeetingByCode`

In `summarizer_handler.go`:
- `userID` is now `string` — all `sessionRepo` queries must accept `string` userID

- [ ] **Step 5: Update ALL repository methods that filter by user_id**

In `summarizer_session_repository.go`, change all methods that take `uint` userID to take `string`:
- `FindByUserID(userID string, ...)`
- `FindByMeetingID(meetingID uint, userID string, ...)`

Also audit `audio_chunk_repository.go` and `transcript_repository.go` — if any methods filter by user_id, change from `uint` to `string`. Also remove any `Preload("Meeting")` calls from `summarizer_session_repository.go` (e.g., `FindByID`, `FindByMeetingID`) since the `Meeting` model is being deleted.

- [ ] **Step 6: Verify build**

```bash
cd backend && go build ./...
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/models/ backend/pkg/utils/ backend/internal/middleware/ backend/internal/handlers/ backend/internal/repositories/
git commit -m "feat: atomic uint→string user ID migration across model, middleware, handlers, repos"
```

---

### Task 13: Fix SummarizationService dependencies

**Files:**
- Modify: `backend/internal/services/summarization_service.go`

- [ ] **Step 1: Replace UserService dependency**

The service currently calls `s.userService.GetUserByID(session.UserID)` to get the user's email for notification. Change to store the user email in the session at creation time, or look it up from JWT claims, or call InsForge REST API.

Simplest approach — add `UserEmail string` field to `SummarizerSession` and populate it when creating the session. Then:
```go
go func() {
    if err := s.emailService.SendSessionReadyEmail(session.UserEmail, "", sessionID); err != nil {
        fmt.Printf("EmailNotification: failed for session %d: %v\n", sessionID, err)
    }
}()
```

Remove the `userService` field from `SummarizationService` and update the constructor.

- [ ] **Step 2: Verify build**

```bash
cd backend && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/services/summarization_service.go
git commit -m "feat: remove UserService dependency from SummarizationService"
```

---

### Task 14: Remove dead code — auth, user, meeting, attendee handlers/services/repositories/models

**Files:**
- Delete: `backend/internal/handlers/auth_handler.go`
- Delete: `backend/internal/handlers/user_handler.go`
- Delete: `backend/internal/handlers/meeting_handler.go`
- Delete: `backend/internal/handlers/saved_attendee_handler.go`
- Delete: `backend/internal/handlers/dto/user.go`
- Delete: `backend/internal/handlers/dto/meeting.go`
- Delete: `backend/internal/services/user_service.go`
- Delete: `backend/internal/services/meeting_service.go`
- Delete: `backend/internal/services/saved_attendee_service.go`
- Delete: `backend/internal/repositories/user_repository.go`
- Delete: `backend/internal/repositories/meeting_repository.go`
- Delete: `backend/internal/repositories/saved_attendee_repository.go`
- Delete: `backend/internal/models/user.go`
- Delete: `backend/internal/models/meeting.go`
- Delete: `backend/internal/models/saved_attendee.go`
- Delete: `backend/internal/cache/oauth_state.go`
- Delete: `backend/internal/routes/auth.go`
- Delete: `backend/internal/routes/user.go`
- Modify: `backend/internal/routes/meeting.go` — extract summarizer/session routes into `summarizer_routes.go`, then delete meeting.go
- Delete: `backend/pkg/oauth/` (all 4 files)
- Delete: `backend/pkg/utils/password.go`

- [ ] **Step 1: Delete all listed files**

```bash
cd backend
rm internal/handlers/auth_handler.go
rm internal/handlers/user_handler.go
rm internal/handlers/meeting_handler.go
rm internal/handlers/saved_attendee_handler.go
rm internal/handlers/dto/user.go
rm internal/handlers/dto/meeting.go
rm internal/services/user_service.go
rm internal/services/meeting_service.go
rm internal/services/saved_attendee_service.go
rm internal/repositories/user_repository.go
rm internal/repositories/meeting_repository.go
rm internal/repositories/saved_attendee_repository.go
rm internal/models/user.go
rm internal/models/meeting.go
rm internal/models/saved_attendee.go
rm internal/cache/oauth_state.go
rm internal/routes/auth.go
rm internal/routes/user.go
rm -rf pkg/oauth/
rm pkg/utils/password.go
```

**Note:** Do NOT delete `internal/routes/meeting.go` yet — it contains `setupSavedAttendeeRoutes` and summarizer sub-routes. Handle it in Step 2.

- [ ] **Step 2: Extract summarizer routes from meeting.go, then delete meeting.go**

Read `backend/internal/routes/meeting.go`. It contains:
- `setupMeetingRoutes` — DELETE (meeting CRUD moves to InsForge)
- `setupSavedAttendeeRoutes` — DELETE (moves to InsForge)
- Summarizer sub-routes (e.g., `POST /meetings/:id/summarizer/start`) — KEEP

Create `backend/internal/routes/summarizer.go` with the summarizer and session routes extracted from `meeting.go`. Then delete `meeting.go`.

- [ ] **Step 3: Update routes.go — remove dead route setup calls**

Remove references to `authHandler`, `userHandler`, `meetingHandler`, `savedAttendeeHandler` from `SetupRoutes`. Keep only livekit, lobby, and summarizer routes.

- [ ] **Step 4: Update main.go — remove dead wiring**

Remove:
- `userRepo`, `meetingRepo`, `savedAttendeeRepo` initialization
- `meetingService`, `userService`, `savedAttendeeService` initialization
- `userHandler`, `authHandler`, `meetingHandler`, `savedAttendeeHandler` initialization
- All email service wiring for meeting invitations

Update handler constructors that no longer take `meetingService`, `userService`:
- `NewLiveKitHandler` — remove `meetingService`, `userService` params
- `NewLobbyHandler` — remove `meetingService`, `userService` params
- `NewLobbyWSHandler` — remove `meetingService`, `userService` params

- [ ] **Step 4: Fix remaining compilation errors**

Run `go build ./...` and fix any remaining imports referencing deleted packages.

- [ ] **Step 5: Verify build**

```bash
cd backend && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add -A backend/
git commit -m "feat: remove dead auth, user, meeting, attendee code from Go backend"
```

---

### Task 15: Update Go database migration — only manage summarizer tables

**Files:**
- Modify: `backend/internal/database/db.go`

- [ ] **Step 1: Update Migrate() to only manage summarizer tables**

Remove AutoMigrate calls for `User`, `Meeting`, `SavedAttendee`. Keep only:
```go
func Migrate() error {
    log.Println("Running database migrations...")
    db := DB.Session(&gorm.Session{})
    err := db.AutoMigrate(&models.SummarizerSession{})
    if err != nil {
        return fmt.Errorf("failed to create summarizer_sessions table: %w", err)
    }
    err = db.AutoMigrate(&models.AudioChunk{})
    if err != nil {
        return fmt.Errorf("failed to create audio_chunks table: %w", err)
    }
    err = db.AutoMigrate(&models.Transcript{})
    if err != nil {
        return fmt.Errorf("failed to create transcripts table: %w", err)
    }
    if err := runSQLMigrations(); err != nil {
        log.Printf("Warning: SQL migrations failed: %v", err)
    }
    log.Println("Database migrations completed")
    return nil
}
```

- [ ] **Step 2: Verify build and commit**

```bash
cd backend && go build ./...
git add backend/internal/database/db.go
git commit -m "feat: update DB migration to only manage summarizer tables"
```

---

## Phase 5: Integration Testing

### Task 16: End-to-end test

- [ ] **Step 1: Start Go backend**

```bash
cd backend && go run ./cmd/server/main.go
```

Expected: Server starts on port 3002, no errors about missing tables/models.

- [ ] **Step 2: Start frontend dev server**

```bash
cd frontend && npm run dev
```

- [ ] **Step 3: Test registration flow**

1. Open `http://localhost:5173/register`
2. Register a new user
3. Verify InsForge creates the user (check `npx @insforge/cli db tables` or dashboard)
4. Complete email verification if required
5. Confirm redirect to dashboard

- [ ] **Step 4: Test meeting creation**

1. Create a new meeting from dashboard
2. Verify `meetings` table has a row in InsForge DB:
   ```bash
   npx @insforge/cli db query "SELECT * FROM meetings;"
   ```
3. Copy meeting code

- [ ] **Step 5: Test meeting join (LiveKit)**

1. Open meeting code in incognito window
2. Verify Go backend generates LiveKit token
3. Confirm video call connects

- [ ] **Step 6: Test lobby flow**

1. Enable lobby for a meeting
2. Join as guest → should see waiting room
3. Approve from creator tab → guest should get LiveKit token

- [ ] **Step 7: Fix any issues found and commit**

---

### Task 17: Environment cleanup

**Files:**
- Modify: `backend/.env` — remove OAuth vars
- Modify: `frontend/.env` — add InsForge vars

- [ ] **Step 1: Clean up backend .env**

Remove:
```
GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REDIRECT_URL
GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, GITHUB_REDIRECT_URL
```

- [ ] **Step 2: Update frontend .env**

Ensure:
```
VITE_INSFORGE_URL=https://7c85mvkc.us-east.insforge.app
VITE_INSFORGE_ANON_KEY=<key>
VITE_API_BASE_URL=http://localhost:3002/api/v1
VITE_LIVEKIT_URL=wss://noaman-rgn77n8s.livekit.cloud
```

- [ ] **Step 3: Final commit**

```bash
git add backend/.env frontend/.env
git commit -m "chore: clean up environment variables for InsForge migration"
```
