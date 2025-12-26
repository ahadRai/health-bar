FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy shared code
COPY shared/ ./shared/

# Copy auth service code
COPY services/auth/ ./services/auth/

# Build the application
WORKDIR /app/services/auth
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/auth/main .
COPY --from=builder /app/services/auth/.env .env

EXPOSE 8001

CMD ["./main"]