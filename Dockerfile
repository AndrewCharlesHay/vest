FROM golang:1.24-alpine AS builder


WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Change ownership of the directory/binary to the new user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

EXPOSE 8080
CMD ["./main"]
