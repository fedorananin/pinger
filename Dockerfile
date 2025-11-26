# Stage 1: Builder
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy files
COPY go.mod ./
COPY main.go ./

# Compile static binary (CGO_ENABLED=0 decouples from system libs)
# -ldflags="-s -w" strips debug info to reduce size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server main.go

# Stage 2: Final Image (Runner)
FROM alpine:latest

# Install ping (iputils)
RUN apk add --no-cache iputils

WORKDIR /root/

# Copy only the compiled file from the first stage
COPY --from=builder /app/server .

EXPOSE 80

CMD ["./server"]