# Multi-stage build for Go app
FROM golang:1.20-alpine AS builder
WORKDIR /src

# Use a GOPROXY that works in restricted networks (adjust if needed)
ENV GOPROXY=https://goproxy.cn,direct

# Cache modules
COPY go.mod go.sum ./
RUN go env -w GOPROXY=${GOPROXY} && go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

# Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
