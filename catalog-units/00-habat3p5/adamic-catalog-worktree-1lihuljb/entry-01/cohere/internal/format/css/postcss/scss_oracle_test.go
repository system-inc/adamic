package postcss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/corpus"
)

// postcssScssParseScript is the scss oracle: postcss-scss 4.0.9's lib/scss-parse from the fork's
// node_modules, called as Prettier's parseScss calls it, postcssScssParse(text, { map: false }), dumped
// as postcssParseScript dumps postcss's tree. An input line is { name, text } or { name, js }: a js line
// names a JavaScript file whose css-in-js templates the script builds the way Prettier's
// src/language-js/embed/css.js does (each quasi's raw text, joined by "@prettier-placeholder-N-id") and
// parses each, printing the text with the result so the Go side parses the same string. A refusal that
// is not a CssSyntaxError (postcss-scss's TypeError) dumps as notCssSyntaxError, "name: message".
const postcssScssParseScript = `
import scssParse from "postcss-scss/lib/scss-parse";
import { parse as babelParse } from "@babel/parser";
import { readFileSync } from "node:fs";
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
function run(name, text) {
  try {
    const root = scssParse(text, { map: false });
    return { name, text, root: dump(root) };
  } catch (e) {
    if (e.name !== "CssSyntaxError") return { name, text, error: { notCssSyntaxError: e.name + ": " + e.message } };
    const error = { name: e.name, reason: e.reason, line: e.line, column: e.column, offset: e.input.offset };
    if (e.endLine !== undefined) Object.assign(error, { endLine: e.endLine, endColumn: e.endColumn, endOffset: e.input.endOffset });
    return { name, text, error };
  }
}
// The template literals Prettier's embed would hand the scss parser: a tagged template's quasi, or a
// template that is a JSX attribute's or child's expression (css={...}, <style jsx>{...}</style>).
function templates(path) {
  const ast = babelParse(readFileSync(path, "utf8"), { sourceType: "module", plugins: ["jsx", "typescript"], errorRecovery: true });
  const found = [];
  (function walk(node, parent) {
    if (!node || typeof node.type !== "string") return;
    if (node.type === "TemplateLiteral" && parent && (parent.type === "TaggedTemplateExpression" || parent.type === "JSXExpressionContainer")) {
      let text = "";
      node.quasis.forEach((quasi, index) => {
        if (index > 0) text += "@prettier-placeholder-" + (index - 1) + "-id";
        text += quasi.value.raw;
      });
      found.push(text);
    }
    for (const key of Object.keys(node)) {
      if (key === "loc" || key === "start" || key === "end" || key === "extra") continue;
      const value = node[key];
      if (Array.isArray(value)) value.forEach((child) => walk(child, node));
      else if (value && typeof value.type === "string") walk(value, node);
    }
  })(ast.program, null);
  return found;
}
let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", () => {
  for (const line of input.split("\n").filter(Boolean)) {
    const { name, text, js } = JSON.parse(line);
    if (js === undefined) { process.stdout.write(JSON.stringify(run(name, text)) + "\n"); continue; }
    templates(js).forEach((each, index) => process.stdout.write(JSON.stringify(run(name + " #" + index, each)) + "\n"));
  }
});
`

// scssParseFixtures are the scss oracle's inline inputs, with malformed as in parseFixtures.
var scssParseFixtures = []parseFixture{
	// Inline comments in every position.
	{name: "inline comment at top", text: "// top\na{}"},
	{name: "inline comment ending the file", text: "a{}\n// last"},
	{name: "inline comment ending the file after a rule", text: "a{b:c}//x"},
	{name: "inline comment before a declaration", text: "a {\n  // c\n  b: c;\n}"},
	{name: "inline comment after a declaration", text: "a { b: c; // trailing\n}"},
	{name: "inline comment after a value with no semicolon", text: "a { b: c // x\n}"},
	{name: "inline comment inside a value", text: "a { b: c // x\n d; }"},
	{name: "inline comment between words of a value", text: "a{b:c// x\nd}"},
	{name: "inline comment in a selector list", text: "a, // x\nb { }"},
	{name: "inline comment before a brace", text: "a // x\n{ }"},
	{name: "inline comment inside a selector", text: "a// x\nb{}"},
	{name: "inline comment in at-rule params", text: "@include foo( // x\n  $a\n);"},
	{name: "inline comment before an at-rule block", text: "@media screen // x\n{}"},
	{name: "inline comment inside at-rule params", text: "@media screen// x\nand (y){}"},
	{name: "blank inline comments", text: "//\n// \t\na{}"},
	{name: "inline comment holding block comment delimiters", text: "// a /* b */ c\n//*/\n///*\na{}"},
	{name: "inline comment with delimiters in a value", text: "a { b: c /* x */ // y */ z\n; }"},
	{name: "inline comment with delimiters inside a value", text: "a { b: c // /* y */ \nd; }"},
	{name: "slashes then an unclosed comment", text: "a { b: c// /* y\nd; }", malformed: true},
	{name: "inline comment padding", text: "//   padded  \t\n//\xc2\xa0x\xe3\x80\x80\na{}"},
	{name: "inline comment with CRLF", text: "// x\r\na{}\r\n// y\r\n"},
	{name: "inline comment ended by a form feed", text: "// x\fa{}"},
	{name: "inline comment ended by CR", text: "// x\ra{}"},
	{name: "slashes inside a word", text: "a { b: x//y }"},
	{name: "slashes in url", text: "a { b: url(http://x.com/a); c: url( //x) }"},
	{name: "slashes in a string", text: "a { b: 'http://x' }"},
	{name: "inline comment in a custom property", text: "a { --x: a // b\n c; }"},
	{name: "inline comment after an at-rule statement", text: "@import 'a'; // x\n@import 'b' // y\n;"},
	{name: "inline comment in a map", text: "$map: (\n  a: 1, // one\n  b: 2 // two\n);"},
	{name: "inline comment non-ASCII", text: "// \xc3\xa9\xf0\x9f\x98\x80\na { b: c // \xf0\x9f\x98\x80\n}"},

	// Interpolation.
	{name: "interpolation in selectors", text: ".a-#{$b} { c: d }\n#{$sel} > a {}\n.a #{&}__b {}\n#{'.x'} {}"},
	{name: "interpolation in property names", text: "a { #{$prop}: 1px; border-#{$side}: 0; #{$a}-#{$b}: c; }"},
	{name: "interpolation in values", text: "a { width: calc(100% - #{$x}); content: \"#{$a}\"; b: #{$c}px; d: a#{$e}b; }"},
	{name: "interpolation in at-rule params", text: "@media #{$q} and (min-width: #{$w}) {}\n@include x(#{$y});\n@if #{$a} == b {}"},
	{name: "interpolation in at-rule names", text: "@#{$name} foo;\n@media#{$q}{}\n@a#{$b}#{$c} d;\n@a #{$b};"},
	{name: "nested interpolation", text: "a { b: #{ #{$x} } }"},
	{name: "interpolation holding a brace in a string", text: "a { b: #{'}'}; c: #{\"a\\\"}\"}; }"},
	{name: "string holding interpolation with a quote", text: "a { b: \"x #{'\"'} y\" }"},
	{name: "string with escaped quote and interpolation", text: "a { b: 'it\\'s #{$x}' }"},
	{name: "interpolation in url", text: "a { b: url(#{$x}/a.png); c: url(\"#{$y}\") }"},
	{name: "interpolation in a placeholder selector", text: "%#{$x}-y { a: b }"},
	{name: "interpolation in a comment", text: "/* #{$x} */ a{}"},
	{name: "hash without a brace", text: "a { color: #fff; b: c#d; }"},
	{name: "interpolation non-ASCII", text: "a { b: \"#{\xf0\x9f\x98\x80}\" }\n.\xc3\xa9-#{$x} {}"},

	// url() as postcss-scss reads it.
	{name: "url with nested parens", text: "a { b: url(foo(bar)) }"},
	{name: "url with escaped paren", text: "a { b: url(a\\)b) }"},
	{name: "url with spaces and newline", text: "a { b: url( x\n) }"},
	{name: "url empty", text: "a { b: url() }"},
	{name: "url after another word", text: "a{b:url x (y)}"},
	{name: "bad paren content", text: "a{b:x(a\n);c:y(\"q\");d:z(a/b);e:w(a\\b)}"},

	// Variables.
	{name: "variables with flags", text: "$a: 1px;\n$b: red !default;\n$c: 2 !global;\n$d: 3 !default !global;\n.x { $e: 1 !global; $f : 2; }"},
	{name: "maps", text: "$map: (key1: value1, key2: value2, key3: value3);"},
	{name: "multiline nested map", text: "$map: (\n  key1: value1,\n  key2: (a: b, c: d),\n  'key3': value3,\n);"},
	{name: "lists", text: "$list: 1px 2px, 3px;\n$l2: (a, b, c);\n$l3: [a b];"},
	{name: "variable at end of file", text: "$a: b"},
	{name: "variables in values", text: "a { width: $a + $b; height: -$x; c: $a+$b; d: math.div(10px, 2); e: 1 - 2 }"},

	// Nested properties.
	{name: "nested property with no value", text: "a { font: { family: x; size: 1px; } }"},
	{name: "nested property with a value", text: "a { font: 12px/30px { family: x; } }"},
	{name: "nested property margin", text: "a { margin: 0 { left: 1px } }"},
	{name: "nested property with a word value", text: "a { font: bold { family: x } }"},
	{name: "nested property without space", text: "a { font:bold{ family: x } }"},
	{name: "nested property important", text: "a { margin: 0 !important { left: 1px } }"},
	{name: "nested property important with spaces", text: "a { margin: 0 ! important { left: 1px } }"},
	{name: "nested property important alone", text: "a { margin: !important { left: 1px } b: x !important { c: d } }"},
	{name: "nested property important uppercase", text: "a { margin: 0 !IMPORTANT { left: 1px } }"},
	{name: "nested property important word", text: "a { margin: 0 important { left: 1px } b: 0 ! /* c */ important { d: e } }"},
	{name: "nested property hack", text: "a { _margin: 0 { left: 1px } *b: 1 { c: d } }"},
	{name: "nested property with comments", text: "a { margin /* c */ : /* d */ 0 { left: 1px } }"},
	{name: "nested property with an inline comment", text: "a { margin: 0 // c\n { left: 1px } }"},
	{name: "nested property spaced colon", text: "a { b : c { d: e } }"},
	{name: "nested property value on a new line", text: "a { b: c\n { d: e } }"},
	{name: "nested property in parens", text: "a { b: (c) { d: e } }"},
	{name: "nested property at top", text: "b: 1px { c: d }"},
	{name: "nested property at offset zero", text: "b:1{c:d}"},
	{name: "nested property ending in a comma", text: "a { b: c, { d: e } }"},
	{name: "selector broken by a newline before the colon", text: "a,\nb:hover {}\na\n:b {}"},
	{name: "pseudo selectors", text: "a { &:hover { } &:not(.b):hover { } a:nth-child(2n + 1) { } :is(a) {} }"},
	{name: "colon inside parens", text: "a { b(c: d) { } }"},

	// Mixins, functions and control directives.
	{name: "mixin and include", text: "@mixin foo($a, $b: 10px) { width: $a; }\n.x { @include foo(1px, $b: 2px); @include bar; @include baz { color: red; } @include qux($args...); }"},
	{name: "include with a block and arguments", text: ".a { @include breakpoint(md) { b: c } @include x() using ($y) { z: $y } }"},
	{name: "function and return", text: "@function double($n) { @return $n * 2; }"},
	{name: "content, use, forward, extend", text: "@use 'sass:math';\n@use \"a\" as b with ($c: d);\n@forward 'src/list' hide list-reset;\n@mixin m { @content; @content(1); }\n.a { @extend %placeholder; @extend .b !optional; }"},
	{name: "debug, warn, error", text: "@debug 'x';\n@warn \"y #{$z}\";\n@error 'bad';"},
	{name: "if and else", text: "@if $a == 1 { b: c } @else if $a == 2 { d: e } @else { f: g }\n@if not index($list, $x) {}\n.a{@if $x{b:c}@else{d:e}}"},
	{name: "each", text: "@each $name, $glyph in $icons { .icon-#{$name}:before { content: $glyph; } }\n@each $key, $value in (a: 1, b: 2) {}"},
	{name: "for and while", text: "@for $i from 1 through 3 { .item-#{$i} { width: 2em * $i; } }\n@while $i > 0 { .item-#{$i} { width: 2em * $i; } $i: $i - 2; }"},
	{name: "at-root and nested media", text: ".a { @at-root .b { c: d } @media (min-width: $md) { e: f } }"},
	{name: "import list", text: "@import 'a', 'b';\n@import url(foo.css);"},

	// Selectors.
	{name: "placeholder selectors", text: "%message-shared { border: 1px; }\n.message { @extend %message-shared; }"},
	{name: "parent selector suffixes", text: ".a { &-b { } &__c { } & + & {} &.d {} .e & {} &:hover {} }"},
	{name: "escapes", text: ".a\\:b {}\n.\\31 0{}"},
	{name: "keyframes", text: "@keyframes x { 0% { a: b } 100% { c: d } }"},

	// Commas, which postcss-scss makes their own word tokens ending one past them.
	{name: "trailing comma in a declaration", text: "a{b:c,}"},
	{name: "trailing comma after a space", text: "a{b:c ,}"},
	{name: "trailing comma in an at-rule", text: "a{@include x,}"},
	{name: "trailing comma in a selector", text: "a,{}"},
	{name: "comma in media params", text: "@media screen,print{}"},
	{name: "comma first", text: ",a{}"},
	{name: "comma at end of file", text: "a: b,"},
	{name: "commas in values", text: "a { b: c,d , e; f: g(h,i) }"},

	// Prettier's css-in-js placeholders.
	{name: "placeholder as a selector", text: "@prettier-placeholder-0-id {\n  color: red;\n}"},
	{name: "placeholder with a pseudo class", text: "@prettier-placeholder-0-id:hover { }\n@prettier-placeholder-1-id @prettier-placeholder-2-id {}"},
	{name: "placeholder inside a selector", text: ".a @prettier-placeholder-0-id { }\n.b@prettier-placeholder-1-id, .c { }"},
	{name: "placeholder as a value", text: "a { color: @prettier-placeholder-0-id; margin: @prettier-placeholder-0-id @prettier-placeholder-1-id; width: calc(100% - @prettier-placeholder-2-id); transition: opacity @prettier-placeholder-3-idms; }"},
	{name: "placeholder as a property", text: "@prettier-placeholder-0-id: 1px;\na { @prettier-placeholder-1-id: 1px; @prettier-placeholder-2-id-x: y }"},
	{name: "placeholder after a word in a property", text: "a { border-@prettier-placeholder-1-id: 1px; }", malformed: true},
	{name: "placeholder standalone", text: "@prettier-placeholder-0-id;\n@prettier-placeholder-1-id\n"},
	{name: "placeholder standalone in a block", text: "a { @prettier-placeholder-0-id; color: red; @prettier-placeholder-1-id }"},
	{name: "placeholder in media params", text: "@media (max-width: @prettier-placeholder-0-id) { a { b: c } }"},
	{name: "placeholder in a comment", text: "/* @prettier-placeholder-0-id */\n// @prettier-placeholder-1-id\n"},
	{name: "placeholder template", text: "\n  display: flex;\n  @prettier-placeholder-0-id;\n  color: @prettier-placeholder-1-id;\n  &:hover { @prettier-placeholder-2-id }\n"},

	// Malformed.
	{name: "unclosed interpolation", text: "a { b: #{c", malformed: true},
	{name: "interpolation closing the block brace", text: "a { b: #{c }", malformed: true},
	{name: "unclosed interpolation in a string", text: "a { b: \"#{c\" }", malformed: true},
	{name: "unclosed string", text: "a { b: \"c }", malformed: true},
	{name: "unclosed string at the last character", text: "a{b:\"", malformed: true},
	{name: "unclosed url", text: "a{b:url(x}", malformed: true},
	{name: "unclosed block", text: "a {", malformed: true},
	{name: "unclosed comment", text: "/* x", malformed: true},
	{name: "nested property missing a semicolon", text: "a { b: c d:e { } }", malformed: true},
	{name: "nested property with no word", text: "a { :\"x\" {} }", malformed: true},
	{name: "unknown word", text: "a { b c }", malformed: true},
	{name: "unexpected close", text: "a{}}", malformed: true},
}

// TestParseSCSSAgreesWithPostcssScss is the scss parser's acceptance test, TestParseAgreesWithPostcss's
// twin: every Prettier tests/format/scss/**/*.scss file under COHERE_PRETTIER_FORK, the css-in-js
// templates of Prettier's styled-components and styled-jsx tests (built as Prettier's embed builds
// them, @prettier-placeholder-N-id and all), and every inline fixture must parse to postcss-scss's tree
// or be refused by both with the same error. Off unless COHERE_PRETTIER_FORK names the fork.
//
// A .scss file postcss-scss refuses is counted, and the test fails if that passes a tenth of them.
func TestParseSCSSAgreesWithPostcssScss(t *testing.T) {
	t.Parallel()
	root := corpus.PrettierFork.Root(t)

	type input struct {
		Name string `json:"name"`
		Text string `json:"text,omitempty"`
		Js   string `json:"js,omitempty"`
	}
	var inputs []input
	malformed := map[string]bool{}
	corpusFiles := map[string]bool{}
	err := filepath.WalkDir(filepath.Join(root, "tests", "format", "scss"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".scss" {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(source) {
			return fmt.Errorf("%s is not UTF-8, which the comparison cannot carry through JSON", path)
		}
		inputs = append(inputs, input{Name: path, Text: string(source)})
		corpusFiles[path] = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(corpusFiles) == 0 {
		t.Fatalf("no .scss files under %s/tests/format/scss", root)
	}
	javaScriptFiles := []string{}
	for _, pattern := range []string{"tests/format/js/multiparser-css/*.js", "tests/format/js/template-literals/*.js"} {
		matches, _ := filepath.Glob(filepath.Join(root, pattern))
		for _, match := range matches {
			if filepath.Base(match) != "format.test.js" {
				javaScriptFiles = append(javaScriptFiles, match)
				inputs = append(inputs, input{Name: match, Js: match})
			}
		}
	}
	for _, fixture := range scssParseFixtures {
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
	command := exec.Command("node", "--input-type=module", "-e", postcssScssParseScript)
	command.Dir = root
	command.Stdin = &stdin
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("node: %v\n%s", err, stderr.String())
	}
	var records []map[string]any
	seen := map[string]bool{}
	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
		seen[record["name"].(string)] = true
	}
	for _, each := range inputs {
		if each.Js == "" && !seen[each.Name] {
			t.Fatalf("the oracle printed nothing for %s", each.Name)
		}
	}

	identical, refusedByBoth, corpusRefused, templates, different := 0, 0, 0, 0, 0
	fail := func(format string, arguments ...any) {
		different++
		if different <= 20 {
			t.Errorf(format, arguments...)
		}
	}
	for _, want := range records {
		name := want["name"].(string)
		text := want["text"].(string)
		if !corpusFiles[name] && !strings.HasPrefix(name, "fixture: ") {
			templates++
		}
		wantError, oracleRefused := want["error"]
		if expectMalformed, isFixture := malformed[name]; isFixture && expectMalformed != oracleRefused {
			fail("%s: the fixture says malformed=%v and postcss-scss says %v", name, expectMalformed, want)
			continue
		}
		got, gotError := dumpParseWith(ParseSCSS, text)
		switch {
		case oracleRefused && gotError != nil:
			if difference := firstParseDifference(wantError, gotError, "error"); difference != "" {
				fail("%s: both refuse, differently: %s", name, difference)
				continue
			}
			refusedByBoth++
			if corpusFiles[name] {
				corpusRefused++
			}
		case oracleRefused:
			fail("%s: postcss-scss refuses it (%v) and Go parsed it", name, wantError)
		case gotError != nil:
			fail("%s: postcss-scss parses it and Go refused: %v", name, gotError)
		default:
			if difference := firstParseDifference(want["root"], got, "root"); difference != "" {
				fail("%s: %s", name, difference)
				continue
			}
			identical++
		}
	}
	t.Logf("%d compared (%d .scss files, %d css-in-js templates from %d JavaScript files, %d fixtures): %d identical, %d refused by both with the same error, %d different; %d .scss files refused",
		len(records), len(corpusFiles), templates, len(javaScriptFiles), len(scssParseFixtures), identical, refusedByBoth, different, corpusRefused)
	if corpusRefused > len(corpusFiles)/10 {
		t.Errorf("postcss-scss refused %d of %d .scss files; the comparison is thinner than it looks", corpusRefused, len(corpusFiles))
	}
	if templates == 0 {
		t.Errorf("no css-in-js templates came out of %d JavaScript files", len(javaScriptFiles))
	}

	// What postcss-scss adds must be in the trees compared, or agreement says nothing about it.
	features := map[string]int{}
	for _, want := range records {
		if root, parsed := want["root"]; parsed {
			countScssFeatures(root, features)
		} else if wantError, isMap := want["error"].(map[string]any); isMap && wantError["notCssSyntaxError"] != nil {
			features["TypeError refusal"]++
		}
	}
	t.Logf("postcss-scss features compared: %v", features)
	for _, feature := range []string{"inline comment", "inline comment raws.text", "nested declaration", "raws.selector.scss", "raws.value.scss", "raws.params.scss", "TypeError refusal"} {
		if features[feature] == 0 {
			t.Errorf("no input exercises %s", feature)
		}
	}
}

// countScssFeatures counts, in an oracle dump, the nodes and raws only postcss-scss makes.
func countScssFeatures(value any, features map[string]int) {
	switch typed := value.(type) {
	case []any:
		for _, each := range typed {
			countScssFeatures(each, features)
		}
	case map[string]any:
		raws, _ := typed["raws"].(map[string]any)
		if typed["type"] == "comment" && raws["inline"] == true {
			features["inline comment"]++
			if _, hasText := raws["text"]; hasText {
				features["inline comment raws.text"]++
			}
		}
		if typed["type"] == "decl" && typed["isNested"] == true {
			features["nested declaration"]++
		}
		for _, field := range []string{"selector", "value", "params"} {
			if raw, isMap := raws[field].(map[string]any); isMap {
				if _, hasScss := raw["scss"]; hasScss {
					features["raws."+field+".scss"]++
				}
			}
		}
		countScssFeatures(typed["nodes"], features)
	}
}
