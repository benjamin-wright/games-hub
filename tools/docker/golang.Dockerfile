# ─── Build Stage ───────────────────────────────────────────────────────────────
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder

# CMD_PATH is the path to the Go command directory (e.g. ./cmd)
ARG CMD_PATH=./cmd

WORKDIR /workspace

# Cache module downloads separately from source
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/app ${CMD_PATH}

# ─── Runtime Stage ─────────────────────────────────────────────────────────────
FROM gcr.io/distroless/static:nonroot AS runtime

# Copy the compiled binary and always expose it under a fixed name inside the image
COPY --from=builder /out/app /usr/local/bin/app

USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/app"]