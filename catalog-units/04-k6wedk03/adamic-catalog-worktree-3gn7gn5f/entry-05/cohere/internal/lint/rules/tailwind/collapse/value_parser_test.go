package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixture is the tree the shipped Tailwind 4.3.3 value parser produced for every shape in the
// corpus, captured by internal/lint/rules/tailwind/tools/generate_value_parser and checked in next to this test.
//
// Measured rather than transcribed. Several of this parser's behaviors fall out of JavaScript
// semantics rather than intent, and a careful reader writing expectations by hand would get them
// wrong in the direction of a well-formed tree: `a)b` parses to a single word `b` with the `a`
// silently discarded, a trailing `\` produces the literal word `\undefined`, and `foo(bar` puts
// `bar` beside the function rather than inside it. Each is asserted directly below as well as
// covered by the corpus, because a behavior that only exists inside a bulk comparison is easy to
// regress and hard to read.

type valueParserCorpus struct {
	TailwindVersion string `json:"tailwindVersion"`
	// Bundle and ParseSymbol record where the parser was extracted from, so a fixture regenerated
	// against a different Tailwind build, or against a differently-mangled symbol, is visible in
	// the diff rather than silently replacing the measurements.
	Bundle      string            `json:"bundle"`
	ParseSymbol string            `json:"parseSymbol"`
	ToCssSymbol string            `json:"toCssSymbol"`
	Cases       []valueParserCase `json:"cases"`
	// ExhaustiveAlphabet and ExhaustiveCases hold every string of length 1 to 3 over an alphabet
	// chosen to stress the interacting branches. They answer what the hand-written corpus cannot:
	// questions about shapes nobody thought to name.
	ExhaustiveAlphabet []string          `json:"exhaustiveAlphabet"`
	ExhaustiveCases    []valueParserCase `json:"exhaustiveCases"`
}

// valueParserCase is one value, the tree the engine built for it, and what that tree prints as.
type valueParserCase struct {
	Value string      `json:"value"`
	Ast   []ValueNode `json:"ast"`
	// Css is the engine's own reprint of the tree. Structure and printed form are different
	// claims: a tree that merged two words and a tree that dropped a separator can both print
	// correctly, and a structurally correct tree can still print wrong if whitespace moved
	// between nodes. Both are compared.
	Css string `json:"css"`
}

func loadValueParserCorpus(t *testing.T) valueParserCorpus {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "valueparser_fixtures.json"))
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var corpus valueParserCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	// A corpus that quietly shrank is the failure this approach exists to prevent, and it would
	// otherwise read as a passing run.
	if len(corpus.Cases) < 250 {
		t.Fatalf("fixture holds only %d cases; expected several hundred shapes", len(corpus.Cases))
	}
	if corpus.ParseSymbol == "" {
		t.Fatalf("fixture does not record which bundle symbol it was generated from")
	}
	if len(corpus.ExhaustiveCases) < 2000 {
		t.Fatalf("fixture holds only %d exhaustive cases; the enumeration is meant to be thousands", len(corpus.ExhaustiveCases))
	}
	return corpus
}

// TestParseValueMatchesEngine is the whole point of this suite: every tree, for every shape,
// compared node for node against what the shipped engine built.
func TestParseValueMatchesEngine(t *testing.T) {
	t.Parallel()
	corpus := loadValueParserCorpus(t)

	comparedValues := 0
	comparedNodes := 0
	maximumDepth := 0

	for _, testCase := range corpus.Cases {
		got := ParseValue(testCase.Value)
		if difference := diffValueNodes(got, testCase.Ast, ""); difference != "" {
			t.Errorf("ParseValue(%q):\n  %s\n  got:    %s\n  engine: %s",
				testCase.Value, difference, renderValueNodes(got), renderValueNodes(testCase.Ast))
		}
		comparedValues++
		comparedNodes += countValueNodes(testCase.Ast)
		if depth := valueNodeDepth(testCase.Ast); depth > maximumDepth {
			maximumDepth = depth
		}

		// The engine's own reprint of its own tree, compared against this port's reprint of this
		// port's tree. Two ports agreeing on structure but disagreeing on printing would pass the
		// check above and fail here.
		if printed := ValueToCss(got); printed != testCase.Css {
			t.Errorf("ValueToCss(ParseValue(%q)) = %q, engine prints %q", testCase.Value, printed, testCase.Css)
		}
	}

	t.Logf("compared %d parse trees (%d nodes, max depth %d) and %d reprints against Tailwind %s (%s, symbols %s/%s)",
		comparedValues, comparedNodes, maximumDepth, comparedValues,
		corpus.TailwindVersion, corpus.Bundle, corpus.ParseSymbol, corpus.ToCssSymbol)
}

// TestParseValueMatchesEngineExhaustively parses every string of length 1 to 3 over an alphabet of
// the characters whose branches interact: the escape, both quote kinds, both parens, separators,
// the slash, and multi-byte characters of two, three and four UTF-8 bytes.
//
// The hand-written corpus covers shapes someone thought of, and it is where the interesting values
// live. This covers the ones nobody thought of, and it is what settles questions the corpus cannot
// see. Whether the escape branch consumes one byte or one whole character is invisible across every
// value in the corpus, and both rules are correct only because a UTF-8 continuation byte can match
// no separator, quote, or paren and so reaches the buffer either way. That is an argument; this is
// the measurement behind it.
func TestParseValueMatchesEngineExhaustively(t *testing.T) {
	t.Parallel()
	corpus := loadValueParserCorpus(t)

	disagreements := 0
	for _, testCase := range corpus.ExhaustiveCases {
		got := ParseValue(testCase.Value)
		if difference := diffValueNodes(got, testCase.Ast, ""); difference != "" {
			// Bounded output: a broken branch disagrees on hundreds of these at once, and the
			// first handful name the shape as well as all of them would.
			if disagreements < 10 {
				t.Errorf("ParseValue(%q):\n  %s\n  got:    %s\n  engine: %s",
					testCase.Value, difference, renderValueNodes(got), renderValueNodes(testCase.Ast))
			}
			disagreements++
		}
		if printed := ValueToCss(got); printed != testCase.Css {
			if disagreements < 10 {
				t.Errorf("ValueToCss(ParseValue(%q)) = %q, engine prints %q", testCase.Value, printed, testCase.Css)
			}
			disagreements++
		}
	}
	if disagreements > 10 {
		t.Errorf("%d disagreements in total; only the first were printed", disagreements)
	}

	t.Logf("compared %d exhaustive inputs over the alphabet %q against Tailwind %s",
		len(corpus.ExhaustiveCases), corpus.ExhaustiveAlphabet, corpus.TailwindVersion)
}

// TestCorpusExercisesEveryBranch guards the corpus itself.
//
// The bulk comparison above is only as good as what it covers, and coverage is the property most
// likely to erode silently: a corpus can lose every unbalanced case and still report several
// hundred green comparisons. These are the branches whose absence would not otherwise be visible.
func TestCorpusExercisesEveryBranch(t *testing.T) {
	t.Parallel()
	corpus := loadValueParserCorpus(t)

	kinds := map[ValueNodeKind]int{}
	for _, testCase := range corpus.Cases {
		walkValueNodes(testCase.Ast, func(node ValueNode) { kinds[node.Kind]++ })
	}
	for _, kind := range []ValueNodeKind{ValueNodeKindWord, ValueNodeKindFunction, ValueNodeKindSeparator} {
		if kinds[kind] < 10 {
			t.Errorf("corpus produced only %d %s nodes; it is not exercising that branch", kinds[kind], kind)
		}
	}

	// Values whose reprint differs from the input are the malformed ones, and they are the whole
	// reason this parser needed measuring rather than reading. A corpus holding none of them would
	// pass every comparison while testing nothing interesting.
	lossy := 0
	for _, testCase := range corpus.Cases {
		if testCase.Css != testCase.Value {
			lossy++
		}
	}
	if lossy < 10 {
		t.Errorf("corpus holds only %d values that do not round-trip; the malformed shapes are the point", lossy)
	}

	if valueNodeDepthAcross(corpus.Cases) < 4 {
		t.Error("corpus never nests functions more than three deep; nested arguments are untested")
	}
}

// TestUnmatchedCloseParenDiscardsPendingText states the first behavior that a transcription would
// get wrong.
//
// Upstream flushes the pending buffer into `tail?.nodes`, and `tail` is undefined when the stack is
// empty, so the optional chain drops the word instead of emitting it. This is not a rounding error:
// `foo\(bar)` parses to nothing at all, because the escaped paren never opened a function and the
// close paren then discarded the entire buffer.
func TestUnmatchedCloseParenDiscardsPendingText(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  []ValueNode
	}{
		{"a)b", []ValueNode{{Kind: ValueNodeKindWord, Value: "b"}}},
		{")", []ValueNode{}},
		{"a)", []ValueNode{}},
		{`foo\(bar)`, []ValueNode{}},
		{"a)b,c", []ValueNode{
			{Kind: ValueNodeKindWord, Value: "b"},
			{Kind: ValueNodeKindSeparator, Value: ","},
			{Kind: ValueNodeKindWord, Value: "c"},
		}},
	}

	for _, testCase := range cases {
		got := ParseValue(testCase.input)
		if difference := diffValueNodes(got, testCase.want, ""); difference != "" {
			t.Errorf("ParseValue(%q) = %s, want %s (%s)",
				testCase.input, renderValueNodes(got), renderValueNodes(testCase.want), difference)
		}
	}
}

// TestTrailingBackslashReadsPastTheEnd states the second.
//
// `buffer += input[i] + input[i + 1]` on the last character indexes one past the end. JavaScript
// yields `undefined` there, and string concatenation stringifies it, so the word really does carry
// the ten characters `\undefined`.
func TestTrailingBackslashReadsPastTheEnd(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct{ input, want string }{
		{`\`, `\undefined`},
		{`a\`, `a\undefined`},
		{`foo(a\`, `a\undefined`},
	} {
		got := ParseValue(testCase.input)
		last := got[len(got)-1]
		if last.Kind != ValueNodeKindWord || last.Value != testCase.want {
			t.Errorf("ParseValue(%q) last node = %+v, want word %q", testCase.input, last, testCase.want)
		}
	}
}

// TestUnclosedFunctionKeepsRemainderAtTopLevel states the third.
//
// The final buffer flush pushes to `ast`, never to the still-open parent, so text that visually sits
// inside an unclosed function ends up beside it. A port that flushed into the open function would
// produce the tree a reader expects and disagree with the engine.
func TestUnclosedFunctionKeepsRemainderAtTopLevel(t *testing.T) {
	t.Parallel()
	got := ParseValue("foo(bar")
	want := []ValueNode{
		{Kind: ValueNodeKindFunction, Value: "foo", Nodes: []ValueNode{}},
		{Kind: ValueNodeKindWord, Value: "bar"},
	}
	if difference := diffValueNodes(got, want, ""); difference != "" {
		t.Errorf("ParseValue(\"foo(bar\") = %s, want %s (%s)", renderValueNodes(got), renderValueNodes(want), difference)
	}

	// Anything already separated lands inside, so only the trailing buffer escapes: the separator
	// after `a` was pushed while the function was open.
	got = ParseValue("foo(a b")
	if len(got) != 2 || got[0].Kind != ValueNodeKindFunction || len(got[0].Nodes) != 2 {
		t.Fatalf("ParseValue(\"foo(a b\") = %s", renderValueNodes(got))
	}
	if got[1].Kind != ValueNodeKindWord || got[1].Value != "b" {
		t.Errorf("trailing word = %+v, want word \"b\" at top level", got[1])
	}
}

// TestSlashIsItsOwnWord covers the branch that exists only for `theme(colors.red.500/10)`, where a
// slash carries meaning without spaces around it. It is a word, never a separator, so a caller
// filtering separators still sees it.
func TestSlashIsItsOwnWord(t *testing.T) {
	t.Parallel()
	got := ParseValue("theme(colors.red.500/10)")
	if len(got) != 1 || got[0].Kind != ValueNodeKindFunction {
		t.Fatalf("got %s", renderValueNodes(got))
	}
	want := []ValueNode{
		{Kind: ValueNodeKindWord, Value: "colors.red.500"},
		{Kind: ValueNodeKindWord, Value: "/"},
		{Kind: ValueNodeKindWord, Value: "10"},
	}
	if difference := diffValueNodes(got[0].Nodes, want, ""); difference != "" {
		t.Errorf("theme() arguments = %s, want %s (%s)", renderValueNodes(got[0].Nodes), renderValueNodes(want), difference)
	}
}

// TestSeparatorRunsArePreservedVerbatim covers whitespace fidelity, which decides whether values
// round-trip. A parser that normalised runs would agree on structure and disagree on every reprint.
func TestSeparatorRunsArePreservedVerbatim(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"a  b", "a, ,b", "foo(bar,  baz)", "a , b", "a\t\tb", "a,,b"} {
		if printed := ValueToCss(ParseValue(input)); printed != input {
			t.Errorf("round-trip of %q gave %q", input, printed)
		}
	}

	got := ParseValue("a, ,b")
	if len(got) != 3 || got[1].Kind != ValueNodeKindSeparator || got[1].Value != ", ," {
		t.Errorf("ParseValue(\"a, ,b\") = %s, want the whole run as one separator", renderValueNodes(got))
	}
}

// TestValueArgumentListsParseAsNodes covers the shape this port exists to serve: the `--value(...)`
// argument lists the @utility evaluator reads. The arguments are nodes, and the separator between
// them is a node too, so a caller walking the list sees the same tree the engine does.
func TestValueArgumentListsParseAsNodes(t *testing.T) {
	t.Parallel()
	got := ParseValue("--value(--percentage-*, [*])")
	if len(got) != 1 || got[0].Kind != ValueNodeKindFunction || got[0].Value != "--value" {
		t.Fatalf("got %s", renderValueNodes(got))
	}
	want := []ValueNode{
		{Kind: ValueNodeKindWord, Value: "--percentage-*"},
		{Kind: ValueNodeKindSeparator, Value: ", "},
		{Kind: ValueNodeKindWord, Value: "[*]"},
	}
	if difference := diffValueNodes(got[0].Nodes, want, ""); difference != "" {
		t.Errorf("--value() arguments = %s (%s)", renderValueNodes(got[0].Nodes), difference)
	}
}

// TestNestedFunctionsNestStructurally covers the case the top-level splitter in segment.go cannot
// answer, and the reason this parser is not redundant with it.
func TestNestedFunctionsNestStructurally(t *testing.T) {
	t.Parallel()
	got := ParseValue("calc(var(--a) * 2)")
	if len(got) != 1 || got[0].Value != "calc" {
		t.Fatalf("got %s", renderValueNodes(got))
	}
	inner := got[0].Nodes
	if len(inner) != 5 || inner[0].Kind != ValueNodeKindFunction || inner[0].Value != "var" {
		t.Fatalf("calc arguments = %s", renderValueNodes(inner))
	}
	if len(inner[0].Nodes) != 1 || inner[0].Nodes[0].Value != "--a" {
		t.Errorf("var arguments = %s", renderValueNodes(inner[0].Nodes))
	}

	// A bare `(` is a function with an empty name, which is how upstream represents grouping.
	grouped := ParseValue("((a))")
	if len(grouped) != 1 || grouped[0].Kind != ValueNodeKindFunction || grouped[0].Value != "" {
		t.Fatalf("ParseValue(\"((a))\") = %s", renderValueNodes(grouped))
	}
	if len(grouped[0].Nodes) != 1 || grouped[0].Nodes[0].Kind != ValueNodeKindFunction {
		t.Errorf("nested group = %s", renderValueNodes(grouped[0].Nodes))
	}
}

// TestQuotedStringsHoldTheirContents covers the branch that keeps a quoted font name intact, and
// the boundary where a quote stops protecting anything.
func TestQuotedStringsHoldTheirContents(t *testing.T) {
	t.Parallel()
	got := ParseValue(`"Comic Sans, Bold", sans-serif`)
	if len(got) != 3 || got[0].Value != `"Comic Sans, Bold"` {
		t.Errorf("quoted family = %s", renderValueNodes(got))
	}

	// An unterminated quote consumes only itself: parsing resumes at the next character rather
	// than swallowing the tail. `"unterminated` is one word because it holds no separator, and
	// that single shape is not enough to pin the behavior. `"a b` is the case that discriminates,
	// and its absence from the corpus hid a real bug in this port: an implementation that jumped
	// to the end of input agreed with the engine on the first and disagreed on the second.
	for _, testCase := range []struct {
		input string
		want  []ValueNode
	}{
		{`"unterminated`, []ValueNode{{Kind: ValueNodeKindWord, Value: `"unterminated`}}},
		{`"a b`, []ValueNode{
			{Kind: ValueNodeKindWord, Value: `"a`},
			{Kind: ValueNodeKindSeparator, Value: " "},
			{Kind: ValueNodeKindWord, Value: "b"},
		}},
		{`"a,b`, []ValueNode{
			{Kind: ValueNodeKindWord, Value: `"a`},
			{Kind: ValueNodeKindSeparator, Value: ","},
			{Kind: ValueNodeKindWord, Value: "b"},
		}},
		// The quote does not suppress the paren either: an unterminated string can still open a
		// function, whose name carries the quote.
		{`"a(b`, []ValueNode{
			{Kind: ValueNodeKindFunction, Value: `"a`, Nodes: []ValueNode{}},
			{Kind: ValueNodeKindWord, Value: "b"},
		}},
	} {
		got := ParseValue(testCase.input)
		if difference := diffValueNodes(got, testCase.want, ""); difference != "" {
			t.Errorf("ParseValue(%q) = %s, want %s (%s)",
				testCase.input, renderValueNodes(got), renderValueNodes(testCase.want), difference)
		}
	}

	// A quote of the other kind does not close the string. The inner quote is followed by a
	// separator in each case below, so a parser closing on either kind would reparse the tail and
	// split one word into three. Without that trailing separator the two behaviors produce the
	// same text and the case proves nothing, which is how this gap was found: a mutation closing
	// on either quote survived a corpus holding only `"a'b"`.
	for _, testCase := range []struct{ input, want string }{
		{`"a'b c"`, `"a'b c"`},
		{`"a'b,c"`, `"a'b,c"`},
		{`'a"b c'`, `'a"b c'`},
		{`"it's a test"`, `"it's a test"`},
		{`'don"t, stop'`, `'don"t, stop'`},
	} {
		got := ParseValue(testCase.input)
		if len(got) != 1 || got[0].Kind != ValueNodeKindWord || got[0].Value != testCase.want {
			t.Errorf("ParseValue(%q) = %s, want one word %q", testCase.input, renderValueNodes(got), testCase.want)
		}
	}
}

// TestCarriageReturnNewlineNormalises covers the first statement of the parser. A `\r\n` becomes a
// one-character separator; a lone `\r` is not a separator at all and stays inside the word.
func TestCarriageReturnNewlineNormalises(t *testing.T) {
	t.Parallel()
	got := ParseValue("a\r\nb")
	if len(got) != 3 || got[1].Kind != ValueNodeKindSeparator || got[1].Value != "\n" {
		t.Errorf("ParseValue(\"a\\r\\nb\") = %s", renderValueNodes(got))
	}

	got = ParseValue("a\rb")
	if len(got) != 1 || got[0].Kind != ValueNodeKindWord || got[0].Value != "a\rb" {
		t.Errorf("ParseValue(\"a\\rb\") = %s, want a lone CR to stay inside the word", renderValueNodes(got))
	}
}

// TestParseValueAcceptsEveryInput covers totality. Real class strings contain arbitrary text, and
// the parser must never panic on any of it.
func TestParseValueAcceptsEveryInput(t *testing.T) {
	t.Parallel()
	corpus := loadValueParserCorpus(t)
	for _, testCase := range corpus.Cases {
		// Every prefix and suffix of every corpus value, which is a cheap way to reach truncations
		// the corpus does not name: half a string literal, half an escape, a stack left open.
		for cut := 0; cut <= len(testCase.Value); cut++ {
			ValueToCss(ParseValue(testCase.Value[:cut]))
			ValueToCss(ParseValue(testCase.Value[cut:]))
		}
	}
}

// countValueNodes counts every node in the tree, including nested ones.
func countValueNodes(nodes []ValueNode) int {
	total := 0
	walkValueNodes(nodes, func(ValueNode) { total++ })
	return total
}

func walkValueNodes(nodes []ValueNode, visit func(ValueNode)) {
	for _, node := range nodes {
		visit(node)
		if node.Kind == ValueNodeKindFunction {
			walkValueNodes(node.Nodes, visit)
		}
	}
}

func valueNodeDepth(nodes []ValueNode) int {
	deepest := 0
	for _, node := range nodes {
		if node.Kind != ValueNodeKindFunction {
			continue
		}
		if depth := 1 + valueNodeDepth(node.Nodes); depth > deepest {
			deepest = depth
		}
	}
	return deepest
}

func valueNodeDepthAcross(cases []valueParserCase) int {
	deepest := 0
	for _, testCase := range cases {
		if depth := valueNodeDepth(testCase.Ast); depth > deepest {
			deepest = depth
		}
	}
	return deepest
}

// diffValueNodes returns the first structural disagreement between two trees, described by path, or
// the empty string when they match. A path rather than a boolean, because a mismatch four levels
// into a nested value is otherwise unreadable in the failure output.
func diffValueNodes(got, want []ValueNode, path string) string {
	if len(got) != len(want) {
		return fmt.Sprintf("at %s: %d nodes, engine has %d", pathOrRoot(path), len(got), len(want))
	}
	for index := range want {
		here := fmt.Sprintf("%s[%d]", path, index)
		if got[index].Kind != want[index].Kind {
			return fmt.Sprintf("at %s: kind %q, engine says %q", here, got[index].Kind, want[index].Kind)
		}
		if got[index].Value != want[index].Value {
			return fmt.Sprintf("at %s: value %q, engine says %q", here, got[index].Value, want[index].Value)
		}
		if want[index].Kind == ValueNodeKindFunction {
			if difference := diffValueNodes(got[index].Nodes, want[index].Nodes, here); difference != "" {
				return difference
			}
		}
	}
	return ""
}

func pathOrRoot(path string) string {
	if path == "" {
		return "root"
	}
	return path
}

// renderValueNodes prints a tree compactly enough to read in a failure line.
func renderValueNodes(nodes []ValueNode) string {
	var builder strings.Builder
	builder.WriteByte('[')
	for index, node := range nodes {
		if index > 0 {
			builder.WriteString(" ")
		}
		switch node.Kind {
		case ValueNodeKindWord:
			fmt.Fprintf(&builder, "word(%q)", node.Value)
		case ValueNodeKindSeparator:
			fmt.Fprintf(&builder, "sep(%q)", node.Value)
		case ValueNodeKindFunction:
			fmt.Fprintf(&builder, "fn(%q)%s", node.Value, renderValueNodes(node.Nodes))
		}
	}
	builder.WriteByte(']')
	return builder.String()
}
