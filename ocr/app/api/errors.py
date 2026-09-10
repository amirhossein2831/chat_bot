from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from app.api.schema import ErrorResponse
from app.domain.errors import OCRError, PermanentError

# The status code is the retry instruction for the job that called us: 4xx will
# fail the same way every time, 5xx is worth trying again.
_CODES: dict[str, int] = {
    "EmptyRequestError": 400,
    "TooManyPagesError": 413,
    "PageTooLargeError": 413,
    "UnreadablePageError": 415,
    "RecognizerUnavailableError": 503,
    "RecognitionFailedError": 500,
}


def install(app: FastAPI) -> None:
    @app.exception_handler(OCRError)
    async def handle(_: Request, error: OCRError) -> JSONResponse:
        name = type(error).__name__
        status = _CODES.get(name, 400 if isinstance(error, PermanentError) else 500)
        return JSONResponse(
            status_code=status,
            content=ErrorResponse(error=str(error), code=_slug(name)).model_dump(),
        )


def _slug(name: str) -> str:
    name = name.removesuffix("Error") or name
    out: list[str] = []
    for index, char in enumerate(name):
        if char.isupper() and index:
            out.append("_")
        out.append(char.lower())
    return "".join(out)
