const FOLD: Record<string, string> = {
  "ي": "ی",
  "ى": "ی",
  "ك": "ک",
  "أ": "ا",
  "إ": "ا",
  "آ": "ا",
  "ٱ": "ا",
  "ة": "ه",
};

// Escapes rather than literals: these codepoints are invisible in an editor,
// and several are rejected outright inside a character class.
const DROP = /[\u200c\u200d\u200e\u200f\ufeff\u0640\u064b-\u0655\u0670]/gu;
const PERSIAN_DIGITS = /[\u06f0-\u06f9]/gu;
const ARABIC_DIGITS = /[\u0660-\u0669]/gu;

// A port of gateway/internal/util/textnorm. If the two diverge, client-side
// filtering contradicts server-side search. Spaces survive on purpose:
// stripping them would merge کتاب خانه with کتابخانه.
export function normalize(input: string): string {
  return input
    .replace(DROP, "")
    .replace(PERSIAN_DIGITS, (d) => String(d.charCodeAt(0) - 0x06f0))
    .replace(ARABIC_DIGITS, (d) => String(d.charCodeAt(0) - 0x0660))
    .replace(/./gu, (c) => FOLD[c] ?? c)
    .toLowerCase()
    .replace(/\s+/g, " ")
    .trim();
}
