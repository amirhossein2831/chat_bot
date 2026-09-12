from app.domain.models import Line, Page


class SuryaRecognizer:
    """Surya, loaded once and held.

    Imported lazily so the service starts, and every path except recognition
    itself works, on a machine where torch is not installed.
    """

    def __init__(self) -> None:
        self._predictor: object | None = None

    def load(self) -> None:
        from surya.recognition import RecognitionPredictor

        self._predictor = RecognitionPredictor()

    @property
    def ready(self) -> bool:
        return self._predictor is not None

    def recognize(self, pages: list[bytes], languages: list[str]) -> list[Page]:
        from PIL import Image

        predictor = self._predictor
        if predictor is None:
            raise RuntimeError("recognizer was not loaded")

        import io

        images = [Image.open(io.BytesIO(page)).convert("RGB") for page in pages]
        results = predictor(images, [languages] * len(images))  # type: ignore[operator]

        recognized: list[Page] = []
        for index, result in enumerate(results):
            lines = [
                Line(text=line.text, top=float(line.bbox[1])) for line in result.text_lines
            ]
            recognized.append(Page.of(index, lines))
        return recognized
