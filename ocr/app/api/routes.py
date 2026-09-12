from typing import Annotated

from fastapi import APIRouter, Form, Request, UploadFile

from app.api.mapper import to_recognize_response
from app.api.schema import HealthResponse, RecognizeResponse
from app.service.recognize import RecognizeService

router = APIRouter()


@router.get("/healthz")
async def health(request: Request) -> HealthResponse:
    service: RecognizeService = request.app.state.service
    name: str = request.app.state.recognizer_name
    return HealthResponse(status="ok", recognizer=name, ready=service.ready)


@router.post("/v1/recognize")
async def recognize(
    request: Request,
    # Optional so that sending nothing is answered by this service's own error
    # shape, which the calling job reads, rather than the framework's.
    pages: list[UploadFile] | None = None,
    languages: Annotated[str | None, Form()] = None,
) -> RecognizeResponse:
    service: RecognizeService = request.app.state.service

    wanted = (
        [code.strip() for code in languages.split(",") if code.strip()] if languages else None
    )
    body = [await page.read() for page in pages or []]

    return to_recognize_response(await service.run(body, wanted))
