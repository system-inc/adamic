package selector

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/estree"
)

// selectorParseScript is the oracle: postcss-selector-parser 2.2.3 from the fork's node_modules, called
// the way Prettier's parse-selector.js calls it, and dumped in the shape dumpSelector prints: every own
// key but parent, sorted, a key the library wrote as undefined dumping as "<undefined>". A css input is
// parsed with the fork's postcss 8.5.16 first, and each selector Prettier would hand parseSelector (a
// rule's raw selector, and the params of custom-selector, extend, nest and at-root) becomes a record of
// its own, so the Go side reads the selector text from the record. Offsets are UTF-16 there; the Go
// side converts its byte offsets back before comparing. An error dumps as its message. Strings dump
// through toWellFormed, the one place Go cannot follow (see wtf8.go), and NaN dumps as JSON's null.
const selectorParseScript = `
import { createRequire } from "node:module";
import { writeSync } from "node:fs";
// Synchronous, so the records before a selector the library hangs on reach the Go side.
function out(s) {
  const b = Buffer.from(s); let o = 0;
  while (o < b.length) { try { o += writeSync(1, b, o); } catch (e) { if (e.code !== "EAGAIN") throw e; } }
}
const require = createRequire(process.cwd() + "/");
const Processor = require("postcss-selector-parser/dist/processor.js");
const postcss = require("postcss");
function dump(v) {
  if (v === undefined) return "<undefined>";
  if (v === null) return null;
  if (Array.isArray(v)) return v.map(dump);
  if (typeof v === "string") return v.toWellFormed();
  if (typeof v !== "object") return v;
  const o = {};
  for (const k of Object.keys(v).sort()) { if (k === "parent") continue; o[k] = dump(v[k]); }
  return o;
}
function parseOne(text) {
  try {
    let result;
    new Processor((selectors) => { result = selectors; }).process(text);
    return { ast: dump(result) };
  } catch (e) {
    return { error: String(e && e.message) };
  }
}
function selectorsOf(css) {
  const out = [];
  postcss.parse(css).walk((node) => {
    if (typeof node.selector === "string") {
      let selector = node.raws.selector ? (node.raws.selector.scss ?? node.raws.selector.raw) : node.selector;
      if (node.raws.between && node.raws.between.trim().length > 0) selector += node.raws.between;
      if (selector.trim().length > 0) out.push(selector);
    }
    if (node.type === "atrule") {
      let params = node.raws.params ? (node.raws.params.scss ?? node.raws.params.raw) : node.params;
      if (node.raws.afterName && node.raws.afterName.trim().length > 0) params = node.raws.afterName + params;
      if (node.raws.between && node.raws.between.trim().length > 0) params += node.raws.between;
      params = params.trim();
      if (node.name === "custom-selector") {
        const match = node.params.match(/:--\S+\s+/);
        if (match) out.push(node.params.slice(match[0].trim().length).trim());
      } else if (params.length > 0 && (node.name === "extend" || node.name === "nest" ||
          (node.name === "at-root" && !/^\(\s*(?:without|with)\s*:.+\)$/s.test(params)))) {
        out.push(params);
      }
    }
  });
  return out;
}
let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", () => {
  for (const line of input.split("\n").filter(Boolean)) {
    const { name, text, css } = JSON.parse(line);
    if (css) {
      selectorsOf(text).forEach((selector, index) => {
        out(JSON.stringify({ name: name + " #" + index, text: selector, corpus: true, ...parseOne(selector) }) + "\n");
      });
    } else {
      out(JSON.stringify({ name, text, ...parseOne(text) }) + "\n");
    }
  }
});
`

// selectorFixture is an inline input. malformed says which way the library must go on it, so an oracle
// that refuses a valid fixture, or accepts a malformed one, fails the test rather than quietly agreeing.
type selectorFixture struct {
	name      string
	text      string
	malformed bool
}

var selectorFixtures = []selectorFixture{
	// Simple selectors, every node type.
	{name: "empty", text: ""},
	{name: "only whitespace", text: "   "},
	{name: "tag", text: "div"},
	{name: "class", text: ".a"},
	{name: "id", text: "#id"},
	{name: "universal", text: "*"},
	{name: "nesting", text: "&"},
	{name: "nesting compound", text: "&.active > &:hover"},
	{name: "compound", text: "a.b#c.d"},
	{name: "string", text: "a 'b' \"c\""},
	{name: "at-word", text: "@media a"},
	{name: "at-word then word", text: "@b.c"},
	{name: "comment", text: "a /* c */ b"},
	{name: "leading comment", text: "/* c */ a"},
	{name: "comment glued", text: "a/**/b"},
	{name: "multi-line comment", text: "a /*\n multi\n line */ > b"},
	{name: "keyframe selectors", text: "0%, 50.5%, to"},

	// Combinators and spacing.
	{name: "descendant", text: "a b"},
	{name: "child", text: "a > b"},
	{name: "adjacent glued", text: "a+b"},
	{name: "sibling", text: "a ~ b"},
	{name: "wide spacing", text: "a   >   b"},
	{name: "newline and tab", text: "a\n>\tb"},
	{name: "space before only", text: "a >b"},
	{name: "space after only", text: "a> b"},
	{name: "doubled combinator", text: "a>>b"},
	{name: "column combinator", text: "a || b"},
	{name: "leading combinator", text: "> a"},
	{name: "newline descendant", text: "a\n  b"},
	{name: "crlf", text: "a,\r\nb \r\n c"},
	{name: "form feed", text: "a\fb"},
	{name: "leading and trailing spaces", text: "  a  "},
	{name: "spaces around commas", text: " a , b ,c"},
	{name: "trailing comma", text: "a,b,"},
	{name: "trailing comma after space", text: "a ,"},
	{name: "empty middle selector", text: "a, ,b"},
	{name: "leading comma", text: ",a"},
	{name: "double comma", text: "a,,b"},

	// Attributes.
	{name: "attribute", text: "[a]"},
	{name: "attribute equals", text: "[a=b]"},
	{name: "attribute single quoted", text: "[a='b']"},
	{name: "attribute double quoted", text: "[a=\"b\"]"},
	{name: "attribute operators", text: "[a~=b][a|=b][a^=b][a$=b][a*=b]"},
	{name: "attribute insensitive", text: "[a=b i]"},
	{name: "attribute quoted insensitive", text: "[a='b' i]"},
	{name: "attribute insensitive tab and trailing space", text: "[a=b\ti ]"},
	{name: "attribute uppercase I", text: "[a=\"b\"  I]"},
	{name: "attribute spaces", text: "[ a = b ]"},
	{name: "attribute padded name", text: "[ a ]"},
	{name: "attribute namespace", text: "[ns|a=b]"},
	{name: "attribute empty namespace", text: "[|a]"},
	{name: "attribute any namespace", text: "[*|a]"},
	{name: "attribute empty value", text: "[a=\"\"]"},
	{name: "attribute no value", text: "[a=]"},
	{name: "attribute double equals", text: "[a==b]"},
	{name: "attribute bracket in string", text: "[a='b]c']"},
	{name: "attribute value with spaces", text: "[data-foo=\"bar baz\" i]"},
	{name: "attribute lone quote value", text: "[a=b\"c\"]"},
	{name: "attribute only i", text: "[a= i]"},
	{name: "attribute on a tag", text: "input[type=\"checkbox\"]:checked + label"},
	{name: "empty attribute", text: "[]"},

	// Pseudos.
	{name: "pseudo class", text: ":hover"},
	{name: "pseudo element", text: "::before"},
	{name: "tag pseudo", text: "a:hover::before"},
	{name: "triple colon", text: ":::a"},
	{name: "not", text: ":not(.a)"},
	{name: "not list", text: ":not(.a, .b)"},
	{name: "is", text: ":is(a, b)"},
	{name: "where tailwind dark", text: ":where(.dark, .dark *)"},
	{name: "nth-child", text: ":nth-child(2n + 1)"},
	{name: "nth-child of", text: ":nth-child(2n + 1 of .x)"},
	{name: "nth-child padded", text: ":nth-child( 2n+1 )"},
	{name: "nth-child negative", text: "li:nth-last-child(-n+3)"},
	{name: "pseudo then class", text: ":hover.foo"},
	{name: "nested pseudos", text: ":is(:not(a))"},
	{name: "deeply nested pseudos", text: ":is(:not(:has(> img, + a)), b) c"},
	{name: "attribute and pseudos", text: "a:not([href]):hover"},
	{name: "host", text: ":host(.x) ::slotted(span)"},
	{name: "part and lang", text: "::part(foo):lang(en):dir(rtl)"},
	{name: "has relative", text: "a:has(> img)"},
	{name: "empty pseudo arguments", text: ":not()"},
	{name: "space-only pseudo arguments", text: ":not( )"},
	{name: "pseudo argument comma", text: ":not(,)"},
	{name: "pseudo trailing comma inside", text: "a, :not(b,)"},
	{name: "custom selector name", text: ":--heading"},
	{name: "pseudo multi-line", text: ":is(\n  a,\n  b\n)"},
	{name: "pseudo with comment", text: ":not(/* x */ a)"},
	{name: "pseudo with string", text: ":contains('a b')"},

	// Parentheses after something that is not a pseudo.
	{name: "tag parentheses", text: "a(b)"},
	{name: "nested parentheses", text: ":not(a(b))"},
	{name: "attribute parentheses", text: "[a](x)"},
	{name: "nesting parentheses", text: "&(x (y))"},

	// Namespaces.
	{name: "namespace tag", text: "ns|a"},
	{name: "empty namespace", text: "|a"},
	{name: "any namespace", text: "*|a"},
	{name: "namespace universal", text: "ns|*"},
	{name: "any namespace universal", text: "*|*"},
	{name: "chained namespaces", text: "a|b|c"},
	{name: "namespace with class", text: "ns|a.b"},
	{name: "namespace after combinator", text: "a > ns|b"},
	{name: "bar combinator with spaces", text: "a | b"},
	{name: "bar after space", text: "a |b"},

	// Escapes.
	{name: "tailwind hover", text: ".hover\\:bg-red:hover"},
	{name: "escape then class", text: ".a\\:b.c"},
	{name: "tailwind fraction", text: ".md\\:w-1\\/2"},
	{name: "escaped digit id", text: "#\\31 23"},
	{name: "escaped digit class", text: ".\\31 0"},
	{name: "escaped space", text: ".a\\ b"},
	{name: "escaped at", text: "\\@media"},
	{name: "tailwind arbitrary value", text: ".w-\\[10px\\]"},
	{name: "tailwind group hover", text: ".group:hover .group-hover\\:opacity-100"},
	{name: "escaped backslash", text: ".a\\\\b"},
	{name: "tailwind arbitrary variant", text: ".\\[\\&_svg\\]\\:size-4"},
	{name: "escaped trailing backslash", text: "a\\"},
	{name: "escape before slash", text: "a\\/b"},
	{name: "escape before newline", text: "a\\\nb"},
	{name: "leading escape then space", text: "\\ b"},
	{name: "escaped dot inside word", text: ".a\\.b.c#d"},
	// A word ending in an escaped space leaves splitWord on the space token: NaN and undefined positions.
	{name: "escaped space then universal", text: ".\\ *"},
	{name: "escaped space in a pseudo", text: ":a\\ ,b"},
	{name: "escaped space at the end", text: "a.b\\ "},
	{name: "escaped space then word", text: "a\\ b c"},
	{name: "important bang", text: "a!b"},
	{name: "braces in a word", text: "a{b}"},

	// Sass interpolation, whose hash is not an id.
	{name: "interpolation", text: "#{$a}"},
	{name: "interpolation in a class", text: ".a#{b}#c"},

	// Non-ASCII text, where offsets differ between UTF-16 and bytes.
	{name: "accented class", text: ".caf\xc3\xa9 > .\xc3\xbc"},
	{name: "emoji classes", text: ".\xf0\x9f\x98\x80 > .\xc3\xbc, #\xf0\x9f\x98\x80"},
	{name: "non-ASCII attribute", text: "[title=\"\xe6\x97\xa5\xe6\x9c\xac\"] a"},
	{name: "emoji then compound", text: ".a\xf0\x9f\x98\x80.b#c"},
	{name: "emoji then escape merge", text: ".\xf0\x9f\x98\x80\\:x.y"},
	{name: "non-ASCII comment", text: "a /* \xc3\xa9\xf0\x9f\x98\x80 */ b"},
	{name: "non-breaking space before i", text: "[a=b\xc2\xa0i]"},
	// A backslash takes one UTF-16 unit, half of an emoji's pair; the halves meet again in the value.
	{name: "escaped emoji in a word", text: ".a\\\xf0\x9f\x98\x80b:hover"},
	{name: "escaped emoji in an attribute", text: "[a=\\\xf0\x9f\x98\x80]"},
	{name: "escaped emoji in word parentheses", text: "a(\\\xf0\x9f\x98\x80)"},
	{name: "escaped emoji before a namespace bar", text: "\\\xf0\x9f\x98\x80|a"},
	{name: "quoted value ending in an emoji", text: "[a=\"x\"\xf0\x9f\x98\x80]"},

	// Malformed: each must be refused by both sides with the same message.
	{name: "unclosed double quote", text: "\"unclosed", malformed: true},
	{name: "unclosed single quote", text: "a 'x", malformed: true},
	{name: "escaped closing quote", text: "a 'x\\'", malformed: true},
	{name: "unclosed comment", text: "a /* x", malformed: true},
	{name: "unclosed attribute", text: "[a", malformed: true},
	{name: "unclosed attribute with value", text: "[a=b", malformed: true},
	{name: "closing bracket", text: "a]", malformed: true},
	{name: "closing parenthesis", text: ")", malformed: true},
	{name: "closing parenthesis after tag", text: "a)", malformed: true},
	{name: "extra closing parenthesis", text: ":not(a))", malformed: true},
	{name: "unclosed pseudo", text: ":not(a", malformed: true},
	{name: "unclosed word parentheses", text: "a(b", malformed: true},
	{name: "semicolon", text: "a;b", malformed: true},
	{name: "lone colon", text: ":", malformed: true},
	{name: "trailing colon", text: "a:", malformed: true},
	{name: "misplaced parenthesis", text: ":hover.foo(", malformed: true},
	{name: "colon then parenthesis", text: "::(a)", malformed: true},
	{name: "colon then attribute", text: "a:[x]", malformed: true},
	{name: "colon then at-word", text: "a:@b", malformed: true},
	{name: "colon then space", text: "a: b", malformed: true},
	{name: "trailing namespace bar", text: ".a|", malformed: true},
	{name: "lone bar", text: "|", malformed: true},
	{name: "leading parentheses", text: "(a)", malformed: true},
	{name: "empty nested parentheses", text: ":not(())", malformed: true},
	{name: "bracket only in a string", text: "[a=\"]\"", malformed: true},
}

// TestParseAgreesWithPostcssSelectorParser is the parser's acceptance test: for every selector
// Prettier would parse in the corpus's .css files, and every inline fixture, the Go tree must equal
// postcss-selector-parser 2.2.3's, property for property, and a refusal must be a refusal on both sides
// with the same message. Off unless COHERE_PRETTIER_FORK names the fork (whose node_modules hold the
// library and postcss) and COHERE_CSS_CORPORA lists .css files or directories to walk (colon-separated,
// node_modules skipped).
//
// A corpus selector the library cannot parse is counted, and the test fails if that count passes a
// tenth of the corpus: an oracle that rejects everything would otherwise read as perfect agreement.
// The inline fixtures each say which way the library must go.
func TestParseAgreesWithPostcssSelectorParser(t *testing.T) {
	t.Parallel()
	corpora := os.Getenv("COHERE_CSS_CORPORA")
	if corpora == "" {
		t.Skip("set COHERE_CSS_CORPORA to compare the parser against postcss-selector-parser")
	}
	root := corpus.PrettierFork.Root(t)

	type input = oracleInput
	var inputs []input
	corpusFiles := 0
	for _, corpus := range strings.Split(corpora, ":") {
		corpus = strings.TrimSpace(corpus)
		if corpus == "" {
			continue
		}
		if strings.HasPrefix(corpus, "~/") {
			home, _ := os.UserHomeDir()
			corpus = filepath.Join(home, corpus[2:])
		}
		err := filepath.WalkDir(corpus, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".git") {
				return filepath.SkipDir
			}
			if entry.IsDir() || filepath.Ext(path) != ".css" {
				return nil
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !utf8.Valid(source) {
				return fmt.Errorf("%s is not UTF-8, which the comparison cannot carry through JSON", path)
			}
			inputs = append(inputs, input{Name: path, Text: string(source), CSS: true})
			corpusFiles++
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", corpus, err)
		}
	}
	if corpusFiles == 0 {
		t.Fatalf("no .css files under %s", corpora)
	}
	malformed := map[string]bool{}
	for _, fixture := range selectorFixtures {
		name := "fixture: " + fixture.name
		inputs = append(inputs, input{Name: name, Text: fixture.text})
		malformed[name] = fixture.malformed
	}

	records := runSelectorOracle(t, root, inputs)

	seenFixtures := map[string]bool{}
	seenCorpusText := map[string]bool{}
	corpusSelectors, corpusDuplicates := 0, 0
	identical, refusedByBoth, corpusRefused, different := 0, 0, 0, 0
	fail := func(format string, arguments ...any) {
		different++
		if different <= 20 {
			t.Errorf(format, arguments...)
		}
	}
	for _, each := range records {
		name, _ := each["name"].(string)
		text, _ := each["text"].(string)
		isCorpus, _ := each["corpus"].(bool)
		if isCorpus {
			if seenCorpusText[text] {
				corpusDuplicates++
				continue
			}
			seenCorpusText[text] = true
			corpusSelectors++
		} else {
			seenFixtures[name] = true
		}
		wantError, oracleRefused := each["error"].(string)
		if expectMalformed, isFixture := malformed[name]; isFixture && expectMalformed != oracleRefused {
			fail("%s: the fixture says malformed=%v and the library says %v", name, expectMalformed, each)
			continue
		}
		got, err := dumpSelector(text)
		switch {
		case oracleRefused && err != nil:
			if err.Error() != wantError {
				fail("%s %q: both refuse, with different messages:\n  library %s\n  Go      %s", name, text, wantError, err.Error())
				continue
			}
			refusedByBoth++
			if isCorpus {
				corpusRefused++
			}
		case oracleRefused:
			fail("%s %q: the library refuses it (%s) and Go parsed it", name, text, wantError)
		case err != nil:
			fail("%s %q: the library parses it and Go refused: %v", name, text, err)
		default:
			if difference := firstSelectorDifference(oracleInBytes(each["ast"], text), got, "root"); difference != "" {
				fail("%s %q: %s", name, text, difference)
				continue
			}
			identical++
		}
	}
	for _, fixture := range selectorFixtures {
		if !seenFixtures["fixture: "+fixture.name] {
			t.Errorf("the oracle printed nothing for fixture %q", fixture.name)
		}
	}
	t.Logf("%d corpus files gave %d distinct selectors (%d repeats skipped); with %d fixtures: %d identical, %d refused by both with the same message, %d different; %d corpus selectors refused",
		corpusFiles, corpusSelectors, corpusDuplicates, len(selectorFixtures), identical, refusedByBoth, different, corpusRefused)
	if corpusSelectors == 0 {
		t.Errorf("the corpus gave no selectors")
	}
	if corpusRefused > corpusSelectors/10 {
		t.Errorf("the library refused %d of %d corpus selectors; the comparison is thinner than it looks", corpusRefused, corpusSelectors)
	}
}

// oracleInput is one line of the oracle script's input: a selector, or with CSS a stylesheet whose
// selectors the script extracts.
type oracleInput struct {
	Name string `json:"name"`
	Text string `json:"text"`
	CSS  bool   `json:"css,omitempty"`
}

// runSelectorOracle runs the oracle script on inputs and returns its records in order.
func runSelectorOracle(t *testing.T, root string, inputs []oracleInput) []map[string]any {
	t.Helper()
	var stdin bytes.Buffer
	for _, each := range inputs {
		encoded, err := json.Marshal(each)
		if err != nil {
			t.Fatal(err)
		}
		stdin.Write(encoded)
		stdin.WriteByte('\n')
	}
	// The library hangs on some selectors (see ErrLoopsForever), so the oracle runs on a deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--input-type=module", "-e", selectorParseScript)
	command.Dir = root
	command.Stdin = &stdin
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()

	var records []map[string]any
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var each map[string]any
		if err := json.Unmarshal([]byte(line), &each); err != nil {
			if runError != nil {
				break // the last line a killed process was writing
			}
			t.Fatal(err)
		}
		records = append(records, each)
	}
	if runError != nil {
		// Records come in input order, so the first input without one is where the library stopped.
		answered := map[string]bool{}
		for _, each := range records {
			name, _ := each["name"].(string)
			answered[strings.SplitN(name, " #", 2)[0]] = true
		}
		for _, each := range inputs {
			if !answered[each.Name] {
				t.Fatalf("node: %v, stopping at %s %q\n%s", runError, each.Name, each.Text, stderr.String())
			}
		}
		t.Fatalf("node: %v\n%s", runError, stderr.String())
	}
	return records
}

// randomSelectorPieces are the fragments TestParseAgreesOnRandomSelectors strings together: every
// character the tokenizer and parser branch on, plus words, escapes and non-ASCII text.
var randomSelectorPieces = []string{
	"a", "b", "div", "i", "n", "1", "-", "_", ".", "#", ":", "::", "(", ")", "[", "]", "=", "~=", "|=",
	"^=", "$=", "*=", "|", "*", "&", ",", " ", "  ", "\n", "\t", "\f", "\r\n", ">", "+", "~", "\\", "\\:",
	"\\ ", "'", "\"", "'x'", "\"y z\"", "/*", "*/", "/* c */", "/", "@", "@m", ";", "{", "}", "#{", "!", "%",
	"\xc3\xa9", "\xf0\x9f\x98\x80", "\xc2\xa0", ":not(", ":is(", ":nth-child(", " of ", "2n+1",
}

// TestParseAgreesOnRandomSelectors strings random fragments into selectors, most of them malformed in
// some way, and requires the same tree or the same error from Go and the library on each. Off unless
// COHERE_PRETTIER_FORK names the fork; COHERE_SELECTOR_RANDOM sets the count (default 20000). The seed is
// fixed, so a failure reproduces. A selector Go refuses with ErrLoopsForever is not sent to the library,
// which would hang on it; the run's deadline catches one Go wrongly lets through.
func TestParseAgreesOnRandomSelectors(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)
	count := 20000
	if value := os.Getenv("COHERE_SELECTOR_RANDOM"); value != "" {
		if _, err := fmt.Sscan(value, &count); err != nil {
			t.Fatalf("COHERE_SELECTOR_RANDOM: %v", err)
		}
	}
	random := rand.New(rand.NewPCG(2, 3))
	seen := map[string]bool{}
	var inputs []oracleInput
	loopsForever := 0
	for len(inputs) < count {
		var text strings.Builder
		for range 1 + random.IntN(10) {
			text.WriteString(randomSelectorPieces[random.IntN(len(randomSelectorPieces))])
		}
		if seen[text.String()] {
			continue
		}
		seen[text.String()] = true
		if _, err := Parse(text.String()); errors.Is(err, ErrLoopsForever) {
			loopsForever++
			continue
		}
		inputs = append(inputs, oracleInput{Name: fmt.Sprint(len(inputs)), Text: text.String()})
	}
	records := runSelectorOracle(t, root, inputs)
	if len(records) != len(inputs) {
		t.Fatalf("%d inputs and %d records", len(inputs), len(records))
	}
	parsed, refused, different := 0, 0, 0
	messages := map[string]int{}
	for index, each := range records {
		text := inputs[index].Text
		wantError, oracleRefused := each["error"].(string)
		got, err := dumpSelector(text)
		difference := ""
		switch {
		case oracleRefused && err != nil:
			if err.Error() != wantError {
				difference = fmt.Sprintf("both refuse: library %q, Go %q", wantError, err.Error())
			}
			messages[wantError]++
		case oracleRefused:
			difference = fmt.Sprintf("the library refuses it (%s) and Go parsed it", wantError)
		case err != nil:
			difference = fmt.Sprintf("the library parses it and Go refused: %v", err)
		default:
			difference = firstSelectorDifference(oracleInBytes(each["ast"], text), got, "root")
		}
		if difference != "" {
			different++
			if different <= 20 {
				t.Errorf("%q: %s", text, difference)
			}
			continue
		}
		if oracleRefused {
			refused++
		} else {
			parsed++
		}
	}
	t.Logf("%d random selectors (%d more skipped as ErrLoopsForever): %d parsed alike, %d refused alike, %d different", len(inputs), loopsForever, parsed, refused, different)
	for message, times := range messages {
		t.Logf("  %5d refused with %q", times, message)
	}
	if parsed < len(inputs)/10 {
		t.Errorf("only %d of %d random selectors parsed; the comparison is mostly of refusals", parsed, len(inputs))
	}
}

// dumpSelector parses one selector and renders it the way the oracle script does, through JSON, so
// both sides compare as the same Go values. sourceIndex stays in bytes; oracleInBytes converts the
// library's side.
func dumpSelector(text string) (any, error) {
	root, err := Parse(text)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(dumpSelectorValue(root))
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(encoded, &result)
	return result, err
}

func dumpSelectorValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return "<undefined>"
	case *estree.Node:
		if typed.Range != [2]int{} {
			return fmt.Sprintf("a %s with a Range, which the parser never sets", typed.Type())
		}
		result := map[string]any{"type": typed.Type()}
		for _, key := range typed.Keys() {
			result[key] = dumpSelectorValue(typed.Get(key))
		}
		return result
	case []*estree.Node:
		if typed == nil {
			return "a nil list, which no library field holds"
		}
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = dumpSelectorValue(child)
		}
		return result
	case map[string]any:
		result := map[string]any{}
		for key, each := range typed {
			result[key] = dumpSelectorValue(each)
		}
		return result
	case float64:
		if typed != typed {
			return nil // NaN, which JSON.stringify writes as null
		}
		return typed
	case string, bool, int:
		return typed
	default:
		return fmt.Sprintf("an unexpected %T", typed)
	}
}

// oracleInBytes converts every sourceIndex in the library's dump from a UTF-16 index into text to the
// byte offset Parse promises: the start of the character the index falls in (the middle of a surrogate
// pair included, which splitWord can produce), and past the end by the same distance. It is written
// apart from Parse's conversion so that a bug there cannot hide here.
func oracleInBytes(value any, text string) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, each := range typed {
			if index, isNumber := each.(float64); key == "sourceIndex" && isNumber {
				typed[key] = float64(utf16IndexToByte(text, int(index)))
				continue
			}
			typed[key] = oracleInBytes(each, text)
		}
	case []any:
		for index, each := range typed {
			typed[index] = oracleInBytes(each, text)
		}
	}
	return value
}

func utf16IndexToByte(text string, index int) int {
	units := 0
	for position, character := range text {
		width := len(utf16.Encode([]rune{character}))
		if index < units+width {
			return position
		}
		units += width
	}
	return len(text) + index - units
}

// firstSelectorDifference names the first path where two decoded JSON values differ, or "".
func firstSelectorDifference(want any, got any, path string) string {
	switch typed := want.(type) {
	case map[string]any:
		other, isMap := got.(map[string]any)
		if !isMap {
			return fmt.Sprintf("%s: want an object, got %v", path, got)
		}
		keys := map[string]bool{}
		for key := range typed {
			keys[key] = true
		}
		for key := range other {
			keys[key] = true
		}
		sorted := make([]string, 0, len(keys))
		for key := range keys {
			sorted = append(sorted, key)
		}
		sort.Strings(sorted)
		for _, key := range sorted {
			wantValue, wantPresent := typed[key]
			gotValue, gotPresent := other[key]
			if wantPresent != gotPresent {
				return fmt.Sprintf("%s.%s: present in the library %v, in Go %v (%v)", path, key, wantPresent, gotPresent, typed["type"])
			}
			if difference := firstSelectorDifference(wantValue, gotValue, path+"."+key); difference != "" {
				return difference
			}
		}
		return ""
	case []any:
		other, isList := got.([]any)
		if !isList || len(other) != len(typed) {
			return fmt.Sprintf("%s: want %d elements, got %v", path, len(typed), got)
		}
		for index := range typed {
			if difference := firstSelectorDifference(typed[index], other[index], fmt.Sprintf("%s[%d]", path, index)); difference != "" {
				return difference
			}
		}
		return ""
	}
	if fmt.Sprintf("%T %v", want, want) != fmt.Sprintf("%T %v", got, got) {
		return fmt.Sprintf("%s: want %#v, got %#v", path, want, got)
	}
	return ""
}
