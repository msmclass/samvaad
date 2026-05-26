# Copyright 2026 Samvaad Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# ==========================================
# STAGE 1: Build static Next.js frontend
# ==========================================
FROM node:18-alpine AS next-builder

WORKDIR /app

# Copy dependency manifests
COPY samvaad/frontend/package.json samvaad/frontend/package-lock.json* ./

# Install dependencies (ignoring scripts for security)
RUN npm install --legacy-peer-deps --ignore-scripts

# Copy frontend source files and public assets
COPY samvaad/frontend/ ./

# Compile Next.js and perform static export (generates "out/" directory)
RUN npm run build

# ==========================================
# STAGE 2: Build embedded Go backend server
# ==========================================
FROM golang:1.26-alpine AS builder

ARG TARGETPLATFORM
ARG TARGETARCH
RUN echo building for "$TARGETPLATFORM"

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY test/ test/
COPY tools/ tools/
COPY version/ version/

# Clear the placeholder frontend assets folder and copy compiled Next.js assets
RUN rm -rf pkg/service/frontend_out && mkdir -p pkg/service/frontend_out
COPY --from=next-builder /app/out/ pkg/service/frontend_out/

# Compile the Go orchestrator into a static, single binary executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH GO111MODULE=on go build -a -o samvaad-server ./cmd/server

# ==========================================
# STAGE 3: Pack lightweight final runner
# ==========================================
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /workspace/samvaad-server /samvaad-server

# Expose ports for HTTP and TURN/WebRTC
EXPOSE 7880 7881 7882/udp

# Run the binary.
ENTRYPOINT ["/samvaad-server"]
