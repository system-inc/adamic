package selector

import (
	"errors"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// These run without node. The oracle test is the real check; this pins the shapes Prettier's printer
// leans on, the byte conversion, and the refusals, so a regression shows in a plain `go test`.

func mustParse(t *testing.T, text string) *estree.Node {
	t.Helper()
	root, err := Parse(text)
	if err != nil {
		t.Fatalf("Parse(%q): %v", text, err)
	}
	if !root.Is("root") {
		t.Fatalf("Parse(%q) returned a %q", text, root.Type())
	}
	return root
}

func TestParseTailwindEscapedClass(t *testing.T) {
	t.Parallel()
	root := mustParse(t, ".hover\\:bg-red:hover")
	selectors := root.List("nodes")
	if len(selectors) != 1 || !selectors[0].Is("selector") {
		t.Fatalf("want one selector, got %v", selectors)
	}
	nodes := selectors[0].List("nodes")
	if len(nodes) != 2 || !nodes[0].Is("class") || !nodes[1].Is("pseudo") {
		t.Fatalf("want a class then a pseudo, got %d nodes", len(nodes))
	}
	if nodes[0].String("value") != "hover\\:bg-red" || nodes[1].String("value") != ":hover" {
		t.Errorf("values %q and %q", nodes[0].String("value"), nodes[1].String("value"))
	}
	// splitWord positions the merged word from its last token, "bg-red" at 8: upstream's quirk, kept.
	if nodes[0].Get("sourceIndex") != 8 {
		t.Errorf("class sourceIndex %v, want upstream's 8", nodes[0].Get("sourceIndex"))
	}
	if nodes[1].Get("sourceIndex") != 14 {
		t.Errorf("pseudo sourceIndex %v, want 14", nodes[1].Get("sourceIndex"))
	}
}

func TestParseAttribute(t *testing.T) {
	t.Parallel()
	root := mustParse(t, "[ns|data-x='b' i]")
	attribute := root.List("nodes")[0].List("nodes")[0]
	if !attribute.Is("attribute") {
		t.Fatalf("got a %q", attribute.Type())
	}
	checks := map[string]any{
		"attribute": "data-x", "namespace": "ns", "operator": "=", "value": "'b'",
		"insensitive": true, "quoted": true, "sourceIndex": 0,
	}
	for key, want := range checks {
		if attribute.Get(key) != want {
			t.Errorf("%s = %#v, want %#v", key, attribute.Get(key), want)
		}
	}
	raws := rawsOf(attribute)
	if raws["insensitive"] != " i" || raws["unquoted"] != "b" {
		t.Errorf("raws %v", raws)
	}

	bare := mustParse(t, "[a]").List("nodes")[0].List("nodes")[0]
	if !bare.Has("operator") || bare.Get("operator") != nil || !bare.Has("value") || bare.Get("value") != nil {
		t.Errorf("an attribute without an operator keeps operator and value as own undefined fields")
	}
	if bare.Has("quoted") || bare.Has("insensitive") {
		t.Errorf("an attribute without a value has no quoted or insensitive")
	}
}

func TestParseCombinatorAndSpaces(t *testing.T) {
	t.Parallel()
	root := mustParse(t, " a  >\tb , c")
	selectors := root.List("nodes")
	if len(selectors) != 2 {
		t.Fatalf("want two selectors, got %d", len(selectors))
	}
	nodes := selectors[0].List("nodes")
	if len(nodes) != 3 || !nodes[1].Is("combinator") {
		t.Fatalf("want tag, combinator, tag")
	}
	if spacesOf(nodes[0])["before"] != " " {
		t.Errorf("leading space %q", spacesOf(nodes[0])["before"])
	}
	combinator := nodes[1]
	if combinator.String("value") != ">" || spacesOf(combinator)["before"] != "  " || spacesOf(combinator)["after"] != "\t" {
		t.Errorf("combinator %q spaces %v", combinator.String("value"), spacesOf(combinator))
	}
	if spacesOf(nodes[2])["after"] != " " {
		t.Errorf("space before the comma %q", spacesOf(nodes[2])["after"])
	}
	if spacesOf(selectors[1].List("nodes")[0])["before"] != " " {
		t.Errorf("space after the comma not on c")
	}
	if mustParse(t, "a,").Get("trailingComma") != true {
		t.Errorf("trailingComma not set")
	}
}

func TestParsePseudoWithSelectors(t *testing.T) {
	t.Parallel()
	root := mustParse(t, ":nth-child(2n + 1 of .x)")
	pseudo := root.List("nodes")[0].List("nodes")[0]
	if !pseudo.Is("pseudo") || pseudo.String("value") != ":nth-child" {
		t.Fatalf("got a %q %q", pseudo.Type(), pseudo.String("value"))
	}
	inner := pseudo.List("nodes")
	if len(inner) != 1 || !inner[0].Is("selector") {
		t.Fatalf("want one inner selector")
	}
	var types []string
	for _, node := range inner[0].List("nodes") {
		types = append(types, node.Type()+":"+node.String("value"))
	}
	want := "tag:2n combinator:+ tag:1 combinator:  tag:of combinator:  class:x"
	if strings.Join(types, " ") != want {
		t.Errorf("inner nodes\n got %s\nwant %s", strings.Join(types, " "), want)
	}
	end := sourceOf(pseudo)["end"].(map[string]any)
	if end["line"] != 1 || end["column"] != 24 {
		t.Errorf("pseudo end %v, want the closing parenthesis at 1:24 (columns are 1-based)", end)
	}
}

func TestParseOffsetsAreBytes(t *testing.T) {
	t.Parallel()
	// ".e-acute > .u-umlaut": the second class starts at UTF-16 index 5 and byte 6. Columns stay UTF-16.
	root := mustParse(t, ".\xc3\xa9 > .\xc3\xbc")
	nodes := root.List("nodes")[0].List("nodes")
	second := nodes[2]
	if second.Get("sourceIndex") != 6 {
		t.Errorf("sourceIndex %v, want byte 6", second.Get("sourceIndex"))
	}
	if column := sourceOf(second)["start"].(map[string]any)["column"]; column != 6 {
		t.Errorf("column %v, want UTF-16 column 6", column)
	}
	// An emoji is two UTF-16 units and four bytes; splitWord's sourceIndex past the end stays past it.
	merged := mustParse(t, ".\xf0\x9f\x98\x80\\:x.y").List("nodes")[0].List("nodes")
	if merged[1].Get("sourceIndex") != len(".\xf0\x9f\x98\x80\\:x.y")+3 {
		t.Errorf("past-the-end sourceIndex %v", merged[1].Get("sourceIndex"))
	}
}

func TestParseRefusals(t *testing.T) {
	t.Parallel()
	refusals := map[string]string{
		"\"unclosed":  "Unclosed quote",
		"a /* x":      "Unclosed comment",
		"[a":          "Expected a closing square bracket.",
		"a]":          "Expected opening square bracket.",
		")":           "Expected opening parenthesis.",
		":not(a":      "Expected closing parenthesis.",
		"a;b":         "Expected a backslash preceding the semicolon.",
		"a:":          "Expected pseudo-class or pseudo-element",
		"a:[x]":       "Unexpected \"[\" found.",
		":hover.foo(": "Misplaced parenthesis.",
		".a|":         "Cannot read properties of undefined (reading '0')",
		"(a)":         "Cannot read properties of undefined (reading 'value')",
		"[a=\"]\"":    "Cannot read properties of undefined (reading '2')",
	}
	for text, message := range refusals {
		_, err := Parse(text)
		if err == nil || err.Error() != message || !printing.IsSyntax(err) {
			t.Errorf("Parse(%q) = %v, want the syntax error %q", text, err, message)
		}
	}
	for _, text := range []string{"a| b", ":is(a|)", "a|@x", "*|)"} {
		_, err := Parse(text)
		if !errors.Is(err, ErrLoopsForever) || !printing.IsSyntax(err) {
			t.Errorf("Parse(%q) = %v, want ErrLoopsForever marked as a syntax error", text, err)
		}
	}
}

// TestParseNeverPanics feeds random fragments through Parse: every outcome must be a tree or a syntax
// error, never a Go panic turned into a plain error.
func TestParseNeverPanics(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewPCG(5, 7))
	pieces := []string{"a", ".", "#", ":", "(", ")", "[", "]", "=", "|", "*", "&", ",", " ", ">", "\\", "'", "\"", "/*", "*/", "@", ";", "i", "\xf0\x9f\x98\x80", "#{"}
	for range 20000 {
		var text strings.Builder
		for range 1 + random.IntN(8) {
			text.WriteString(pieces[random.IntN(len(pieces))])
		}
		if _, err := Parse(text.String()); err != nil && !printing.IsSyntax(err) {
			t.Fatalf("Parse(%q): %v", text.String(), err)
		}
	}
}
