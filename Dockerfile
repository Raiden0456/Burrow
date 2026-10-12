FROM --platform=$BUILDPLATFORM node:24-alpine AS web
WORKDIR /src/web
RUN corepack enable
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY web/embed.go ./web/embed.go
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/gonteractive ./cmd/gonteractive

FROM golang:1.27-alpine
ARG ACP_AGENTS="@agentclientprotocol/codex-acp @agentclientprotocol/claude-agent-acp @openai/codex"
RUN apk add --no-cache ca-certificates tzdata nodejs npm su-exec \
    && if [ -n "$ACP_AGENTS" ]; then npm install -g $ACP_AGENTS && npm cache clean --force; fi \
    && adduser -D -h /home/app app \
    && mkdir -p /data /home/app/.codex \
    && chown -R app:app /data /home/app

ENV HOME=/home/app \
    APP_DATA_DIR=/data \
    APP_GO_CACHE=/home/app/.cache/go-build \
    CODEX_HOME=/home/app/.codex

USER app
RUN mkdir -p /tmp/warm && cd /tmp/warm \
    && printf 'module warm\n\ngo 1.27\n' > go.mod \
    && printf 'package warm\n\nfunc F() int { return 1 }\n' > warm.go \
    && printf 'package warm\n\nimport "testing"\n\nfunc TestF(t *testing.T) { if F() != 1 { t.Fatal() } }\n' > warm_test.go \
    && GOCACHE=$APP_GO_CACHE GOFLAGS=-mod=mod GOPROXY=off GOTOOLCHAIN=local CGO_ENABLED=0 go test ./... \
    && rm -rf /tmp/warm

USER root
COPY deploy/docker-entrypoint.sh /usr/local/bin/docker-entrypoint
COPY --from=build /out/gonteractive /usr/local/bin/gonteractive
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["docker-entrypoint"]
CMD ["gonteractive"]
