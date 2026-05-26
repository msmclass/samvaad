# ==========================================
# STAGE 1: Build static Next.js frontend
# ==========================================
FROM node:18-alpine AS next-builder

WORKDIR /app

# Copy dependency manifests
COPY frontend/package.json frontend/package-lock.json* ./

# Install dependencies (ignoring scripts for security)
RUN npm install --frozen-lockfile --ignore-scripts

# Copy frontend source files
COPY frontend/ ./

# Compile Next.js and perform static export (generates "out/" directory)
RUN npm run build

# ==========================================
# STAGE 2: Build embedded Go backend server
# ==========================================
FROM golang:1.26-alpine AS go-builder

WORKDIR /app

# Copy the backend source
COPY backend/main.go ./

# Clear placeholder files and copy compiled Next.js assets to the embed folder
RUN rm -rf frontend_out && mkdir -p frontend_out
COPY --from=next-builder /app/out/ ./frontend_out/

# Compile the Go orchestrator into a static, single binary executable
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o samvaad-app main.go

# ==========================================
# STAGE 3: Pack lightweight final runner
# ==========================================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

COPY --from=go-builder /app/samvaad-app /samvaad-app

# Default environment configuration
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/samvaad-app"]
