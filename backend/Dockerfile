# syntax=docker/dockerfile:1

ARG commit_id=dev
ARG GOSMEE_VERSION=0.28.0

FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy only necessary source files
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o gosmee-web-server ./cmd/server

FROM alpine:latest AS runtime-base

ARG commit_id
ARG GOSMEE_VERSION
ENV COMMIT_ID=${commit_id}

# Install gosmee, configure non-root user
RUN apk add --no-cache ca-certificates wget tar \
    && addgroup -g 65532 -S nonroot \
    && adduser -u 65532 -S nonroot -G nonroot \
    && wget -q https://github.com/chmouel/gosmee/releases/download/v${GOSMEE_VERSION}/gosmee_${GOSMEE_VERSION}_linux_x86_64.tar.gz \
    && tar -xzf gosmee_${GOSMEE_VERSION}_linux_x86_64.tar.gz -C /usr/local/bin gosmee \
    && rm gosmee_${GOSMEE_VERSION}_linux_x86_64.tar.gz \
    && chmod +x /usr/local/bin/gosmee \
    && apk del wget tar

WORKDIR /app

# Prepare data directory and gosmee cache directory with restrictive permissions
RUN mkdir -p /data /home/nonroot/.cache/gosmee \
    && chown -R 65532:65532 /data /home/nonroot \
    && chmod 700 /data \
    && chmod 755 /home/nonroot/.cache/gosmee

USER 65532:65532

ENV HOME=/home/nonroot

EXPOSE 8080

ENTRYPOINT ["./gosmee-web-server"]

FROM runtime-base AS dev

# Copy pre-built binary from local build context
COPY --chown=65532:65532 gosmee-web-server /app/gosmee-web-server

FROM runtime-base AS prod

# Copy binary from builder stage for production image
COPY --from=builder --chown=65532:65532 /app/gosmee-web-server /app/gosmee-web-server
