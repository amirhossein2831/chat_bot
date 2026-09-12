from app.api.schema import PageResponse, RecognizeResponse
from app.domain.models import Recognition


def to_recognize_response(recognition: Recognition) -> RecognizeResponse:
    return RecognizeResponse(
        pages=[PageResponse(index=page.index, text=page.text) for page in recognition.pages],
        text=recognition.text,
        characters=recognition.characters,
    )
