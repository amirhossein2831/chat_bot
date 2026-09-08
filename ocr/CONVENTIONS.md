# OCR service conventions (Python)

Rationale lives in spec section 13.

## Why this service exists

Persian OCR, and nothing else. Surya is PyTorch with no Go port, and
Tesseract's `fas` model is poor enough on Persian that bad OCR is worse than
none — the model then answers confidently from garbage. Every other format
(PDF text, DOCX, XLSX, charset detection) is handled in Go.

Scope: **receive page images, return text.** No database, no queue, no
credentials, no object-store access.

## Stack

Python 3.12 · uv with a committed lockfile · FastAPI · pydantic v2 ·
pydantic-settings · ruff · mypy strict · pytest

## Structure

    app/
      main.py         composition root + lifespan
      settings.py
      api/      schema.py  mapper.py  routes.py  errors.py
      domain/   models.py  errors.py
      service/  ports.py  recognize.py
      adapters/ surya.py  images.py
    tests/  fakes.py  test_recognize.py  test_api.py

`typing.Protocol` is the Go interface analogue: the adapter satisfies it
without importing it, so the service is testable with no model loaded.

## Rules

- **`domain` imports nothing** — no fastapi, no pydantic, no torch. This is
  what lets service tests run in milliseconds with a fake recognizer and no GPU.
- **Blocking work never runs directly in an async handler.** One blocking call
  stalls the whole event loop and freezes every concurrent request. Inference
  goes through `anyio.to_thread.run_sync`.
- **Load models in `lifespan`**, never at import time and never per request.
- **Bake weights into the image**, `HF_HUB_OFFLINE=1`. Runtime HuggingFace
  downloads fail from Iran.
- **One GPU job at a time** via a semaphore. The GPU is shared with vLLM; a
  scanned 200-page document must degrade OCR latency, not starve chat inference.
- **Store logical order, never visual order.** `python-bidi` is for display;
  storing reordered text corrupts search, breaks the normaliser and confuses
  the model. Sort lines top-to-bottom, keep each line as the recognizer
  produced it.
- **4xx is permanent, 5xx is transient.** The status code carries retry
  semantics for the calling River job.
- Frozen dataclasses with `slots=True` for domain types.
- `Any` banned in `domain` and `service`. mypy strict runs in CI.
- **Fakes, not `unittest.mock`.** A fake implementing the Protocol is
  type-checked; `mock.patch` passes after a rename.
- No mutable default arguments.

## Hardening

This service parses untrusted files. Non-root, read-only root filesystem, no
network egress beyond what it needs, container memory and CPU limits, hard caps
on page count and output size before any parser runs.
