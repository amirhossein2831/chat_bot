# Working rules for this repository

## Read first

The design spec at `docs/superpowers/specs/2026-09-08-persian-ai-chat-design.md`
is the source of truth. It records not just what was decided but why, and the
alternatives that were rejected. Do not re-litigate a decision without reading
its section first.

Per-component conventions live in `gateway/CONVENTIONS.md`,
`web/CONVENTIONS.md` and `ocr/CONVENTIONS.md`. They are enforceable rules; the
spec carries the reasoning.

## Project-wide rules

- **Persian is the primary language, English is secondary.** Every user-facing
  string is translated, and layout works in both directions.
- **Never log message content**, in any component. Request, user and
  conversation IDs only.
- **Money is `int64` USD micro-cents.** Never a float, anywhere.
- **All persisted Persian text gets a normalised twin** (`*_norm`) written
  through the shared normaliser. Search queries are normalised the same way.
- **Cost metering is never skipped**, even with billing out of scope. Cost data
  cannot be backfilled.
- **Provider API keys exist only on the relay.** Nothing inside Iran holds them.
- Commit messages: imperative mood, one logical change.

## The traps that have already been identified

These are documented because each one is invisible until it is expensive:

1. `http.Server.WriteTimeout` bounds the whole response and silently truncates
   SSE streams. It must be `0`; use per-write deadlines instead.
2. The generation context must be detached from the request context, or a
   dropped mobile connection destroys a generation that was already paid for.
3. `httputil.ReverseProxy` buffers by default — the relay needs
   `FlushInterval: -1` or every answer arrives in one burst at the end.
4. Blocking work inside an async FastAPI handler stalls the entire event loop,
   not just that request.
5. Persian PDFs frequently carry a text layer with a broken `ToUnicode` CMap
   that returns confident mojibake. Check Arabic-block character ratio before
   trusting extracted text.
6. Iranian `.txt` and `.csv` files are frequently Windows-1256. Never assume
   UTF-8.
7. HuggingFace model downloads fail from Iran. Bake weights into images and set
   `HF_HUB_OFFLINE=1`.
8. Google Fonts is unreliable from Iran. Self-host Vazirmatn.
9. OpenAI omits token usage from streamed responses unless
   `stream_options.include_usage` is set — every request then bills as zero.
10. Never run `python-bidi` over recognised text. Store logical order; bidi is
    for display only.

## Working method

- Tests first. A task is not done because the code exists.
- Run the verification command and read its output before claiming anything
  passes.
- When a design decision turns out wrong during implementation, update the spec
  in the same commit. A spec that disagrees with the code is worse than none.
