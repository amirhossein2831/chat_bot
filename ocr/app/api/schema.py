from pydantic import BaseModel


class PageResponse(BaseModel):
    index: int
    text: str


class RecognizeResponse(BaseModel):
    pages: list[PageResponse]
    text: str
    characters: int


class HealthResponse(BaseModel):
    status: str
    recognizer: str
    ready: bool


class ErrorResponse(BaseModel):
    error: str
    code: str
