import anyio

from app.domain.errors import (
    EmptyRequestError,
    PageTooLargeError,
    RecognitionFailedError,
    RecognizerUnavailableError,
    TooManyPagesError,
)
from app.domain.models import Recognition
from app.service.ports import Recognizer


class RecognizeService:
    def __init__(
        self,
        recognizer: Recognizer,
        *,
        max_pages: int,
        max_page_bytes: int,
        default_languages: list[str],
    ) -> None:
        self._recognizer = recognizer
        self._max_pages = max_pages
        self._max_page_bytes = max_page_bytes
        self._default_languages = default_languages
        # One at a time. The GPU is shared with chat inference, so a scanned
        # two-hundred-page document must slow OCR down rather than starve
        # everyone's chat.
        self._turn = anyio.Semaphore(1)

    @property
    def ready(self) -> bool:
        return self._recognizer.ready

    async def run(self, pages: list[bytes], languages: list[str] | None = None) -> Recognition:
        self._check(pages)

        if not self._recognizer.ready:
            raise RecognizerUnavailableError("no model is loaded")

        wanted = languages or self._default_languages

        async with self._turn:
            try:
                # Inference blocks. Run in a thread or every other request on
                # this worker freezes with it, not just this one.
                recognized = await anyio.to_thread.run_sync(
                    self._recognizer.recognize, pages, wanted
                )
            except Exception as error:
                raise RecognitionFailedError(str(error)) from error

        return Recognition(pages=tuple(recognized))

    def _check(self, pages: list[bytes]) -> None:
        if not pages:
            raise EmptyRequestError

        if len(pages) > self._max_pages:
            raise TooManyPagesError(len(pages), self._max_pages)

        for index, page in enumerate(pages):
            if len(page) > self._max_page_bytes:
                raise PageTooLargeError(index, len(page), self._max_page_bytes)
