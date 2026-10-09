package mediaquery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/estree"
)

// mediaQueryScript is the oracle: postcss-media-query-parser 0.2.3 from the fork's node_modules. Each
// input line is a fixture's params, or a corpus file, which the script parses with postcss 8 and walks
// for the at-rules Prettier's parser-postcss.js sends to parseMediaQuery: @media and @custom-media
// (case-insensitively), with params rebuilt as parseNestedCSS rebuilds them (raws.params, a
// non-blank afterName and between, trimmed), non-empty and without "#{". Every other at-rule's params
// go to parseValue or stay a string, so @supports, @container, @custom-variant, @import and the rest do
// not reach this parser. A file postcss refuses yields no params (Prettier would not get this far on it
// either). Each node dumps with its own keys sorted, parent left out; undefined dumps as
// "<undefined>" so presence is compared too. sourceIndex is UTF-16 there; the Go side converts.
const mediaQueryScript = `
import { createRequire } from "node:module";
const require = createRequire(process.cwd() + "/package.json");
const parseMedia = require("postcss-media-query-parser").default;
const postcssParse = require("postcss/lib/parse");
const fs = require("node:fs");
function dump(v) {
  if (v === undefined) return "<undefined>";
  if (v === null) return null;
  if (Array.isArray(v)) return v.map(dump);
  if (typeof v !== "object") return v;
  const o = {};
  for (const k of Object.keys(v).sort()) { if (k === "parent") continue; o[k] = dump(v[k]); }
  return o;
}
function run(name, params) {
  try {
    return { name, params, tree: dump(parseMedia(params)) };
  } catch (e) {
    return { name, params, error: String(e && e.message) };
  }
}
let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", () => {
  for (const line of input.split("\n").filter(Boolean)) {
    const { name, params, file } = JSON.parse(line);
    if (file === undefined) { process.stdout.write(JSON.stringify(run(name, params)) + "\n"); continue; }
    let root;
    try { root = postcssParse(fs.readFileSync(file, "utf8"), { map: false }); } catch { continue; }
    let count = 0;
    root.walkAtRules((node) => {
      if (!["media", "custom-media"].includes(node.name.toLowerCase())) return;
      let p = node.raws.params ? (node.raws.params.scss ?? node.raws.params.raw) : node.params;
      if (node.raws.afterName && node.raws.afterName.trim().length > 0) p = node.raws.afterName + p;
      if (node.raws.between && node.raws.between.trim().length > 0) p += node.raws.between;
      p = p.trim();
      if (p.length === 0 || p.includes("#{")) return;
      process.stdout.write(JSON.stringify(run(name + " @" + node.name + " #" + (++count), p)) + "\n");
    });
  }
});
`

// mediaQueryFixture is an inline params string. throws says which way the library must go on it, so an
// oracle that throws on a fixture meant to parse, or parses one meant to throw, fails the test rather
// than quietly agreeing.
type mediaQueryFixture struct {
	name   string
	params string
	throws bool
}

var mediaQueryFixtures = []mediaQueryFixture{
	// Media types, keywords and features.
	{name: "type and feature", params: "screen and (min-width: 768px)"},
	{name: "not", params: "not print"},
	{name: "only", params: "only screen"},
	{name: "bare type", params: "print"},
	{name: "bare feature", params: "(color)"},
	{name: "prefers color scheme", params: "(prefers-color-scheme: dark)"},
	{name: "prefers reduced motion", params: "(prefers-reduced-motion: reduce)"},
	{name: "vendor feature", params: "(-webkit-min-device-pixel-ratio: 2), (min-resolution: 192dpi)"},
	{name: "only with two features", params: "only screen and (min-width: 100px) and (max-width: 200px)"},
	{name: "not with feature", params: "not screen and (color), print and (orientation: landscape)"},
	{name: "features joined", params: "(min-width: 100px) and (max-width: 200px)"},
	{name: "upper case", params: "SCREEN AND (COLOR), NOT print, ONLY tv"},
	{name: "keywords alone", params: "and"},
	{name: "keywords in a row", params: "only and not"},

	// Comma lists.
	{name: "comma list", params: "screen, print"},
	{name: "comma list with spacing", params: "screen , print and (color)  ,tv"},
	{name: "empty queries in a list", params: "a,,b"},
	{name: "only commas", params: " , "},
	{name: "trailing comma", params: "screen,"},
	{name: "comma inside parens", params: "(a: f(1, 2)), print"},

	// Range syntax.
	{name: "range both sides", params: "(400px <= width <= 700px)"},
	{name: "range one side", params: "(width >= 600px)"},
	{name: "range strict", params: "screen and (400px < width < 700px)"},
	{name: "range with colon free value", params: "(height > 30em) and (width = 50em)"},

	// Spacing.
	{name: "no space inside", params: "screen and (max-width:100px)"},
	{name: "wide spacing", params: "  screen  and   (  max-width  :  100px  )   "},
	{name: "tabs and newlines", params: "screen\tand\n(color)\r\n,\n\tprint"},
	{name: "no space before paren", params: "screen and(color)"},
	{name: "no space after paren", params: "(color)and screen"},
	{name: "paren then word", params: "(a)b"},
	{name: "colon with empty value", params: "(a:)"},
	{name: "colon with spaces only", params: "(a:   )"},
	{name: "two colons", params: "(a: b: c)"},
	{name: "empty parens", params: "()"},

	// Comments (the library has no notion of them).
	{name: "comment between", params: "screen /* comment */ and (color)"},
	{name: "comment first", params: "/* c */ screen"},
	{name: "comment in feature", params: "(min-width: /* x */ 100px)"},

	// Custom media and functions.
	{name: "custom media", params: "--small-viewport (max-width: 30em)"},
	{name: "custom media list", params: "--narrow-window (max-width: 30em), print"},
	{name: "custom media reference", params: "(--small-viewport)"},
	{name: "calc value", params: "(max-width: calc(100px + 2em))"},
	{name: "var value", params: "screen and (min-width: var(--bp))"},
	{name: "function first", params: "fn(a, b) and (color)"},
	{name: "function alone", params: "fn(a b)"},
	{name: "nested parens", params: "(display: grid) and (not (display: inline-grid))"},
	{name: "style query", params: "style(--responsive: true)"},
	{name: "container params", params: "sidebar (min-width: 400px)"},
	{name: "custom variant params", params: "dark (&:where(.dark, .dark *))"},

	// The post-pass that types what the first pass left undefined.
	{name: "two words", params: "a b"},
	{name: "three words", params: "a b c"},
	{name: "four words", params: "a b c d"},
	{name: "five words", params: "a b c d e"},
	{name: "word paren word", params: "a (b) c"},
	{name: "word word paren", params: "a b (c)"},
	{name: "word word word paren", params: "a b c (d)"},
	{name: "word keyword word word", params: "a and b c"},
	{name: "paren then words", params: "(a) b c"},
	{name: "not then words", params: "not a b"},
	{name: "type then words", params: "only a b c"},

	// Interpolation and strings, which only the feature parser tracks.
	{name: "interpolation type", params: "#{$query} and (color)"},
	{name: "interpolation value", params: "screen and (max-width: #{$w})"},
	{name: "interpolation with colon", params: "(#{$a: b}: c)"},
	{name: "string in interpolation", params: "({'a}' : b}: c)"},
	{name: "escaped quote in interpolation", params: "({'a\\'b'}: c)"},
	{name: "double quote in interpolation", params: "({\"x:y\"}: z)"},
	{name: "quote outside interpolation", params: "(a: \"b:c\")"},
	{name: "braces first", params: "{a} b"},
	{name: "brace element", params: "{a b} (c)"},

	// url(), which the library peels off the front.
	{name: "url then type", params: "url(foo.css) screen"},
	{name: "url with spacing", params: "  url ( 'a(b)' )  print, tv"},
	{name: "url alone", params: "url(x)"},
	{name: "url not first", params: "screen url(x)"},
	{name: "url upper case", params: "URL(x) screen"},

	// Unbalanced parentheses: elements that never close are dropped.
	{name: "stray close paren", params: "screen)"},
	{name: "close paren mid word", params: "a)b c"},
	{name: "unclosed feature", params: "print and (color"},
	{name: "unclosed alone", params: "(a"},
	{name: "close paren first", params: ")"},
	{name: "open paren alone", params: "("},
	{name: "brace closing a feature", params: "(a} b", throws: true},
	{name: "brace closing a feature at the end", params: "screen and (color}", throws: true},
	{name: "brace closing a feature mid-word drops it", params: "(a}:b)"},
	{name: "interpolation then an extra brace", params: "(a{}}", throws: true},
	{name: "brace alone", params: "}"},
	{name: "brace pair element", params: "{a}"},

	// Empty and blank.
	{name: "empty", params: ""},
	{name: "blank", params: "   "},

	// Non-ASCII: offsets differ between UTF-16 and bytes, and JavaScript's \s is wider than ASCII.
	{name: "non-ASCII comment", params: "screen and (min-width: 1em) /* \xc3\xa9 */"},
	{name: "non-ASCII type", params: "\xc3\xa9cran and (color)"},
	// Before a feature's colon, where the feature is gathered byte by byte (@system_adamic's stream P2).
	{name: "no-break space before a feature", params: "(\u00a0x: 1px)"},
	{name: "non-ASCII in a feature", params: "(\xc3\xa9-width: 1px)"},
	{name: "non-ASCII after a feature's colon", params: "(x: \xc3\xa9)"},
	{name: "no-break space", params: "(max-width:\xc2\xa0100px)\xc2\xa0and\xc2\xa0print"},
	{name: "ideographic space", params: "\xe3\x80\x80screen\xe3\x80\x80and (color)"},
	{name: "emoji", params: "\xf0\x9f\x98\x80 and (a: \xf0\x9f\x98\x80), \xf0\x9f\x98\x80"},
	{name: "byte order mark", params: "\xef\xbb\xbfscreen"},
	{name: "next line is not whitespace", params: "a\xc2\x85b"},
	{name: "em space and line separator", params: "a\xe2\x80\x83b\xe2\x80\xa8c"},
}

// TestParseAgreesWithPostcssMediaQueryParser is the parser's acceptance test: for the params of every
// @media and @custom-media in the corpus, and every inline fixture, the Go tree must equal the
// library's, property for property, and a throw must be a refusal here with the same message. Off
// unless COHERE_PRETTIER_FORK names the fork (whose node_modules hold postcss-media-query-parser 0.2.3
// and postcss) and COHERE_CSS_CORPORA lists .css files or directories to walk (colon-separated,
// node_modules skipped).
//
// A corpus params string the library throws on is counted, and the test fails if that count passes a
// tenth of the corpus params, or if the corpus yields none: an oracle that rejects everything, or a
// corpus that reaches nothing, would otherwise read as agreement.
func TestParseAgreesWithPostcssMediaQueryParser(t *testing.T) {
	t.Parallel()
	corpora := os.Getenv("COHERE_CSS_CORPORA")
	if corpora == "" {
		t.Skip("set COHERE_CSS_CORPORA to compare the parser against postcss-media-query-parser")
	}
	root := corpus.PrettierFork.Root(t)

	type input struct {
		Name   string `json:"name"`
		Params string `json:"params"`
		File   string `json:"file,omitempty"`
	}
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
			inputs = append(inputs, input{Name: path, File: path})
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
	throws := map[string]bool{}
	for _, fixture := range mediaQueryFixtures {
		name := "fixture: " + fixture.name
		inputs = append(inputs, input{Name: name, Params: fixture.params})
		throws[name] = fixture.throws
	}

	var stdin bytes.Buffer
	for _, each := range inputs {
		encoded, err := json.Marshal(each)
		if err != nil {
			t.Fatal(err)
		}
		stdin.Write(encoded)
		stdin.WriteByte('\n')
	}
	command := exec.Command("node", "--input-type=module", "-e", mediaQueryScript)
	command.Dir = root
	command.Stdin = &stdin
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}

	var records []map[string]any
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var each map[string]any
		if err := json.Unmarshal([]byte(line), &each); err != nil {
			t.Fatal(err)
		}
		records = append(records, each)
	}

	identical, refusedByBoth, corpusParams, corpusRefused, fixturesSeen, different := 0, 0, 0, 0, 0, 0
	fail := func(format string, arguments ...any) {
		different++
		if different <= 20 {
			t.Errorf(format, arguments...)
		}
	}
	for _, want := range records {
		name := want["name"].(string)
		params := want["params"].(string)
		wantError, oracleRefused := want["error"].(string)
		expectThrow, isFixture := throws[name]
		if isFixture {
			fixturesSeen++
		} else {
			corpusParams++
		}
		if isFixture && expectThrow != oracleRefused {
			fail("%s: the fixture says throws=%v and the library says %v", name, expectThrow, want)
			continue
		}
		got, err := dumpMediaQuery(params)
		switch {
		case oracleRefused && err != nil:
			if err.Error() != wantError {
				fail("%s: both refuse, with different messages:\n  library %s\n  Go      %s", name, wantError, err.Error())
				continue
			}
			refusedByBoth++
			if !isFixture {
				corpusRefused++
			}
		case oracleRefused:
			fail("%s: the library throws (%s) on %q and Go parsed it", name, wantError, params)
		case err != nil:
			fail("%s: the library parses %q and Go refused: %v", name, params, err)
		default:
			if difference := firstMediaQueryDifference(want["tree"], got, "tree"); difference != "" {
				fail("%s (%q): %s", name, params, difference)
				continue
			}
			identical++
		}
	}
	t.Logf("%d compared (%d params from %d corpus files, %d fixtures): %d identical, %d refused by both with the same message, %d different; %d corpus params refused",
		len(records), corpusParams, corpusFiles, fixturesSeen, identical, refusedByBoth, different, corpusRefused)
	if fixturesSeen != len(mediaQueryFixtures) {
		t.Errorf("the oracle answered %d of %d fixtures", fixturesSeen, len(mediaQueryFixtures))
	}
	if corpusParams == 0 {
		t.Errorf("the corpus reached parseMediaQuery with no params; the comparison rests on fixtures alone")
	}
	if corpusRefused > corpusParams/10 {
		t.Errorf("the library refused %d of %d corpus params; the comparison is thinner than it looks", corpusRefused, corpusParams)
	}
}

// dumpMediaQuery parses params and renders the tree the way the oracle script does, through JSON, so
// both sides compare as the same Go values.
func dumpMediaQuery(params string) (any, error) {
	tree, err := Parse(params)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(dumpMediaQueryValue(tree, utf16OffsetsOf(params)))
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(encoded, &result)
	return result, err
}

func dumpMediaQueryValue(value any, offsets []int) any {
	switch typed := value.(type) {
	case nil:
		return "<undefined>"
	case *estree.Node:
		result := map[string]any{"type": typed.Type()}
		if typed.Type() == "" {
			result["type"] = "<undefined>"
		}
		if typed.Range != [2]int{0, 0} {
			result["range"] = "a range other than [0, 0], which the library has no use for"
		}
		for _, key := range typed.Keys() {
			result[key] = dumpMediaQueryValue(typed.Get(key), offsets)
		}
		return result
	case []*estree.Node:
		if typed == nil {
			return "a nil list, which no library field holds"
		}
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = dumpMediaQueryValue(child, offsets)
		}
		return result
	case int:
		// Only sourceIndex is an int: a byte offset here, a UTF-16 index there.
		if typed < 0 || typed >= len(offsets) {
			return fmt.Sprintf("an offset %d outside the input", typed)
		}
		return offsets[typed]
	case string, bool:
		return typed
	default:
		return fmt.Sprintf("an unexpected %T", typed)
	}
}

// utf16OffsetsOf maps every byte offset that starts a character (and the end) to its UTF-16 index.
func utf16OffsetsOf(text string) []int {
	offsets := make([]int, len(text)+1)
	index := 0
	for position, character := range text {
		size := utf8.RuneLen(character)
		if character == utf8.RuneError {
			size = 1
		}
		for byteIndex := position; byteIndex < position+size; byteIndex++ {
			offsets[byteIndex] = index
		}
		index += len(utf16.Encode([]rune{character}))
	}
	offsets[len(text)] = index
	return offsets
}

// firstMediaQueryDifference names the first path where two decoded JSON values differ, or "".
func firstMediaQueryDifference(want any, got any, path string) string {
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
			if difference := firstMediaQueryDifference(wantValue, gotValue, path+"."+key); difference != "" {
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
			if difference := firstMediaQueryDifference(typed[index], other[index], fmt.Sprintf("%s[%d]", path, index)); difference != "" {
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
