package postcss

import (
	"errors"
	"reflect"
	"testing"

	"github.com/system-inc/cohere/internal/format/printing"
)

// TestParseSCSSTree checks what postcss-scss adds, without node: an inline comment's raws and rewritten
// text, a nested declaration's fields and their order, raws.value.scss, and an at-rule name joined with
// the interpolation after it.
func TestParseSCSSTree(t *testing.T) {
	t.Parallel()
	root, err := ParseSCSS("// a /* b\n.x { margin: 0 // c\n { left: 1px } }\n@media#{$q} {}")
	if err != nil {
		t.Fatal(err)
	}
	nodes := root.List("nodes")
	if len(nodes) != 3 {
		t.Fatalf("%d nodes", len(nodes))
	}

	comment := nodes[0]
	commentRaws := comment.Get("raws").(map[string]any)
	if comment.Type() != "comment" || comment.String("text") != "a *//* b" || commentRaws["inline"] != true || commentRaws["text"] != "a /* b" || commentRaws["left"] != " " || commentRaws["right"] != "" {
		t.Errorf("comment %q %v", comment.String("text"), commentRaws)
	}
	if comment.Range != [2]int{0, 9} {
		t.Errorf("comment range %v", comment.Range)
	}

	nested := nodes[1].List("nodes")[0]
	if nested.Type() != "decl" || !nested.Bool("isNested") || nested.String("prop") != "margin" || len(nested.List("nodes")) != 1 {
		t.Errorf("nested declaration %s %v %q", nested.Type(), nested.Get("isNested"), nested.String("prop"))
	}
	if keys := nested.Keys(); !reflect.DeepEqual(keys, []string{"raws", "isNested", "nodes", "source", "prop", "value"}) {
		t.Errorf("nested declaration keys %v", keys)
	}
	value := nested.Get("raws").(map[string]any)["value"]
	if want := map[string]any{"value": "0 ", "raw": "0 /* c*/\n ", "scss": "0 // c\n "}; !reflect.DeepEqual(value, want) {
		t.Errorf("raws.value %#v", value)
	}

	media := nodes[2]
	if media.String("name") != "media#{$q}" || media.String("params") != "" {
		t.Errorf("at-rule %q %q", media.String("name"), media.String("params"))
	}
}

// TestParseSCSSErrors checks the refusals postcss-scss adds, and that its TypeError is marked too.
func TestParseSCSSErrors(t *testing.T) {
	t.Parallel()
	for _, each := range []struct {
		text   string
		reason string
		column int
	}{
		{"a { b: #{c", "Unclosed interpolation", 8},
		{"a { b: \"#{c\" }", "Unclosed interpolation", 8},
		{"a{b:\"", "Unclosed string", 5},
	} {
		_, err := ParseSCSS(each.text)
		var syntaxError *CssSyntaxError
		if !printing.IsSyntax(err) || !errors.As(err, &syntaxError) {
			t.Errorf("%q: %v is not a marked CssSyntaxError", each.text, err)
			continue
		}
		if syntaxError.Reason != each.reason || syntaxError.Line != 1 || syntaxError.Column != each.column {
			t.Errorf("%q: %+v", each.text, *syntaxError)
		}
	}

	_, err := ParseSCSS("a { :\"x\" {} }")
	var typeError *TypeError
	if !printing.IsSyntax(err) || !errors.As(err, &typeError) {
		t.Errorf("a nested property with no word: %v is not a marked TypeError", err)
	}

	// The css parser reads none of it: `//` is a word there and #{ is not interpolation.
	if _, err := Parse("a { b: #{c }"); err == nil {
		t.Errorf("Parse accepted an scss-only input")
	}
}

// TestParseSCSSNeverPanics parses every prefix of every scss fixture: each must parse or be refused with
// a marked error, never fail any other way.
func TestParseSCSSNeverPanics(t *testing.T) {
	t.Parallel()
	for _, fixture := range scssParseFixtures {
		for end := 0; end <= len(fixture.text); end++ {
			_, err := ParseSCSS(fixture.text[:end])
			if err != nil && !printing.IsSyntax(err) {
				t.Errorf("%s, first %d bytes: %v", fixture.name, end, err)
			}
		}
	}
}
