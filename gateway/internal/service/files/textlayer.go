package files

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// A page carrying fewer characters than this is a scan with a stray header,
	// not a page of text.
	minCharsPerPage = 50

	// Below this share of pages carrying text, the document is a scan.
	minPagesWithText = 0.3

	// A Persian document whose text is barely Arabic-block is not Persian text.
	minArabicShare = 0.3

	// Above this share of Latin-1 Supplement letters, the text is garbled: they
	// are rare in real English and absent from Persian.
	maxGarbledShare = 0.2

	// Enough characters to judge the ratio on. Below it, judge nothing.
	minCharsToJudge = 200
)

// UsableTextLayer reports whether a PDF's own text can be trusted, which is two
// questions rather than one.
//
// The second is Persian-specific and is the expensive one to get wrong. Many
// Iranian PDFs from older publishing tools carry a text layer with a broken
// ToUnicode map: extraction returns confident-looking mojibake rather than
// failing. Feeding that to a model produces a confident answer drawn from
// garbage, which is the worst outcome available here because nothing looks
// broken at any point.
func UsableTextLayer(pages []string) bool {
	if len(pages) == 0 {
		return false
	}

	withText := 0
	total := 0
	for _, page := range pages {
		trimmed := strings.TrimSpace(page)
		total += utf8.RuneCountInString(trimmed)
		if utf8.RuneCountInString(trimmed) >= minCharsPerPage {
			withText++
		}
	}

	if withText == 0 || float64(withText)/float64(len(pages)) < minPagesWithText {
		return false
	}
	if total/len(pages) < minCharsPerPage {
		return false
	}

	return !isMojibake(strings.Join(pages, "\n"))
}

// isMojibake judges whether a text layer that extracted successfully is
// nonetheless nonsense. A broken CMap maps Persian letters into the Latin-1
// Supplement range, which is where the signature lies: those letters are rare
// in genuine English and do not appear in Persian at all. Judging by the
// absence of Arabic instead would condemn every English document.
func isMojibake(text string) bool {
	var arabic, garbled, letters, replacement int

	for _, r := range text {
		switch {
		case r == unicode.ReplacementChar:
			replacement++
		case r >= 0x0600 && r <= 0x06FF, r >= 0xFB50 && r <= 0xFDFF, r >= 0xFE70 && r <= 0xFEFF:
			arabic++
			letters++
		case r >= 0x00C0 && r <= 0x024F:
			garbled++
			letters++
		case unicode.IsLetter(r):
			letters++
		}
	}

	if replacement > letters/20 {
		return true
	}
	if letters < minCharsToJudge {
		return false
	}
	if float64(garbled)/float64(letters) > maxGarbledShare {
		return true
	}
	// Some Persian survived, but not enough of it to be the document's language.
	return arabic > 0 && float64(arabic)/float64(letters) < minArabicShare
}
