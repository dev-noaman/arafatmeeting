.PHONY: test test-backend test-frontend test-all build-backend build-frontend

test: test-all

test-backend:
	cd backend && go test ./... -v -count=1

test-frontend:
	cd frontend && npx vitest run

test-e2e:
	cd frontend && npx playwright test

test-all: test-backend test-frontend

build-backend:
	cd backend && go build -o server.exe ./cmd/server

build-frontend:
	cd frontend && npm run build
