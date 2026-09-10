FROM python:3.12-slim AS base

ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy

COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

WORKDIR /srv

COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev --extra recognizer

COPY app ./app

# Weights are baked in at build time, where HuggingFace is reachable, and the
# runtime is then forbidden from reaching for them: a download from Iran fails,
# and failing at startup is better than failing on the first scanned page.
COPY weights /opt/weights
ENV HF_HUB_OFFLINE=1 \
    HF_HOME=/opt/weights \
    PATH="/srv/.venv/bin:$PATH"

RUN useradd --system --uid 10001 ocr && chown -R ocr /srv
USER ocr

EXPOSE 8100
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8100"]
