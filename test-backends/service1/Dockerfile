FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY main.go .
COPY go.mod go.sum ./
RUN go mod init service1 && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service1 .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory and set ownership
RUN mkdir -p /app && chown -R appuser:appgroup /app

WORKDIR /app

# Copy binary with correct permissions
COPY --from=builder --chown=appuser:appgroup /app/service1 .

# Ensure the binary is executable
RUN chmod +x service1

# Use non-root user
USER appuser

EXPOSE 8080
CMD ["./service1"]