# Gonteractive

Self-hosted, single-user PWA for learning Go: curriculum topics with AI-generated lessons, quizzes (single/multi choice + AI-graded open questions), coding drills verified with `go test`, FSRS spaced repetition, a daily quiz with Web Push, and an "ask anything" chat. Everything AI goes through pluggable providers.

## Stack

- Backend: Go (stdlib `net/http` routing), SQLite via `modernc.org/sqlite` (no CGO), `sqlc` for queries, `goose` migrations run on startup.
- Frontend: React 19 + Vite + TypeScript + Tailwind v4 + shadcn/ui (radix-nova), TanStack Query, React Router v8, CodeMirror 6, shiki, vite-plugin-pwa (injectManifest, custom `src/sw.ts` handles push).
- The built SPA is embedded into the Go binary (`web/embed.go` → `web/dist`). One binary + one `.db` file.
- Deploy: `compose.yaml` = app + Tailscale sidecar (`network_mode: service:tailscale`, HTTPS via `deploy/tailscale/serve.json`).

## Layout

```
cmd/gonteractive/      main: wiring, graceful shutdown
internal/
  ai/                  Provider interface, Schema + GenerateJSON (validate + one retry)
    anthropic/         anthropic-sdk-go, beta messages, output_config.format, server-side fallbacks
    openaicompat/      openai-go chat completions + json_schema (OpenAI, OpenRouter, Ollama…)
    acp/               Agent Client Protocol client (coder/acp-go-sdk): long-lived agent process, session per call
  providers/           provider CRUD, per-task routing (chat/generate/grade), live instance cache
  curriculum/          topics.yaml (embedded) — the course outline
  lesson/ quiz/ drill/ chat/ progress/ daily/   feature packages: service + http.go
  store/               Open + migrations/, queries/ (sqlc input), db/ (sqlc output, generated)
  httpx/               JSON/error helpers, SSE writer
  auth/ settings/ tutor/ server/ config/
web/src/
  app/                 router, providers, layout
  components/          shared (Markdown, CodeBlock, …); components/ui = shadcn (generated)
  features/<name>/     index.ts + api.ts at root; components/ hooks/ types/ utils/ inside
```

## Commands

- `make dev-api` (API on :8080, password `dev`) + `make dev-web` (Vite on :5173, proxies `/api`).
- `make test` — `go test -race ./...`, oxlint, `tsc -b`.
- `make build` — builds web then `bin/gonteractive`.
- `make generate` — regenerate `internal/store/db` after editing `internal/store/queries/*.sql` or migrations.
- `make docker` — `docker compose up -d --build` (needs `.env`, see `.env.example`).

## Conventions

- No code comments by default; doc comments only for genuinely non-obvious contracts.
- Never edit `internal/store/db/*` by hand. Schema changes = new goose migration file, never edit applied ones.
- Feature packages own their prompts and JSON schemas. Schemas must stay compatible with OpenAI strict mode and Anthropic structured outputs: every property required, `additionalProperties: false`, no numeric/length constraints — enforce those in a `Validate()` method instead (GenerateJSON retries on failure).
- Every AI call picks its provider via `providers.Service.For(ctx, task)`; return `providers.ErrNotConfigured` (412) when none exists.
- Streaming endpoints use `httpx.NewSSE` with events `delta` / `done` / `error`; the web client consumes them via `stream()` in `web/src/lib/api.ts`.
- Times are unix seconds in SQLite; local dates (streak, daily quiz) use the server `TZ`.
- UI is bilingual (en/ru): every user-facing string goes through `t()` from `@/i18n` with keys in `web/src/i18n/locales/en.ts` (source of truth, typed) and `ru.ts`. Plurals use `{ one, few, many, other }` objects. The chosen language is stored server-side (`/api/settings/interface`, also used for push texts) and cached in localStorage. Curriculum titles/objectives need `title_ru`/`objectives_ru` in `topics.yaml` — loading panics otherwise. Backend errors meant for users carry a `code` (`httpx.Coded`) that the client maps to a translation in `lib/api.ts`.
- Theme is One Light / One Dark: tokens in `web/src/index.css`, extra Tailwind colors `hue-*`, `success`, `warning`, `heat-0..4`. Data marks (mastery/module bars) use the single primary hue; the activity heatmap uses the validated `heat-*` ramp; colorful accents are for icons and badges only.
- ACP agents in the image: `codex-acp` (login: `docker compose exec -u app app codex login --device-auth`) and `claude-agent-acp` (`CLAUDE_CODE_OAUTH_TOKEN` from `claude setup-token`).
