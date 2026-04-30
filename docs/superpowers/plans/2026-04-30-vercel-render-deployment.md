# Vercel + Render Deployment Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deploy Mini-Meeting to Vercel (frontend) and Render (Go backend) with zero VPS, using git-based deployments.

**Architecture:** Two separate deployments connected via env vars. Frontend on Vercel serves static React SPA. Go backend on Render handles LiveKit tokens, lobby WebSocket, and AI summarizer. Both communicate over HTTPS. InsForge remains as external BaaS.

**Tech Stack:** React 19 + Vite (Vercel Static Site), Go 1.24 + Fiber (Render Web Service via Docker), InsForge SDK (external BaaS)

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `docs/superpowers/plans/2026-04-30-vercel-render-deployment.md` | Create | This plan document |
| `backend/Dockerfile` | Modify | Update existing Dockerfile for Render compatibility (ports, healthcheck) |
| `render.yaml` | Create | Render Blueprint config (Static Site + Web Service) |
| `vercel.json` | Modify | Add SPA rewrites + CORS headers for Vercel |
| `.gitignore` | Modify | Add frontend node_modules, dist, and other deploy artifacts |
| `backend/internal/config/config.go` | Modify | Add `FRONTEND_URL` env var for dynamic CORS |
| `backend/cmd/server/main.go` | Modify | Use dynamic CORS origins from env var |
| `backend/.env.example` | Modify | Add `FRONTEND_URL` to example |

---

### Task 1: Initialize Git Repository and Configure .gitignore

**Files:**
- Modify: `.gitignore` (root)
- Create: `.github/` (empty, for git)

- [ ] **Step 1: Update root `.gitignore` to cover frontend artifacts**

Append to existing `.gitignore`:

```gitignore
# Frontend
frontend/node_modules
frontend/dist
frontend/.vite
frontend/.cache

# General
*.log
.cache
```

- [ ] **Step 2: Verify `.gitignore` is comprehensive**

Run: `cat .gitignore`
Expected: Both Go and frontend exclusion patterns present.

- [ ] **Step 3: Initialize git and make first commit**

Run:
```bash
git init
git add .
git commit -m "chore: initial project setup"
```

- [ ] **Step 4: Push to GitHub**

Run:
```bash
git remote add origin <your-github-repo-url>
git branch -M main
git push -u origin main
```

---

### Task 2: Make Backend CORS Configurable for Production

**Files:**
- Modify: `backend/internal/config/config.go`
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/.env.example`

The CORS config in `main.go:58` is hardcoded to `http://localhost:5173,http://127.0.0.1:5173,https://noaman.cloud`. It needs a dynamic `FRONTEND_URL` env var so we can point it to the Vercel deployment URL.

- [ ] **Step 1: Add `FrontendURL` to ServerConfig**

In `backend/internal/config/config.go`, confirm `ServerConfig` already has `FrontendURL string` (line 28 — it does).

- [ ] **Step 2: Update CORS middleware to use `FrontendURL` from config**

In `backend/cmd/server/main.go`, replace the hardcoded CORS origins:

```go
// Before:
app.Use(cors.New(cors.Config{
    AllowOrigins:     "http://localhost:5173,http://127.0.0.1:5173,https://noaman.cloud",
    AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
    ExposeHeaders:    "Content-Length,Content-Type",
}))
```

```go
// After — comma-separate localhost + production frontend
origins := cfg.Server.FrontendURL
if origins == "" {
    origins = "http://localhost:5173,http://127.0.0.1:5173"
} else {
    origins += ",http://localhost:5173,http://127.0.0.1:5173"
}
app.Use(cors.New(cors.Config{
    AllowOrigins:     origins,
    AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
    AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
    AllowCredentials: true,
    ExposeHeaders:    "Content-Length,Content-Type",
}))
```

- [ ] **Step 3: Verify `FRONTEND_URL` is already in config loading**

In `backend/internal/config/config.go` line 109: `FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173")` — already present. No change needed.

- [ ] **Step 4: Update `.env.example`**

Add `FRONTEND_URL` to the example file under Server Configuration section.

- [ ] **Step 5: Verify build**

Run: `cd backend && go build ./cmd/server`
Expected: No errors.

- [ ] **Step 6: Commit**

```bash
git add backend/
git commit -m "feat: make CORS origins configurable via FRONTEND_URL env var"
```

---

### Task 3: Update Backend Dockerfile for Render

**Files:**
- Modify: `backend/Dockerfile`

The existing Dockerfile is solid but needs minor adjustments: copy the `.env` config to support Render's environment variable injection, and ensure it listens on the correct port.

- [ ] **Step 1: Update Dockerfile to use Render's PORT env var**

Render assigns a dynamic port via `$PORT`. The backend already reads `PORT` from env (config.go line 107). Confirm this works.

- [ ] **Step 2: Update the Dockerfile for Render compatibility**

```dockerfile
# Multi-stage build for Go backend
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy migrations if they exist
COPY --from=builder /app/migrations ./migrations

# Expose port (Render will use $PORT)
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/api/v1/health || exit 1

# Run the binary
CMD ["./main"]
```

- [ ] **Step 3: Verify Dockerfile builds locally**

Run: `cd backend && docker build -t mini-meeting-backend .`
Expected: Build succeeds without errors.

- [ ] **Step 4: Commit**

```bash
git add backend/Dockerfile
git commit -m "chore: update Dockerfile for Render deployment"
```

---

### Task 4: Create Render Blueprint Config

**Files:**
- Create: `render.yaml`

This file enables one-click deployment to Render via Blueprint. Defines two services: a Static Site for the frontend and a Web Service for the Go backend.

- [ ] **Step 1: Create `render.yaml`**

```yaml
services:
  # Go Backend
  - type: web
    name: mini-meeting-backend
    env: docker
    plan: free
    region: oregon
    healthCheckPath: /api/v1/health
    envVars:
      - key: PORT
        fromService:
          type: web
          name: mini-meeting-backend
          property: port
      - key: ENV
        value: production
      - key: FRONTEND_URL
        fromService:
          type: web
          name: mini-meeting-frontend
          property: domain
      - key: JWT_SECRET
        sync: false
      - key: JWT_EXPIRATION
        value: 24h
      - key: DATABASE_URL
        fromDatabase:
          name: mini-meeting-db
          property: connectionString
      - key: REDIS_HOST
        fromRedis
        property: host
      - key: REDIS_PORT
        fromRedis
        property: port
      - key: REDIS_USERNAME
        fromRedis
        property: username
      - key: REDIS_PASSWORD
        fromRedis
        property: password
      - key: LIVEKIT_API_KEY
        sync: false
      - key: LIVEKIT_API_SECRET
        sync: false
      - key: LIVEKIT_URL
        sync: false
      - key: OPEN_ROUTER_API_KEY
        sync: false
      - key: OPEN_ROUTER_BASE_URL
        value: https://openrouter.ai/api/v1
      - key: OPEN_ROUTER_MODEL
        value: nvidia/nemotron-3-nano-30b-a3b:free
      - key: OPEN_ROUTER_TIMEOUT
        value: 300s
      - key: OPEN_ROUTER_MAX_TOKENS
        value: 4096
      - key: BRAVO_API_KEY
        sync: false
      - key: BREVO_SENDER_EMAIL
        sync: false
      - key: BREVO_SENDER_NAME
        value: Mini Meeting
      - key: SUMMARIZER_CHUNK_DURATION_SECONDS
        value: "20"
      - key: SUMMARIZER_TEMP_DIR
        value: /tmp/summarizer

databases:
  - name: mini-meeting-db
    plan: free
    region: oregon
    ipAllowList: []

redis:
  - name: mini-meeting-redis
    plan: free
    region: oregon
```

> **Note:** Render Blueprints don't support cross-service env vars (like frontend domain → backend CORS). You'll need to set `FRONTEND_URL` manually in the Render dashboard after deploying the frontend.

- [ ] **Step 2: Commit**

```bash
git add render.yaml
git commit -m "chore: add Render Blueprint config"
```

---

### Task 5: Configure Vercel Deployment

**Files:**
- Modify: `vercel.json`
- Create: `frontend/.vercelignore`

Vercel auto-detects Vite projects, but we need explicit SPA routing and to exclude unnecessary files.

- [ ] **Step 1: Update root `vercel.json`**

The existing `vercel.json` is in the `frontend/` directory. We need a root-level one that tells Vercel where the frontend lives. However, since Vercel deploys from a single directory, we'll configure it via the Vercel dashboard instead.

Create root `vercel.json` with empty config (Vercel auto-detects):

```json
{
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

- [ ] **Step 2: Create `frontend/.vercelignore`**

```
node_modules
dist
.git
*.local
```

- [ ] **Step 3: Commit**

```bash
git add vercel.json frontend/.vercelignore
git commit -m "chore: add Vercel config"
```

---

### Task 6: Push and Deploy to Vercel

**Files:** None (dashboard operations)

- [ ] **Step 1: Push all changes to GitHub**

```bash
git push origin main
```

- [ ] **Step 2: Connect to Vercel**

1. Go to https://vercel.com/new
2. Import your GitHub repo
3. Framework preset: **Vite** (auto-detected)
4. Root Directory: `frontend`
5. Build Command: `npm run build`
6. Output Directory: `dist`
7. Install Command: `npm install`

- [ ] **Step 3: Set environment variables in Vercel dashboard**

| Variable | Value |
|----------|-------|
| `VITE_API_BASE_URL` | `https://mini-meeting-backend.onrender.com/api/v1` (your Render URL) |
| `VITE_INSFORGE_URL` | `https://7c85mvkc.us-east.insforge.app` |
| `VITE_INSFORGE_ANON_KEY` | `eyJhbGciOiJIUzI1NiIs...` (your key) |

- [ ] **Step 4: Deploy**

Click "Deploy". Wait for build to complete. Note the deployed URL (e.g., `https://mini-meeting.vercel.app`).

---

### Task 7: Deploy to Render

**Files:** None (dashboard operations)

- [ ] **Step 1: Deploy via Render Blueprint or Dashboard**

**Option A — Dashboard (easiest):**

1. Go to https://dashboard.render.com
2. Click "New +" → "Web Service"
3. Connect your GitHub repo
4. Configure:
   - **Name:** `mini-meeting-backend`
   - **Region:** Oregon (same as Vercel if possible)
   - **Branch:** `main`
   - **Root Directory:** `backend`
   - **Runtime:** Docker
   - **Plan:** Free

5. Set environment variables:

| Variable | Value |
|----------|-------|
| `PORT` | `3000` |
| `ENV` | `production` |
| `FRONTEND_URL` | `https://mini-meeting.vercel.app` (your Vercel URL) |
| `JWT_SECRET` | Same as InsForge's JWT secret |
| `JWT_EXPIRATION` | `24h` |
| `DATABASE_URL` | Your PostgreSQL connection string |
| `REDIS_HOST` | Your Redis host |
| `REDIS_PORT` | Your Redis port |
| `REDIS_USERNAME` | Your Redis username |
| `REDIS_PASSWORD` | Your Redis password |
| `LIVEKIT_API_KEY` | Your LiveKit API key |
| `LIVEKIT_API_SECRET` | Your LiveKit API secret |
| `LIVEKIT_URL` | `wss://your-livekit-server.com` (LiveKit Cloud URL) |
| `OPEN_ROUTER_API_KEY` | Your OpenRouter key |
| `OPEN_ROUTER_BASE_URL` | `https://openrouter.ai/api/v1` |
| `OPEN_ROUTER_MODEL` | `nvidia/nemotron-3-nano-30b-a3b:free` |
| `BRAVO_API_KEY` | Your Brevo API key |
| `BREVO_SENDER_EMAIL` | Your sender email |

**Option B — Blueprint (advanced):**

1. Go to https://dashboard.render.com/select-repo?repo=...
2. Select your repo
3. Render will read `render.yaml` and provision all services

- [ ] **Step 2: Deploy**

Click "Create Web Service". Render will build the Docker image and deploy.

- [ ] **Step 3: Note the Render backend URL**

e.g., `https://mini-meeting-backend-xxxx.onrender.com`

---

### Task 8: Update Frontend API URL and Redeploy

- [ ] **Step 1: Update `VITE_API_BASE_URL` in Vercel dashboard**

Go to Vercel → Project Settings → Environment Variables → Update `VITE_API_BASE_URL` to your actual Render URL:

```
https://mini-meeting-backend-xxxx.onrender.com/api/v1
```

- [ ] **Step 2: Redeploy frontend**

In Vercel → Deployments → Click the latest deployment → "Redeploy".

---

### Task 9: Update `FRONTEND_URL` in Render Dashboard

- [ ] **Step 1: Set `FRONTEND_URL` env var on Render**

Go to Render → Your service → Environment → Add:

```
FRONTEND_URL = https://mini-meeting.vercel.app (or your custom domain)
```

- [ ] **Step 2: Trigger redeploy**

Render → Deploys → "Manual Deploy" → "Deploy latest commit"

---

### Task 10: Verify Deployment

- [ ] **Step 1: Check health endpoint**

Open: `https://mini-meeting-backend-xxxx.onrender.com/api/v1/health`
Expected: `{"status":"ok","message":"Server is running"}`

- [ ] **Step 2: Check frontend loads**

Open: `https://mini-meeting.vercel.app`
Expected: Landing page with no console errors.

- [ ] **Step 3: Test login flow**

1. Navigate to `/login`
2. Login with existing account
3. Verify no CORS errors in console

- [ ] **Step 4: Test registration flow**

1. Navigate to `/register`
2. Create new account
3. Verify auto-login and redirect to `/dashboard`

- [ ] **Step 5: Test meeting creation**

1. Create a meeting
2. Copy the meeting code
3. Join from another browser
4. Verify video/audio works (requires LiveKit credentials)

---

## Post-Deployment Notes

### Render Free Tier Limitations
- Web Service sleeps after 15 min of inactivity
- First request after sleep takes ~30-60 seconds to cold-start
- Consider upgrading to paid plan ($7/mo) for always-on

### Vercel Free Tier
- Unlimited deployments and bandwidth
- Global CDN (excellent performance)
- Preview deployments for every PR branch

### LiveKit Cloud
- Use `wss://` URL (not `ws://`) for production
- Example: `wss://your-project.us-east.livekit.cloud`
- Free tier: 250 GB data, 500 concurrent users

### CORS
- Backend CORS now reads from `FRONTEND_URL` env var
- If you add a custom domain, update both `FRONTEND_URL` (Render) and the domain in Vercel

### Database
- InsForge handles auth + meeting tables automatically
- Go backend's PostgreSQL only stores summarizer data (sessions, chunks, transcripts)
- If you don't need summarization, you can use SQLite for the backend instead

---

## Quick Env Var Checklist

### Vercel (Frontend)
- [ ] `VITE_API_BASE_URL` = Render backend URL + `/api/v1`
- [ ] `VITE_INSFORGE_URL` = InsForge project URL
- [ ] `VITE_INSFORGE_ANON_KEY` = InsForge anon key
- [ ] `VITE_LIVEKIT_URL` = LiveKit Cloud wss:// URL

### Render (Backend)
- [ ] `PORT` = `3000`
- [ ] `FRONTEND_URL` = Vercel domain
- [ ] `JWT_SECRET` = Same as InsForge
- [ ] `DATABASE_URL` = PostgreSQL connection string (or free Render DB)
- [ ] `REDIS_HOST/PORT/USERNAME/PASSWORD` = Redis Cloud credentials
- [ ] `LIVEKIT_API_KEY` + `LIVEKIT_API_SECRET` + `LIVEKIT_URL`
- [ ] `OPEN_ROUTER_API_KEY`
- [ ] `BRAVO_API_KEY` + `BREVO_SENDER_EMAIL`

### External Services
- [ ] LiveKit Cloud account created
- [ ] Redis Cloud account created
- [ ] OpenRouter account with credits
- [ ] Brevo account with sender email verified
