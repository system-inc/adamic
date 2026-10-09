package unist

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatfiles"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/format/markdown/mdast"
	"github.com/system-inc/cohere/internal/format/prettier"
)

// fixtures reach every transform and every branch of attach and updatePositions: each node type with
// and without tag, anchor and comments around it; comments in every slot; documents with and without
// markers, directives and trailing comments; the !!pairs and !!omap shapes upstream fails on; and the
// edges: end of input in each place, \r\n, tabs, non-ASCII and astral characters.
var fixtures = []string{
	// Empty and trivial streams.
	"", "\n", "\n\n\n", " ", "  \n  ", "\t", "a", "a\n", "  a  \n", "# c", "# c\n", "#", "#\n#", "---", "---\n",
	"...", "...\n", "\ufeff", "\ufeffa", "\ufeff# c\n", "\r\n", "a\r\n", "a\r", "\r",

	// Plain scalars.
	"a b  c", "a:    b", "a: b   c  ", "a: b\tc", "a:\n  b\n  c", "a: b\n  c\n\n  d", "a\n b\n  c", "a: é\nb: 🍐 x\nc: 中文",
	"a: \u00a0b", "key: value # c", "1", "a: 1\nb: 2.5\nc: true\nd: null\ne: ~", "a: -1\nb: .inf", "a:\u3000b",

	// Quoted scalars.
	`"a"`, `'a'`, `a: "b"`, `a: 'b'`, `"a'b"`, `'a"b'`, `"a\"b"`, `'a''b'`, `"a\nb"`, `"\ud800"`, `"\udc00x\ud83d"`,
	`"\ud83d\ude00"`, `"\x41\u00e9\U0001F600"`, "\"a\n  b\"", "'a\n\n  b'", "\"a \\\n  b\"", `""`, `''`, `"a" # c`,
	"'a' # c\n", `"🍐"`, "a: \"\u2028\"",

	// Block scalars.
	"a: |\n  x\n  y\n", "a: >\n  x\n  y\n", "a: |-\n  x\n", "a: |+\n  x\n", "a: >-\n  x\n", "a: >+\n  x\n\n",
	"a: |2\n   x\n  y\n", "a: |1\n  x\n", "a: >2-\n   x\n", "a: |-2\n   x\n", "a: |+\n  x\n\nb: c", "a: |\n  x\n\n\nb: c",
	"a: | # c\n  x\n", "a: >- # c\n  x\n  y", "|\n  x\n", ">\n  x\n\n  y\n", "- |\n  x\n- y", "a: |\nb: c", "a: |\n",
	"a: |\n  x\n# c\n", "a: |\n  x\n\n# c\n", "a: |+\n  x\n# c\n", "--- |\n  x\n", "--- >-\n  x\n...\n", "|", ">",
	"a: !t |\n  x\n", "a: &x |\n  x\n", "- !t &a | # c\n  x", "? |\n  a\n: b", "a: |\n  x\n  # not a comment\n",

	// Anchors, aliases, tags.
	"a: &x 1\nb: *x", "*x : a", "&a a: b", "!!str a", "!t &a b", "&a !t b", "!!map\na: b", "&a\n- b", "a: !!seq\n- b",
	"a: &x\n  b: c", "a: !t\n  - b", "--- !t\na: b", "!!set\n? a\n? b", "a: !!str", "a: &x", "- !t\n- &a\n- *a",
	"a: !<tag:x> b", "a: !local b", "[*a, &b c, !t d]", "{*a : b, &c d: e}", "a: !t # c\n  b", "a: &x # c\n  # d\n  b: c",
	"!t\n# c\nb", "! a", "!! a", "a: ! b", "&a", "!t", "&a # c", "!t # c\n", "*a", "*a # c", "&a !t # c\n# d\nb",
	"%TAG !e! tag:example.com,2000:\n---\n!e!x a", "%TAG ! tag:x,2000:\n---\n!a b", "a: !!binary aGVsbG8=",
	"a: !!timestamp 2001-12-14", "- !!int 1\n- !!float 1.5\n- !!null\n- !!bool true", "<<: *a", "a: &a\n  b: c\nd:\n  <<: *a",

	// Comments in every slot.
	"# c\na: b", "a: b # c", "a: # c\n  b", "a:\n  # c\n  b", "a:\n  - b\n  # c\n", "a:\n  b: c\n  # d", "a: b\n# c",
	"a: b\n\n# c", "a:\n  b: c\n\n# d", "# c\n---\na: b", "--- # c\na", "a\n... # c", "a: b\n# c\nd: e", "- a # c\n- b",
	"- # c\n  a", "- # c", "a: # c", "? a # c\n: b", "? a\n# c\n: b", "a: # c\n  - b", "a:\n  - b # c\n  - c",
	"- a\n  # c\n- b", "- - a\n  # c\n- b", "a:\n  b:\n    c: d\n    # e\n  # f\n# g", "# a\n\n# b\n\nc: d", "a: b # c\n# d",
	"a: [b] # c", "a: {b: c} # d", "a:\n  # b\n\n  # c\n  d: e", "%YAML 1.2 # c\n---\na", "# c\n%YAML 1.2\n---\na",
	"a:\n- b\n# c\n\n# d", "a: b\n  # c", "? # c\n  a\n: b", "key: # c\n  value", "a: !t # c\n  - b", "a:\n  b:\n   #b\n #a",
	"a:\n  b:\n #a\n   #a", "a:\n  - b:\n      c\n    # d\n  # e", "? a\n  # b\n: c\n  # d", "? a\n: b\n  # c\n# d",
	"? |\n  a\n  # b\n: c", "a: |\n  x\n  # c\nb: d", "[\n  a\n  # c\n]", "{\n  a: b,\n  # c\n}", "[a, # c\n b]",
	"{a: 1, # c\n b: 2}", "[\n# c\n]", "{\n# c\n}", "[a # c\n]", "{a: b # c\n}", "a:\n  - [b,\n    # c\n    d]",
	"- a\n # c\n- b", "-\n  # c\n  a", "a:\n  # c\n", "a:\n# c\nb:", "? a\n# c", "? a\n  # c", "- ? a\n  # c\n  : b",
	"# c\n# d\n---\n# e\na\n# f\n...\n# g\n---\n# h\nb\n# i", "a # c\n---\nb # d", "--- # c\n--- # d\n",

	// Flow collections.
	"[a, b, c]", "{a: 1, b: 2}", "[]", "{}", "[ ]", "{ }", "a: []", "a: {}", "[a, b, ]", "{a, b}", "{a: }", "[a: b]",
	"{? a}", "{: b}", "[? a : b]", "[a: b, c: d]", "{a: [1, 2], b: {c: d}}", "[[a], [b, [c]]]", "[{a: b}, {c: d}]",
	"[\n  a,\n\n  b\n]", "[a\n b]", "{a: b\n c}", "[\"a\", 'b']", "[a, b]: c", "{a: b}: c", "? [a, b]\n: c", "{a: b, c: }",
	"{a, b: c, }", "[a, {b: c}, ]", "a: {\n}", "a: { } # c\n", "{a: [b]}: c", "[,]", "{a: 1, : }", "[a, : ]", "{ : }",
	"[ : ]", "{a, :}", "[!t a: b]", "[&x a: b]", "[!t a]", "[&a [b]]", "{!t a: !u b}", "[a: !t ]", "[: b]", "[a:]",
	"[\"a\": b]", "['a':b]", "[a, b: c, d]", "!t [a]", "&a {b: c}", "!t # c\n[a]", "[*a : b]", "[? a]", "[? ]",
	"{a: b, c}", "{\"a\":b}", "[a,\n# c\nb]", "{a: [b, {c: [d]}]}", "[[[[[[a]]]]]]", "[a, [b, c], {d: e}, f: g]",

	// Mapping items: explicit and implicit keys, empty keys and values.
	"? a\n: b", "? - a\n: b", "? a\n", ": b", "? \n: ", ":", "? a\n  b\n: c", "? !t a\n: b", "a:\n  # c\n  b",
	"a: !!map\n  b: c", "a: &x\n  - b", "a:\n  b:\n    c:\n      d: e", "a b: c", "a:b: c", "a:\n  b", "a:\n\n  b",
	"a: \n  - b\n  - c", "a:\n- b\n- c", "a:\n  -\n  - b", "? a\n: - b\n  - c", "*a :", "*a : b", "? *a\n: b", "&a a:",
	"a: [b]\nc: {d: e}", "? a\n: b\n? c\n: d", "? \"a\n  b\"\n: c", "'a\n  b': c", "? a", "?", "? # c", ": # c", "!t a: b",
	"!t : b", "&a : b", "? &a\n: b", "a:\n  !t\n  b: c", "a: !t\n", "? !t\n: !u\n", "a:\n  ? b\n  : c\n  d: e",

	// Sequences.
	"- a\n- b", "- - a\n  - b", "- a: b\n  c: d", "-\n  a", "- ", "-", "- - - a", "- a\n\n- b", "-   a\n-   b",
	"- [a]\n- {b: c}", "- ? a\n  : b", "- &a\n  - b", "- !t\n  a: b", "a:\n  - b\n  -\n    - c", "- !t", "- &a",
	"- !t # c\n- b", "- &a !t\n  a", "-\n-\n-", "- # c\n- # d", "- a\n-", "- : b", "- ? a",

	// The !!pairs and !!omap shapes: sequences of pairs, and the items upstream cannot read.
	"!!pairs\n- a: b\n- c: d", "!!omap\n- a: b\n- c: d", "- !!pairs\n  - &x a: b\n", "!!pairs\n- a\n", "!!pairs\n- {}\n",
	"!!pairs\n- a: b\n- c\n", "!!omap\n- !!map {}\n", "!!pairs [a]", "!!pairs [a: b]", "!!pairs\n- &a a\n", "!!pairs\n- !t\n",
	"!!omap\n- !t a: b", "!!pairs\n- # c\n  a: b", "!!pairs\n- ? a\n  : b", "!!pairs\n- : b", "!!omap\n- a: b\n  # c",
	"!!pairs\n- &x a: b", "!!pairs\n- !t a: b", "a: !!pairs\n  - b: c", "!!set\n? a\n? b: c", "!!omap\n- a: 1\n- a: 2",
	"!!pairs\n- a:", "!!pairs\n- a: b # c\n", "!!pairs\n- &x\n  a: b",

	// Documents, markers and directives.
	"---\na\n---\nb", "a\n...\n---\nb", "%YAML 1.2\n---\na", "%YAML 1.1\n---\na: yes", "---\n---", "--- a",
	"a\n--- # c\nb", "a\n...\n# c\n---\nb", "%YAML 1.2\n---\na\n...\n%YAML 1.2\n---\nb", "a\n...\n", "---\na\n...\n",
	"--- # c\n", "---\n# c\n", "a\n---\n", "---\n...\n---\n...\n", "# c\n---\n# d\n...", "%FOO bar baz\n---\na",
	"%FOO\tbar  baz \n---", "%FOO \n---", "%YAML 1.2\n# c\n%TAG ! !\n# d\n---", "%YAML 1.2 # c\n# d\n---\n",
	"---\n- a\n---\n- b\n...\n", "a\n---\n\n\nb", "--- !!map\na: b\n--- !!seq\n- c", "a: b\n... # c\n---\nc",
	"a\n... # c\n", "--- a\n... # c\n--- b", "...\n...\n", "...\n# c\n", "# c\n...\n", "--- # c\n...\n",
	"--- !t\n...", "--- &a\n...", "--- !t # c\n", "---\n!t\n", "--- # c\n# d\na", "--- |\n  a\n... # c\n",
	"---\n\n\n...\n", "---\r\na\r\n...\r\n", "a\n...\n\n\n", "--- a\n...\n--- b\n...\n", "--- # a\n--- # b\n...",

	// Errors: the composer's messages, and their positions.
	"a: b: c", "a:\n  b\n c", "[a", "{a", "'a", "\"a", "a: 'b", "- a\nb: c", "a: &", "*", "%YAML 1.2\na", "a\n%YAML",
	"--- a\nb: c", "a: [b, c", "a: {b", "? a\n? b\nc", "a: |\n x\n  y\n z", "\ta: b", "a:\n\t- b", "{a: b}}", "]", "}",
	"a: b\n  - c", "&a &b c", "!a !b c", "%TAG\n---", "@a", "`a", "a: @b", "a: ]", "a: \"\\q\"", "a: !e!x b",
	"%YAML 2.0\n---\na", "%YAML 1.2\n%YAML 1.2\n---\na", "a: b\n...\nc: d\n... e", "[a]: b: c", "- a: b\n - c",
	"a: |0\n  x", "a: >10\n  x", "a: |x\n  y", "*a\n*b", "a: 'b\n\nc", "a: 1\na: 2", "{a: 1, a: 2}",

	// Line endings, tabs, non-ASCII and astral characters.
	"a: b\r\nc: d\r\n", "a: |\r\n  x\r\n  y\r\n", "\ufeffa: b\n", "a: b\rc: d", "a:\tb", "- é\n- 中文\n- 🍐🍐",
	"\"é 🍐\": '中'", "a: |\n  🍐 x\n  中\n", "🍐: 🍐 # 🍐\n# 🍐\n- x", "é:\n  - 🍐\n  # é\n", "a: b # c\r\n# d\r\n",
	"- a\r\n  # c\r\n- b\r\n", "--- # c\r\n...\r\n", "[🍐, é]: {中: 文}", "? 🍐\n: é", "a:\t# c\n  b", "a: b\t# c",
}

// generatedInputs draws inputs from YAML's indicators, properties, scalar styles, spaces, tabs, line
// breaks, comments and a few letters and non-ASCII characters, seeded so every run compares the same
// set.
func generatedInputs(count int) []string {
	random := rand.New(rand.NewPCG(5, 6))
	pieces := []string{
		" ", " ", " ", "  ", "\n", "\n", "\n", "\n  ", "\n    ", "\t", "\r\n",
		"-", "- ", "- ", "? ", ": ", ": ", ":", "#", " #", " # c", " # c", "\n# c", "\n  # c", "[", "]", "{", "}", ",",
		", ", "'", "\"", "|", ">", "|-", ">+", "|2", "&a ", "&", "!", "!!str ", "!!set ", "!!omap ", "!!pairs ",
		"!!map ", "!!seq ", "!foo ", "*a", "<<", "%YAML 1.2\n", "%TAG !e! tag:x,\n", "---", "...", "--- ", "... ",
		"a", "b", "ab", "x y", "1", "\u00e9", "\U0001F600", "\u4e2d",
	}
	inputs := make([]string, 0, count)
	for range count {
		length := 1 + random.IntN(24)
		var builder strings.Builder
		for range length {
			builder.WriteString(pieces[random.IntN(len(pieces))])
		}
		inputs = append(inputs, builder.String())
	}
	return inputs
}

func TestParseMatchesUpstream(t *testing.T) {
	t.Parallel()
	failures, errored := compareAll(t, "fixture", nil, fixtures, 40)
	t.Logf("%d of %d fixtures match upstream (%d of them errors)", len(fixtures)-failures, len(fixtures), errored)
}

func TestGeneratedParseMatchesUpstream(t *testing.T) {
	t.Parallel()
	count := 20000
	if testing.Short() {
		count = 2000
	}
	failures, errored := compareAll(t, "generated", nil, generatedInputs(count), 30)
	t.Logf("%d of %d generated inputs match upstream (%d of them errors)", count-failures, count, errored)
}

// TestOracleCanFail proves the comparison sees a difference: the port's tree for one text against
// upstream's for another, where the trees differ in a single position, value, comment slot or error.
func TestOracleCanFail(t *testing.T) {
	t.Parallel()
	cases := []struct{ mine, theirs string }{
		{"a:  b", "a: b"},         // a position
		{"a: c", "a: b"},          // a value
		{"a: b\n# c", "a: b # c"}, // a comment's slot
		{"a: [b", "a: {b"},        // an error's message
		{"- a\n- b", "- a\n- b\n"},
	}
	theirs := make([]string, len(cases))
	for index, each := range cases {
		theirs[index] = each.theirs
	}
	expected := oracleParse(t, theirs)
	for index, each := range cases {
		if differences := compareParse(each.mine, expected[index]); len(differences) == 0 {
			t.Errorf("the comparison found no difference between %q and %q", each.mine, each.theirs)
		}
	}
}

const suiteScript = `
import suite from "FORK/node_modules/yaml-test-suite/index.js";
const inputs = [];
for (const test of suite) {
  for (const testCase of test.cases) {
    if (typeof testCase.yaml === "string") inputs.push(testCase.yaml);
    if (typeof testCase.dump === "string") inputs.push(testCase.dump);
  }
}
process.stdout.write(JSON.stringify(inputs));
`

// TestYamlTestSuiteParseMatchesUpstream compares every input of the yaml-test-suite in the fork, and each
// case's canonical dump.
func TestYamlTestSuiteParseMatchesUpstream(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	script := filepath.Join(t.TempDir(), "suite.mjs")
	if err := os.WriteFile(script, []byte(strings.ReplaceAll(suiteScript, "FORK", forkRoot(t))), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command("node", script).Output()
	if err != nil {
		t.Skipf("the yaml-test-suite is not loadable: %v", err)
	}
	var texts []string
	if err := json.Unmarshal(output, &texts); err != nil {
		t.Fatal(err)
	}
	failures, errored := compareAll(t, "suite", nil, texts, 30)
	t.Logf("%d of %d yaml-test-suite inputs match upstream (%d of them errors)", len(texts)-failures, len(texts), errored)
}

// javaScriptSpaceText is the set String.prototype.trim removes, for markdown's front matter value.
const javaScriptSpaceText = "\t\n\v\f\r \u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000\ufeff"

// corpusSizeLimit skips the generated multi-megabyte files the printer's corpus test skips.
const corpusSizeLimit = 1 << 20

// TestCorpusParseMatchesUpstream compares every .yaml and .yml file under COHERE_YAML_CORPUS, and the
// YAML front matter value of every .md file there, enumerated as the printer's corpus test enumerates
// them: through the engine's ignore layers, into nested repositories. A measuring run.
func TestCorpusParseMatchesUpstream(t *testing.T) {
	t.Parallel()
	roots := os.Getenv("COHERE_YAML_CORPUS")
	if roots == "" {
		t.Skip("set COHERE_YAML_CORPUS to measure; this is a measuring run, not a unit test")
	}
	enumerator, err := prettier.New(formatoptions.Default())
	if err != nil {
		t.Fatal(err)
	}
	var names, texts []string
	yamlFiles, frontMatters, tooLarge := 0, 0, 0
	pending := strings.Split(roots, ":")
	seen := map[string]bool{}
	// Every corpus walked, logged once the loop ends, so a run that reaches fewer repositories says so.
	var measured []string
	for len(pending) > 0 {
		root := strings.TrimSpace(pending[0])
		pending = pending[1:]
		if strings.HasPrefix(root, "~/") {
			home, _ := os.UserHomeDir()
			root = filepath.Join(home, root[2:])
		}
		if root == "" || seen[root] {
			continue
		}
		seen[root] = true
		// Nested repositories regardless of root's own ignore lists, as the differential finds them
		// (#k6vebep): a repository under a path root never formats is still a corpus.
		nestedRepositories, err := formatfiles.NestedRepositoriesBelow(root)
		if err != nil {
			t.Fatalf("finding the repositories nested in %s: %v", root, err)
		}
		for _, nested := range nestedRepositories {
			pending = append(pending, filepath.Join(root, nested))
		}

		enumeration, err := enumerator.Enumerate(root)
		if errors.Is(err, formatoptions.ErrPrettierConfigRemains) {
			t.Logf("skipping %s, not yet adopted: %v", root, err)
			continue
		}
		if err != nil {
			t.Fatalf("enumerating %s: %v", root, err)
		}
		measured = append(measured, fmt.Sprintf("%s (%d)", root, len(enumeration.Files)))
		for _, file := range enumeration.Files {
			if !filepath.IsAbs(file) {
				file = filepath.Join(root, file)
			}
			extension := strings.ToLower(filepath.Ext(file))
			if extension != ".yaml" && extension != ".yml" && extension != ".md" {
				continue
			}
			source, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if len(source) > corpusSizeLimit {
				tooLarge++
				continue
			}
			// Prettier's core: no byte order mark, \n line endings.
			text := strings.TrimPrefix(string(source), "\ufeff")
			text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
			if extension == ".md" {
				frontMatter, _ := mdast.ParseFrontMatter(text)
				if frontMatter == nil || frontMatter.Language != "yaml" {
					continue
				}
				value := strings.Trim(frontMatter.Value, javaScriptSpaceText)
				if value == "" {
					continue
				}
				frontMatters++
				names = append(names, file+" (front matter)")
				texts = append(texts, value)
				continue
			}
			yamlFiles++
			names = append(names, file)
			texts = append(texts, text)
		}
	}
	t.Logf("corpora walked: %d\n  %s", len(measured), strings.Join(measured, "\n  "))
	failures, errored := compareAll(t, "corpus", names, texts, 30)
	t.Logf("%d of %d inputs match upstream (%d yaml files, %d front matter values; %d of them errors); %d files over %d bytes not measured",
		len(texts)-failures, len(texts), yamlFiles, frontMatters, errored, tooLarge, corpusSizeLimit)
}

// TestComparisonSeesEveryField corrupts upstream's correct tree one field at a time and checks that the
// comparison reports each corruption: every field the comparison reads, positions, identities and
// parents, and null told apart from an empty array.
func TestComparisonSeesEveryField(t *testing.T) {
	t.Parallel()
	text := "%YAML 1.2 # d\n---\n- &a !t |2- # i\n   x\n- *a # t\n- ? k\n  : v\n  # e\n# l\n- [b, {c: d}]\n"
	expected := oracleParse(t, []string{text})[0]
	if expected.Error != nil {
		t.Fatalf("the fixture does not parse: %s", expected.Error.Message)
	}
	if differences := compareParse(text, expected); len(differences) != 0 {
		t.Fatalf("the uncorrupted tree differs: %v", differences)
	}
	// find returns the first object, depth first, of the type.
	var find func(node any, nodeType string) map[string]any
	find = func(node any, nodeType string) map[string]any {
		switch value := node.(type) {
		case map[string]any:
			if value["type"] == nodeType {
				return value
			}
			for _, key := range []string{"children", "tag", "anchor", "middleComments", "leadingComments", "indicatorComment", "trailingComment", "endComments"} {
				if found := find(value[key], nodeType); found != nil {
					return found
				}
			}
		case []any:
			for _, each := range value {
				if found := find(each, nodeType); found != nil {
					return found
				}
			}
		}
		return nil
	}
	point := func(node map[string]any, which string) map[string]any {
		return node["position"].(map[string]any)[which].(map[string]any)
	}
	corruptions := map[string]func(root map[string]any){
		"start offset": func(root map[string]any) {
			point(find(root, "plain"), "start")["offset"] = point(find(root, "plain"), "start")["offset"].(float64) + 1
		},
		"end line":      func(root map[string]any) { point(find(root, "mapping"), "end")["line"] = 99.0 },
		"start column":  func(root map[string]any) { point(find(root, "comment"), "start")["column"] = 7.0 },
		"document end":  func(root map[string]any) { point(find(root, "document"), "end")["offset"] = 3.0 },
		"value":         func(root map[string]any) { find(root, "alias")["value"] = "b" },
		"tag value":     func(root map[string]any) { find(root, "tag")["value"] = "!u" },
		"anchor value":  func(root map[string]any) { find(root, "anchor")["value"] = "" },
		"comment value": func(root map[string]any) { find(root, "comment")["value"] = " x" },
		"chomping":      func(root map[string]any) { find(root, "blockLiteral")["chomping"] = "keep" },
		"indent":        func(root map[string]any) { find(root, "blockLiteral")["indent"] = nil },
		"type":          func(root map[string]any) { find(root, "blockLiteral")["type"] = "blockFolded" },
		"marker":        func(root map[string]any) { find(root, "document")["directivesEndMarker"] = false },
		"end marker":    func(root map[string]any) { find(root, "document")["documentEndMarker"] = true },
		"name":          func(root map[string]any) { find(root, "directive")["name"] = "TAG" },
		"parameters":    func(root map[string]any) { find(root, "directive")["parameters"] = []any{} },
		"parent":        func(root map[string]any) { find(root, "plain")["parent"] = 0.0 },
		"identity":      func(root map[string]any) { find(root, "comment")["id"] = 9999.0 },
		"keys":          func(root map[string]any) { find(root, "plain")["keys"] = []any{"type"} },
		"null for []":   func(root map[string]any) { find(root, "mappingValue")["leadingComments"] = nil },
		"no tag":        func(root map[string]any) { find(root, "blockLiteral")["tag"] = nil },
		"no indicator":  func(root map[string]any) { find(root, "blockLiteral")["indicatorComment"] = nil },
		"trailing moved": func(root map[string]any) {
			alias := find(root, "alias")
			alias["leadingComments"] = []any{alias["trailingComment"]}
			alias["trailingComment"] = nil
		},
		"end comment": func(root map[string]any) {
			var corrupt func(node any) bool
			corrupt = func(node any) bool {
				switch value := node.(type) {
				case map[string]any:
					if comments, isList := value["endComments"].([]any); isList && len(comments) > 0 {
						value["endComments"] = []any{}
						return true
					}
					for _, key := range []string{"children", "tag", "anchor"} {
						if corrupt(value[key]) {
							return true
						}
					}
				case []any:
					for _, each := range value {
						if corrupt(each) {
							return true
						}
					}
				}
				return false
			}
			if !corrupt(root) {
				t.Error("the fixture has no end comment to corrupt")
			}
		},
		"root comments": func(root map[string]any) { root["comments"] = root["comments"].([]any)[1:] },
		"child dropped": func(root map[string]any) {
			flow := find(root, "flowSequence")
			flow["children"] = flow["children"].([]any)[:1]
		},
		"middle comments": func(root map[string]any) { find(root, "plain")["middleComments"] = []any{find(root, "comment")} },
	}
	for name, corrupt := range corruptions {
		var root map[string]any
		if err := json.Unmarshal(expected.Root, &root); err != nil {
			t.Fatal(err)
		}
		corrupt(root)
		encoded, err := json.Marshal(root)
		if err != nil {
			t.Fatal(err)
		}
		if differences := compareParse(text, oracleResult{Root: encoded}); len(differences) == 0 {
			t.Errorf("corrupting the %s went unnoticed", name)
		}
	}
	// And the errors: a different message, name or position.
	errorText := "a: [b"
	expectedError := oracleParse(t, []string{errorText})[0]
	for name, corrupt := range map[string]func(*jsonError){
		"message":  func(e *jsonError) { e.Message += "." },
		"name":     func(e *jsonError) { e.Name = "TypeError" },
		"position": func(e *jsonError) { e.Position.End.Offset++ },
	} {
		copied := *expectedError.Error
		position := *copied.Position
		copied.Position = &position
		corrupt(&copied)
		if differences := compareParse(errorText, oracleResult{Error: &copied}); len(differences) == 0 {
			t.Errorf("corrupting the error's %s went unnoticed", name)
		}
	}
}
