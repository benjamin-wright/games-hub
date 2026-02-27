# syntax=docker/dockerfile:1

# ─── Build Stage ───────────────────────────────────────────────────────────────
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder

# CMD_PATH is the path to the Go command directory (e.g. ./cmd)
ARG CMD_PATH=./cmd
# BINARY is the name of the compiled output binary
ARG BINARY=app

WORKDIR /workspace

# Cache module downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/${BINARY} ${CMD_PATH}

# ─── Runtime Stage ─────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot AS runtime

ARG BINARY=app

# Copy the compiled binary and always expose it under a fixed name inside the image
COPY --from=builder /out/${BINARY} /usr/local/bin/app

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/app"]
