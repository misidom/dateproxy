FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dateproxy .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user and group
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Create app directory and set ownership
RUN mkdir -p /app && chown -R appuser:appgroup /app

WORKDIR /app

# Copy binary and config with correct permissions
COPY --from=builder --chown=appuser:appgroup /app/dateproxy .
COPY --chown=appuser:appgroup config.yaml .

# Ensure the binary is executable
RUN chmod +x dateproxy

# Use non-root user
USER appuser

EXPOSE 8080

CMD ["./dateproxy"]