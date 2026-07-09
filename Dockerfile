# Copyright 2025 Google LLC
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

# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc libc-dev

# Set the current working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the web launcher/agent server executable
# Compiles cmd/launcher/universal or standard binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/kronos-agent ./examples/web/main.go

# Stage 2: Minimal run container
FROM alpine:3.21

# Add standard CA certs for outbound API-Sports and Gemini API calls
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /

# Copy the compiled Go binary from Stage 1
COPY --from=builder /app/kronos-agent /kronos-agent

# Expose port for REST/JSON API service
EXPOSE 8080

# Environment variables with safe defaults (No Pathos)
ENV PORT=8080
ENV GOOGLE_API_KEY=""

# Set the entrypoint to run our Go backend web launcher
ENTRYPOINT ["/kronos-agent", "web", "api"]
