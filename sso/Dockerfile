# syntax=docker/dockerfile:1

FROM golang:1.22.1-alpine AS builder

# Set destination for COPY
WORKDIR /

COPY . .

# Download Go modules
COPY go.* ./
RUN go mod download

# Build
RUN GOOS=linux go build -o sso ./cmd/sso/main.go

FROM alpine:3.18

WORKDIR /

COPY --from=builder . .

EXPOSE 44044

CMD ["./sso"]