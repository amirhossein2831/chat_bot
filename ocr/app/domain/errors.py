class OCRError(Exception):
    """Base for every failure this service reports."""


class PermanentError(OCRError):
    """Retrying will fail the same way: the request itself is the problem."""


class TransientError(OCRError):
    """The request was fine and may succeed later."""


class TooManyPagesError(PermanentError):
    def __init__(self, given: int, limit: int) -> None:
        super().__init__(f"{given} pages exceeds the limit of {limit}")


class PageTooLargeError(PermanentError):
    def __init__(self, index: int, size: int, limit: int) -> None:
        super().__init__(f"page {index} is {size} bytes, over the limit of {limit}")


class EmptyRequestError(PermanentError):
    def __init__(self) -> None:
        super().__init__("no pages were sent")


class UnreadablePageError(PermanentError):
    def __init__(self, index: int) -> None:
        super().__init__(f"page {index} is not a readable image")


class RecognizerUnavailableError(TransientError):
    def __init__(self, reason: str) -> None:
        super().__init__(f"the recogniser is not available: {reason}")


class RecognitionFailedError(TransientError):
    def __init__(self, reason: str) -> None:
        super().__init__(f"recognition failed: {reason}")
