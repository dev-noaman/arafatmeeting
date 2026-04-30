# Mini Meeting - Setup Instructions

## Quick Start

### 1. Install Go (Required for Backend)

Download and install Go from: https://go.dev/dl/

**Windows Installation:**
1. Download the `.msi` installer
2. Run the installer and follow the prompts
3. Restart your terminal after installation
4. Verify installation: `go version`

### 2. Install Node.js and pnpm (if not already installed)

**Node.js:** https://nodejs.org/ (LTS version recommended)

**pnpm:**
```bash
npm install -g pnpm
```

### 3. Ensure PostgreSQL is Running

PostgreSQL must be running on port 5432. The application will create the `mini_meeting` database automatically.

### 4. Setup Backend

```bash
cd backend

# Install Go dependencies
go mod tidy
```

### 5. Setup Frontend

```bash
cd frontend

# Install dependencies
pnpm install
```

### 6. Start the Application

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/server/main.go
```

The backend will:
- Connect to PostgreSQL database
- Run migrations to create tables
- Start on http://localhost:3002

**Terminal 2 - Frontend:**
```bash
cd frontend
pnpm run dev
```

The frontend will start on http://localhost:5173

## Testing the Application

1. Navigate to http://localhost:5173
2. You should see the Landing page
3. Click "Sign In" to go to the login page
4. Register a new account with email/password
5. Create a meeting and test the features

## Troubleshooting

### "Go not found"
Make sure Go is installed and added to your PATH. Restart your terminal after installation.

### "Module not found" errors
Run `go mod tidy` in the backend folder to install missing dependencies.

### Frontend shows blank page
1. Make sure backend is running on port 3002
2. Check browser console for errors (F12)
3. Verify the frontend build: `pnpm run build`

### Database errors
Make sure PostgreSQL is running on port 5432. The database `mini_meeting` will be created automatically on first run.

## Database

PostgreSQL database: `mini_meeting` on localhost:5432

To view the database:
- **Command line:** `psql -U postgres -d mini_meeting`
- **GUI Tools:** pgAdmin, DBeaver, TablePlus
