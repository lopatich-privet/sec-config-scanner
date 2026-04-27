# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /config-analyzer ./cmd/main.go

# Stage 2: Runtime (distroless)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /config-analyzer /config-analyzer

USER nonroot:nonroot

ENTRYPOINT ["/config-analyzer"]
