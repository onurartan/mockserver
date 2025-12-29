# Step 1: Build
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o mockserver .

# Step 2: Runtime
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /root/

COPY --from=builder /app/mockserver .

EXPOSE 5000

# Start Server
ENTRYPOINT ["./mockserver", "start", "--config", "/example/mockserver.yaml"]