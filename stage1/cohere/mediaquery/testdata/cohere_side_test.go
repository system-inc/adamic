// This file is not built here. mediaquery_test.go lays it into cohere's own package,
// cohere/internal/format/css/mediaquery, with `go test -overlay`, so that it can call cohere's Parse and
// cohere's own test helpers (mediaQueryFixtures and utf16OffsetsOf, from oracle_test.go) exactly as
// cohere's tests do, without a byte of the submodule changing.
//
// It writes every case the port is asked, as the cases file main.ts reads, and Go cohere's answer to
// each, in the words main.ts prints. The cases are cohere's: every inline fixture of its oracle test
// (whose trees cohere holds to postcss-media-query-parser 0.2.3 itself), the params of its golden trees
// and of its refusals; then generated params, from a seeded generator over the pieces media queries
// are made of and the characters the parser treats specially.

package mediaquery

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
)

// adamicRequest is what mediaquery_test.go asks.
type adamicRequest struct {
	Seed      int64  `json:"seed"`
	Generated int    `json:"generated"`
	Cases     string `json:"cases"`
	Answers   string `json:"answers"`
}

// adamicGoldenAndRefusalParams are the params of parse_test.go's TestParseTrees and TestParseRefusals,
// which hold them in function bodies this file can't reach.
var adamicGoldenAndRefusalParams = []string{
	"screen and (min-width: 768px)",
	" url(a) print , (b)",
	"a b c d",
	"\xc3\xa9 and (x)",
	"(a} b",
	"screen and (color}",
	"url(",
	"  url (a(b)",
	// Reviewer R, round six: a string closed only by the quote that opened it.
	"({'a\"b'}:x)",
}

// adamicWhitespace is JavaScript's whitespace, all 25 characters: what \s matches and trim strips
// (ECMAScript WhiteSpace and LineTerminator), and what the port's isWhitespace and stage 0's trim must
// each agree on. Reviewer R2 found the pieces held 15 of them, so a mutant that forgot U+2029, or
// U+2000, survived.
var adamicWhitespace = []string{
	"\t", "\n", "\v", "\f", "\r", " ", "\u00a0", "\u1680",
	"\u2000", "\u2001", "\u2002", "\u2003", "\u2004", "\u2005", "\u2006", "\u2007", "\u2008", "\u2009", "\u200a",
	"\u2028", "\u2029", "\u202f", "\u205f", "\u3000", "\ufeff",
}

// adamicNotWhitespace are characters near JavaScript's whitespace that aren't in it: Go's
// unicode.IsSpace takes U+0085, U+200B is a zero-width space, and U+180E was whitespace in Unicode
// before 6.3. DEL and U+001F are the edges of what the output escapes.
var adamicNotWhitespace = []string{"\u0085", "\u200b", "\u180e", "\u007f", "\u001f"}

// adamicPieces are what generated params are made of: the words, punctuation and whitespace the parser
// decides on, and characters whose UTF-16 length differs from their UTF-8 length. Every character of
// adamicWhitespace and adamicNotWhitespace is among them too.
var adamicPieces = append(append([]string{
	"screen", "print", "all", "and", "not", "only", "AND", "url", "URL", "url(", "url (", "x.css",
	"(", "(", ")", ")", "{", "}", "#{", "$query", ",", ",", ":", ": ", "'", "\"", "\\", ";",
	" ", " ", " ", "  ", "\r\n",
	"min-width", "max-width", "color", "100px", "1em", "2", "-webkit-min-device-pixel-ratio",
	"--custom", "calc(1px + 2em)", "var(--bp)", "<=", ">", "=", "/*", "*/", "a", "b",
	"\u00e9", "\u4e16", "\U0001f600", "\U0001d11e",
	// Every keyword the parser compares, in three cases (reviewer R, round six), and strings holding the
	// other quote, which only a string closed by its own quote keeps.
	// A string holding the other quote reaches the feature parser only inside parentheses, before a
	// colon, so the pieces carry that shape too.
	"not", "NOT", "Not", "only", "ONLY", "Only", "And", "url(", "URL(", "Url(", "{'a\"b'}", "{\"a'b\"}",
	"({'a\"b'}:", "({\"a'b\"}:", "{'\"'}:", "{\"'\"}:",
}, adamicWhitespace...), adamicNotWhitespace...)

// TestAdamicPortCases writes the cases file and Go cohere's answers.
// Not parallel: writes the fixed output paths supplied by the environment request.
func TestAdamicPortCases(t *testing.T) {
	path := os.Getenv("ADAMIC_PORT_REQUEST")
	if path == "" {
		t.Skip("run by stage1/cohere/mediaquery/mediaquery_test.go in the adamic repository")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		t.Fatal(err)
	}

	var params []string
	for _, fixture := range mediaQueryFixtures {
		params = append(params, fixture.params)
	}
	params = append(params, adamicGoldenAndRefusalParams...)
	// One case for each character of adamicWhitespace and adamicNotWhitespace, wherever the parser
	// trims, splits or tests for whitespace: around a query, between a type and a keyword, inside a
	// feature, around its colon and its value, and around a comma. These don't rest on the generator's
	// draw.
	for _, character := range append(append([]string{}, adamicWhitespace...), adamicNotWhitespace...) {
		params = append(params, strings.ReplaceAll("_screen_and_(_min-width_:_1px_)_,_print_", "_", character))
	}
	random := rand.New(rand.NewSource(request.Seed))
	for range request.Generated {
		var builder strings.Builder
		for range random.Intn(14) {
			builder.WriteString(adamicPieces[random.Intn(len(adamicPieces))])
		}
		params = append(params, builder.String())
	}

	escape := strings.NewReplacer("\\", "\\\\", "\t", "\\t", "\n", "\\n", "\r", "\\r")
	var cases, answers strings.Builder
	for index, each := range params {
		if !utf8.ValidString(each) {
			t.Fatalf("%q is not UTF-8, which the port reads its cases as", each)
		}
		cases.WriteString(escape.Replace(each))
		cases.WriteByte('\n')
		fmt.Fprintf(&answers, "case %d\n", index)
		tree, err := Parse(each)
		if err != nil {
			fmt.Fprintf(&answers, "error %s\n", err.Error())
			continue
		}
		adamicRender(t, tree, 0, utf16OffsetsOf(each), &answers)
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// adamicRender writes a node and its children at depth, as main.ts does, with sourceIndex as the
// library's UTF-16 index. A node whose keys aren't the ones the port's MediaNode has fails the test,
// since the port would have no way to show the difference.
func adamicRender(t *testing.T, node *estree.Node, depth int, offsets []int, builder *strings.Builder) {
	t.Helper()
	keys := slices.Sorted(slices.Values(node.Keys()))
	_, isContainer := node.Get("nodes").([]*estree.Node)
	want := []string{"after", "before", "sourceIndex", "value"}
	if isContainer {
		want = []string{"after", "before", "nodes", "sourceIndex", "value"}
	}
	if !slices.Equal(keys, want) {
		t.Fatalf("a %q node has the keys %v, where the port's MediaNode has %v", node.Type(), keys, want)
	}
	nodeType := node.Type()
	if nodeType == "" {
		nodeType = "<undefined>"
	}
	sourceIndex, isInt := node.Get("sourceIndex").(int)
	if !isInt || sourceIndex < 0 || sourceIndex >= len(offsets) {
		t.Fatalf("a %q node's sourceIndex is %v", node.Type(), node.Get("sourceIndex"))
	}
	fmt.Fprintf(builder, "%s%s %s @%d before=%s after=%s", strings.Repeat("  ", depth), nodeType,
		adamicQuote(node.Get("value")), offsets[sourceIndex], adamicQuote(node.Get("before")), adamicQuote(node.Get("after")))
	children := node.List("nodes")
	if isContainer {
		fmt.Fprintf(builder, " nodes=%d", len(children))
	}
	builder.WriteByte('\n')
	for _, child := range children {
		adamicRender(t, child, depth+1, offsets, builder)
	}
}

// adamicQuote is main.ts's quote for a string, and <undefined> for anything else, which the port can't
// print, so that a property the port takes for a string and the Go leaves out shows as a difference.
func adamicQuote(value any) string {
	text, isString := value.(string)
	if !isString {
		return "<undefined>"
	}
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range text {
		switch {
		case character == '\\':
			builder.WriteString("\\\\")
		case character == '"':
			builder.WriteString("\\\"")
		case character == '\n':
			builder.WriteString("\\n")
		case character == '\r':
			builder.WriteString("\\r")
		case character == '\t':
			builder.WriteString("\\t")
		case character < 0x20 || character == 0x7f:
			fmt.Fprintf(&builder, "\\u%04x", character)
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
