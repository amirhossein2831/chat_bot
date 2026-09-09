package textnorm

import "testing"

func TestNormalize(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"arabic yeh":            {"يک", "یک"},
		"alef maksura":          {"کتابى", "کتابی"},
		"arabic kaf":            {"كتاب", "کتاب"},
		"alef with hamza above": {"أسان", "اسان"},
		"alef with hamza below": {"إيران", "ایران"},
		"alef with madda":       {"آب", "اب"},
		"teh marbuta":           {"مدرسة", "مدرسه"},

		"zwnj removed": {"می\u200cروم", "میروم"},
		"zwj removed":  {"می\u200dروم", "میروم"},

		"persian digits":      {"۱۲۳", "123"},
		"arabic indic digits": {"١٢٣", "123"},
		"mixed digits":        {"۱2٣", "123"},

		"diacritics stripped": {"کِتابُ", "کتاب"},
		"tatweel stripped":    {"کتـــاب", "کتاب"},
		"bom stripped":        {"\ufeffکتاب", "کتاب"},
		"bidi marks stripped": {"\u200fکتاب\u200e", "کتاب"},

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

func TestSpellingVariantsConverge(t *testing.T) {
	variants := []string{"می\u200cروم", "میروم", "مي\u200cروم", "مي\u200cرٌوم"}
	want := Normalize(variants[0])
	for _, v := range variants[1:] {
		if got := Normalize(v); got != want {
			t.Fatalf("Normalize(%q) = %q, want %q", v, got, want)
		}
	}
}

// Stripping spaces would merge کتاب خانه with کتابخانه.
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
