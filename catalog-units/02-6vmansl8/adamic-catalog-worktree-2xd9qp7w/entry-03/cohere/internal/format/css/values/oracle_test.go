package values

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
	"github.com/system-inc/cohere/internal/format/estree"
)

// valuesParseScript is the oracle: postcss-values-parser 2.0.1 from the fork's node_modules, loaded
// the way Prettier's parse-value.js loads it (lib/parser.js), run with { loose: true } as Prettier
// runs it and with { loose: false }, the library's default. Corpus files are read with postcss 8 and
// every value parser-postcss.js would hand to parseValue under the css parser is extracted the way it
// extracts it: declaration values (raws.value.raw, cut before !default and !global, progid: skipped,
// custom properties holding a {} block reparsed as rules), and the params of at-root, import, use,
// forward, namespace, supports and the other at-rules it routes there. Every other at-rule's params are
// added as "extra" inputs, which Prettier does not parse as values but which exercise the tokenizer.
//
// A node dumps as its own enumerable fields minus parent, sorted; NaN dumps as "NaN" and undefined as
// "<undefined>". A throw dumps as String(error).
const valuesParseScript = `
const { createRequire } = require("node:module");
const fs = require("node:fs");
const path = require("node:path");
const requireFromFork = createRequire(path.join(process.cwd(), "package.json"));
const Parser = requireFromFork("postcss-values-parser/lib/parser.js");
const postcssParse = requireFromFork("postcss/lib/parse");

function dump(v) {
  if (v === undefined) return "<undefined>";
  if (v === null) return null;
  if (typeof v === "number" && Number.isNaN(v)) return "NaN";
  if (Array.isArray(v)) return v.map(dump);
  if (typeof v !== "object") return v;
  const o = {};
  for (const k of Object.keys(v).sort()) { if (k === "parent") continue; o[k] = dump(v[k]); }
  return o;
}

function run(text, loose) {
  try {
    return { tree: dump(new Parser(text, { loose }).parse()) };
  } catch (e) {
    return { error: String(e) };
  }
}

const moduleRuleNames = new Set(["import", "use", "forward"]);
const valueAtRules = ["namespace", "supports", "if", "else", "for", "each", "while", "debug", "mixin", "include", "function", "return", "define-mixin", "add-mixin"];

function extract(root, originalText, out) {
  root.walk((node) => {
    if (node.type === "decl") {
      if (typeof node.prop === "string" && node.prop.startsWith("--") && typeof node.value === "string" && node.value.startsWith("{")) {
        if (node.value.trimEnd().endsWith("}")) {
          const textBefore = originalText.slice(0, node.source.start.offset).replace(/[^\n]/g, " ");
          const nodeText = "a".repeat(node.prop.length) + originalText.slice(node.source.start.offset + node.prop.length, node.source.end.offset);
          let ast;
          try { ast = postcssParse(textBefore + nodeText); } catch {}
          if (ast?.nodes?.length === 1 && ast.nodes[0].type === "rule") extract(ast.nodes[0], textBefore + nodeText, out);
        }
        return;
      }
      let value = node.raws.value ? (node.raws.value.scss ?? node.raws.value.raw) : node.value;
      if (value.trim().length > 0) {
        const defaultDirective = value.match(/(\s*)(!default).*$/);
        if (defaultDirective) value = value.slice(0, defaultDirective.index);
        const globalDirective = value.match(/(\s*)(!global).*$/);
        if (globalDirective) value = value.slice(0, globalDirective.index);
        if (value.startsWith("progid:")) return;
        out.push({ source: "decl", text: value });
      }
      return;
    }
    if (node.type !== "atrule") return;
    let params = node.raws.params ? (node.raws.params.scss ?? node.raws.params.raw) : node.params;
    if (node.raws.afterName && node.raws.afterName.trim().length > 0) params = node.raws.afterName + params;
    if (node.raws.between && node.raws.between.trim().length > 0) params += node.raws.between;
    params = params.trim();
    if (node.name === "custom-selector" || params.length === 0) return;
    const name = node.name;
    const lowercasedName = name.toLowerCase();
    if (name === "warn" || name === "error" || name === "extend" || name === "nest") return;
    if (name === "at-root") {
      if (/^\(\s*(?:without|with)\s*:.+\)$/s.test(params)) out.push({ source: "atrule", text: params });
      return;
    }
    if (moduleRuleNames.has(lowercasedName)) { out.push({ source: "atrule", text: params }); return; }
    if (valueAtRules.includes(name)) {
      params = params.replace(/(\$\S+?)(\s+)?\.{3}/, "$1...$2");
      params = params.replace(/^(?!if)([^"'\s(]+)(\s+)\(/, "$1($2");
      out.push({ source: "atrule", text: params });
      return;
    }
    out.push({ source: "extra", text: params });
  });
}

let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", () => {
  for (const line of input.split("\n").filter(Boolean)) {
    const request = JSON.parse(line);
    let items;
    if (request.path !== undefined) {
      const text = fs.readFileSync(request.path, "utf8");
      items = [];
      extract(postcssParse(text), text, items);
    } else {
      items = [{ source: "fixture", text: request.text }];
    }
    const seen = new Set();
    for (const item of items) {
      if (seen.has(item.text)) continue;
      seen.add(item.text);
      process.stdout.write(JSON.stringify({ name: request.name, source: item.source, text: item.text, loose: run(item.text, true), strict: run(item.text, false) }) + "\n");
    }
    process.stdout.write(JSON.stringify({ name: request.name, done: true, count: seen.size }) + "\n");
  }
});
`

// valuesFixture is an inline input. malformed says whether the library refuses it with
// { loose: true }, the way Prettier runs it, so an oracle that refuses a valid fixture, or accepts a
// malformed one, fails the test rather than quietly agreeing.
type valuesFixture struct {
	name      string
	text      string
	malformed bool
}

var valuesFixtures = []valuesFixture{
	// Functions.
	{name: "calc with var fallback", text: "calc(100% - var(--x, 2px))"},
	{name: "color-mix", text: "color-mix(in oklab, var(--color-red-500) 50%, transparent)"},
	{name: "nested functions three deep", text: "translate(calc(var(--a) * -1), min(10px, max(1vw, 2em)))"},
	{name: "function after a function", text: "rgb(0 0 0 / 0.5) linear-gradient(to right, #fff 0%, rgba(0, 0, 0, .5) 100%)"},
	{name: "parenthesized group", text: "(min-width: 100px) and (max-width: 200px)"},
	{name: "nested parens in calc", text: "calc((1px + 2px) * (3 - 4))"},
	{name: "empty function", text: "foo()"},
	{name: "function with spaces inside", text: "var( --x , 1px )"},
	{name: "atword then paren", text: "@foo(bar)"},
	{name: "number then paren", text: "5(a)"},

	// Numbers.
	{name: "signed and dotted numbers", text: "-.5em +2 1e3 .5e+3 1e-10 -1E+2px 10.px 0.0 +.5"},
	{name: "numbers with units", text: "10px 1.5rem 100% 3s 45deg 2fr 1x"},
	{name: "number then letters with e", text: "1e1e 2em 3e"},
	{name: "minus number after a word", text: "a -5px"},
	{name: "minus after a minus", text: "a - -5px"},
	{name: "number followed by an at", text: "5@x"},
	{name: "custom property name", text: "--foo"},
	{name: "dash dash number", text: "--5"},
	{name: "vendor prefixed word", text: "-webkit-box -moz-calc(1px)"},
	{name: "lone minus and dashes", text: "- -- ---"},

	// Words and colors.
	{name: "hex colors", text: "#fff #ffff #ffffff #ffffff80 #abcdeg #12345 #FFF"},
	{name: "hash alone", text: "# #- #{a}"},
	{name: "braces", text: "{a} { } }"},
	{name: "backslash", text: "a\\b \\@x \\"},
	{name: "word with form feed inside", text: "a\fb"},
	{name: "important-free keywords", text: "solid red !ie"},
	{name: "atwords", text: "@media @a@b x@y"},

	// Strings and urls.
	{name: "strings with escapes", text: `"a\"b" 'c\'d' "e\\" 'f\\\'g' "multi\
line"`},
	{name: "empty strings", text: `"" ''`},
	{name: "url unquoted", text: "url(foo/bar.png)"},
	{name: "url quoted", text: `url("foo bar.png") url('x')`},
	{name: "url with a double slash", text: "url(https://example.com/a.png)"},
	{name: "url with spaces", text: "url( a.png )"},
	{name: "url empty", text: "url()"},
	{name: "url with nested parens", text: "url(a(b)c)"},
	{name: "url after url", text: "url(a) url(//b)"},
	{name: "inline comment after url", text: "url(a) // trailing"},

	// Separators and operators.
	{name: "commas and slashes", text: "1px/2px, 3px / 4px,5px"},
	{name: "grid area", text: "1 / 2 / span 3"},
	{name: "colons", text: "a:b : c"},
	{name: "operators with spaces", text: "calc(1px + 2px - 3px * 4 / 5)"},
	{name: "operators without spaces", text: "calc(1px+2px-3px*4/5)"},
	{name: "operator after operator", text: "1 + -2 * +3"},
	{name: "operator then word", text: "a +b"},
	{name: "calc minus before number", text: "calc(1 -2)"},
	{name: "calc doubled operators", text: "calc(1 -+2)"},
	{name: "calc leading signed", text: "calc(-0.5 + 2)"},
	{name: "doubled operators outside calc", text: "+-4px 5+ 5"},
	{name: "asterisk", text: "* a*b"},
	{name: "unicode ranges", text: "U+0025-00FF, u+4??, U+0-7F, u+"},
	{name: "u then plus inside word", text: "menu+u+1"},
	{name: "unicode range stops at a letter past f", text: "u+00fg U+0A-0G"},
	{name: "hash before a line separator", text: "#\xe2\x80\xa8a #a\xe2\x80\xa8"},

	// Comments and whitespace.
	{name: "comments inside values", text: "1px /* one */ 2px/* two */,/**/3px"},
	{name: "comment with stars", text: "a /* * / ** */ b /*/ c */"},
	{name: "multi-line comment", text: "a /* one\ntwo\nthree */ b"},
	{name: "inline comment", text: "a // rest of line\nb"},
	{name: "inline comment at end", text: "a //tail"},
	{name: "multi-line value", text: "a,\n  b\n  c(\n    d\n  )"},
	{name: "tabs and carriage returns", text: "a\t\tb\r\nc"},
	{name: "trailing space", text: "a b "},
	{name: "space before close paren", text: "foo(a )"},
	{name: "space before comma", text: "a , b"},

	// Text that is not ASCII.
	{name: "non-ASCII words", text: "caf\xc3\xa9 \xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e 1px"},
	{name: "emoji in a string", text: "\"\xf0\x9f\x98\x80\" a"},
	{name: "emoji before an operator", text: "\xf0\x9f\x98\x80 + 1"},
	{name: "non-ASCII before a function", text: "\xc3\xa9 calc(1px + 2px) \xf0\x9f\x98\x80x(a)"},
	{name: "non-ASCII on a later line", text: "\xc3\xa9\n\xf0\x9f\x98\x80 b /* \xc3\xa9\n\xc3\xa9 */ c"},

	// Malformed with { loose: true }: each must be refused by both sides with the same message.
	{name: "only whitespace", text: " ", malformed: true},
	{name: "lone plus", text: "+", malformed: true},
	{name: "trailing doubled operator", text: "a + +", malformed: true},
	{name: "unclosed function", text: "a(b", malformed: true},
	{name: "stray close paren", text: "a) b", malformed: true},
	{name: "unclosed double quote", text: "\"abc", malformed: true},
	{name: "unclosed single quote", text: "'abc", malformed: true},
	{name: "escaped closing quote", text: "\"abc\\\"", malformed: true},
	{name: "unclosed comment", text: "a /* x", malformed: true},
	{name: "unclosed comment on a later line", text: "a\nb /* x", malformed: true},
	{name: "space before trailing close paren", text: " )", malformed: true},
	{name: "leading space then comma", text: " ,a"},
}

// TestParseAgreesWithPostcssValuesParser is the parser's acceptance test: for every value Prettier
// would hand the library from the CSS corpus, every extra at-rule param, and every inline fixture, the
// Go tree must equal postcss-values-parser's, field for field, under both { loose: true } and
// { loose: false }, and a refusal must be a refusal on both sides with the same message. Off unless
// COHERE_PRETTIER_FORK names the fork (whose node_modules hold postcss-values-parser 2.0.1 and postcss)
// and COHERE_CSS_CORPORA lists .css files or directories to walk (colon-separated, node_modules
// skipped).
//
// A corpus value the library refuses under { loose: true } is counted, and the test fails if that
// count passes a tenth of the corpus values: an oracle that rejects everything would otherwise read as
// perfect agreement.
func TestParseAgreesWithPostcssValuesParser(t *testing.T) {
	t.Parallel()
	corpora := os.Getenv("COHERE_CSS_CORPORA")
	if corpora == "" {
		t.Skip("set COHERE_CSS_CORPORA to compare the parser against postcss-values-parser")
	}
	root := corpus.PrettierFork.Root(t)

	type request struct {
		Name string  `json:"name"`
		Path *string `json:"path,omitempty"`
		Text *string `json:"text,omitempty"`
	}
	var requests []request
	malformed := map[string]bool{}
	for _, corpus := range strings.Split(corpora, ":") {
		corpus = strings.TrimSpace(corpus)
		if strings.HasPrefix(corpus, "~/") {
			home, _ := os.UserHomeDir()
			corpus = filepath.Join(home, corpus[2:])
		}
		err := filepath.WalkDir(corpus, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".git" || entry.Name() == ".next") {
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
			requests = append(requests, request{Name: path, Path: &path})
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", corpus, err)
		}
	}
	corpusFiles := len(requests)
	if corpusFiles == 0 {
		t.Fatalf("no .css files under %s", corpora)
	}
	for index := range valuesFixtures {
		name := "fixture: " + valuesFixtures[index].name
		requests = append(requests, request{Name: name, Text: &valuesFixtures[index].text})
		malformed[name] = valuesFixtures[index].malformed
	}

	var stdin bytes.Buffer
	for _, each := range requests {
		encoded, err := json.Marshal(each)
		if err != nil {
			t.Fatal(err)
		}
		stdin.Write(encoded)
		stdin.WriteByte('\n')
	}
	command := exec.Command("node", "-e", valuesParseScript)
	command.Dir = root
	command.Stdin = &stdin
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}

	type record struct {
		Name   string         `json:"name"`
		Source string         `json:"source"`
		Text   string         `json:"text"`
		Loose  map[string]any `json:"loose"`
		Strict map[string]any `json:"strict"`
		Done   bool           `json:"done"`
	}
	var records []record
	answered := map[string]bool{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var each record
		if err := json.Unmarshal([]byte(line), &each); err != nil {
			t.Fatal(err)
		}
		if each.Done {
			answered[each.Name] = true
			continue
		}
		records = append(records, each)
	}
	for _, each := range requests {
		if !answered[each.Name] {
			t.Fatalf("the oracle did not finish %s", each.Name)
		}
	}

	compared, identical, refusedByBoth, different := 0, 0, 0, 0
	corpusValues, corpusRefused, extraValues := 0, 0, 0
	fail := func(format string, arguments ...any) {
		different++
		if different <= 25 {
			t.Errorf(format, arguments...)
		}
	}
	for _, each := range records {
		label := fmt.Sprintf("%s (%s) %q", each.Name, each.Source, each.Text)
		switch each.Source {
		case "decl", "atrule":
			corpusValues++
			if _, refused := each.Loose["error"]; refused {
				corpusRefused++
			}
		case "extra":
			extraValues++
		}
		if expectMalformed, isFixture := malformed[each.Name]; isFixture {
			if _, refused := each.Loose["error"]; refused != expectMalformed {
				fail("%s: the fixture says malformed=%v and the library says %v", label, expectMalformed, each.Loose)
				continue
			}
		}
		for _, mode := range []struct {
			name  string
			want  map[string]any
			loose bool
		}{{"loose", each.Loose, true}, {"strict", each.Strict, false}} {
			compared++
			got, err := dumpValues(each.Text, Options{Loose: mode.loose})
			wantError, oracleRefused := mode.want["error"].(string)
			switch {
			case oracleRefused && err != nil:
				var thrown *Error
				if !asError(err, &thrown) || thrown.Error() != wantError {
					fail("%s %s: both refuse, with different messages:\n  library %s\n  Go      %v", label, mode.name, wantError, err)
					continue
				}
				refusedByBoth++
			case oracleRefused:
				fail("%s %s: the library refuses it (%s) and Go parsed it", label, mode.name, wantError)
			case err != nil:
				fail("%s %s: the library parses it and Go refused: %v", label, mode.name, err)
			default:
				if difference := firstValuesDifference(mode.want["tree"], got, "root"); difference != "" {
					fail("%s %s: %s", label, mode.name, difference)
					continue
				}
				identical++
			}
		}
	}
	t.Logf("%d corpus files: %d values Prettier parses (%d refused in loose mode) and %d extra at-rule params; %d fixtures; %d comparisons (loose and strict): %d identical, %d refused by both with the same message, %d different",
		corpusFiles, corpusValues, corpusRefused, extraValues, len(valuesFixtures), compared, identical, refusedByBoth, different)
	if corpusValues == 0 {
		t.Errorf("the corpus yielded no values")
	}
	if corpusRefused > corpusValues/10 {
		t.Errorf("the library refused %d of %d corpus values; the comparison is thinner than it looks", corpusRefused, corpusValues)
	}
}

func asError(err error, target **Error) bool {
	for err != nil {
		if thrown, isThrown := err.(*Error); isThrown {
			*target = thrown
			return true
		}
		unwrapper, isUnwrapper := err.(interface{ Unwrap() error })
		if !isUnwrapper {
			return false
		}
		err = unwrapper.Unwrap()
	}
	return false
}

// dumpValues parses one value and renders it the way the oracle script does, through JSON, with
// sourceIndex converted back to UTF-16.
func dumpValues(text string, options Options) (any, error) {
	tree, err := Parse(text, options)
	if err != nil {
		return nil, err
	}
	toUTF16 := utf16IndexOf(text)
	encoded, err := json.Marshal(dumpValuesValue(tree, toUTF16))
	if err != nil {
		return nil, err
	}
	var result any
	err = json.Unmarshal(encoded, &result)
	return result, err
}

func dumpValuesValue(value any, toUTF16 func(int) int) any {
	switch typed := value.(type) {
	case nil:
		return "<undefined>"
	case *estree.Node:
		if typed.Range != [2]int{0, 0} {
			return fmt.Sprintf("a node with range %v, which the port leaves at [0, 0]", typed.Range)
		}
		result := map[string]any{"type": typed.Type()}
		for _, key := range typed.Keys() {
			if key == "sourceIndex" {
				index, isInt := typed.Get(key).(int)
				if !isInt {
					result[key] = fmt.Sprintf("a sourceIndex of type %T", typed.Get(key))
					continue
				}
				result[key] = toUTF16(index)
				continue
			}
			result[key] = dumpValuesValue(typed.Get(key), toUTF16)
		}
		return result
	case []*estree.Node:
		if typed == nil {
			return "a nil list, which no field of the library holds"
		}
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = dumpValuesValue(child, toUTF16)
		}
		return result
	case map[string]any:
		result := map[string]any{}
		for key, each := range typed {
			result[key] = dumpValuesValue(each, toUTF16)
		}
		return result
	case float64:
		if math.IsNaN(typed) {
			return "NaN"
		}
		return typed
	case string, bool, int:
		return typed
	default:
		return fmt.Sprintf("an unexpected %T", typed)
	}
}

// utf16IndexOf inverts the byte offsets utf16Units assigns, extrapolating one unit per byte past the
// end the way Parser.sourceIndex does.
func utf16IndexOf(text string) func(int) int {
	_, offsets := utf16Units(text)
	byByte := map[int]int{}
	for index, offset := range offsets {
		if _, present := byByte[offset]; !present {
			byByte[offset] = index
		}
	}
	return func(offset int) int {
		if index, present := byByte[offset]; present {
			return index
		}
		if offset > len(text) {
			return len(offsets) - 1 + offset - len(text)
		}
		return -1000000 - offset
	}
}

// firstValuesDifference names the first path where two decoded JSON values differ, or "".
func firstValuesDifference(want any, got any, path string) string {
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
			if difference := firstValuesDifference(wantValue, gotValue, path+"."+key); difference != "" {
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
			if difference := firstValuesDifference(typed[index], other[index], fmt.Sprintf("%s[%d]", path, index)); difference != "" {
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
