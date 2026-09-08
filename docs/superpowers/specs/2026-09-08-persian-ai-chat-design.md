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
- File upload with text extraction, including Persian OCR
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
      |-- Extraction (in-process River job, Go)
      |     +-- OCR service (Python + Surya, GPU) for scans and images
      +-- Relay (foreign VPS) --> OpenAI / Anthropic

The gateway is the only publicly reachable component. Inference, relay,
extraction and storage are internal.

## 4. Stack

| Layer | Choice |
|---|---|
| Web | React 19 + Vite + TypeScript |
| UI | Tailwind v4 + shadcn/ui, logical properties only |
| i18n | i18next; Vazirmatn self-hosted; Intl for Jalali dates and Persian digits |
| API | Go 1.24, chi, fx (DI), viper (config) |
| Data | Postgres 17 (pgx + RowToStructByName), Redis 7 |
|  Objects | MinIO (self-hosted, S3-compatible) |
| Jobs | River (Postgres-backed) |
| Migrations | goose |
| Inference | vLLM (Qwen3), relay to OpenAI/Anthropic |
| Extraction | Go-native (pdfium via wazero, excelize) |
| OCR | Python service: FastAPI + Surya, GPU |
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
    migrations  configs  web
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
`attachments` goes pending -> client uploads -> confirm -> River job -> text
object + status update.

**Extraction is Go-native.** An earlier draft put the whole extraction pipeline
in Python; that was wrong on scope. Go matches or beats Python on every format
here: PDF text via pdfium, XLSX via excelize, DOCX by unzipping and parsing its
XML, charset detection via `x/text`. Only OCR has no Go path.

Use **pdfium**, not MuPDF. Both are C engines of comparable quality, but MuPDF
(and therefore PyMuPDF) is AGPLv3, which raises a source-disclosure question
for a commercial network service. pdfium is BSD-3, and
`github.com/klippa-app/go-pdfium` offers a wazero-based WebAssembly mode, so it
needs no cgo and the single static binary survives.

Routing:

    pdf   -> pdfium text extraction
               |- usable text layer  -> done (milliseconds, ~95% of files)
               +- absent or garbage  -> render at 250 DPI -> OCR
    image -> OCR
    docx  -> unzip + parse XML
    xlsx  -> excelize
    txt/csv/md -> charset detection, then decode

Two heuristics pick the branch, and the second is Persian-specific:

- **Is there a text layer?** Mean characters per page under ~50, or fewer than
  30% of pages carrying text.
- **Is the text layer lying?** Many Iranian PDFs produced by older Persian
  publishing tools carry a text layer with a broken `ToUnicode` CMap: extraction
  returns confident-looking mojibake. Compare the fraction of characters in the
  Arabic block (U+0600–U+06FF) against replacement and control characters; a
  Persian document whose text is 5% Arabic-block characters is lying and must
  fall through to OCR. Skipping this check feeds garbage to the model, which
  then answers confidently from garbage — the worst failure mode available,
  because nothing looks broken.

Plain-text and CSV files from Iranian sources are frequently **Windows-1256**,
not UTF-8. Never assume UTF-8; detect.

**OCR ships in Phase 1**, as a separate Python service. It uses **Surya**, not
Tesseract: Tesseract's `fas` model is poor on Persian, and bad OCR is worse
than none because the model then answers confidently from garbage. Iranian
official documents, contracts and older books are disproportionately scans, so
this branch carries more traffic here than it would in other markets.

The Go job renders pages to PNG at 250 DPI with pdfium and POSTs them to the
OCR service, which returns text. Go keeps the queue, the retries, the database
and every credential; the OCR service is a stateless function. Instrument the
share of uploads taking this branch — it drives GPU capacity planning.

Limits: 20MB, 200 pages, 10 minute timeout, 1/day free, 20/day granted tier.

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

## 11. Web application

React 19 + Vite + TypeScript, single-page, served from the same origin as the
API so there is no CORS and no cookie-domain problem.

### Structure

Feature-first, mirroring the Go domain packages rather than splitting by
technical layer:

    web/src/
      app/          router, providers, layout shell
      api/          typed client, SSE parser, error mapping
      features/
        auth/       login, register
        chat/       ChatView, MessageList, Composer, useChatStream
        conversations/  sidebar, search
        memory/     facts list, custom instructions
        prompts/    Persian library
        files/      upload, attachment chips
      components/   shadcn/ui primitives
      i18n/         fa.json, en.json, direction handling
      lib/          textnorm.ts, format.ts
      styles/

A feature owns its components, hooks and API calls. Nothing in `features/`
imports from another feature; shared pieces move to `components/` or `lib/`.

### State

- **TanStack Query** for anything the server owns: conversations, messages,
  models, memory, prompts, current user.
- **Zustand** for the live token buffer, the composer draft, and UI preferences
  (locale, theme, sidebar collapsed).
- No Redux. The only genuinely hard state here is one in-flight stream.

### The streaming hook

`useChatStream()` returns `{ send, stop, streaming, buffer, error }` and is the
heart of the client.

- `fetch` + `ReadableStream` + `TextDecoder`, **never `EventSource`** — it is
  GET-only and cannot send headers or a body.
- Frames split on the blank line, then `event:` and `data:` parsed per frame.
- `AbortController` powers the stop button and also fires `POST
  /chat/{id}/stop` so the server cancels generation instead of merely
  detaching.
- On `done`, invalidate the messages query so the persisted message replaces
  the local buffer; the buffer is never the source of truth.
- On mount, if the newest message has `status = streaming`, open
  `GET /chat/{id}/stream` to reattach — this is what makes a refresh
  mid-answer work.
- Reconnect with backoff on transport failure, showing a connection notice
  rather than silently stalling.

**Batch renders, do not render per token.** Setting state on every delta
re-renders the message list dozens of times per second and makes long answers
visibly janky on the mid-range Android devices most users have. Accumulate
into a ref and flush on `requestAnimationFrame`.

### Markdown while streaming

Partial markdown is malformed markdown: a half-arrived code fence renders as
broken output that flickers when it completes. Before rendering, count fence
markers and append a virtual closing fence when the count is odd.

Syntax highlighting is deferred to completion — plain `<pre>` while streaming,
Shiki once `done` arrives. Highlighting every partial frame is expensive and
the result is thrown away microseconds later.

### The message tree in the UI

Assistant messages whose parent has more than one child show a branch switcher
(`< 2/3 >`); selecting a branch moves the conversation head. Editing a user
message creates a new sibling rather than mutating history. This is the client
half of the tree from section 7.

### RTL and Persian rendering

- `dir` and `lang` set on `<html>` from the active locale.
- **Logical properties only** — `ms-*`, `me-*`, `ps-*`, `pe-*`, `start`, `end`.
  An ESLint rule bans `ml-*`, `mr-*`, `pl-*`, `pr-*`, `left-*`, `right-*`,
  `text-left`, `text-right` from day one. Retrofitting RTL into physical-property
  CSS is miserable; preventing it costs one rule.
- `dir="ltr"` on `<pre>` and inline code, or code blocks inside Persian text
  render right-aligned and unreadable.
- Inline Latin runs wrapped in `<bdi>`. Without isolation, a Persian sentence
  containing an English product name renders with the punctuation in the wrong
  place — subtly wrong in a way that reads as amateurish to native speakers.
- **Jalali dates via `Intl.DateTimeFormat('fa-IR-u-ca-persian')`** and Persian
  digits via `Intl.NumberFormat('fa-IR')`. The platform handles the Persian
  calendar natively, so no date library is needed. Formatting is display-only —
  never store or copy Persian digits.
- **Vazirmatn self-hosted and subset**, `font-display: swap`. Google Fonts is
  unreliable from Iran, and a webfont that never arrives leaves Persian text in
  a fallback that renders poorly.

### Composer

Auto-growing textarea; Enter sends, Shift+Enter breaks; IME composition events
respected so the send key does not fire mid-composition. Model picker reads
`GET /models` and greys out unavailable models with a reason. Attachment chips
show extraction progress. A notice appears when the server reports the thread
was truncated.

### Cross-cutting

- Auth is cookie-based; the token never touches JavaScript. A 401 anywhere
  redirects to login.
- Streaming text lives in an `aria-live="polite"` region so screen readers
  follow the answer.
- Route-level code splitting; Shiki lazy-loaded. Message lists are virtualized
  only if a real conversation gets slow — not preemptively.
- `lib/textnorm.ts` is a direct port of the Go normaliser. If the two diverge,
  client-side filtering contradicts server-side search results.

### Testing

Vitest + Testing Library, MSW for API mocking. The stream hook is tested
against **recorded SSE fixtures**, mirroring the Go golden fixtures — including
frames split mid-line, since that is where client parsers break too. One
Playwright end-to-end covering login, send, stream, refresh-mid-stream.

A pure SPA is not indexable. If organic search matters, add a small prerendered
marketing and prompt-library site later; do not convert the app to SSR.

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

## 13. Python: scope and conventions

### Why Python exists at all

Exactly one reason: **Persian OCR.** Surya is PyTorch, has no Go port, and
exporting it to ONNX for `onnxruntime-go` is its own project with uncertain
accuracy. Tesseract has Go bindings but its Persian is bad enough to be worse
than nothing. Everything else — PDF text, DOCX, XLSX, charset detection — is
equal or better in Go.

The OCR service therefore **ships in Phase 1**. It is the only Python in the
system, and the only Python in the request path — everything else Python could
plausibly do here is either better in Go or belongs offline.

### The rule for future Python

> Python belongs offline (training, evaluation, analysis) and at custom-ML
> boundaries where no server exists. It never belongs in the request path.

"It needs a model" is not "it needs Python in the codebase." vLLM,
text-embeddings-inference and Whisper servers all expose HTTP APIs that Go
calls directly. Python enters only where custom pre- and post-processing wraps
raw model output, which is why OCR qualifies and embeddings do not.

Later candidates beyond the OCR service: an evaluation harness for Persian model quality (arguably
higher value than OCR — it answers whether the free tier is good enough with
numbers rather than impressions), Persian STT for voice input, cost and pricing
analysis over `usage_events`, LoRA fine-tuning, and Phase 2 financial
indicators and backtesting.

Explicit non-candidates: embeddings and reranking (serve behind TEI, call from
Go), web scraping for Phase 2 (`goquery` and `chromedp` are adequate, and one
language is worth more), and Persian NLP libraries such as hazm (we chose
trigram matching over stemming, so there is nothing for them to do).

### Queue: River stays

River is Go-only, so an earlier draft's "Go enqueues, Python consumes" did not
actually work. The resolution is not a broker.

The deciding property is **transactional enqueue**. Every job here acts on a
row that was just written, so with any external broker the two writes cannot be
atomic: Postgres commits and the publish fails (job silently lost, document
stuck at "processing" forever), or the publish succeeds and the transaction
rolls back (job fires for a row that does not exist). The standard fix is a
transactional outbox — a Postgres jobs table plus a relay process — which is a
Postgres-backed queue with a broker bolted behind it. River is that outbox
without the second hop.

Rejected alternatives, with the trigger that would change the decision:

- **Redis (asynq/Celery)** — the apparent cross-language win is illusory:
  asynq's payload format is Go-specific and Celery's protocol is Python-specific,
  so interop means hand-rolling a shared envelope over a Redis list and
  discarding the retry, backoff and scheduling those libraries exist to
  provide. Revisit only if sub-10ms job latency is ever required.
- **RabbitMQ** — genuinely language-agnostic and durable, but a stateful
  service to operate for a few thousand jobs a day, and still no transactional
  enqueue. Revisit at three or more services in three or more languages needing
  real routing and fanout.
- **Kafka** — wrong shape, not merely oversized: no per-message ack or
  redelivery, no delayed retry, and a slow consumer blocks its whole partition,
  so one multi-minute OCR job stalls everything behind it. Its legitimate future
  home is streaming `usage_events` to analytics consumers that replay a log —
  alongside River, not replacing it.

If a Python consumer is ever genuinely needed, the cheapest correct option is
Postgres polling with `FOR UPDATE SKIP LOCKED`: safe concurrent consumers, no
coordination, roughly forty lines, still transactional, nothing new to operate.

### Python service conventions

The Go layering maps onto Python idioms rather than being transliterated:

| Go | Python |
|---|---|
| interface declared by consumer | `typing.Protocol` — structural, implicit satisfaction |
| viper + typed struct | `pydantic-settings`, validated at startup |
| fx | plain composition root in FastAPI `lifespan` |
| dto + mapper | pydantic schemas + mapper functions |
| model / domain | frozen `@dataclass(slots=True)` |
| sentinel errors + error.go | exception types + one exception handler |
| table-driven tests, fakes | `pytest.mark.parametrize`, fakes |

`Protocol` is the important one: like a Go interface, the adapter satisfies it
without importing it, so the service is testable with no model loaded.

Structure, per service:

    ocr/
      pyproject.toml         uv, ruff, mypy strict, pytest
      app/
        main.py              composition root + lifespan
        settings.py
        api/    schema.py  mapper.py  routes.py  errors.py
        domain/ models.py  errors.py
        service/ ports.py  recognize.py
        adapters/ surya.py  images.py
      tests/  fakes.py  test_recognize.py  test_api.py

Rules:

- **`domain` imports nothing** — no fastapi, no pydantic, no torch. Same rule as
  `model` in Go, and it is what lets service tests run in milliseconds with a
  fake recognizer and no GPU.
- **Blocking work must not run in an async handler.** FastAPI handlers are
  async, and one blocking call stalls the entire event loop — every concurrent
  request freezes, not just that one. Model inference goes through
  `anyio.to_thread.run_sync`. This is the Python analogue of the `WriteTimeout`
  trap: invisible in development, catastrophic under load.
- **Load models in `lifespan`**, never at import time and never per request.
- **Bake model weights into the image**, `HF_HUB_OFFLINE=1`. Runtime HuggingFace
  downloads fail from Iran.
- **One GPU job at a time** via a semaphore: the GPU is shared with vLLM, so a
  scanned 200-page document must degrade OCR latency rather than starve chat
  inference for everyone.
- **Store logical order, never visual order.** Running `python-bidi` over
  recognized text stores visually-reordered characters, which corrupts search,
  breaks `textnorm` and confuses the model. Sort lines top-to-bottom and keep
  each line's logical order as the recognizer produced it.
- **mypy strict in CI**, `Any` banned in `domain` and `service`. Without it the
  property that makes the Go design work — wrong type is a build failure — is
  lost.
- **`uv` with a committed lockfile.** Not pip, not poetry.
- **Fakes, not `unittest.mock`.** A fake class implementing the Protocol gives
  type-checked doubles; `mock.patch` gives tests that still pass after a rename.
- **4xx means permanent, 5xx means transient.** The status code carries retry
  semantics for the calling River job: unsupported format and page-limit
  breaches must not be retried; OOM and model-load failures must be.

When a second Python service arrives (STT), copy this skeleton rather than
extracting a shared framework. Two services is not enough evidence to know what
they would share.

### Document text is untrusted

Extracted text is wrapped in explicit delimiters with a system instruction
stating that document content is data and never instructions. Without it, any
uploaded PDF can rewrite the system prompt.

## 14. Relay server

A second, deliberately small Go binary (`cmd/relay`) hosted outside Iran. It is
the only component that holds provider credentials.

### Responsibilities, complete

1. Authenticate the gateway.
2. Enforce a host allowlist.
3. Inject the provider key.
4. Stream the response back unbuffered.
5. Emit metrics and metadata-only logs.

Nothing else. It does not normalise request bodies, does not retry, does not
know what a conversation is. Normalisation churn stays in Iran where deploys
are cheap and frequent; the relay is the component that is hardest to update
and riskiest to touch, so it ships rarely.

### Routing and key injection

    POST /relay/{provider}/{path...}
      -> allowlist lookup: openai -> https://api.openai.com
                           anthropic -> https://api.anthropic.com
      -> strip every inbound Authorization / x-api-key header
      -> inject the provider key for that provider
      -> proxy

The upstream is chosen from a static map keyed by the path segment, **never
from a header or query parameter**. Taking the upstream from the request would
turn the relay into an open proxy the moment the gateway credential leaks.

Hold **several keys per provider**, selected round-robin with per-key error
tracking. A key that starts returning 401 or persistent 429 is taken out of
rotation without taking the tier down. Keys reload on SIGHUP so rotation is a
routine operation, not an emergency redeploy.

### Authentication

**mTLS**, with the relay verifying a client certificate against its own CA
(`tls.Config{ClientAuth: RequireAndVerifyClientCert}`). A rotatable HMAC bearer
is the fallback where certificate distribution is impractical; accept two valid
credentials at once so rotation has no downtime.

The relay terminates TLS itself via autocert rather than sitting behind another
proxy — one less hop, and the client certificate is verified by the process
that acts on it. Add an IP allowlist for the gateway's egress addresses as
defence in depth where those addresses are stable.

### Streaming

`httputil.ReverseProxy` with **`FlushInterval: -1`**, which flushes every write
immediately. The default buffers, which would hold tokens until the response
completes and make every answer arrive in one burst — the product appears
frozen and the cause is invisible.

Server `WriteTimeout` is 0 for the same reason it is on the gateway. The
upstream transport sets `ResponseHeaderTimeout` but no overall deadline, since
a legitimate generation can run for minutes.

### No retries here

Upstream 429s and 5xxs pass through unchanged so the gateway's circuit breaker
observes reality. The gateway already fails over across relays; retrying in
both places multiplies load against a provider that is already rate-limiting
you, which is how a temporary throttle becomes a terminated account.

### Logging and metrics

Metadata only: request ID, provider, model, status, byte count, duration, and
which key was used. **Never request or response bodies.** If this VPS is seized
or subpoenaed, user conversations were never on it.

The model name arrives as an `X-Relay-Model` header from the gateway, so the
relay never parses a request body — cheaper, and it keeps the "no bodies" rule
structurally true rather than merely intended.

Metrics: requests by provider and status, upstream time-to-first-byte, bytes
streamed, per-key error rate, and keys currently in rotation.

### Health

`/healthz` is liveness only. `/readyz` reports whether upstreams were reachable
on the last cached probe (30s), so the gateway's health checks do not turn into
traffic against the providers.

### Deployment

Distroless image, non-root, read-only root filesystem, egress firewalled to the
two provider hosts. **At least two relays** on different hosts, regions and
provider accounts — a single relay is a single point of total frontier failure,
and being blocked or flagged is the expected steady state rather than an
incident.

### Load shaping

Per-key rate limiting keeps each key's volume within a plausible range. All of
your users arriving at a provider through one key looks like one implausibly
heavy account, and that pattern is what gets keys terminated. Spreading load
across keys and relays is a availability requirement, not an optimisation.

## 15. Deliberate omissions

- Kubernetes. Compose on 2-3 VMs until a second gateway node earns it.
- `pkg/`, `test/`, `tools/`, and other empty layout directories.
- OpenAPI spec. The controller DTOs are the contract at this size; a
  hand-maintained spec drifts silently, which is worse than none.
- More than two relays at launch. Build the failover interface, deploy two.
- Semantic/vector search anywhere.
