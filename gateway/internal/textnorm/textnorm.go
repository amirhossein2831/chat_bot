// Package textnorm makes Persian text comparable.
//
// Postgres has no Persian stemmer, so search relies on trigram matching over
// normalised text. Both stored values and queries must pass through Normalize
// or they will not match. The same rules are ported to TypeScript in the web
// client; if the two diverge, client-side filtering contradicts server-side
// search results.
package textnorm

import (
	"strings"
	"unicode"
)

// Codepoints that carry no meaning for matching and are removed outright.
const (
	zwnj    = '\u200c' // نیم‌فاصله: می‌روم / می روم / میروم are one word
	zwj     = '\u200d'
	lrm     = '\u200e'
	rlm     = '\u200f'
	bom     = '\ufeff'
	tatweel = '\u0640' // kashida, used purely to stretch justified text
)

// folds maps Arabic codepoints, and Persian letters that carry a hamza, onto
// the plain Persian form users are most likely to type.
var folds = map[rune]rune{
	'ي': 'ی', // Arabic yeh
	'ى': 'ی', // alef maksura
	'ك': 'ک', // Arabic kaf
	'أ': 'ا',
	'إ': 'ا',
	'آ': 'ا',
	'ٱ': 'ا',
	'ة': 'ه', // teh marbuta
}

// Normalize folds codepoint variants, strips joiners, diacritics and bidi
// marks, converts Persian and Arabic-Indic digits to Latin, lowercases, and
// collapses whitespace. It is idempotent.
func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	lastWasSpace := false
	for _, r := range s {
		switch {
		case r == zwnj, r == zwj, r == lrm, r == rlm, r == bom, r == tatweel:
			continue
		case isDiacritic(r):
			continue
		case unicode.IsSpace(r):
			// Collapse any run of whitespace, including newlines, to one gap.
			if !lastWasSpace {
				b.WriteRune(' ')
				lastWasSpace = true
			}
			continue
		}

		lastWasSpace = false
		b.WriteRune(foldRune(r))
	}

	return strings.TrimSpace(b.String())
}

func foldRune(r rune) rune {
	if folded, ok := folds[r]; ok {
		return folded
	}
	switch {
	case r >= '۰' && r <= '۹': // Persian digits ۰-۹
		return r - '۰' + '0'
	case r >= '٠' && r <= '٩': // Arabic-Indic digits ٠-٩
		return r - '٠' + '0'
	}
	return unicode.ToLower(r)
}

// isDiacritic covers the Arabic harakat plus the superscript alef, none of
// which most users type and all of which break matching when present.
func isDiacritic(r rune) bool {
	return (r >= 'ً' && r <= 'ٕ') || r == 'ٰ'
}
