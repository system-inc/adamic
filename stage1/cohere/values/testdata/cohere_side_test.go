// This file is not built here. values_test.go lays it into cohere's own package,
// cohere/internal/format/css/values, with `go test -overlay`, so that it can call cohere's Parse and
// cohere's own test helpers (valuesFixtures and utf16IndexOf, from oracle_test.go) exactly as cohere's
// tests do, without a byte of the submodule changing.
//
// It writes every case the port is asked, as the cases file main.ts reads, and Go cohere's answer to
// each, with { Loose: true } and then { Loose: false }, in the words main.ts prints. The cases are
// cohere's: every inline fixture of its oracle test (whose trees cohere holds to postcss-values-parser
// 2.0.1 itself), and the values of its shape and refusal tests; then generated values, from a seeded
// generator over the pieces CSS values are made of and the characters the tokenizer treats specially.

package values

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
)

// adamicRequest is what values_test.go asks.
type adamicRequest struct {
	Seed      int64  `json:"seed"`
	Generated int    `json:"generated"`
	Cases     string `json:"cases"`
	Answers   string `json:"answers"`
}

// adamicShapeAndRefusalValues are the values of values_test.go's TestParseShapes and
// TestParseRefusals, which hold them in function bodies this file can't reach.
var adamicShapeAndRefusalValues = []string{
	"-5px --foo calc(1px + 2px) #fff \"a\" /* c */",
	"(a)",
	"\xc3\xa9 \xf0\x9f\x98\x80x b",
	"a // b",
	"url(a b)",
	"a(b",
	"a) b",
	"\"abc",
	"a /*",
	" ",
	"a + +",
	"calc(1 -2)",
}

// adamicPieces are what generated values are made of: the words, numbers, functions, operators and
// punctuation the tokenizer and parser decide on, and characters whose UTF-16 length differs from
// their UTF-8 length. Strings, comments and parenthesized groups come mostly whole, so most values
// parse; one raw quote, comment opener and parenthesis each keep the refusals among the cases.
var adamicPieces = []string{
	"a", "b", "red", "solid", "url", "calc", "var", "--x", "-webkit-box", "#fff", "#abcdeg", "#", "@media", "@a",
	"1", "10px", "-5", "+2", ".5", "1e3", "1e-10", "2E+2px", "0.", "5%", "u+", "U+00", "u+4??", "-0A", "e", "E",
	"(a)", "calc(1px + 2px)", "calc(1 -2)", "calc(-0.5 + 2)", "url(x.png)", "url( a b )", "url(//c)", "var(--x, 1px)", "f()", "(1 / 2)",
	"\"a b\"", "'c'", "\"e\\\"f\"", "''", "\"\u00e9\"", "/* c */", "/**/", "/* x\ny */", "// d\n",
	"(", ")", "'", "\"", "/*",
	"{", "}", "[", "]", ",", ",", ":", ";", "!", "&", "|", "~", ">", "?", ".",
	"+", "-", "*", "/", "//", "*/", "\\",
	" ", " ", " ", " ", "  ", "\t", "\n", "\r\n", "\f",
	"\u00e9", "\u4e16", "\U0001f600", "\u2028", "\u2029", "\u00a0", "\u007f", "\u001f",
}

// adamicClassEdgePieces start at every edge of every character class the tokenizer and parser test,
// and just outside it, and spell every keyword they compare in both cases (reviewer R, round six): '#'
// before a digit, a letter of either case or a character next to them (alphaNum, rNoFollow), colors of
// every length with digits and letters at both ends (isColor), unicode ranges at each edge of theirs,
// exponents at both digits, every character that ends a word and its ASCII neighbours, and calc and
// url in three cases each.
var adamicClassEdgePieces = []string{
	"#0", "#9", "#a", "#z", "#A", "#Z", "#/", "#:", "#`", "#{", "#@", "#[", "#_", "#999", "#000", "#fff", "#FFF",
	"#09af", "#09AF", "#abcdef09", "#ABCDEF90", "#12345", "#1234567", "#123459", "#g", "#G",
	"u+0", "u+9", "u+a", "u+f", "u+A", "u+F", "u+g", "u+G", "u+?", "u+-", "U+0-9", "u+99",
	"0", "9", "e0", "e9", "1e+0", "1e-9", "1E+9", "9.0e-9", "0e", "9E",
	"=", "<", "^", "$", "%", "a=b", "x<y", "@0", "@z",
	"calc", "CALC", "Calc", "url", "URL", "Url", "calc(1 -2)", "CALC(1 -2)", "Calc(-1 + 2)", "url(a b)", "URL(a b)", "Url(//x)", "URL(//x)",
	"--", "-", "---", "-- ",
}

// adamicEdgeValues don't rest on the generator's draw: a '#' before the line terminators isHex would
// stop at (the tokenizer makes each such '#' a token of its own first); DEL and U+001F, the edges of
// what the output escapes, in a word, a string and a comment; and reviewer R's round six, each value
// one that catches a plausible bug at the edge of a character class or in a keyword's case.
var adamicEdgeValues = []string{
	"#999", "#z1", "#Z1", "#_a", "a=b", "CALC(1 -2)", "URL(a b)", "Url(//x)", "u+99", "--",
	"#\u2028a #\u2029b #\u0085c",
	"a\u007fb \"\u007f\" /*\u007f*/ a\u001fb \"\u001f\" /*\u001f*/",
}

// TestAdamicPortCases writes the cases file and Go cohere's answers.
// Not parallel: writes the fixed output paths supplied by the environment request.
func TestAdamicPortCases(t *testing.T) {
	path := os.Getenv("ADAMIC_PORT_REQUEST")
	if path == "" {
		t.Skip("run by stage1/cohere/values/values_test.go in the adamic repository")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		t.Fatal(err)
	}

	var texts []string
	for _, fixture := range valuesFixtures {
		texts = append(texts, fixture.text)
	}
	texts = append(texts, adamicShapeAndRefusalValues...)
	texts = append(texts, adamicEdgeValues...)
	random := rand.New(rand.NewSource(request.Seed))
	for range request.Generated {
		var builder strings.Builder
		for range random.Intn(16) {
			if random.Intn(4) == 0 {
				builder.WriteString(adamicClassEdgePieces[random.Intn(len(adamicClassEdgePieces))])
			} else {
				builder.WriteString(adamicPieces[random.Intn(len(adamicPieces))])
			}
		}
		texts = append(texts, builder.String())
	}

	escape := strings.NewReplacer("\\", "\\\\", "\t", "\\t", "\n", "\\n", "\r", "\\r")
	var cases, answers strings.Builder
	for index, text := range texts {
		if !utf8.ValidString(text) {
			t.Fatalf("%q is not UTF-8, which the port reads its cases as", text)
		}
		cases.WriteString(escape.Replace(text))
		cases.WriteByte('\n')
		toUTF16 := utf16IndexOf(text)
		for _, loose := range []bool{true, false} {
			mode := "strict"
			if loose {
				mode = "loose"
			}
			fmt.Fprintf(&answers, "case %d %s\n", index, mode)
			tree, err := Parse(text, Options{Loose: loose})
			if err != nil {
				var thrown *Error
				if !asError(err, &thrown) {
					// A panic that isn't one of upstream's throws: a fault in the Go, which the port
					// can't print, so it shows as a difference.
					fmt.Fprintf(&answers, "fault %v\n", err)
					continue
				}
				fmt.Fprintf(&answers, "error %s\n", thrown.Error())
				continue
			}
			adamicRender(tree, 0, toUTF16, &answers)
		}
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}

// adamicRender writes a node and its children at depth, as main.ts does: its type, then every key it
// has but nodes, sorted, with sourceIndex as the library's UTF-16 index.
func adamicRender(node *estree.Node, depth int, toUTF16 func(int) int, builder *strings.Builder) {
	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(node.Type())
	keys := slices.Sorted(slices.Values(node.Keys()))
	children, isContainer := node.Get("nodes").([]*estree.Node)
	for _, key := range keys {
		if key == "nodes" {
			continue
		}
		value := node.Get(key)
		if key == "sourceIndex" {
			if index, isInt := value.(int); isInt {
				value = toUTF16(index)
			}
		}
		fmt.Fprintf(builder, " %s=%s", key, adamicValue(value))
	}
	if isContainer {
		fmt.Fprintf(builder, " nodes=%d", len(children))
	}
	builder.WriteByte('\n')
	for _, child := range children {
		adamicRender(child, depth+1, toUTF16, builder)
	}
}

// adamicValue is a field's value as main.ts writes it.
func adamicValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return "<undefined>"
	case string:
		return adamicQuote(typed)
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case float64:
		if math.IsNaN(typed) {
			return "NaN"
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case map[string]any:
		keys := slices.Sorted(func(yield func(string) bool) {
			for key := range typed {
				if !yield(key) {
					return
				}
			}
		})
		var builder strings.Builder
		builder.WriteByte('{')
		for index, key := range keys {
			if index > 0 {
				builder.WriteByte(',')
			}
			fmt.Fprintf(&builder, "%s:%s", key, adamicValue(typed[key]))
		}
		builder.WriteByte('}')
		return builder.String()
	default:
		return fmt.Sprintf("<a %T>", typed)
	}
}

// adamicQuote is main.ts's quote.
func adamicQuote(text string) string {
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
