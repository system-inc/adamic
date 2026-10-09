package values

import (
	"math"
	"testing"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// parseNodes parses with Prettier's options and returns the Value's children.
func parseNodes(t *testing.T, text string) []*estree.Node {
	t.Helper()
	root, err := Parse(text, Options{Loose: true})
	if err != nil {
		t.Fatalf("Parse(%q): %v", text, err)
	}
	if !root.Is("root") || len(root.List("nodes")) != 1 || !root.List("nodes")[0].Is("value") {
		t.Fatalf("Parse(%q): want a root holding one value, got %v", text, root.Keys())
	}
	return root.List("nodes")[0].List("nodes")
}

func sourceColumn(node *estree.Node, edge string) any {
	return node.Get("source").(map[string]any)[edge].(map[string]any)["column"]
}

// TestParseShapes checks, without node, the fields the oracle found that are easiest to get wrong:
// the shapes Prettier's printer and loc.js lean on, and upstream's quirks.
func TestParseShapes(t *testing.T) {
	t.Parallel()
	nodes := parseNodes(t, "-5px --foo calc(1px + 2px) #fff \"a\" /* c */")
	types := []string{"number", "word", "func", "word", "string", "comment"}
	if len(nodes) != len(types) {
		t.Fatalf("want %d nodes, got %d", len(types), len(nodes))
	}
	for index, nodeType := range types {
		if !nodes[index].Is(nodeType) {
			t.Errorf("node %d: want %s, got %s", index, nodeType, nodes[index].Type())
		}
	}

	// A signed number splits into value and unit, and its sourceIndex is the merged word's last token.
	if nodes[0].String("value") != "-5" || nodes[0].String("unit") != "px" || nodes[0].Get("sourceIndex") != 1 {
		t.Errorf("-5px: got value %q unit %q sourceIndex %v", nodes[0].String("value"), nodes[0].String("unit"), nodes[0].Get("sourceIndex"))
	}
	// --foo is "--" and "foo" merged, indexed at "foo"; loc.js's fixValueWordLoc undoes that.
	if nodes[1].String("value") != "--foo" || nodes[1].Get("sourceIndex") != 7 || raws(nodes[1])["before"] != " " {
		t.Errorf("--foo: got value %q sourceIndex %v raws %v", nodes[1].String("value"), nodes[1].Get("sourceIndex"), raws(nodes[1]))
	}
	// A function holds its parens and arguments, and is balanced again at the end.
	function := nodes[2].List("nodes")
	if len(function) != 5 || !function[0].Is("paren") || !function[2].Is("operator") || !function[4].Is("paren") || nodes[2].Get("unbalanced") != float64(0) {
		t.Errorf("calc: got %d children, unbalanced %v", len(function), nodes[2].Get("unbalanced"))
	}
	// The operator's sourceIndex is its token's end line, 1, not its index.
	if function[2].Get("sourceIndex") != 1 {
		t.Errorf("operator sourceIndex: want 1 (the line), got %v", function[2].Get("sourceIndex"))
	}
	if !nodes[3].Bool("isHex") || !nodes[3].Bool("isColor") {
		t.Errorf("#fff: want isHex and isColor")
	}
	if nodes[4].String("value") != "a" || !nodes[4].Bool("quoted") || raws(nodes[4])["quote"] != "\"" {
		t.Errorf("string: got %q quoted %v raws %v", nodes[4].String("value"), nodes[4].Get("quoted"), raws(nodes[4]))
	}
	if nodes[5].String("value") != " c " || nodes[5].Bool("inline") {
		t.Errorf("comment: got %q inline %v", nodes[5].String("value"), nodes[5].Get("inline"))
	}

	// A value opening with a paren ends that paren at column NaN; the next paren reuses a stale end.
	parens := parseNodes(t, "(a)")
	if column, isFloat := sourceColumn(parens[0], "end").(float64); !isFloat || !math.IsNaN(column) {
		t.Errorf("first paren end column: want NaN, got %v", sourceColumn(parens[0], "end"))
	}
	if sourceColumn(parens[2], "end") != 2 {
		t.Errorf("closing paren end column: want 2, got %v", sourceColumn(parens[2], "end"))
	}

	// sourceIndex is bytes; columns stay UTF-16 units.
	multibyte := parseNodes(t, "\xc3\xa9 \xf0\x9f\x98\x80x b")
	if multibyte[1].Get("sourceIndex") != 3 || multibyte[2].Get("sourceIndex") != 9 || sourceColumn(multibyte[2], "start") != 7 {
		t.Errorf("non-ASCII: got sourceIndex %v and %v, column %v", multibyte[1].Get("sourceIndex"), multibyte[2].Get("sourceIndex"), sourceColumn(multibyte[2], "start"))
	}

	// Inline comments only in loose mode.
	if comment := parseNodes(t, "a // b")[1]; !comment.Is("comment") || !comment.Bool("inline") || comment.String("value") != " b" {
		t.Errorf("inline comment: got %s %q", comment.Type(), comment.String("value"))
	}
	// Outside loose mode url() arguments become one word.
	strict, err := Parse("url(a b)", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if url := strict.List("nodes")[0].List("nodes")[0].List("nodes"); len(url) != 3 || url[1].String("value") != "a b" || url[1].Has("isHex") {
		t.Errorf("strict url: got %d children", len(url))
	}
}

func TestParseRefusals(t *testing.T) {
	t.Parallel()
	for text, want := range map[string]string{
		"a(b":   "ParserError: Expected closing parenthesis at line: 1, column 2",
		"a) b":  "ParserError: Expected opening parenthesis at line: 1, column 2",
		"\"abc": "TokenizeError: Unclosed quote at line: 1, column: 1, token: 0",
		"a /*":  "TokenizeError: Unclosed comment at line: 1, column: 3, token: 2",
		" ":     "TypeError: Cannot read properties of undefined (reading 'raws')",
		"a + +": "TypeError: Cannot read properties of undefined (reading '0')",
	} {
		_, err := Parse(text, Options{Loose: true})
		if err == nil || !printing.IsSyntax(err) || err.Error() != want {
			t.Errorf("Parse(%q): want the syntax error %q, got %v", text, want, err)
		}
	}
	if _, err := Parse("calc(1 -2)", Options{}); err == nil {
		t.Errorf("calc(1 -2) without loose: want a refusal")
	}
	if _, err := Parse("calc(1 -2)", Options{Loose: true}); err != nil {
		t.Errorf("calc(1 -2) with loose: %v", err)
	}
}
