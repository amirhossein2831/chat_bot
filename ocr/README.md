# OCR service

Persian OCR for the gateway: page images in, text out. Conventions and the
reasoning behind them are in `CONVENTIONS.md` and spec section 13.

It holds no database, no queue, no credentials and no bucket access. The Go
job renders PDF pages at 250 DPI, posts them here, and owns everything else.

## API

    GET  /healthz          -> {status, recognizer, ready}
    POST /v1/recognize     multipart: pages=<file> (repeatable), languages=fa,en
                           -> {pages: [{index, text}], text, characters}

**The status code is the retry instruction** for the calling job. 4xx will fail
the same way every time and must not be retried; 5xx is worth retrying.

| status | code | meaning |
|---|---|---|
| 400 | `empty_request` | no pages were sent |
| 413 | `too_many_pages`, `page_too_large` | over a hard limit |
| 415 | `unreadable_page` | not a readable image |
| 500 | `recognition_failed` | inference failed |
| 503 | `recognizer_unavailable` | no model loaded |

Text comes back in **logical order**, never visually reordered. Lines are
sorted down the page; each line is left as the recogniser produced it.

## Running it

    uv sync
    uv run uvicorn app.main:app --port 8100

    uv run ruff check app && uv run ruff format --check app
    uv run --with mypy mypy app

`OCR_RECOGNIZER=surya` is the real thing and needs torch plus baked-in weights,
so it only runs from the image. `OCR_RECOGNIZER=echo` reports what it received
without pretending to read it, which is how the upload and extraction path is
exercised on a machine where torch cannot be installed — the weight CDN is
unreachable from Iran and torch alone is several gigabytes.

Settings are `OCR_`-prefixed: `HOST`, `PORT`, `RECOGNIZER`, `MAX_PAGES`,
`MAX_PAGE_BYTES`, `LANGUAGES`.
