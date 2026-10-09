package regexp

import "testing"

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern, flags string
		valid          bool
	}{
		{"[^[\\q{ab|a}]&&[a]]", "v", true}, {"[^[\\q{ab|a}]--[\\q{ab}]]", "v", false},
		{"[🌍-🌎]", "", false}, {"[🌍-🌎]", "u", true},
		{"[a-🌍]", "", true}, {"[[]", "", true}, {"[^\\q{}]", "v", false},
		{"(?:a|b)+?", "gi", true}, {"(?<word>\\p{Letter}+)\\k<word>", "u", true},
		{"(?<=a)b(?<!c)", "", true}, {"[a-z&&[^aeiou]]", "v", false}, {"[[a-z]&&[^aeiou]]", "v", true}, {"[ab&&c]", "v", false}, {"[a&&b-z]", "v", false},
		{"[\\q{ab|cd}]", "v", true}, {"(a)\\1", "u", true},
		{"a{3,2}", "", false}, {"(?<x>a)(?<x>b)", "", false},
		{"a{2,2}", "", true},
		{"\\0\\cA\\p{sc=Latn}", "u", true}, {"\\01", "u", false},
		{"]", "u", false}, {"}", "u", false},
		{"\\p{not_a_property}", "u", false}, {"\\8", "u", false},
		{"\\8", "", true}, {"(?=a)+", "u", false}, {"[z-a]", "", false},
	}
	for _, test := range tests {
		_, err := Parse(test.pattern, test.flags)
		if (err == nil) != test.valid {
			t.Errorf("Parse(%q,%q) error=%v, want valid=%v", test.pattern, test.flags, err, test.valid)
		}
	}
}

func TestFlags(t *testing.T) {
	t.Parallel()
	for _, flags := range []string{"uu", "uv", "z"} {
		if _, err := Parse("", flags); err == nil {
			t.Errorf("flags %q accepted", flags)
		}
	}
}

func TestQuantifierBounds(t *testing.T) {
	t.Parallel()
	p, err := Parse("a{2,4}?", "")
	if err != nil {
		t.Fatal(err)
	}
	q := p.Body.Alternatives[0].Terms[0].(*Quantifier)
	if q.Min.String() != "2" || q.Max.String() != "4" || q.Greedy {
		t.Fatalf("quantifier = %#v", q)
	}
}
