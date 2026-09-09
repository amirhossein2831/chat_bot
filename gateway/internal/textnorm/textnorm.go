package textnorm

import (
	"strings"
	"unicode"
)

const (
	zwnj    = '\u200c'
	zwj     = '\u200d'
	lrm     = '\u200e'
	rlm     = '\u200f'
	bom     = '\ufeff'
	tatweel = '\u0640'
)

var folds = map[rune]rune{
	'ي': 'ی',
	'ى': 'ی',
	'ك': 'ک',
	'أ': 'ا',
	'إ': 'ا',
	'آ': 'ا',
	'ٱ': 'ا',
	'ة': 'ه',
}

// Stored values and search queries must both pass through this or they will
// not match. Spaces are left alone on purpose: stripping them would merge
// کتاب خانه with کتابخانه.
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
	case r >= '۰' && r <= '۹':
		return r - '۰' + '0'
	case r >= '٠' && r <= '٩':
		return r - '٠' + '0'
	}
	return unicode.ToLower(r)
}

func isDiacritic(r rune) bool {
	return (r >= 'ً' && r <= 'ٕ') || r == 'ٰ'
}
