package tailwind

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The fixture is what the shipped Tailwind 4.3.3 parser built for every stylesheet in the corpus,
// captured by internal/lint/rules/tailwind/tools/generate_css_parser and checked in next to this test.
//
// Measured rather than transcribed, because the two are different claims and the custom-property
// branch is where they part. Reading `css-parser.ts` suggests a `--foo` declaration is normalized
// like any other; asking the engine says its value keeps the literal newlines and indentation of
// the source, because that branch slices the input directly instead of routing bytes through the
// whitespace-collapsing buffer. A hand-written expectation set would have encoded the plausible
// version and agreed with the engine on every single-line theme entry, which is most of them.

type cssParserCorpus struct {
	TailwindVersion string `json:"tailwindVersion"`
	// ParseExportName records the minified name the parser was found under, so a fixture
	// regenerated against a different Tailwind build is visible in the diff.
	ParseExportName string `json:"parseExportName"`
	// ParseResolvedByIdentity records that the generator located the parser by behavior rather than
	// by name. A false here would mean the corpus was captured from something unverified.
	ParseResolvedByIdentity bool `json:"parseResolvedByIdentity"`

	// The counts the exit criterion asks to be reported, measured by the engine over the real
	// stylesheets rather than by this port.
	StylesheetCount    int `json:"stylesheetCount"`
	ThemeEntryCount    int `json:"themeEntryCount"`
	UtilityBlockCount  int `json:"utilityBlockCount"`
	CustomVariantCount int `json:"customVariantCount"`

	Cases      []cssParserCase      `json:"cases"`
	ErrorCases []cssParserErrorCase `json:"errorCases"`
}

// cssParserCase is one stylesheet and the tree the engine built from it.
type cssParserCase struct {
	Name   string            `json:"name"`
	Source string            `json:"source"`
	Input  string            `json:"input"`
	Nodes  []*cssFixtureNode `json:"nodes"`
}

// cssParserErrorCase is a stylesheet the engine rejected. A parser that accepts these is not
// more permissive, it is wrong: any tree it invented is one the engine never produces.
//
// Message is null when the engine unexpectedly accepted the input, in which case
// UnexpectedlyParsed holds what it built. That combination is a finding about the corpus rather
// than about the port, so the test reports it rather than asserting on it.
type cssParserErrorCase struct {
	Name               string            `json:"name"`
	Source             string            `json:"source"`
	Input              string            `json:"input"`
	Message            *string           `json:"message"`
	UnexpectedlyParsed []*cssFixtureNode `json:"unexpectedlyParsed"`
}

// cssFixtureNode mirrors the engine's node shape as JSON, so converting it to a Node is an explicit
// step this test controls rather than something struct tags do invisibly.
type cssFixtureNode struct {
	Kind         string            `json:"kind"`
	Selector     string            `json:"selector"`
	Name         string            `json:"name"`
	Params       string            `json:"params"`
	Property     string            `json:"property"`
	Value        *string           `json:"value"`
	ValuePresent bool              `json:"valuePresent"`
	Important    bool              `json:"important"`
	Nodes        []*cssFixtureNode `json:"nodes"`
}

func loadCSSParserCorpus(t *testing.T) cssParserCorpus {
	t.Helper()

	path := filepath.Join("testdata", "cssparser_fixtures.json")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var corpus cssParserCorpus
	if err := json.Unmarshal(contents, &corpus); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if len(corpus.Cases) == 0 {
		t.Fatalf("%s holds no cases", path)
	}
	if !corpus.ParseResolvedByIdentity {
		t.Fatalf("%s was captured without resolving parse() by identity", path)
	}
	return corpus
}

// describeNodes renders a tree as indented text, so a mismatch reports the shape that differed
// rather than a pointer address. Values are quoted because the differences that matter here are
// whitespace, and an unquoted diff of two strings that differ by a newline is unreadable.
func describeNodes(nodes []*Node, indent string) string {
	var builder strings.Builder
	for _, node := range nodes {
		builder.WriteString(indent)
		switch node.Kind {
		case KindRule:
			builder.WriteString("rule " + quoteForDiff(node.Selector) + "\n")
		case KindAtRule:
			builder.WriteString("at-rule " + quoteForDiff(node.Name) + " params=" + quoteForDiff(node.Params) + "\n")
		case KindDeclaration:
			value := "<absent>"
			if node.ValuePresent {
				value = quoteForDiff(node.Value)
			}
			important := ""
			if node.Important {
				important = " !important"
			}
			builder.WriteString("declaration " + quoteForDiff(node.Property) + " = " + value + important + "\n")
		case KindComment:
			builder.WriteString("comment " + quoteForDiff(node.Value) + "\n")
		default:
			builder.WriteString(string(node.Kind) + "\n")
		}
		if len(node.Nodes) > 0 {
			builder.WriteString(describeNodes(node.Nodes, indent+"  "))
		}
	}
	return builder.String()
}

// quoteForDiff quotes a string so whitespace differences are visible in a failure message.
func quoteForDiff(value string) string {
	quoted := "\""
	for _, r := range value {
		switch r {
		case '\n':
			quoted += "\\n"
		case '\t':
			quoted += "\\t"
		case '\r':
			quoted += "\\r"
		case '"':
			quoted += "\\\""
		default:
			quoted += string(r)
		}
	}
	return quoted + "\""
}

// nodesFromFixture converts the engine's serialized tree into the port's Node type.
func nodesFromFixture(fixtures []*cssFixtureNode) []*Node {
	if fixtures == nil {
		return nil
	}
	nodes := make([]*Node, 0, len(fixtures))
	for _, fixture := range fixtures {
		node := &Node{
			Kind:         NodeKind(fixture.Kind),
			Selector:     fixture.Selector,
			Name:         fixture.Name,
			Params:       fixture.Params,
			Property:     fixture.Property,
			ValuePresent: fixture.ValuePresent,
			Important:    fixture.Important,
			Nodes:        nodesFromFixture(fixture.Nodes),
		}
		if fixture.Value != nil {
			node.Value = *fixture.Value
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// TestParseCSSMatchesEngine is the whole contract: for every stylesheet in the corpus, the tree
// this port builds is the tree the engine built.
//
// The count is printed on success rather than only on failure, because a suite that ran six cases
// and a suite that ran ninety are indistinguishable from a green line.
func TestParseCSSMatchesEngine(t *testing.T) {
	t.Parallel()
	corpus := loadCSSParserCorpus(t)

	comparedNodes := 0
	var countNodes func(nodes []*Node)
	countNodes = func(nodes []*Node) {
		for _, node := range nodes {
			comparedNodes++
			countNodes(node.Nodes)
		}
	}

	for _, testCase := range corpus.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			parsed, err := ParseCSS(testCase.Input)
			if err != nil {
				t.Fatalf("ParseCSS returned %v; the engine parsed this input successfully", err)
			}

			expected := nodesFromFixture(testCase.Nodes)
			got := describeNodes(parsed, "")
			want := describeNodes(expected, "")
			if got != want {
				t.Errorf("tree mismatch for %q\n--- input ---\n%s\n--- got ---\n%s\n--- want (engine) ---\n%s",
					testCase.Name, testCase.Input, got, want)
			}
			countNodes(parsed)
		})
	}

	t.Logf(
		"tailwind %s: compared %d stylesheets and %d nodes against the engine's parser (found by identity as %s)",
		corpus.TailwindVersion, len(corpus.Cases), comparedNodes, corpus.ParseExportName,
	)
}

// TestParseCSSRejectsWhatTheEngineRejects pins the error behaviour.
//
// A parser that quietly accepts malformed input is the failure mode that hides: it builds a
// plausible tree from a stylesheet the engine refuses, and every reading taken from that tree is
// invented. Only the fact of rejection is asserted, not the message text, because the messages
// carry a source location upstream that this port deliberately does not reproduce.
func TestParseCSSRejectsWhatTheEngineRejects(t *testing.T) {
	t.Parallel()
	corpus := loadCSSParserCorpus(t)

	rejected := 0
	for _, testCase := range corpus.ErrorCases {
		// Not parallel: its subtests add to totals the test checks after them, and a parallel subtest would run
		// only after the test had returned, so the check would pass on nothing.
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.Message == nil {
				t.Skipf("the engine accepted this input, so it pins nothing about rejection")
			}

			_, err := ParseCSS(testCase.Input)
			if err == nil {
				t.Fatalf("ParseCSS accepted %q; the engine rejected it with %q", testCase.Input, *testCase.Message)
			}

			var syntaxError *CSSSyntaxError
			if !errors.As(err, &syntaxError) {
				t.Fatalf("ParseCSS returned %T, want *CSSSyntaxError", err)
			}
			rejected++
		})
	}

	t.Logf("rejected %d malformed stylesheets that the engine also rejects", rejected)
}

// TestParseCSSCustomPropertyWhitespaceIsRaw states, on its own, the behaviour this component exists
// to get right.
//
// It is covered by the corpus above, but a mismatch there reads as one of ninety trees differing.
// Here it reads as what it is: a custom property's value is sliced from the source and keeps its
// interior whitespace, while an ordinary declaration's value is collapsed. A port that treats
// `--foo` as an ordinary declaration passes every other test in this file.
func TestParseCSSCustomPropertyWhitespaceIsRaw(t *testing.T) {
	t.Parallel()
	input := ":root {\n  --font-sans:\n        var(--f), -apple-system,\n        sans-serif;\n  font-family:\n        var(--f),\n        sans-serif;\n}"

	nodes, err := ParseCSS(input)
	if err != nil {
		t.Fatalf("ParseCSS: %v", err)
	}
	if len(nodes) != 1 || len(nodes[0].Nodes) != 2 {
		t.Fatalf("expected one rule holding two declarations, got:\n%s", describeNodes(nodes, ""))
	}

	custom, ordinary := nodes[0].Nodes[0], nodes[0].Nodes[1]

	// The custom property keeps the newline and the indentation that followed it.
	if !strings.Contains(custom.Value, "\n        sans-serif") {
		t.Errorf("custom property value lost its raw whitespace: %s", quoteForDiff(custom.Value))
	}

	// The ordinary declaration does not, which is what makes the difference above load-bearing
	// rather than incidental.
	if strings.Contains(ordinary.Value, "\n") {
		t.Errorf("ordinary declaration value kept a newline it should have collapsed: %s", quoteForDiff(ordinary.Value))
	}
}

// TestParseCSSReadsRealStylesheets states the exit criterion as a test: the port reproduces what
// the engine sees across a real repository's theme and its whole @import graph.
//
// The counts are the engine's, recorded at generation time, and are recomputed here from this
// port's own trees. Asserting they agree is what makes "it parsed without erroring" mean something.
func TestParseCSSReadsRealStylesheets(t *testing.T) {
	t.Parallel()
	corpus := loadCSSParserCorpus(t)

	stylesheets, themeEntries, utilityBlocks, customVariants := 0, 0, 0, 0

	for _, testCase := range corpus.Cases {
		if testCase.Source != "repository" {
			continue
		}
		stylesheets++

		nodes, err := ParseCSS(testCase.Input)
		if err != nil {
			t.Fatalf("ParseCSS(%s): %v", testCase.Name, err)
		}

		Walk(nodes, func(node *Node) WalkAction {
			if node.Kind != KindAtRule {
				return WalkContinue
			}
			switch node.Name {
			case "@theme":
				for _, child := range node.Nodes {
					if child.Kind == KindDeclaration && strings.HasPrefix(child.Property, "--") {
						themeEntries++
					}
				}
			case "@utility":
				utilityBlocks++
			case "@custom-variant":
				customVariants++
			}
			return WalkContinue
		})
	}

	if stylesheets != corpus.StylesheetCount {
		t.Errorf("read %d stylesheets, corpus holds %d", stylesheets, corpus.StylesheetCount)
	}
	if themeEntries != corpus.ThemeEntryCount {
		t.Errorf("found %d @theme entries, the engine found %d", themeEntries, corpus.ThemeEntryCount)
	}
	if utilityBlocks != corpus.UtilityBlockCount {
		t.Errorf("found %d @utility blocks, the engine found %d", utilityBlocks, corpus.UtilityBlockCount)
	}
	if customVariants != corpus.CustomVariantCount {
		t.Errorf("found %d @custom-variant rules, the engine found %d", customVariants, corpus.CustomVariantCount)
	}

	t.Logf(
		"across %d real stylesheets: %d @theme entries, %d @utility blocks, %d @custom-variant rules, each matching the engine",
		stylesheets, themeEntries, utilityBlocks, customVariants,
	)
}
