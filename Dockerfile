# --- build stage ---
FROM docker.io/golang:1.25 as builder
WORKDIR /src
COPY src/ ./
RUN go mod download && \
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/action .

# --- runtime stage ---
FROM debian:bookworm-slim
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tar git && rm -rf /var/lib/apt/lists/*

# kustomize is downloaded at runtime by the action according to env KUSTOMIZE_VERSION
COPY --from=builder /out/action /usr/local/bin/action
ENTRYPOINT ["/usr/local/bin/action"]
