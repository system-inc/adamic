package css

import (
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/printing"
)

// The always-run checks on the parser glue, without node: the shapes the printer relies on and byte
// offsets worked out by hand. parse_oracle_test.go is the full comparison against the fork.

func mustParse(t *testing.T, text string) *estree.Node {
	t.Helper()
	root, err := parse(text)
	if err != nil {
		t.Fatalf("parse(%q): %v", text, err)
	}
	return root
}

// firstDeclaration is the first css-decl under the root's first rule.
func firstDeclaration(t *testing.T, root *estree.Node) *estree.Node {
	t.Helper()
	rule := root.List("nodes")[0]
	declaration := rule.List("nodes")[0]
	if !declaration.Is("css-decl") {
		t.Fatalf("want a css-decl, got %s", declaration.Type())
	}
	return declaration
}

// valueChain is the value-root's group, the value-value's group: the single node of a one-word value.
func valueChain(declaration *estree.Node) []*estree.Node {
	root := declaration.Child("value")
	value := root.Child("group")
	return []*estree.Node{root, value, value.Child("group")}
}

func TestParseGlueTypesAndOffsets(t *testing.T) {
	t.Parallel()
	text := "a { b: x --c }"
	root := mustParse(t, text)
	if !root.Is("css-root") || locStart(root) != 0 || locEnd(root) != len(text) {
		t.Fatalf("root: %s [%d, %d]", root.Type(), locStart(root), locEnd(root))
	}
	rule := root.List("nodes")[0]
	if !rule.Is("css-rule") || !rule.Child("selector").Is("selector-root") || rawsIn(rule)["selector"] != "a" {
		t.Fatalf("rule: %s, selector %s, raws.selector %v", rule.Type(), rule.Child("selector").Type(), rawsIn(rule)["selector"])
	}

	declaration := firstDeclaration(t, root)
	valueRoot := declaration.Child("value")
	if !valueRoot.Is("value-root") || valueRoot.String("text") != "x --c" {
		t.Fatalf("value root: %s %q", valueRoot.Type(), valueRoot.String("text"))
	}
	// The value-root starts after the prop and raws.between: "b" + ": ".
	if want := strings.Index(text, "x --c"); locStart(valueRoot) != want {
		t.Errorf("value-root starts at %d, want %d", locStart(valueRoot), want)
	}
	words := valueRoot.Child("group").Child("group").List("groups")
	if len(words) != 2 || !words[1].Is("value-word") || words[1].String("value") != "--c" {
		t.Fatalf("words: %v", words)
	}
	// fixValueWordLoc moves a word that starts with -- back over its dashes.
	if want := strings.Index(text, "--c"); locStart(words[1]) != want || locEnd(words[1]) != want+3 {
		t.Errorf("--c is [%d, %d], want [%d, %d]", locStart(words[1]), locEnd(words[1]), want, want+3)
	}
	for _, node := range []*estree.Node{root, rule, declaration, valueRoot, words[1]} {
		if node.Range != [2]int{locStart(node), locEnd(node)} {
			t.Errorf("%s: Range %v, source offsets [%d, %d]", node.Type(), node.Range, locStart(node), locEnd(node))
		}
	}
}

func TestParseGlueOffsetsAreBytes(t *testing.T) {
	t.Parallel()
	// "\xc3\xa9" is one UTF-16 unit and two bytes, "\xf0\x9f\x98\x80" two units and four bytes.
	text := "/* \xf0\x9f\x98\x80 */\na { b: \"\xc3\xa9\" c }\n.\xc3\xa9 { d: e }"
	root := mustParse(t, text)
	rule := root.List("nodes")[1]
	words := rule.List("nodes")[0].Child("value").Child("group").Child("group").List("groups")
	if len(words) != 2 || !words[1].Is("value-word") {
		t.Fatalf("words: %v", words)
	}
	if want := strings.Index(text, " c ") + 1; locStart(words[1]) != want || locEnd(words[1]) != want+1 {
		t.Errorf("c is [%d, %d], want [%d, %d]", locStart(words[1]), locEnd(words[1]), want, want+1)
	}
	class := root.List("nodes")[2].Child("selector").List("nodes")[0].List("nodes")[0]
	if want := strings.Index(text, ".\xc3\xa9"); !class.Is("selector-class") || locStart(class) != want || locEnd(class) != want+3 {
		t.Errorf("%s is [%d, %d], want [%d, %d]", class.Type(), locStart(class), locEnd(class), want, want+3)
	}
}

func TestParseGlueUrlArgumentsAreAString(t *testing.T) {
	t.Parallel()
	root := mustParse(t, "a { b: url( x.png ) }")
	function := valueChain(firstDeclaration(t, root))[2]
	if !function.Is("value-func") || function.String("value") != "url" {
		t.Fatalf("want the url function, got %s", function.Type())
	}
	groups, isList := function.Child("group").Get("groups").([]any)
	if !isList || len(groups) != 1 || groups[0] != "x.png" {
		t.Fatalf("url groups: %#v", function.Child("group").Get("groups"))
	}
}

func TestParseGlueCustomPropertyBlock(t *testing.T) {
	t.Parallel()
	text := ":root { --x: { a: b } }"
	declaration := firstDeclaration(t, mustParse(t, text))
	value := declaration.Child("value")
	if !value.Is("css-rule") || len(value.List("nodes")) != 1 {
		t.Fatalf("value: %s with %d nodes", value.Type(), len(value.List("nodes")))
	}
	inner := value.List("nodes")[0]
	if want := strings.Index(text, "a: b"); !inner.Is("css-decl") || locStart(inner) != want || locStart(value) != want {
		t.Errorf("inner %s at %d, rule at %d, want %d", inner.Type(), locStart(inner), locStart(value), want)
	}
}

func TestParseGlueFrontMatter(t *testing.T) {
	t.Parallel()
	text := "---\ntitle: x\n---\na { b: c }"
	root := mustParse(t, text)
	frontMatter := root.Child("frontMatter")
	if !frontMatter.Is("front-matter") || frontMatter.String("raw") != "---\ntitle: x\n---" || frontMatter.String("value") != "title: x" {
		t.Fatalf("front matter: %s raw %q value %q", frontMatter.Type(), frontMatter.String("raw"), frontMatter.String("value"))
	}
	if want := strings.Index(text, "b: c"); locStart(firstDeclaration(t, root)) != want {
		t.Errorf("declaration at %d, want %d", locStart(firstDeclaration(t, root)), want)
	}
}

func TestParseGlueErrors(t *testing.T) {
	t.Parallel()
	if _, err := parse("a {"); err == nil || !printing.IsSyntax(err) {
		t.Errorf("an unclosed block: want a syntax error, got %v", err)
	}
	// Upstream throws a TypeError reading node.raws.value.raw: a crash, not a parse failure.
	if _, err := parse(":root { --x: {a} b }"); err == nil || printing.IsSyntax(err) {
		t.Errorf("a custom property upstream cannot read: want a plain error, got %v", err)
	}
	if _, err := parse("@custom-selector foo;"); err == nil || printing.IsSyntax(err) {
		t.Errorf("a custom selector without a name: want a plain error, got %v", err)
	}
}

func TestParseGlueVisitorKeys(t *testing.T) {
	t.Parallel()
	if keys := visitorKeys(estree.New("css-root", 0, 0)); !slices.Equal(keys, []string{"frontMatter", "nodes"}) {
		t.Errorf("css-root: %v", keys)
	}
	if keys := visitorKeys(estree.New("value-paren_group", 0, 0)); !slices.Equal(keys, []string{"open", "groups", "close"}) {
		t.Errorf("value-paren_group: %v", keys)
	}
	if _, known := lookupVisitorKeys(estree.New("front-matter", 0, 0)); known {
		t.Errorf("front-matter has no visitor keys upstream")
	}
}
