package graphql

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

// graphqlParseScript is the oracle: graphql-js 17.0.2 from the fork's node_modules, called the way
// Prettier's parser-graphql.js calls it, with parseComments's token walk, dumped in the shape
// dumpParse prints. A field graphql-js wrote as undefined dumps as "<undefined>" and a field it never
// wrote is absent, so property presence is compared too. Offsets are UTF-16 there; the Go side converts.
// A parse error dumps as Prettier's createError message: graphql-js's message and its first location.
const graphqlParseScript = `
import { parse } from "graphql";
function dump(v) {
  if (v === undefined) return "<undefined>";
  if (v === null) return null;
  if (Array.isArray(v)) return v.map(dump);
  if (typeof v !== "object") return v;
  const o = { kind: v.kind, loc: [v.loc.start, v.loc.end] };
  for (const k of Object.keys(v).sort()) { if (k === "loc" || k === "kind") continue; o[k] = dump(v[k]); }
  return o;
}
let input = ""; process.stdin.setEncoding("utf8"); process.stdin.on("data", (c) => input += c);
process.stdin.on("end", () => {
  for (const line of input.split("\n").filter(Boolean)) {
    const { name, text } = JSON.parse(line);
    try {
      const ast = parse(text, { experimentalFragmentArguments: true });
      const comments = [];
      for (let t = ast.loc.startToken; t !== ast.loc.endToken; t = t.next) {
        if (t.kind === "Comment") comments.push({ loc: [t.start, t.end], value: t.value });
      }
      process.stdout.write(JSON.stringify({ name, ast: dump(ast), comments }) + "\n");
    } catch (e) {
      const at = e.locations?.[0];
      process.stdout.write(JSON.stringify({ name, error: at ? e.message + " (" + at.line + ":" + at.column + ")" : "not a GraphQLError: " + String(e) }) + "\n");
    }
  }
});
`

// parseFixture is an inline input. malformed says which way graphql-js must go on it, so an oracle that
// refuses a valid fixture, or accepts a malformed one, fails the test rather than quietly agreeing.
type parseFixture struct {
	name      string
	text      string
	malformed bool
}

var parseFixtures = []parseFixture{
	// Executable definitions.
	{name: "shorthand query", text: "{ a }"},
	{name: "operations of every type", text: "query { a } mutation M { b } subscription S @live { c } query Q($a: Int, $b: [String!]! = [\"x\"] @d) @e { f }"},
	{name: "fields", text: "{ alias: field(a: 1, b: \"two\", c: THREE, d: null, e: true, f: false, g: 1.5e3, h: $v, i: [1, 2], j: {k: 1, l: {m: 2}}) @include(if: $x) @skip(if: false) { nested { deeper } } }"},
	{name: "fragments", text: "{ ...Spread @defer ... on Type { x } ... @include(if: true) { y } ... { z } }\nfragment Spread on Type @dir { a }"},
	{name: "fragment arguments", text: "fragment F($a: Int = 1, $b: String @deprecated) on T @dir { f(a: $a) ...G(x: $a, y: {z: [$b]}) }\nquery { ...F(a: 2, b: \"x\") @skip(if: false) ...F }"},
	{name: "descriptions on executable definitions and variables", text: "\"operation\" query Q(\"variable\" $a: Int, \"\"\"block variable\"\"\" $b: Int) { a }\n\"\"\"fragment\"\"\" fragment F on T { a }"},
	{name: "values", text: "{ a(l: [], o: {}, n: {a: [{b: 1}], c: {}}, m: [[1], []], neg: -0, f1: 0.0, f2: 1E-3, f3: -12.5e+10, i: 1234567890123456789012) }"},
	{name: "types", text: "query($a: [[Int!]!]!, $b: T, $c: [T]) { a }"},
	{name: "commas and whitespace are ignored", text: ",,,{\ta,,b\t,c}\t,"},

	// Type system definitions.
	{name: "schema definition", text: "\"schema description\" schema @a { query: Query mutation: Mutation subscription: Subscription }"},
	{name: "scalar", text: "\"\"\"scalar\"\"\" scalar Date @specifiedBy(url: \"https://example.com\")"},
	{name: "object type", text: "type Query implements A & B @key(fields: \"id\") { \"field\" field(\"argument\" a: Int = 1 @d, b: [String!]!): String! @deprecated(reason: \"no\") other: Int }\ntype Empty\ntype Leading implements & A & B { a: Int }"},
	{name: "interface implementing interfaces", text: "interface I implements J & K @d { a: Int }\ninterface Bare"},
	{name: "union", text: "union U = | A | B\nunion V @d = A\nunion Lone"},
	{name: "enum", text: "enum E { \"value\" A @h B C_D }\nenum Bare"},
	{name: "input", text: "input In { a: Int = 1 b: In2! @f c: [In] = [{a: 1}] }"},
	{name: "directive definition", text: "\"\"\"directive\"\"\" directive @dir(a: Int = 1, b: String) repeatable on | FIELD | FRAGMENT_SPREAD | INLINE_FRAGMENT\ndirective @plain on QUERY\ndirective @all @meta on FRAGMENT_VARIABLE_DEFINITION | VARIABLE_DEFINITION | SCHEMA | SCALAR | OBJECT | FIELD_DEFINITION | ARGUMENT_DEFINITION | INTERFACE | UNION | ENUM | ENUM_VALUE | INPUT_OBJECT | INPUT_FIELD_DEFINITION | DIRECTIVE_DEFINITION | MUTATION | SUBSCRIPTION | FRAGMENT_DEFINITION"},
	{name: "every extension", text: "extend schema @b\nextend schema { query: Q }\nextend scalar Date @c\nextend type Query implements C { added: Int }\nextend type Query @e\nextend type Query implements D\nextend interface I @g { b: Int }\nextend interface I implements J\nextend union U = C\nextend union U @u\nextend enum E { C }\nextend enum E @x\nextend input In { c: String }\nextend input In @y\nextend directive @dir @i"},
	{name: "keywords as names", text: "type type { query: query fragment(on: on): extend schema: scalar }\n{ on: on query mutation }"},

	// Strings.
	{name: "string escapes", text: `{ a(s: "quote\" backslash\\ slash\/ b\b f\f n\n r\r t\t") }`},
	{name: "unicode escapes", text: `{ a(fixed: "\u00e9\u0041\uFFFF", variable: "\u{1F600}\u{0}\u{10FFFF}\u{000041}", pair: "\uD83D\uDE00", lower: "\ud83d\ude00", mixed: "x\u{1f600}y") }`},
	{name: "literal non-ASCII", text: "\"descripción ✓ 😀 日本語\" type T { \"emoji 😀 then more\" a(x: String = \"😀😀\"): Int } # comment 😀 after\n{ a(b: \"é\") }"},
	{name: "block string indentation", text: "type T {\n  \"\"\"\n    first\n      indented\n\n    last\n  \"\"\"\n  a: Int\n}"},
	{name: "block string first line kept", text: "{ a(s: \"\"\"  first line\n      second\n    third\n  \"\"\") }"},
	{name: "block string escaped triple quote", text: "{ a(s: \"\"\"has \\\"\"\" triple and \\\"\"\"\"\"\") b(s: \"\"\"\\\"\"\"\"\"\") }"},
	{name: "block string blank lines and tabs", text: "{ a(s: \"\"\"\n\n\t\tx\n\t  \n\t\ty\n\n   \n\"\"\") b(s: \"\"\"   \"\"\") c(s: \"\"\"\"\"\") d(s: \"\"\"\n\"\"\") }"},
	{name: "block string with CRLF and CR", text: "{ a(s: \"\"\"\r\n  one\r\n  two\r  three\r\n\"\"\") }\r\n# comment\r\n{ b }"},
	{name: "block string backslashes stay raw", text: "{ a(s: \"\"\"\\n \\u0041 \\\\ \\\"\"\" \"\"\") }"},
	{name: "block string with non-ASCII", text: "\"\"\"\n  😀 first\n    é second\n\"\"\" scalar S"},

	// Comments.
	{name: "comments everywhere", text: "# leading\nquery # after keyword\nQ # after name\n( # in variables\n$a # after variable\n: # after colon\nInt # after type\n) # after variables\n{ # in selection\na # after field\n( # in arguments\nb: # after argument name\n1 # after value\n) # after arguments\n} # trailing\n# last"},
	{name: "comments in type system", text: "# schema\ntype T # name\n{ # open\n  # before field\n  a: Int # after field\n  # before close\n} # after\nenum E { # e\n A # a\n}"},
	{name: "comment at end of file without newline", text: "{ a } #end"},
	{name: "empty comment and comment only lines", text: "#\n{ a }\n#\n#\n"},
	{name: "comment with a carriage return", text: "# one\r{ a }\r# two\r"},

	// Byte order marks.
	{name: "leading byte order mark", text: "\xef\xbb\xbf{ a }"},
	{name: "byte order mark between tokens", text: "{ a \xef\xbb\xbf b }"},

	// Malformed: each must be refused by both sides with the same message.
	{name: "empty document", text: "", malformed: true},
	{name: "only a comment", text: "# nothing", malformed: true},
	{name: "unclosed selection set", text: "{ a", malformed: true},
	{name: "empty selection set", text: "query { }", malformed: true},
	{name: "unterminated block string", text: "{ a(s: \"\"\"never closed) }", malformed: true},
	{name: "unterminated string", text: "{ a(s: \"never closed) }", malformed: true},
	{name: "newline in string", text: "{ a(s: \"line\nbreak\") }", malformed: true},
	{name: "unicode escape out of range", text: `{ a(s: "\u{110000}") }`, malformed: true},
	{name: "unicode escape too long", text: `{ a(s: "\u{FFFFFFFFF}") }`, malformed: true},
	{name: "unicode escape empty", text: `{ a(s: "\u{}") }`, malformed: true},
	{name: "unicode escape surrogate in braces", text: `{ a(s: "\u{D800}") }`, malformed: true},
	{name: "lone leading surrogate escape", text: `{ a(s: "\uD83D") }`, malformed: true},
	{name: "lone trailing surrogate escape", text: `{ a(s: "\uDE00") }`, malformed: true},
	{name: "leading surrogate then a non-surrogate", text: `{ a(s: "\uD83D\u0041") }`, malformed: true},
	{name: "short fixed escape", text: `{ a(s: "\u12") }`, malformed: true},
	{name: "fixed escape cut by non-ASCII", text: `{ a(s: "\u12é4") }`, malformed: true},
	{name: "variable escape cut by non-ASCII", text: `{ a(s: "\u{1é}") }`, malformed: true},
	{name: "variable escape cut by an emoji", text: `{ a(s: "\u{1😀}") }`, malformed: true},
	{name: "bad escape character", text: `{ a(s: "\x") }`, malformed: true},
	{name: "bad escape of a non-ASCII character", text: `{ a(s: "\é") }`, malformed: true},
	{name: "two dots", text: "{ ..a }", malformed: true},
	{name: "leading dot number", text: "{ a(b: .5) }", malformed: true},
	{name: "leading zero", text: "{ a(b: 00) }", malformed: true},
	{name: "dangling decimal point", text: "{ a(b: 1.) }", malformed: true},
	{name: "dangling exponent", text: "{ a(b: 1e) }", malformed: true},
	{name: "number then name", text: "{ a(b: 1a) }", malformed: true},
	{name: "number then dot", text: "{ a(b: 1.5.) }", malformed: true},
	{name: "number then non-ASCII", text: "{ a(b: 1é) }", malformed: true},
	{name: "lone minus", text: "{ a(b: -) }", malformed: true},
	{name: "single quote", text: "{ a(b: 'x') }", malformed: true},
	{name: "unexpected punctuation", text: "{ a ? }", malformed: true},
	{name: "unexpected emoji", text: "{ a 😀 }", malformed: true},
	{name: "unexpected non-ASCII after non-ASCII on the line", text: "{ a(b: \"日本語\") é }", malformed: true},
	{name: "error on a later line after CRLF", text: "{\r\n  a(b: \"😀\")\r\n  ?\r\n}", malformed: true},
	{name: "extension with nothing", text: "extend type Foo", malformed: true},
	{name: "scalar extension without directives", text: "extend scalar X", malformed: true},
	{name: "schema extension with nothing", text: "extend schema", malformed: true},
	{name: "directive extension without directives", text: "extend directive @d", malformed: true},
	{name: "extend an unknown kind", text: "extend thing X @d", malformed: true},
	{name: "reserved enum value", text: "enum E { true }", malformed: true},
	{name: "reserved enum value null", text: "enum E { A null }", malformed: true},
	{name: "fragment named on", text: "fragment on on T { a }", malformed: true},
	{name: "spread of a fragment named on", text: "{ ...on }", malformed: true},
	{name: "variable in a constant", text: "type T { f(a: Int = $v): Int }", malformed: true},
	{name: "dollar in a constant", text: "query($a: Int = $) { a }", malformed: true},
	{name: "unknown directive location", text: "directive @d on FOO", malformed: true},
	{name: "description on shorthand query", text: "\"desc\" { a }", malformed: true},
	{name: "description on an extension", text: "\"desc\" extend type T @d", malformed: true},
	{name: "unknown definition", text: "notakeyword X", malformed: true},
	{name: "operation type unknown in schema", text: "schema { other: Q }", malformed: true},
	{name: "missing type condition", text: "fragment F { a }", malformed: true},
	{name: "empty arguments", text: "{ a() }", malformed: true},
	{name: "empty fragment arguments", text: "{ ...F() }", malformed: true},
	{name: "value missing", text: "{ a(b: ) }", malformed: true},
	{name: "list type unclosed", text: "query($a: [Int) { a }", malformed: true},
	{name: "repeatable after on", text: "directive @d on FIELD repeatable", malformed: true},
	{name: "string where a name goes", text: "{ \"a\" }", malformed: true},
	{name: "error positioned at end of file", text: "query Q(", malformed: true},
}

// TestParseAgreesWithGraphqlJs is the parser's acceptance test: for every .graphql and .gql file in the
// corpora, and every inline fixture, the Go tree and comment list must equal graphql-js's, property for
// property, and a refusal must be a refusal on both sides with the same message. Off unless
// COHERE_PRETTIER_FORK names the fork (whose node_modules hold graphql 17.0.2) and COHERE_GRAPHQL_CORPORA
// lists directories to walk (colon-separated, node_modules skipped).
//
// A corpus file graphql-js cannot parse is counted, and the test fails if that count passes a tenth of
// the corpus: an oracle that rejects everything would otherwise read as perfect agreement. The inline
// fixtures each say which way graphql-js must go.
func TestParseAgreesWithGraphqlJs(t *testing.T) {
	t.Parallel()
	corpora := os.Getenv("COHERE_GRAPHQL_CORPORA")
	if corpora == "" {
		t.Skip("set COHERE_GRAPHQL_CORPORA to compare the parser against graphql-js")
	}
	root := corpus.PrettierFork.Root(t)

	type input struct {
		Name string `json:"name"`
		Text string `json:"text"`
	}
	var inputs []input
	malformed := map[string]bool{}
	corpusFiles := 0
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
			if entry.IsDir() || (filepath.Ext(path) != ".graphql" && filepath.Ext(path) != ".gql") {
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
			corpusFiles++
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", corpus, err)
		}
	}
	if corpusFiles == 0 {
		t.Fatalf("no .graphql or .gql files under %s", corpora)
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
	command := exec.Command("node", "--input-type=module", "-e", graphqlParseScript)
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
		wantError, oracleRefused := want["error"].(string)
		if expectMalformed, isFixture := malformed[each.Name]; isFixture && expectMalformed != oracleRefused {
			fail("%s: the fixture says malformed=%v and graphql-js says %v", each.Name, expectMalformed, want)
			continue
		}
		got, err := dumpParse(each.Text)
		switch {
		case oracleRefused && err != nil:
			if err.Error() != wantError {
				fail("%s: both refuse, with different messages:\n  graphql-js %s\n  Go         %s", each.Name, wantError, err.Error())
				continue
			}
			refusedByBoth++
			if _, isFixture := malformed[each.Name]; !isFixture {
				corpusRefused++
			}
		case oracleRefused:
			fail("%s: graphql-js refuses it (%s) and Go parsed it", each.Name, wantError)
		case err != nil:
			fail("%s: graphql-js parses it and Go refused: %v", each.Name, err)
		default:
			difference := firstParseDifference(want["ast"], got["ast"], "ast") + firstParseDifference(want["comments"], got["comments"], "comments")
			if difference != "" {
				fail("%s: %s", each.Name, difference)
				continue
			}
			identical++
		}
	}
	t.Logf("%d compared (%d corpus files, %d fixtures): %d identical, %d refused by both with the same message, %d different; %d corpus files refused",
		len(inputs), corpusFiles, len(parseFixtures), identical, refusedByBoth, different, corpusRefused)
	if corpusRefused > corpusFiles/10 {
		t.Errorf("graphql-js refused %d of %d corpus files; the comparison is thinner than it looks", corpusRefused, corpusFiles)
	}
}

// dumpParse parses one text and renders it the way the oracle script does, through JSON, so both sides
// compare as the same Go values.
func dumpParse(text string) (map[string]any, error) {
	document, comments, err := Parse(text)
	if err != nil {
		return nil, err
	}
	offsets := utf16OffsetsOf(text)
	dumped := make([]any, len(comments))
	for index, comment := range comments {
		if comment.Type() != "Comment" {
			return nil, fmt.Errorf("a comment of type %q", comment.Type())
		}
		dumped[index] = map[string]any{
			"loc":   []int{offsets[comment.Range[0]], offsets[comment.Range[1]]},
			"value": comment.String("value"),
		}
	}
	encoded, err := json.Marshal(map[string]any{"ast": dumpParseValue(document, offsets), "comments": dumped})
	if err != nil {
		return nil, err
	}
	var result map[string]any
	err = json.Unmarshal(encoded, &result)
	return result, err
}

func dumpParseValue(value any, offsets []int) any {
	switch typed := value.(type) {
	case nil:
		return "<undefined>"
	case *estree.Node:
		result := map[string]any{"kind": typed.Type(), "loc": []int{offsets[typed.Range[0]], offsets[typed.Range[1]]}}
		for _, key := range typed.Keys() {
			result[key] = dumpParseValue(typed.Get(key), offsets)
		}
		return result
	case []*estree.Node:
		if typed == nil {
			return "a nil list, which no graphql-js field holds"
		}
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = dumpParseValue(child, offsets)
		}
		return result
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
				return fmt.Sprintf("%s.%s: present in graphql-js %v, in Go %v (%v)", path, key, wantPresent, gotPresent, typed["kind"])
			}
			if difference := firstParseDifference(wantValue, gotValue, path+"."+key); difference != "" {
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
