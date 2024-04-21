# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder

# Set destination for COPY
WORKDIR /

COPY ./cmd/agent ./cmd/agent/
COPY ./pkg ./pkg/
COPY ./internal/agent ./internal/agent/
COPY ./internal/model ./internal/model/
COPY ./.env ./

# Download Go modules
COPY go.* ./
RUN go mod download

# Build
RUN GOOS=linux go build -o agent ./cmd/agent/main.go

FROM alpine:3.18

WORKDIR /

COPY --from=builder . .

CMD ["./agent"]