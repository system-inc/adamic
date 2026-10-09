package postcss

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
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

// postcssParseScript is the oracle: postcss 8.5.16 from the fork's node_modules, called the way
// Prettier's parser-postcss.js calls it, postcssParse.default(text, { map: false }), dumped in the shape
// dumpParse prints. Every node dumps its type, every own enumerable field but parent (source without
// input), with sorted keys, plus "keys", the field order Object.keys gives, so creation order is compared
// too. Offsets are UTF-16 there; the Go side converts. A refusal dumps the CssSyntaxError's name, reason,
// positions and input offsets.
const postcssParseScript = `
import parse from "postcss/lib/parse";
function dump(v) {
  if (v === undefined) return "<undefined>";
  if (v === null) return null;
  if (Array.isArray(v)) return v.map(dump);
  if (typeof v !== "object") return v;
  const o = {};
  if (typeof v.type === "string") o.keys = Object.keys(v).filter((k) => k !== "parent" && k !== "type");
  for (const k of Object.keys(v).sort()) {
    if (k === "parent") continue;
    if (k === "source") { o.source = { start: dump(v.source.start) }; if ("end" in v.source) o.source.end = dump(v.source.end); continue; }
    o[k] = dump(v[k]);
  }
  return o;
}
let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", () => {
  for (const line of input.split("\n").filter(Boolean)) {
    const { name, text } = JSON.parse(line);
    try {
      const root = parse.default(text, { map: false });
      process.stdout.write(JSON.stringify({ name, root: dump(root) }) + "\n");
    } catch (e) {
      if (e.name !== "CssSyntaxError") {
        process.stdout.write(JSON.stringify({ name, error: { notCssSyntaxError: String(e) } }) + "\n");
        continue;
      }
      const error = { name: e.name, reason: e.reason, line: e.line, column: e.column, offset: e.input.offset };
      if (e.endLine !== undefined) Object.assign(error, { endLine: e.endLine, endColumn: e.endColumn, endOffset: e.input.endOffset });
      process.stdout.write(JSON.stringify({ name, error }) + "\n");
    }
  }
});
`

// parseFixture is an inline input. malformed says which way postcss must go on it, so an oracle that
// refuses a valid fixture, or accepts a malformed one, fails the test rather than quietly agreeing.
type parseFixture struct {
	name      string
	text      string
	malformed bool
}

var parseFixtures = []parseFixture{
	// Rules and nesting.
	{name: "one rule", text: "a{b:c}"},
	{name: "nested rules", text: ".a {\n  color: red;\n  .b { color: blue }\n  &:hover { x: y; }\n  > li { a: b }\n}\n"},
	{name: "deep nesting with at-rules", text: ".a { @media (x) { .b { @supports (y) { c: d } } } }"},
	{name: "empty rules", text: "a{} b { } c {\n}\n{}\n{ a: b }"},
	{name: "selector lists", text: "a,\nb , c>d~e+f, [x=\"y\"], :is(a, b)::before {}"},
	{name: "selector with brackets and parens", text: "a[title=\"x}y\"], a:not([x]) , a:nth-child(2n+1){b:c}"},
	{name: "hacks", text: "a{_b:c;*d:e}"},
	{name: "progid", text: "a{filter:progid:DXImageTransform.Microsoft.gradient(startColorstr='#1', endColorstr='#2')}"},

	// At-rules.
	{name: "import", text: "@import url(foo.css);\n@import 'a.css' screen;\n@import \"b.css\" layer(x) supports(display: grid);"},
	{name: "charset", text: "@charset \"utf-8\";"},
	{name: "media", text: "@media (min-width: 100px) and (max-width:200px) { a { b: c } }\n@media screen{}\n@media print,screen{a{b:c}}"},
	{name: "font-face", text: "@font-face{font-family:x;src:url(a.woff2) format('woff2'),url(\"b.woff\") format(\"woff\");font-display:swap}"},
	{name: "layer statements and blocks", text: "@layer base, components, utilities;\n@layer base { html { a: b } }\n@layer { }"},
	{name: "tailwind at-rules", text: "@import \"tailwindcss\";\n@theme inline {\n  --color-x: red;\n  --font-*: initial;\n}\n@custom-variant dark (&:where(.dark, .dark *));\n@utility tab-4 { tab-size: 4; }\n.a { @apply flex items-center; }\n@source \"../node_modules\";"},
	{name: "keyframes", text: "@keyframes spin { from { transform: rotate(0deg) } 50.5% { opacity: .5 } to { transform: rotate(360deg) } }"},
	{name: "supports", text: "@supports (display:grid) and (not (display:inline-grid)){a{b:c}}"},
	{name: "at-rule closed by brace", text: "a { @apply x y }\nb { @apply z   }\nc { @foo /* c */ }"},
	{name: "at-rule at end of file", text: "@foo bar"},
	{name: "bare at-rule at end of file", text: "@foo"},
	{name: "at-rule with trailing spaces at end of file", text: "@foo bar   \n"},
	{name: "at-rule name then spaces at end of file", text: "@foo   "},
	{name: "at-rule with comment params", text: "@foo /* c */;\n@foo bar /* c */ ;\n@foo/* c */bar;"},
	{name: "at-rule params with brackets", text: "@foo [a;b] (c{d}) ;\n@bar (a;{) {}"},
	{name: "at-rule params with braces in brackets", text: "@foo ({ x }) { a: b }"},
	{name: "at-rule then a free semicolon", text: "@foo;;"},
	{name: "at-rule directly followed", text: "@media(x){}@media{}"},
	{name: "at-word ended by each delimiter", text: "@a\"b\";@c'd';@e#f;@g/h;@i\\j;@k[l];"},

	// Declarations and !important.
	{name: "important", text: "a{b:c!important}"},
	{name: "important with spaces", text: "a{b:c ! important}"},
	{name: "important uppercase", text: "a{b:c !IMPORTANT}"},
	{name: "important then semicolon", text: "a{b:c! important ;}"},
	{name: "important then comment", text: "a{b:c !important /* x */}"},
	{name: "comment then important", text: "a{b:c /* x */ !important}"},
	{name: "important then spaces", text: "a{b:c!important  }"},
	{name: "important alone", text: "a{b: !important}"},
	{name: "important without bang", text: "a{b:c important}"},
	{name: "important split by a comment", text: "a{b:c ! /* x */ important}"},
	{name: "important twice", text: "a{b:c !important !important}"},
	{name: "important with tabs and newlines", text: "a{b:c\t!\n\timportant\n}"},
	{name: "important on a custom property", text: "a{--x: y !important}"},
	{name: "between with spaces and comments", text: "a{b /* x */ : /* y */ c}"},
	{name: "value spacing", text: "a{  b  :  c  d  ;  e:f  }"},
	{name: "colon in value after another word", text: "a{b:c url(x:y)}"},

	// Comments everywhere.
	{name: "comments everywhere", text: "/* a */ a /* b */ , c /* c */ { /* d */ b /* e */ : /* f */ c /* g */ ; /* h */ } /* i */"},
	{name: "comment inside a value", text: "a{b:c/* x */d; e:f, /* y */ g; h:i /* z */; j:/* k */l}"},
	{name: "comment with one safe neighbor", text: "a /* x */b, c/* y */ d{e:f /* z */g; h:i/* w */ j}"},
	{name: "comments in selectors", text: "a/* x */,b{} a /* x */{} a,/* x */b{} a/* x */b{}"},
	{name: "empty and blank comments", text: "/**/ /*   */ /*\n\n*/"},
	{name: "comment padding", text: "/*  text  */ /*\ttext\n*/ /* \xc2\xa0text\xe3\x80\x80 */ /*\xef\xbb\xbfx*/"},
	{name: "comment only", text: "/* only */"},
	{name: "comment in at-rule params", text: "@media /* a */ screen /* b */ and (x) /* c */ {}"},
	{name: "comment star slash adjacent", text: "a{b:c/**/}/**//**/"},

	// Strings.
	{name: "strings with escapes", text: "a{content:\"a\\\"b\";b:'c\\'d';c:\"\\\\\";d:\"\\\\\\\"\";}"},
	{name: "string with escaped newline", text: "a{b:\"multi\\\nline\"}"},
	{name: "strings with braces", text: "a{content:\"}\"; b: '{'; c: \";\"}"},

	// url() and parentheses.
	{name: "url unquoted", text: "a{b:url(foo.png)}"},
	{name: "url with spaces", text: "a{b:url( foo.png )}"},
	{name: "url single quoted", text: "a{b:url('a b.png')}"},
	{name: "url double quoted", text: "a{b:url(\"x\")}"},
	{name: "url data", text: "a{b:url(data:image/svg+xml;charset=utf8,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%3E)}"},
	{name: "url with escaped paren", text: "a{b:url(a\\)b)}"},
	{name: "url empty", text: "a{b:url()}"},
	{name: "url with newline", text: "a { b: url(foo\n) }"},
	{name: "url uppercase", text: "a{b:URL(x;y)}"},
	{name: "url after a word elsewhere", text: "a{b:url x (y)}"},
	{name: "bad paren content", text: "a{b:x(a\n);c:y(\"q\");d:z(a/b);e:w(a\\b)}"},
	{name: "line terminator before a bad character", text: "a{b:x(\xe2\x80\xa8\n)}"},
	// Whether the first parenthesis is one brackets token or loose tokens decides whether "url" reaches
	// the tokenizer's buffer, so which of two errors the second, unclosed one raises.
	{name: "line terminator then url", text: "a{b:x(\xe2\x80\xa8\n\nurl) (y}", malformed: true},
	{name: "nested parens", text: "a{b:calc(1px + (2px * var(--x, 3px)))}"},
	{name: "square brackets", text: "a{grid-template-areas: [a] \"b\"; grid-template-columns: [full-start] 1fr [full-end]}"},

	// Custom properties.
	{name: "custom property json-like", text: "a{--x:{a:b}}"},
	{name: "custom property json", text: "a{--x: {\"a\": [1, 2], \"b\": {\"c\": null}}; --y: [1,{\"a\":2}]}"},
	{name: "custom property empty", text: "a{--empty:;}"},
	{name: "custom property space", text: "a{--empty: ;}"},
	{name: "custom property trailing spaces", text: "a{--x:   }"},
	{name: "custom property value trailing spaces", text: "a{--x: 1px  }"},
	{name: "custom property comment", text: "a{--x:a/* c */b; --y: /* only */; --z:a /* c */}"},
	{name: "custom property empty block", text: ":root{--a:{}}"},
	{name: "custom property block with semicolons", text: ":root{--a:{b:c;d:e};--f:g}"},
	{name: "custom property at end of file", text: "--x: y  "},

	// Semicolons.
	{name: "semicolon missing at block end", text: "a{b:c;d:e}"},
	{name: "doubled semicolons", text: "a{b:c;;d:e;;}"},
	{name: "leading semicolon", text: "a{;b:c}"},
	{name: "free semicolon after rule", text: "a{};b{}  ;  c{} ; ; d{}"},
	{name: "free semicolon past the end", text: "a{}  ;"},
	{name: "free semicolon before non-ASCII", text: "a{}   ;\xc3\xa9{}"},
	{name: "semicolon after at-rule block", text: "@media x{};"},
	{name: "declaration at end of file", text: "a: b"},
	{name: "declaration at end of file with trailing space", text: "a: b  \n"},

	// Byte order marks, line endings, non-ASCII.
	{name: "byte order mark", text: "\xef\xbb\xbfa{b:c}"},
	{name: "reversed byte order mark", text: "\xef\xbf\xbea{}"},
	{name: "byte order mark later", text: "a{}\xef\xbb\xbfb{}"},
	{name: "CRLF", text: "a {\r\n  b: c;\r\n}\r\n@media x {\r\n  d { e: f }\r\n}\r\n"},
	{name: "CR only", text: "a{\rb:c\r}\r"},
	{name: "form feed", text: "a{\fb:c\f}"},
	{name: "non-ASCII", text: ".\xc3\xa9 { content: \"\xe6\x97\xa5\xe6\x9c\xac\xe8\xaa\x9e \xf0\x9f\x98\x80\"; --x: \xf0\x9f\x98\x80; } /* \xe2\x9c\x93 */ \xf0\x9f\x98\x80 { a: b }"},
	{name: "non-ASCII positions across lines", text: "a{b:c}\n\xf0\x9f\x98\x80{x:y}\n@media \xf0\x9f\x98\x80{}\n\xf0\x9f\x98\x80\xf0\x9f\x98\x80 { z: \xf0\x9f\x98\x80 }"},
	{name: "escapes", text: "a\\:b{c:d} .\\31 0{} a{b:\\\\} .a\\\\b{} .x\\{{}"},
	{name: "escaped astral character", text: ".a\\\xf0\x9f\x98\x80{b:c}"},
	{name: "hex escape then space", text: ".\\E9 x{}"},

	// Odd shapes postcss accepts.
	{name: "colon first in a block", text: "a{:b}"},
	{name: "string first in a block", text: "a{\"x\":b}"},
	{name: "brackets first", text: "a{(x):b}"},
	{name: "progid before a colon", text: "a{b:progid:c}"},
	{name: "semicolon inside brackets at top", text: "a(b;c){}"},
	{name: "custom property with brace before colon", text: "--x{a:b}"},
	{name: "important inside a custom property block", text: "a{--x:{b:c !important}}"},
	{name: "important lookalikes", text: "a{b:important;c:!important;d:x!imp;e:x !important!important}"},
	{name: "between with a symbol word", text: "a{b + :c}"},

	// Empty input.
	{name: "empty", text: ""},
	{name: "whitespace only", text: "  \n\t "},

	// Malformed: each must be refused by both sides with the same error.
	{name: "unclosed block", text: "a{", malformed: true},
	{name: "unclosed block after a declaration", text: "a{b:c", malformed: true},
	{name: "unclosed nested block", text: "a{b{c:d}", malformed: true},
	{name: "unclosed block on a later line", text: "\xf0\x9f\x98\x80 {\r\n x { y: z }", malformed: true},
	{name: "unexpected close", text: "}", malformed: true},
	{name: "unexpected close later", text: "a{b:c}\n}", malformed: true},
	{name: "unclosed string", text: "a{b:\"c}", malformed: true},
	{name: "unclosed single quoted string", text: "a{b:'c\\'}", malformed: true},
	{name: "unclosed comment", text: "/* x", malformed: true},
	{name: "unclosed url", text: "a{b:url(x}", malformed: true},
	{name: "unclosed paren", text: "a{b:c(d}", malformed: true},
	{name: "unclosed paren with semicolon", text: "a{b:(c;}", malformed: true},
	{name: "unclosed square bracket", text: "a{b:[c}", malformed: true},
	{name: "unknown word at top", text: "a b", malformed: true},
	{name: "unknown word in block", text: "a{b c}", malformed: true},
	{name: "unknown word before colon", text: "a{b c:d}", malformed: true},
	{name: "unknown word non-ASCII", text: "a{\xf0\x9f\x98\x80\xf0\x9f\x98\x80 x}", malformed: true},
	{name: "double colon", text: "a{b::c}", malformed: true},
	{name: "missed semicolon", text: "a{b:c d:e}", malformed: true},
	{name: "missed semicolon after non-ASCII", text: "a{b:\xf0\x9f\x98\x80 c:d}", malformed: true},
	{name: "missed semicolon after a string", text: "a{b:\"x\" c:d}", malformed: true},
	{name: "at-rule without name", text: "@{}", malformed: true},
	{name: "at-rule without name before space", text: "a{} @ b{}", malformed: true},
	{name: "semicolon with no colon", text: "a{b;}", malformed: true},
	{name: "backslash at end", text: "a{b:c}\\", malformed: true},
	{name: "word then brace at top", text: "a b}", malformed: true},
	{name: "unclosed at-rule block", text: "@media x{a{b:c}", malformed: true},
	{name: "unclosed bracket in a custom property block", text: "a{--x:{a:b}", malformed: true},
	{name: "important word with a word character between", text: "a{b x:c}", malformed: true},
}

// TestParseAgreesWithPostcss is the parser's acceptance test: for every .css file in the corpora, and
// every inline fixture, the Go tree must equal postcss's, field for field and in creation order, and a
// refusal must be a refusal on both sides with the same CssSyntaxError. Off unless COHERE_PRETTIER_FORK
// names the fork (whose node_modules hold postcss 8.5.16) and COHERE_CSS_CORPORA lists files or
// directories to walk (colon-separated, node_modules and .git skipped).
//
// The brief's 26 files are the theme styles and two more:
//
//	COHERE_CSS_CORPORA=~/Projects/ahra/app/_theme/styles:~/Projects/ahra/libraries/structure/source/theme/styles:~/Projects/ahra/projects/www-ahra-ai/app/_theme/styles:~/Projects/ahra/projects/www-ahra-ai/libraries/structure/source/theme/styles:~/Projects/phi/www-phi-health/app/_theme/styles:~/Projects/phi/www-phi-health/libraries/structure/source/theme/styles:~/Projects/ahra/app/(os-layout)/_components/kingdom/OsKingdomGraph.css:~/Projects/phi/www-phi-health/public/fonts/inter/inter.css
//
// A corpus file postcss cannot parse is counted, and the test fails if that count passes a tenth of the
// corpus: an oracle that rejects everything would otherwise read as perfect agreement.
func TestParseAgreesWithPostcss(t *testing.T) {
	t.Parallel()
	corpora := os.Getenv("COHERE_CSS_CORPORA")
	if corpora == "" {
		t.Skip("set COHERE_CSS_CORPORA to compare the parser against postcss")
	}
	root := corpus.PrettierFork.Root(t)

	type input struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	var inputs []input
	malformed := map[string]bool{}
	corpusFiles := 0
	seen := map[string]bool{}
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
			if entry.IsDir() && (entry.Name() == "node_modules" || entry.Name() == ".git") {
				return filepath.SkipDir
			}
			if entry.IsDir() || filepath.Ext(path) != ".css" || seen[path] {
				return nil
			}
			seen[path] = true
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if !utf8.Valid(source) {
				return fmt.Errorf("%s is not UTF-8, which the comparison cannot carry through JSON", path)
			}
			inputs = append(inputs, input{Name: path, Text: string(source)})
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
	for _, fixture := range parseFixtures {
		name := "fixture: " + fixture.name
		inputs = append(inputs, input{Name: name, Text: fixture.text})
		malformed[name] = fixture.malformed
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
	command := exec.Command("node", "--input-type=module", "-e", postcssParseScript)
	command.Dir = root
	command.Stdin = &stdin
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	oracle := map[string]map[string]any{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		oracle[record["name"].(string)] = record
	}

	identical, refusedByBoth, corpusRefused, different := 0, 0, 0, 0
	fail := func(format string, arguments ...any) {
		different++
		if different <= 20 {
			t.Errorf(format, arguments...)
		}
	}
	for _, each := range inputs {
		want, present := oracle[each.Name]
		if !present {
			t.Fatalf("the oracle printed nothing for %s", each.Name)
		}
		wantError, oracleRefused := want["error"]
		if expectMalformed, isFixture := malformed[each.Name]; isFixture && expectMalformed != oracleRefused {
			fail("%s: the fixture says malformed=%v and postcss says %v", each.Name, expectMalformed, want)
			continue
		}
		got, gotError := dumpParse(each.Text)
		switch {
		case oracleRefused && gotError != nil:
			if difference := firstParseDifference(wantError, gotError, "error"); difference != "" {
				fail("%s: both refuse, differently: %s", each.Name, difference)
				continue
			}
			refusedByBoth++
			if _, isFixture := malformed[each.Name]; !isFixture {
				corpusRefused++
			}
		case oracleRefused:
			fail("%s: postcss refuses it (%v) and Go parsed it", each.Name, wantError)
		case gotError != nil:
			fail("%s: postcss parses it and Go refused: %v", each.Name, gotError)
		default:
			if difference := firstParseDifference(want["root"], got, "root"); difference != "" {
				fail("%s: %s", each.Name, difference)
				continue
			}
			identical++
		}
	}
	t.Logf("%d compared (%d corpus files, %d fixtures): %d identical, %d refused by both with the same error, %d different; %d corpus files refused",
		len(inputs), corpusFiles, len(parseFixtures), identical, refusedByBoth, different, corpusRefused)
	if corpusRefused > corpusFiles/10 {
		t.Errorf("postcss refused %d of %d corpus files; the comparison is thinner than it looks", corpusRefused, corpusFiles)
	}
}

// dumpParse parses one text and renders it the way the oracle script does, through JSON, so both sides
// compare as the same Go values: the tree, or a non-nil error dump.
func dumpParse(text string) (any, any) {
	return dumpParseWith(Parse, text)
}

// dumpParseWith is dumpParse with the parser named, Parse or ParseSCSS.
func dumpParseWith(parse func(string) (*estree.Node, error), text string) (any, any) {
	offsets := utf16OffsetsOf(stripByteOrderMark(text))
	root, err := parse(text)
	if err != nil {
		var syntaxError *CssSyntaxError
		if !errors.As(err, &syntaxError) {
			return nil, map[string]any{"notCssSyntaxError": err.Error()}
		}
		dumped := map[string]any{
			"name":   syntaxError.Name(),
			"reason": syntaxError.Reason,
			"line":   syntaxError.Line,
			"column": syntaxError.Column,
			"offset": offsets(syntaxError.Offset),
		}
		if syntaxError.HasEnd {
			dumped["endLine"] = syntaxError.EndLine
			dumped["endColumn"] = syntaxError.EndColumn
			dumped["endOffset"] = offsets(syntaxError.EndOffset)
		}
		return nil, throughJSON(dumped)
	}
	return throughJSON(dumpParseValue(root, offsets)), nil
}

func throughJSON(value any) any {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("not JSON: %v", err)
	}
	var result any
	if err := json.Unmarshal(encoded, &result); err != nil {
		return fmt.Sprintf("not JSON: %v", err)
	}
	return result
}

func dumpParseValue(value any, offsets func(int) int) any {
	switch typed := value.(type) {
	case nil:
		return "<undefined>"
	case *estree.Node:
		result := map[string]any{"type": typed.Type()}
		keys := []string{}
		for _, key := range typed.Keys() {
			keys = append(keys, key)
			result[key] = dumpParseValue(typed.Get(key), offsets)
		}
		result["keys"] = keys
		if source, isMap := typed.Get("source").(map[string]any); isMap {
			start, _ := source["start"].(map[string]any)
			end, hasEnd := source["end"].(map[string]any)
			wantRange := [2]int{0, 0}
			if start != nil && hasEnd {
				wantRange = [2]int{start["offset"].(int), end["offset"].(int)}
			}
			if typed.Range != wantRange {
				result["range"] = fmt.Sprintf("Range %v is not source's %v", typed.Range, wantRange)
			}
		}
		return result
	case []*estree.Node:
		if typed == nil {
			return "a nil list, which no postcss field holds"
		}
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = dumpParseValue(child, offsets)
		}
		return result
	case map[string]any:
		result := map[string]any{}
		for key, each := range typed {
			if key == "offset" {
				if offset, isInt := each.(int); isInt {
					result[key] = offsets(offset)
					continue
				}
			}
			result[key] = dumpParseValue(each, offsets)
		}
		return result
	case string, bool, int:
		return typed
	default:
		return fmt.Sprintf("an unexpected %T", typed)
	}
}

// stripByteOrderMark drops what Input drops, so offsets convert against the text postcss parsed.
func stripByteOrderMark(text string) string {
	if strings.HasPrefix(text, "\xef\xbb\xbf") || strings.HasPrefix(text, "\xef\xbf\xbe") {
		return text[3:]
	}
	return text
}

// utf16OffsetsOf converts a byte offset into the text to its UTF-16 index, an offset past the end
// keeping its distance past the end.
func utf16OffsetsOf(text string) func(int) int {
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
		if character > 0xFFFF {
			index += 2
		} else {
			index++
		}
	}
	offsets[len(text)] = index
	return func(offset int) int {
		if offset > len(text) {
			return index + offset - len(text)
		}
		if offset < 0 {
			return offset
		}
		return offsets[offset]
	}
}

// firstParseDifference names the first path where two decoded JSON values differ, or "".
func firstParseDifference(want any, got any, path string) string {
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
				return fmt.Sprintf("%s.%s: present in postcss %v (%v), in Go %v (%v)", path, key, wantPresent, wantValue, gotPresent, gotValue)
			}
			if difference := firstParseDifference(wantValue, gotValue, path+"."+key); difference != "" {
				return difference
			}
		}
		return ""
	case []any:
		other, isList := got.([]any)
		if !isList || len(other) != len(typed) {
			return fmt.Sprintf("%s: want %d elements %v, got %v", path, len(typed), typed, got)
		}
		for index := range typed {
			if difference := firstParseDifference(typed[index], other[index], fmt.Sprintf("%s[%d]", path, index)); difference != "" {
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
