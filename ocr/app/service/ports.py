from typing import Protocol, runtime_checkable

from app.domain.models import Page


@runtime_checkable
class Recognizer(Protocol):
    """What the service needs from a recogniser, declared by the caller.

    An adapter satisfies this without importing it, which is what lets the
    service run with no model loaded.
    """

    @property
    def ready(self) -> bool: ...

    def recognize(self, pages: list[bytes], languages: list[str]) -> list[Page]:
        """Blocking. The service is responsible for keeping it off the loop."""
        ...
