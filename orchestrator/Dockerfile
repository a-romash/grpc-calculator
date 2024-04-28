# syntax=docker/dockerfile:1

FROM golang:1.22.1-alpine AS builder

# Set destination for COPY
WORKDIR /

COPY . .

# Download Go modules
COPY go.* ./
RUN go mod download

# Build
RUN GOOS=linux go build -o server ./cmd/orchestrator/main.go

FROM alpine:3.18

WORKDIR /

COPY --from=builder . .

EXPOSE 8080
EXPOSE 44045

CMD ["./server"]