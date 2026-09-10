from dataclasses import dataclass
from typing import Self


@dataclass(frozen=True, slots=True)
class Line:
    """One recognised line, in the order the recogniser read it.

    `top` is the vertical position on the page and exists only so lines can be
    sorted down the page. The text itself is never reordered: it is stored in
    logical order, and reversing it for display is the reader's job.
    """

    text: str
    top: float


@dataclass(frozen=True, slots=True)
class Page:
    index: int
    lines: tuple[Line, ...]

    @classmethod
    def of(cls, index: int, lines: list[Line]) -> Self:
        return cls(index=index, lines=tuple(sorted(lines, key=lambda line: line.top)))

    @property
    def text(self) -> str:
        return "\n".join(line.text for line in self.lines if line.text)


@dataclass(frozen=True, slots=True)
class Recognition:
    pages: tuple[Page, ...]

    @property
    def text(self) -> str:
        return "\n\n".join(page.text for page in self.pages if page.text)

    @property
    def characters(self) -> int:
        return sum(len(page.text) for page in self.pages)
