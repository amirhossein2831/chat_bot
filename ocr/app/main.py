import contextlib
from collections.abc import AsyncIterator

from fastapi import FastAPI

from app.adapters.echo import EchoRecognizer
from app.adapters.surya import SuryaRecognizer
from app.api import errors, routes
from app.service.ports import Recognizer
from app.service.recognize import RecognizeService
from app.settings import Settings


def build_recognizer(settings: Settings) -> tuple[Recognizer, str]:
    if settings.recognizer == "echo":
        return EchoRecognizer(), "echo"

    surya = SuryaRecognizer()
    # Loading here rather than at import or per request: the weights are
    # gigabytes and are held for the life of the process.
    surya.load()
    return surya, "surya"


def create_app(settings: Settings | None = None) -> FastAPI:
    resolved = settings or Settings()

    @contextlib.asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        recognizer, name = build_recognizer(resolved)
        app.state.recognizer_name = name
        app.state.service = RecognizeService(
            recognizer,
            max_pages=resolved.max_pages,
            max_page_bytes=resolved.max_page_bytes,
            default_languages=resolved.languages,
        )
        yield

    app = FastAPI(title="OCR", version="0.1.0", lifespan=lifespan)
    errors.install(app)
    app.include_router(routes.router)
    return app


app = create_app()
