# Multi-stage build for Go app
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Use a GOPROXY that works in restricted networks
ENV GOPROXY=https://goproxy.cn,direct

# Cache modules
COPY go.mod go.sum ./
RUN go env -w GOPROXY=${GOPROXY} && go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server

# Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates curl
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/openapi.yaml /app/openapi.yaml
EXPOSE 8080
ENTRYPOINT ["/app/server"]
