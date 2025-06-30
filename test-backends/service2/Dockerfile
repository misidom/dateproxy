FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY main.go .
RUN go mod init service2 && go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service2 .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory and set ownership
RUN mkdir -p /app && chown -R appuser:appgroup /app

WORKDIR /app

# Copy binary with correct permissions
COPY --from=builder --chown=appuser:appgroup /app/service2 .

# Ensure the binary is executable
RUN chmod +x service2

# Use non-root user
USER appuser

EXPOSE 8080
CMD ["./service2"]