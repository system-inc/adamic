package postcss

import (
	"errors"
	"testing"

	"github.com/system-inc/cohere/internal/format/printing"
)

// TestParseTree checks one small tree field by field without node: types, fields, raws, and byte
// offsets past non-ASCII text, with column left in UTF-16 units.
func TestParseTree(t *testing.T) {
	t.Parallel()
	// U+00E9 is 2 bytes and 1 unit, the emoji 4 bytes and 2 units.
	root, err := Parse("\xc3\xa9{b:c ! important}\n\xf0\x9f\x98\x80 {/* x */--d:{e}}")
	if err != nil {
		t.Fatal(err)
	}
	if root.Type() != "root" || root.Range != [2]int{0, 41} {
		t.Fatalf("root %s %v", root.Type(), root.Range)
	}
	rules := root.List("nodes")
	if len(rules) != 2 {
		t.Fatalf("%d rules", len(rules))
	}

	first := rules[0]
	if first.Type() != "rule" || first.String("selector") != "\xc3\xa9" || first.Range != [2]int{0, 19} {
		t.Errorf("first rule %s %q %v", first.Type(), first.String("selector"), first.Range)
	}
	declaration := first.List("nodes")[0]
	raws := declaration.Get("raws").(map[string]any)
	if declaration.String("prop") != "b" || declaration.String("value") != "c" || !declaration.Bool("important") || raws["important"] != " ! important" || raws["between"] != ":" {
		t.Errorf("declaration %v %v", declaration.Keys(), raws)
	}
	if keys := declaration.Keys(); len(keys) != 5 || keys[0] != "raws" || keys[1] != "source" || keys[2] != "prop" || keys[3] != "important" || keys[4] != "value" {
		t.Errorf("declaration keys %v", keys)
	}

	second := rules[1]
	start := second.Get("source").(map[string]any)["start"].(map[string]any)
	if start["offset"] != 20 || start["line"] != 2 || start["column"] != 1 {
		t.Errorf("second rule starts at %v", start)
	}
	children := second.List("nodes")
	comment := children[0]
	commentStart := comment.Get("source").(map[string]any)["start"].(map[string]any)
	// The comment starts at column 5 of line 2 (the emoji is two units), 26 bytes into the text.
	if comment.Type() != "comment" || comment.String("text") != "x" || commentStart["column"] != 5 || commentStart["offset"] != 26 {
		t.Errorf("comment %q at %v", comment.String("text"), commentStart)
	}
	custom := children[1]
	if custom.String("prop") != "--d" || custom.String("value") != "{e}" {
		t.Errorf("custom property %q %q", custom.String("prop"), custom.String("value"))
	}
}

// TestParseByteOrderMark checks that offsets count from after the byte order mark, as postcss's do.
func TestParseByteOrderMark(t *testing.T) {
	t.Parallel()
	root, err := Parse("\xef\xbb\xbfa{}")
	if err != nil {
		t.Fatal(err)
	}
	rule := root.List("nodes")[0]
	if rule.Range != [2]int{0, 3} {
		t.Errorf("rule range %v", rule.Range)
	}
}

// TestParseErrors checks that a refusal is marked as a syntax error and carries postcss's fields.
func TestParseErrors(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		text      string
		reason    string
		line      int
		column    int
		endColumn int
	}{
		{"a{", "Unclosed block", 1, 1, 0},
		{"a{b:\"c}", "Unclosed string", 1, 5, 0},
		{"\xf0\x9f\x98\x80 b", "Unknown word \xf0\x9f\x98\x80", 1, 1, 3},
		{"a{b:c d:e}", "Missed semicolon", 1, 6, 0},
		{"}", "Unexpected }", 1, 1, 2},
	} {
		_, err := Parse(each.text)
		var syntaxError *CssSyntaxError
		if !printing.IsSyntax(err) || !errors.As(err, &syntaxError) {
			t.Errorf("%q: %v is not a marked CssSyntaxError", each.text, err)
			continue
		}
		if syntaxError.Reason != each.reason || syntaxError.Line != each.line || syntaxError.Column != each.column || syntaxError.EndColumn != each.endColumn {
			t.Errorf("%q: %+v", each.text, *syntaxError)
		}
	}
}

// TestParseNeverPanics parses every prefix of every fixture: each must parse or be refused with a
// CssSyntaxError, never fail any other way.
func TestParseNeverPanics(t *testing.T) {
	t.Parallel()
	for _, fixture := range parseFixtures {
		for end := 0; end <= len(fixture.text); end++ {
			_, err := Parse(fixture.text[:end])
			if err != nil && !printing.IsSyntax(err) {
				t.Errorf("%s, first %d bytes: %v", fixture.name, end, err)
			}
		}
	}
}
