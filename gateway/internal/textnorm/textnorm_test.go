package textnorm

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		// Users type Arabic and Persian codepoints interchangeably depending
		// on their keyboard; search must not care which.
		"arabic yeh":            {"يک", "یک"},
		"alef maksura":          {"کتابى", "کتابی"},
		"arabic kaf":            {"كتاب", "کتاب"},
		"alef with hamza above": {"أسان", "اسان"},
		"alef with hamza below": {"إيران", "ایران"},
		"alef with madda":       {"آب", "اب"},
		"teh marbuta":           {"مدرسة", "مدرسه"},

		// Half-space: the same word is written three ways in the wild.
		"zwnj removed": {"می\u200cروم", "میروم"},
		"zwj removed":  {"می\u200dروم", "میروم"},

		// Three digit systems in circulation.
		"persian digits":      {"۱۲۳", "123"},
		"arabic indic digits": {"١٢٣", "123"},
		"mixed digits":        {"۱2٣", "123"},

		// Typography that carries no meaning for matching.
		"diacritics stripped": {"کِتابُ", "کتاب"},
		"tatweel stripped":    {"کتـــاب", "کتاب"},
		"bom stripped":        {"\ufeffکتاب", "کتاب"},
		"bidi marks stripped": {"\u200fکتاب\u200e", "کتاب"},

		// Whitespace and case.
		"whitespace collapsed": {"  دو   کلمه ", "دو کلمه"},
		"newlines collapsed":   {"خط\nدوم", "خط دوم"},
		"latin lowercased":     {"Claude AI", "claude ai"},

		"empty": {"", ""},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := Normalize(c.in); got != c.want {
				t.Fatalf("Normalize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// می‌روم written with a ZWNJ, with nothing at all, and with an Arabic yeh are
// the same token and must converge, or search silently fails to find things.
func TestSpellingVariantsConverge(t *testing.T) {
	variants := []string{"می\u200cروم", "میروم", "مي\u200cروم", "مي\u200cرٌوم"}
	want := Normalize(variants[0])
	for _, v := range variants[1:] {
		if got := Normalize(v); got != want {
			t.Fatalf("Normalize(%q) = %q, want %q", v, got, want)
		}
	}
}

// A real space is deliberately NOT removed, even though users sometimes type
// "می روم" for "می‌روم". Stripping spaces would merge genuinely different
// words: کتاب خانه (a book house) and کتابخانه (a library) are not the same
// thing, and collapsing them would return wrong results far more often than
// the spaced half-space spelling would return missing ones.
func TestSpacesAreNotStripped(t *testing.T) {
	if got := Normalize("می روم"); got != "می روم" {
		t.Fatalf("Normalize(%q) = %q, want the space preserved", "می روم", got)
	}
	if Normalize("کتاب خانه") == Normalize("کتابخانه") {
		t.Fatal("distinct words must not collapse into one")
	}
}

func TestNormalizeIsIdempotent(t *testing.T) {
	for _, in := range []string{"می\u200cروم", "۱۲۳ كتاب", "  a  B  "} {
		once := Normalize(in)
		if twice := Normalize(once); twice != once {
			t.Fatalf("Normalize not idempotent for %q: %q then %q", in, once, twice)
		}
	}
}
