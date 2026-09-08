# chat_bot

A Persian-first, Claude-like AI chat platform for the Iranian market, with
selectable models spanning self-hosted open weights and frontier models reached
through a relay hosted abroad.

Phase 2 — web browsing and financial-market advice — is out of scope for now.
The design keeps that seam open but builds none of it.

## Layout

    docs/superpowers/specs/     design specs (the source of truth)
    docs/superpowers/plans/     implementation plans, one per component
    gateway/                    Go: API gateway + relay binary
    web/                        React SPA
    ocr/                        Python: Persian OCR service
    deployments/                docker-compose, Caddy, ops

The relay lives at `gateway/cmd/relay`, not in a top-level folder, because it
shares the provider wire types with the gateway. Splitting it into its own
module would duplicate them, and a wire-format change would then need two
coordinated deploys instead of one. Same repository, two binaries, deployed
independently.

## Documents

| Document | Contents |
|---|---|
| `docs/superpowers/specs/2026-09-08-persian-ai-chat-design.md` | The full Phase 1 design: scope, model access, architecture, data model, Persian text handling, memory, files, web app, OCR, relay |
| `docs/superpowers/plans/2026-09-08-gateway.md` | Gateway implementation plan, 14 TDD tasks |
| `CLAUDE.md` | Working rules for this repository |
| `gateway/CONVENTIONS.md` | Go conventions |
| `web/CONVENTIONS.md` | React conventions |
| `ocr/CONVENTIONS.md` | Python conventions |

## Status

| Component | Design | Plan | Built |
|---|---|---|---|
| Gateway | done | done | not started |
| OCR service | done | — | — |
| Relay | done | — | — |
| Web app | done | — | — |

## Open decision

**The product has no name.** It appears in the Go module path, the repository,
the domain and the UI. `chat_bot` is a placeholder. Settle it before gateway
Task 1, which pins the module path into every import.
