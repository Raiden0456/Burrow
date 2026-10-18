# Burrow

Shown to users as **Burrow**; the repo, Go module, image and env names still say `gonteractive`.

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
  jobs/                background jobs for long AI work (lesson/quiz/drill/daily), detached from requests
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
- Long AI generation (lessons, topic quizzes, drills, the manual daily quiz) never runs inside a request: start it with `jobs.Manager.Start` and return the job (202). One running job per kind+topic; lesson text is exposed as `partial` while running; jobs left running by a restart are marked failed on startup. The web app polls `/api/jobs`, and `JobsWatcher` toasts when a job finishes.
- Short interactive AI output (quiz hints/deep dives, drill help, chat) streams over SSE and is persisted in the `hints` / chat tables. A hint before answering caps the FSRS rating at Hard.
- Quiz answer drafts (`quiz_items.draft`) and drill code (`drills.draft`) are autosaved server-side so work survives closing the tab or switching devices.
- Streaming endpoints use `httpx.NewSSE` with events `delta` / `done` / `error`; the web client consumes them via `stream()` in `web/src/lib/api.ts`.
- Times are unix seconds in SQLite; local dates (streak, daily quiz) use the server `TZ`.
- UI is bilingual (en/ru): every user-facing string goes through `t()` from `@/i18n` with keys in `web/src/i18n/locales/en.ts` (source of truth, typed) and `ru.ts`. Plurals use `{ one, few, many, other }` objects. The chosen language is stored server-side (`/api/settings/interface`, also used for push texts) and cached in localStorage. Curriculum titles/objectives need `title_ru`/`objectives_ru` in `topics.yaml` — loading panics otherwise. Backend errors meant for users carry a `code` (`httpx.Coded`) that the client maps to a translation in `lib/api.ts`.
- Tokens live in `web/src/index.css`, extra Tailwind colors `hue-*`, `success`, `warning`, `*-solid`, `heat-0..4`. Light theme uses Radix Colors (slate neutrals, indigo primary; `hue-*`, `success`, `warning`, `destructive` = step 11 so they read as text and as `/10`–`/20` tints; `success-solid`/`warning-solid` = step 9 for bars, dots and other solid fills, since step 11 amber turns brown as a fill). Dark theme stays One Dark, where the solid tokens equal the regular ones. Data marks (mastery/module bars) use the single primary hue; the activity heatmap uses the validated `heat-*` ramp; colorful accents are for icons and badges only.
- Wallpaper (same as Pulse): one image in `$APP_DATA_DIR/wallpaper`, served at `/api/wallpaper/image?v=<mtime>`; dimming is `wallpaperDim` in `/api/settings/interface`. When set, `<html>` gets `data-wallpaper`: border/muted/accent tokens turn translucent, cards and `[data-glass]` surfaces become blurred glass (`index.css`, outside `@layer` so it beats `bg-card`). Content must never sit bare on the wallpaper — wrap loose blocks in `data-glass` and use the `wallpaper:` variant for extra padding. Glass creates a containing block for `position: fixed`, so fixed bars inside cards go through a portal (`ActionBar`).
- First render staggers in via `enter(i)` from `lib/enter.ts`; spread it straight onto the Card (card components pass `...props` through), never a wrapper — opacity on a wrapper breaks the card's backdrop blur mid-animation.
- ACP agents in the image: `codex-acp` (login: `docker compose exec -u app app codex login --device-auth`) and `claude-agent-acp` (`CLAUDE_CODE_OAUTH_TOKEN` from `claude setup-token`).
