# Persian AI Chat Platform — Phase 1 Design

Date: 2026-09-08
Status: Approved for planning

## 1. Product

A general-purpose Claude-like chat assistant for the Iranian market, bilingual
(Persian RTL / English LTR), with selectable models spanning self-hosted open
models and frontier models reached through a foreign relay.

Phase 2, out of scope here, adds web browsing and financial-market advice. The
design keeps that seam open but builds none of it.

### In scope (Phase 1)

- Email + password auth, sessions
- Conversations with full history, branching, search
- Streaming chat with a model picker (local + frontier)
- Custom instructions and per-user memory
- Persian prompt library
- File upload with text extraction
- Quota and usage metering
- Persian/English UI with direction switching

### Out of scope (later phases)

Payment/Zarinpal, phone OTP, voice input, web browsing, financial features,
mobile app, semantic search, ML moderation.

Dropping payment from Phase 1 removes the requirement for a registered Iranian
legal entity and an enamad certificate, which is a launch blocker, not merely
a feature.

## 2. Model access strategy

Hybrid, and the hybrid is the point:

- **Free tier** runs self-hosted open models (Qwen3 14B/32B on vLLM) inside Iran.
- **Paid-equivalent tier** reaches OpenAI and Anthropic through a relay hosted
  abroad. In Phase 1 access is granted manually to invited users, since there
  is no payment.

The two tiers fail independently. When the relay is unreachable the product
degrades to local models rather than going dark. This is a designed state, not
an incident: relay health is polled, a circuit breaker trips, `GET /models`
reports availability, and the picker greys out frontier models before the user
hits send.

Provider credentials live **on the relay**, never in Iran. The gateway
authenticates to the relay with its own rotatable credential (mTLS preferred).
If the Iranian side is compromised, no provider account is exposed.

Run at least two relays on different hosts and accounts. A single relay is a
single point of total frontier failure, and being blocked or rate-limited is
the expected steady state.

## 3. Architecture

    Browser (React SPA, fa/en RTL)
          | HTTPS + SSE
    Go API Gateway  (only public surface)
      |-- Postgres, Redis, object storage (MinIO)
      |-- vLLM (self-hosted GPU)        free tier
      |-- Extraction worker (Python)    async
      +-- Relay (foreign VPS) --> OpenAI / Anthropic

The gateway is the only publicly reachable component. Inference, relay,
extraction and storage are internal.

## 4. Stack

| Layer | Choice |
|---|---|
| Web | React 19 + Vite + TypeScript |
| UI | Tailwind v4 + shadcn/ui, logical properties only |
| i18n | i18next; Vazirmatn self-hosted; dayjs-jalali |
| API | Go 1.24, chi, fx (DI), viper (config) |
| Data | Postgres 17 (pgx + RowToStructByName), Redis 7 |
|  Objects | MinIO (self-hosted, S3-compatible) |
| Jobs | River (Postgres-backed) |
| Migrations | goose |
| Inference | vLLM (Qwen3), relay to OpenAI/Anthropic |
| Extraction | Python: PyMuPDF fast path, Surya OCR fallback |
| Ops | Docker Compose on 2-3 VMs, Caddy, MinIO, Prometheus/Grafana/Loki |

Rationale for the load-bearing picks:

- **Go over Java reactive.** Not on I/O capability, where they tie. Go wins on
  cost of ownership: no colored-function discipline, linear stack traces,
  explicit context propagation, small footprint. Virtual threads have removed
  reactive's remaining advantage on the JVM anyway.
- **chi over Fiber.** Fiber's pooled `*fiber.Ctx` must not outlive the handler,
  which directly conflicts with detached generation contexts. Fiber is also not
  `net/http`, costing `httptest`, OTel middleware, and `ResponseController`.
  Its raw throughput advantage is irrelevant when requests are held open for
  seconds waiting on tokens.
- **pgx + hand-written models over sqlc.** sqlc emits a struct per query row
  shape, which conflicts with the single-model rule.

## 5. Layering and code rules

    model  <- domain <- dto/mapper <- controller
      ^        ^
    service (declares ports) <- repository, client (implement implicitly)
                                        ^
                                  infrastructure

| Layer | May contain | Must never contain |
|---|---|---|
| Controller | decode DTO, call mapper, call service, map error, write | business branching, SQL, provider calls |
| Mapper | DTO <-> model/domain, pure functions | I/O |
| Service | orchestration, transactions, port calls | SQL, net/http, wire types |
| Repository | SQL, scanning into model | business rules, defaulting |
| Model | entity + logic about that entity | driver types, JSON tags |
| Domain | non-persisted types, cross-entity logic, converters | persistence concerns |

Rules:

- One model end to end. Two mapping boundaries only: DTO <-> model in transport,
  and wire <-> `provider.ChatRequest` inside each client. No others.
- `model` has a table; `domain` does not. `domain` may import `model`;
  `model` imports nothing from this codebase.
- Interfaces are declared by the consumer (`service/*/ports.go`), implemented
  implicitly by `repository` and `client`, bound by fx.
- Accept interfaces, return structs. No interface until there are two reasons.
- Context is always the first parameter, never a struct field.
- Constructors take dependencies. No globals, no `init()` side effects.
- Packages named for the domain, never for the layer. No `utils`.
- Errors are sentinels in `model`, wrapped with `%w`; HTTP status mapping lives
  in exactly one file, `transport/http/error.go`.
- Mappers are pure and unit-tested; that is where field drift hides.
- `viper` is imported only by `config`; everything downstream takes a typed
  `*Config`, validated at load so bad config fails at boot.

### Project layout

    cmd/gateway  cmd/relay
    internal/
      model  domain  dto  mapper  controller
      service/{chat,memory,files,quota,auth}
      repository/{postgres,redis}
      infrastructure/{postgres,redis,objectstore}
      client/{openai,anthropic,relay}
      transport/http/{router,middleware,sse}
      relayserver  jobs  config  textnorm
    migrations  configs  web  extractor
    build/package  deployments  scripts  docs

`dto` is a standalone package so `mapper` can import it without importing
`controller`.

## 6. Gateway internals

### HTTP surface

    POST   /api/v1/auth/register | /login | /logout
    GET    /api/v1/me
    GET    /api/v1/models                  list + live availability
    GET    /api/v1/conversations
    POST   /api/v1/conversations
    GET    /api/v1/conversations/:id/messages
    PATCH  /api/v1/conversations/:id       rename, archive
    POST   /api/v1/chat                    -> SSE
    GET    /api/v1/chat/:msgID/stream      reattach to in-flight generation
    POST   /api/v1/chat/:msgID/stop
    POST   /api/v1/files/presign | /files/:id/confirm
    GET    /api/v1/prompts
    GET/PUT /api/v1/memory | /settings

### Streaming path

    1. auth          session cookie -> user
    2. assemble      system + instructions + memory + thread + attachments
    3. reserve       atomic Redis reservation at maximum possible cost
    4. persist       user message + assistant row (status=streaming), one tx
    5. dispatch      provider.Stream(genCtx, req)
    6. relay         SSE out, accumulate, flush partial to Postgres every ~500ms
    7. finalize      status, token counts, real cost
    8. reconcile     release reservation, debit actual

Four decisions inside it:

- **The generation context is not the request context.** `genCtx` is detached
  with its own timeout, so a dropped mobile connection lets generation finish
  and persist. An explicit stop cancels `genCtx`. These are different events.
- **Partial flushing every ~500ms**, so a crash loses half a second of text,
  not a 2000-token answer already paid for.
- **Reserve then reconcile**, never check then charge: concurrent requests can
  otherwise each pass the check and collectively blow the cap.
- **`http.Server.WriteTimeout` must be 0**, with per-flush deadlines via
  `http.ResponseController`. Otherwise long streams die mid-sentence silently.

### Provider contract

```go
type Provider interface {
    Stream(ctx context.Context, req ChatRequest) (<-chan Delta, error)
    Models() []ModelInfo
    Health(ctx context.Context) error
}
```

The stream terminates with exactly one `Done` (carrying `Usage`) or one
`Error`. Billing depends on that invariant; it gets its own test.

Only two wire formats exist: OpenAI-style and Anthropic-style. vLLM is
OpenAI-compatible, so the `local` provider is the OpenAI client with a
different base URL and no relay.

`ChatRequest.Tools` exists but is unused in Phase 1 — the Phase 2 seam.

### Relay protocol

    POST https://relay-a/relay/openai/v1/chat/completions
      Authorization: <gateway credential>
      body: provider-native JSON, already normalized by the gateway

    relay: authenticate gateway -> strip inbound auth -> inject provider key
         -> forward to allowlisted host -> stream back unbuffered

The relay is deliberately dumb. Normalization churn stays in Iran where deploys
are cheap; the relay ships rarely. It enforces a strict host allowlist so a
compromised gateway cannot turn it into an open proxy, and it logs metadata
only — never message content.

## 7. Data model

    users            id, email, password_hash, locale, dir_pref, status
    auth_identities  user_id, kind, identifier, secret, verified_at
    invites          code, created_by, used_by, used_at
    subscriptions    user_id, plan_code, source, expires_at
    conversations    user_id, title, default_model, head_message_id, archived_at
    messages         id, conversation_id, parent_id, role, content, content_norm,
                     status, model, tokens_in, tokens_out, cost_usd_micros
    attachments      user_id, object_key, mime, extraction_status, text_key
    memory_facts     user_id, text, text_norm, source_message_id, pinned
    user_settings    custom_instructions, locale, theme
    usage_events     user_id, message_id, model, tokens_in, tokens_out,
                     cost_usd_micros, created_at
    prompts          locale, category, title, body

Notes:

- **Messages form a tree**, not a list: `parent_id` plus
  `conversations.head_message_id`. Edit-and-regenerate and regenerate-response
  both create sibling branches, which a flat list cannot represent. Build the
  tree now even if v1's UI only exposes "regenerate last" — retrofitting a
  parent pointer onto live data means inventing history that was never recorded.
- **`usage_events` is append-only and is the source of truth** for cost. Redis
  quota counters are a derived cache, rebuilt nightly from it. Never bill from
  Redis; never gate on Postgres in the hot path.
- **Effective plan is computed on read.** `expires_at` in the past means free
  tier. Nothing breaks if a job fails to run.
- **Cost is stored in USD micro-cents**; rial pricing is configuration. Cost
  data cannot be backfilled, so meter from day one even without billing.

## 8. Persian text handling

Postgres has no Persian stemmer, and Persian text is unusually messy:

- Arabic vs Persian codepoints (`ي`/`ی`, `ك`/`ک`) typed interchangeably
- ZWNJ (U+200C): `می‌روم` / `می روم` / `میروم`
- Three digit systems: Latin, Arabic-Indic, Persian
- Optional diacritics

`internal/textnorm` normalizes on write into `content_norm` (unify codepoints,
strip ZWNJ and diacritics, fold digits to Latin), indexed with **pg_trgm GIN**.
Trigram substring matching sidesteps the missing stemmer and handles Persian
agglutination better than a wrong stemmer would. Queries are normalized the
same way. The same rules are ported to TypeScript for client-side filtering —
if the two diverge, search results contradict the UI.

## 9. Memory

Two separate things, deliberately not conflated:

1. **Custom instructions** — user-authored, always injected, never inferred.
2. **Memory facts** — extracted asynchronously by the *local* model (zero FX
   cost, off the hot path), deduped by `text_norm`.

**All facts are injected, capped at ~50 facts / ~1500 tokens. No embeddings, no
retrieval.** Vector retrieval needs pgvector plus an embedding model, and
Persian embedding quality is the weakest link in the open-model stack — it
would silently retrieve wrong facts, which is the hardest bug class to notice.
Always-inject is deterministic and inspectable. Retrieval becomes worthwhile
around thousands of facts.

Memory is visible and editable: a settings page listing every fact, each
deletable, with an undo toast when one is added. In this market a memory the
user can read and delete is a trust feature, and it is also the debugging tool
when the model behaves oddly.

### Context assembly budget

    system prompt         fixed
    custom instructions   <= 1000 chars
    memory facts          <= 1500 tokens
    attachment text       <= 25% of window
    thread (from head)    fills remainder, newest first
    last user message     always

Budgeted per model, since Qwen3-14B's window is a fraction of Claude's. When
history is trimmed **the UI says so** — silent truncation makes the model look
forgetful for reasons the user cannot see.

## 10. Files and extraction

Presigned direct-to-bucket upload; large files never proxy through the gateway.
`attachments` goes pending -> client uploads -> confirm -> River job -> Python
worker -> text object + status update.

Fast path / slow path: PyMuPDF for PDFs with a text layer (most real
documents, milliseconds); OCR only when that returns empty. Persian OCR uses
**Surya**, not Tesseract — Tesseract's `fas` model is poor on Persian, and bad
OCR is worse than none because the model answers confidently from garbage.

Limits: 20MB, 200 pages, 10 minute timeout, 1/day free, 20/day granted tier.

Extracted text is **untrusted input**. It is wrapped in explicit delimiters with
a system instruction stating that document content is data, never instructions.
Without this, any PDF can rewrite the system prompt.

### Object storage: MinIO

Self-hosted MinIO rather than a managed bucket. It speaks the S3 API, so the
`infrastructure/objectstore` client (minio-go or aws-sdk-go-v2) and presigned
PUT/GET flow are unchanged if it is ever swapped for a managed provider. It
also removes the last Iranian cloud-vendor dependency and runs as one container
in Compose.

Consequences to own: disk capacity and backup are now yours, so put uploads and
extracted text on a dedicated volume with a scheduled `mc mirror` to a second
location. Set a bucket lifecycle rule expiring raw uploads after 90 days —
extracted text is small and worth keeping, original PDFs are not. Presign both
upload and download so the gateway never proxies file bytes. MinIO is AGPLv3:
running it unmodified as infrastructure carries no source-disclosure
obligation, but do not fork it without checking that.

## 11. Frontend

- `useChatStream` over `fetch` + `ReadableStream` + `AbortController`.
  **Not `EventSource`** — it is GET-only and cannot send headers.
- TanStack Query for conversation and message lists; Zustand for the live token
  buffer only.
- Branch switcher (`< 2/3 >`) under a message when siblings exist.
- RTL: `dir` on `<html>`; **logical properties only** (`ms-*`/`me-*`,
  `start`/`end`), enforced by an ESLint rule from day one. Retrofitting RTL
  into physical-property CSS is miserable.
- Self-host Vazirmatn; Google Fonts is unreliable from Iran.
- `dir="ltr"` on code blocks inside RTL text; `unicode-bidi: isolate` on inline
  Latin runs, or Persian sentences containing English words render scrambled.
- Persian digits and Jalali dates in the UI layer only; never in stored data.

A pure SPA is not indexable. If organic search matters, add a small prerendered
marketing/prompt-library site later — do not convert the app to SSR.

## 12. Observability, testing, abuse

**Metrics.** Time-to-first-token per provider is the primary metric: it is what
users perceive as speed and the first thing to degrade when the relay struggles.
Also tokens/sec, stream abort rate, upstream errors by provider, breaker state,
quota rejections, daily cost per user.

**Logs** carry request, user and conversation IDs and never message content.

**Testing.** Pure units for `model`, `domain`, `mapper`. Services against
in-memory fakes, no database. Repositories against testcontainers Postgres.
Provider clients against **golden SSE fixtures** — recorded real streams
replayed through `httptest`. That is the highest-value test in the codebase:
stream parsing is where subtle bugs live and CI cannot call real providers.
An `fx.ValidateApp` test moves DI wiring errors from production startup to CI.

**Abuse.** Per-user and per-IP rate limits, concurrent stream caps, invite-only
signup, `usage_events` as audit trail, a documented takedown process, and a
small keyword blocklist for obvious illegal categories that logs rather than
silently drops. No ML moderation pipeline in Phase 1.

## 13. Deliberate omissions

- Kubernetes. Compose on 2-3 VMs until a second gateway node earns it.
- `pkg/`, `test/`, `tools/`, and other empty layout directories.
- OpenAPI spec. The controller DTOs are the contract at this size; a
  hand-maintained spec drifts silently, which is worse than none.
- More than two relays at launch. Build the failover interface, deploy two.
- Semantic/vector search anywhere.
