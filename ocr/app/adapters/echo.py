from app.domain.models import Line, Page


class EchoRecognizer:
    """Stands in for Surya where it cannot be installed.

    Torch and the Surya weights are several gigabytes and the weight CDN is
    unreachable from Iran, so a machine without them still needs a way to
    exercise the upload, queue and extraction path end to end. It reports what
    it received and never pretends to have read anything.
    """

    @property
    def ready(self) -> bool:
        return True

    def recognize(self, pages: list[bytes], languages: list[str]) -> list[Page]:
        return [
            Page.of(
                index,
                [Line(text=f"[no recogniser: page {index + 1}, {len(page)} bytes]", top=0.0)],
            )
            for index, page in enumerate(pages)
        ]
