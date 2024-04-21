# syntax=docker/dockerfile:1

FROM golang:1.21-alpine AS builder

# Set destination for COPY
WORKDIR /

COPY ./cmd/orchestrator ./cmd/orchestrator/
COPY ./pkg ./pkg/
COPY ./internal/orchestrator ./internal/orchestrator/
COPY ./internal/model ./internal/model/
COPY ./.env ./

# Download Go modules
COPY go.* ./
RUN go mod download

# Build
RUN GOOS=linux go build -o server ./cmd/orchestrator/main.go

FROM alpine:3.18

WORKDIR /

COPY --from=builder . .

EXPOSE 8080

CMD ["./server"]