// This file is not built here. graphql_test.go lays it into cohere's own package,
// cohere/internal/format/graphql, with `go test -overlay`, so that it can call cohere's Parse exactly as
// cohere's tests do, without a byte of the submodule changing.
//
// It writes every text the port is asked to parse, as the cases file main.ts reads, and Go cohere's
// answer to each, in the words main.ts prints. The texts are cohere's own: every string constant in
// this package's tests, read out of their Go with go/ast (the inline fixtures its graphql-js oracle test
// runs, the shapes, string values, comments and errors its parser tests check, and the snippets its
// format tests print, along with their names and expectations, which are texts too), so the cases follow
// cohere's tests as they change. Then generated documents, from a seeded generator that writes every
// kind of definition, selection, value and type from pieces at the edges of each of the lexer's and the
// parser's classes, and breaks some of them.

package graphql

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
)

// adamicRequest is what graphql_test.go asks.
type adamicRequest struct {
	Seed      int64  `json:"seed"`
	Generated int    `json:"generated"`
	Cases     string `json:"cases"`
	Answers   string `json:"answers"`
}

// adamicCohereTexts reads every string constant out of this package's tests: each string literal, and
// each sum of string literals as the whole it spells.
func adamicCohereTexts(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	seen := map[string]bool{}
	var constant func(expression ast.Expr) (string, bool)
	constant = func(expression ast.Expr) (string, bool) {
		switch typed := expression.(type) {
		case *ast.BasicLit:
			if typed.Kind != gotoken.STRING {
				return "", false
			}
			text, err := strconv.Unquote(typed.Value)
			return text, err == nil
		case *ast.ParenExpr:
			return constant(typed.X)
		case *ast.BinaryExpr:
			if typed.Op != gotoken.ADD {
				return "", false
			}
			left, isLeft := constant(typed.X)
			right, isRight := constant(typed.Y)
			return left + right, isLeft && isRight
		}
		return "", false
	}
	files := gotoken.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "adamic_") {
			continue
		}
		file, err := goparser.ParseFile(files, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if _, isImport := node.(*ast.ImportSpec); isImport {
				return false
			}
			expression, isExpression := node.(ast.Expr)
			if !isExpression {
				return true
			}
			text, isConstant := constant(expression)
			if !isConstant {
				return true
			}
			if !seen[text] {
				seen[text] = true
				texts = append(texts, text)
			}
			// A sum is one text; its parts are not texts of their own.
			return false
		})
	}
	return texts
}

// adamicEscape is a backslash and then rest: an escape sequence in a GraphQL string, spelled at run time
// so that no tool that reads this file can take it for one of Go's own.
func adamicEscape(rest string) string {
	return `\` + rest
}

// adamicCharacter is one code point as a string, for the characters a source file shouldn't hold raw.
func adamicCharacter(code rune) string {
	return string(code)
}

// The generator's pieces, each list starting at the edges of a class the lexer or the parser decides
// on, then the near misses just outside it.
var (
	// The 9 ways to be ignored between tokens, then what isn't: a vertical tab, a form feed, no-break and
	// line separators, U+0085, and a zero-width space, each "Unexpected character".
	adamicSeparators = []string{" ", "\t", ",", "\n", "\r\n", "\r", adamicCharacter(0xfeff), " ,\t", "\n\r", "\v", "\f", adamicCharacter(0xa0), adamicCharacter(0x2028),
		adamicCharacter(0x85), adamicCharacter(0x200b)}

	// Names at every edge of NameStart and NameContinue, and the characters just outside them: `@`, `[`,
	// a backquote, `{`, `/`, `:`, and letters that aren't ASCII. The first 32 are names, among them the
	// keywords and the names that are values; the first 13 are the plain ones.
	adamicNames = []string{"a", "z", "A", "Z", "_", "a0", "a9", "_9", "aZ", "Za", "__", "x_y", "Type", "on", "On", "ON", "query", "Query", "QUERY",
		"true", "True", "false", "null", "NULL", "nul", "fragment", "extend", "Extend", "schema", "implements", "repeatable", "subscription",
		"a@", "a[", "a`", "a{", "a/", "a:", adamicCharacter(0xe9), "a" + adamicCharacter(0xe9), "a" + adamicCharacter(0x301), "a" + adamicCharacter(0x1f600)}

	// The 13 keywords that start a definition, then others in another case, and a name that isn't one.
	adamicDefinitionKeywords = []string{"schema", "scalar", "type", "interface", "union", "enum", "input", "directive", "query", "mutation", "subscription", "fragment", "extend",
		"Schema", "SCALAR", "Type", "Query", "Fragment", "EXTEND", "notakeyword"}

	// Directive locations: each of the 21, then near misses.
	adamicLocations = []string{"QUERY", "MUTATION", "SUBSCRIPTION", "FIELD", "FRAGMENT_DEFINITION", "FRAGMENT_SPREAD", "INLINE_FRAGMENT", "VARIABLE_DEFINITION",
		"FRAGMENT_VARIABLE_DEFINITION", "SCHEMA", "SCALAR", "OBJECT", "FIELD_DEFINITION", "ARGUMENT_DEFINITION", "INTERFACE", "UNION", "ENUM", "ENUM_VALUE",
		"INPUT_OBJECT", "INPUT_FIELD_DEFINITION", "DIRECTIVE_DEFINITION", "field", "Field", "FIELDS", "QUERY_", "_QUERY", "FOO"}

	// Numbers: every shape IntValue and FloatValue take, and each way to get one wrong. The first 14 are
	// numbers, and the first 3 integers.
	adamicNumbers = []string{"0", "9", "-0", "-9", "10", "1234567890123456789012", "0.0", "1.5", "-12.5e+10", "1e3", "1E-3", "1e+0", "0e0", "9.9E9",
		"00", "01", "-", "-a", "1.", "1.e", "1e", "1E+", "1a", "1_", "1.5.", "1" + adamicCharacter(0xe9), "0x1", ".5", "..5", "1..2", "-.5", "1.5e3.0", "1E", "0.", "-01"}

	// Inside a string: characters at the edges of UTF-8's and UTF-16's lengths and of the scalar values,
	// controls, and every escape graphql-js takes, at the edges of each: the hex digits' cases and
	// ends, surrogate pairs in both cases and at their ends, and the widest escape.
	adamicStringPieces = []string{
		"x", " ", adamicCharacter(0xe9), adamicCharacter(0x1f600), adamicCharacter(0x65e5) + adamicCharacter(0x672c), adamicCharacter(0x80),
		adamicCharacter(0x7ff), adamicCharacter(0x800), adamicCharacter(0xd7ff), adamicCharacter(0xe000), adamicCharacter(0xffff),
		adamicCharacter(0x10000), adamicCharacter(0x10ffff), "'", "#", "\t", adamicCharacter(0), adamicCharacter(0x1f), adamicCharacter(0x7f),
		adamicEscape(`"`), adamicEscape(`\`), adamicEscape("/"), adamicEscape("b"), adamicEscape("f"), adamicEscape("n"), adamicEscape("r"),
		adamicEscape("t"), adamicEscape("u0000"), adamicEscape("u0041"), adamicEscape("u001F"), adamicEscape("u007f"), adamicEscape("u00e9"),
		adamicEscape("uD7FF"), adamicEscape("uE000"), adamicEscape("uFFFF"), adamicEscape("uffff"), adamicEscape("u9AfF"), adamicEscape("u09af"),
		adamicEscape("uD83D") + adamicEscape("uDE00"), adamicEscape("ud83d") + adamicEscape("ude00"), adamicEscape("uDBFF") + adamicEscape("uDFFF"),
		adamicEscape("uD800") + adamicEscape("uDC00"), adamicEscape("u{0}"), adamicEscape("u{41}"), adamicEscape("u{1F600}"),
		adamicEscape("u{1f600}"), adamicEscape("u{10FFFF}"), adamicEscape("u{D7FF}"), adamicEscape("u{E000}"), adamicEscape("u{00000041}"),
	}

	// Escapes graphql-js refuses: letters it has no escape for, hex escapes cut short or with a
	// character just outside the hex digits, surrogates alone and out of order, the widest escape, one
	// past it, points past U+10FFFF, and points that reach the sign bit of its 32-bit arithmetic.
	adamicStringFaults = []string{
		adamicEscape("x"), adamicEscape(adamicCharacter(0xe9)), adamicEscape("u"), adamicEscape(" "), adamicEscape("0"), adamicEscape("'"),
		adamicEscape("B"), adamicEscape("N"), adamicEscape("u0g00"), adamicEscape("u00G0"), adamicEscape("u12"),
		adamicEscape("u12" + adamicCharacter(0xe9) + "4"), adamicEscape("u/000"), adamicEscape("u:000"), adamicEscape("u@000"),
		adamicEscape("u`000"), adamicEscape("uD800"), adamicEscape("uDBFF"), adamicEscape("uDC00"), adamicEscape("uDFFF"),
		adamicEscape("uD83D") + adamicEscape("u0041"), adamicEscape("uD83D") + adamicEscape("uD83D"), adamicEscape("uD83D") + adamicEscape("u"),
		adamicEscape("uD83D") + adamicEscape("x"), adamicEscape("uD83D"), adamicEscape("uDE00"), adamicEscape("uDE00") + adamicEscape("uD83D"),
		adamicEscape("u{110000}"), adamicEscape("u{D800}"), adamicEscape("u{DFFF}"), adamicEscape("u{000000041}"), adamicEscape("u{7FFFFFF}"),
		adamicEscape("u{8000000}"), adamicEscape("u{7FFFFFFF}"), adamicEscape("u{80000000}"), adamicEscape("u{FFFFFFFF}"),
		adamicEscape("u{FFFFFFFFF}"), adamicEscape("u{}"), adamicEscape("u{1" + adamicCharacter(0xe9) + "}"),
		adamicEscape("u{1" + adamicCharacter(0x1f600) + "}"), adamicEscape("u{12"), adamicEscape("u{g}"), adamicEscape("u{-1}"),
		adamicEscape("u{ 1}"), adamicEscape("u{1F600"), adamicEscape("u{FFFFFFFF"), adamicEscape("u{1234567" + adamicCharacter(0x1f600)),
	}

	// Inside a block string: what dedent and the closing quotes decide on.
	adamicBlockPieces = []string{"x", " ", "  ", "\t", "\n", "\r\n", "\r", "\n  ", "\n\t", "\n    ", "\n \t", adamicCharacter(0xe9), adamicCharacter(0x1f600),
		adamicEscape(`"""`), adamicEscape(`""`), adamicEscape(""), adamicEscape("n"), adamicEscape("u0041"), `"`, `""`, adamicCharacter(0), adamicCharacter(0xa0)}

	// Comments' text.
	adamicCommentPieces = []string{"", " ", "x", adamicCharacter(0xe9), adamicCharacter(0x1f600), "#", `"`, `"""`, "\t", adamicCharacter(0), adamicCharacter(0x2028),
		adamicCharacter(0xfeff), adamicCharacter(0x85)}

	// Characters the lexer refuses wherever a token starts.
	adamicStrays = []string{"?", "'", "%", "^", "*", "+", "~", ";", "<", ">", `\`, "/", "`", "..", ".", "....", adamicCharacter(0x1f600), adamicCharacter(0xe9),
		adamicCharacter(0), adamicCharacter(0x1f), adamicCharacter(0x7f), adamicCharacter(0x80), adamicCharacter(0xfffd), adamicCharacter(0x10ffff)}
)

// adamicGenerator writes documents from a seeded source. A tame document is made of valid pieces only,
// at the edges of what each class takes, so that most of them parse; a wild one also takes the near
// misses just outside each class, and is often broken besides.
type adamicGenerator struct {
	random *rand.Rand
	depth  int
	wild   bool
}

func (generator *adamicGenerator) pick(pieces []string) string {
	return pieces[generator.random.Intn(len(pieces))]
}

func (generator *adamicGenerator) chance(percent int) bool {
	return generator.random.Intn(100) < percent
}

// edge picks from the first valid pieces, or, in a wild document now and then, from all of them.
func (generator *adamicGenerator) edge(pieces []string, valid int) string {
	if generator.wild && generator.chance(15) {
		return generator.pick(pieces)
	}
	return generator.pick(pieces[:valid])
}

// wildly is whether a wild document takes a turn graphql-js refuses here.
func (generator *adamicGenerator) wildly(percent int) bool {
	return generator.wild && generator.chance(percent)
}

// separator is mostly a space, sometimes any of the separators, a comment, or nothing.
func (generator *adamicGenerator) separator() string {
	switch {
	case generator.chance(70):
		return " "
	case generator.chance(30):
		return ""
	case generator.chance(15):
		return "#" + generator.pick(adamicCommentPieces) + generator.pick(adamicCommentPieces) + generator.pick([]string{"\n", "\r", "\r\n"})
	}
	return generator.edge(adamicSeparators, 9)
}

// adamicGlues is whether a character run into another makes one token of two: a name's, a number's,
// a quote, or a dot.
func adamicGlues(character byte) bool {
	return character == '_' || character == '"' || character == '.' || (character >= '0' && character <= '9') || (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z')
}

// join writes tokens with separators between them, never nothing between two that would run together.
func (generator *adamicGenerator) join(tokens ...string) string {
	var builder strings.Builder
	for _, each := range tokens {
		if each == "" {
			continue
		}
		if builder.Len() > 0 {
			separator := generator.separator()
			written := builder.String()
			if separator == "" && adamicGlues(written[len(written)-1]) && adamicGlues(each[0]) {
				separator = " "
			}
			builder.WriteString(separator)
		}
		builder.WriteString(each)
	}
	return builder.String()
}

// name is a valid name mostly, any of the edge names sometimes.
func (generator *adamicGenerator) name() string {
	if generator.chance(85) {
		return generator.pick(adamicNames[:13])
	}
	return generator.edge(adamicNames, 32)
}

func (generator *adamicGenerator) stringValue() string {
	var builder strings.Builder
	builder.WriteString(`"`)
	for count := generator.random.Intn(4); count > 0; count-- {
		if generator.wildly(25) {
			builder.WriteString(generator.pick(adamicStringFaults))
		} else {
			builder.WriteString(generator.pick(adamicStringPieces))
		}
	}
	if !generator.wildly(5) {
		builder.WriteString(`"`)
	}
	return builder.String()
}

func (generator *adamicGenerator) blockString() string {
	var builder strings.Builder
	builder.WriteString(`"""`)
	for count := generator.random.Intn(6); count > 0; count-- {
		piece := generator.pick(adamicBlockPieces)
		builder.WriteString(piece)
		// A quote or a backslash run into the closing quotes ends the block early, or escapes them.
		if !generator.wild && (strings.HasSuffix(piece, `"`) || strings.HasSuffix(piece, `\`)) {
			builder.WriteString("x")
		}
	}
	if !generator.wildly(5) {
		builder.WriteString(`"""`)
	}
	return builder.String()
}

func (generator *adamicGenerator) description() string {
	if generator.chance(70) {
		return ""
	}
	if generator.chance(50) {
		return generator.stringValue() + generator.separator()
	}
	return generator.blockString() + generator.separator()
}

// value is a Value; constant leaves out variables, mostly.
func (generator *adamicGenerator) value(constant bool) string {
	generator.depth++
	defer func() { generator.depth-- }()
	choice := generator.random.Intn(12)
	if generator.depth > 4 && choice >= 9 {
		choice = 0
	}
	switch choice {
	case 0, 1:
		return generator.edge(adamicNumbers, 14)
	case 2:
		return generator.stringValue()
	case 3:
		return generator.blockString()
	case 4:
		return generator.pick([]string{"true", "false", "null", "True", "NULL"})
	case 5, 6:
		return generator.name()
	case 7, 8:
		if constant && (!generator.wild || generator.chance(80)) {
			return generator.pick(adamicNumbers[:3])
		}
		return generator.join("$", generator.name())
	case 9:
		tokens := []string{"["}
		for count := generator.random.Intn(3); count > 0; count-- {
			tokens = append(tokens, generator.value(constant))
		}
		return generator.join(append(tokens, "]")...)
	default:
		tokens := []string{"{"}
		for count := generator.random.Intn(3); count > 0; count-- {
			tokens = append(tokens, generator.name(), ":", generator.value(constant))
		}
		return generator.join(append(tokens, "}")...)
	}
}

func (generator *adamicGenerator) typeReference() string {
	generator.depth++
	defer func() { generator.depth-- }()
	var reference string
	if generator.depth < 4 && generator.chance(30) {
		reference = generator.join("[", generator.typeReference(), "]")
	} else {
		reference = generator.name()
	}
	if generator.chance(35) {
		reference = generator.join(reference, "!")
	}
	return reference
}

// arguments is ( Argument+ ), mostly; empty sometimes, which graphql-js refuses.
func (generator *adamicGenerator) arguments(constant bool) string {
	if generator.chance(55) {
		return ""
	}
	tokens := []string{"("}
	count := 1 + generator.random.Intn(3)
	if generator.wildly(5) {
		count = 0
	}
	for ; count > 0; count-- {
		tokens = append(tokens, generator.name(), ":", generator.value(constant))
	}
	return generator.join(append(tokens, ")")...)
}

func (generator *adamicGenerator) directives(constant bool) string {
	var tokens []string
	for count := generator.random.Intn(3) - 1; count > 0; count-- {
		tokens = append(tokens, generator.join("@", generator.name(), generator.arguments(constant)))
	}
	return generator.join(tokens...)
}

func (generator *adamicGenerator) variableDefinitions() string {
	if generator.chance(60) {
		return ""
	}
	tokens := []string{"("}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		definition := []string{generator.description() + "$", generator.name(), ":", generator.typeReference()}
		if generator.chance(40) {
			definition = append(definition, "=", generator.value(true))
		}
		tokens = append(tokens, generator.join(append(definition, generator.directives(true))...))
	}
	return generator.join(append(tokens, ")")...)
}

func (generator *adamicGenerator) selectionSet() string {
	generator.depth++
	defer func() { generator.depth-- }()
	tokens := []string{"{"}
	count := 1 + generator.random.Intn(3)
	if generator.wildly(5) {
		count = 0
	}
	for ; count > 0; count-- {
		kind := generator.random.Intn(6)
		if generator.depth >= 4 && (kind == 1 || kind == 2) {
			kind = 0
		}
		switch kind {
		case 0:
			tokens = append(tokens, generator.join("...", generator.name(), generator.arguments(false), generator.directives(false)))
		case 1:
			tokens = append(tokens, generator.join("...", "on", generator.name(), generator.directives(false), generator.selectionSet()))
		case 2:
			tokens = append(tokens, generator.join("...", generator.directives(false), generator.selectionSet()))
		default:
			field := []string{}
			if generator.chance(25) {
				field = append(field, generator.name(), ":")
			}
			field = append(field, generator.name(), generator.arguments(false), generator.directives(false))
			if generator.depth < 4 && generator.chance(30) {
				field = append(field, generator.selectionSet())
			}
			tokens = append(tokens, generator.join(field...))
		}
	}
	return generator.join(append(tokens, "}")...)
}

func (generator *adamicGenerator) inputValues(open string, close string) string {
	tokens := []string{open}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		value := []string{generator.description() + generator.name(), ":", generator.typeReference()}
		if generator.chance(30) {
			value = append(value, "=", generator.value(true))
		}
		tokens = append(tokens, generator.join(append(value, generator.directives(true))...))
	}
	return generator.join(append(tokens, close)...)
}

func (generator *adamicGenerator) fields() string {
	tokens := []string{"{"}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		field := []string{generator.description() + generator.name()}
		if generator.chance(30) {
			field = append(field, generator.inputValues("(", ")"))
		}
		tokens = append(tokens, generator.join(append(field, ":", generator.typeReference(), generator.directives(true))...))
	}
	return generator.join(append(tokens, "}")...)
}

func (generator *adamicGenerator) implements() string {
	if generator.chance(60) {
		return ""
	}
	tokens := []string{"implements"}
	if generator.chance(25) {
		tokens = append(tokens, "&")
	}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		if len(tokens) > 1 && tokens[len(tokens)-1] != "&" {
			tokens = append(tokens, "&")
		}
		tokens = append(tokens, generator.name())
	}
	return generator.join(tokens...)
}

func (generator *adamicGenerator) members(delimiter string) string {
	var tokens []string
	if generator.chance(25) {
		tokens = append(tokens, delimiter)
	}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		if len(tokens) > 0 && tokens[len(tokens)-1] != delimiter {
			tokens = append(tokens, delimiter)
		}
		tokens = append(tokens, generator.name())
	}
	return generator.join(tokens...)
}

func (generator *adamicGenerator) optional(percent int, text func() string) string {
	if generator.chance(percent) {
		return text()
	}
	return ""
}

func (generator *adamicGenerator) operationTypes() string {
	tokens := []string{"{"}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		tokens = append(tokens, generator.edge([]string{"query", "mutation", "subscription", "Query", "other"}, 3), ":", generator.name())
	}
	return generator.join(append(tokens, "}")...)
}

func (generator *adamicGenerator) enumValues() string {
	tokens := []string{"{"}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		tokens = append(tokens, generator.join(generator.description()+generator.name(), generator.directives(true)))
	}
	return generator.join(append(tokens, "}")...)
}

// definition is one definition, of the kind its keyword says.
func (generator *adamicGenerator) definition() string {
	keyword := generator.edge(adamicDefinitionKeywords, 13)
	description := generator.description()
	switch keyword {
	case "schema", "Schema":
		return description + generator.join(keyword, generator.directives(true), generator.operationTypes())
	case "scalar", "SCALAR":
		return description + generator.join(keyword, generator.name(), generator.directives(true))
	case "type", "Type", "interface":
		return description + generator.join(keyword, generator.name(), generator.implements(), generator.directives(true), generator.optional(70, generator.fields))
	case "union":
		return description + generator.join(keyword, generator.name(), generator.directives(true), generator.optional(70, func() string { return generator.join("=", generator.members("|")) }))
	case "enum":
		return description + generator.join(keyword, generator.name(), generator.directives(true), generator.optional(70, generator.enumValues))
	case "input":
		return description + generator.join(keyword, generator.name(), generator.directives(true), generator.optional(70, func() string { return generator.inputValues("{", "}") }))
	case "directive":
		tokens := []string{keyword, "@", generator.name(), generator.optional(40, func() string { return generator.inputValues("(", ")") }), generator.directives(true)}
		if generator.chance(30) {
			tokens = append(tokens, "repeatable")
		}
		tokens = append(tokens, "on")
		if generator.chance(25) {
			tokens = append(tokens, "|")
		}
		for count := 1 + generator.random.Intn(3); count > 0; count-- {
			if tokens[len(tokens)-1] != "on" && tokens[len(tokens)-1] != "|" {
				tokens = append(tokens, "|")
			}
			tokens = append(tokens, generator.edge(adamicLocations, 21))
		}
		return description + generator.join(tokens...)
	case "query", "mutation", "subscription", "Query":
		if generator.chance(20) {
			// A description here is one graphql-js refuses, on a shorthand query.
			if !generator.wild {
				description = ""
			}
			return description + generator.selectionSet()
		}
		return description + generator.join(keyword, generator.optional(60, generator.name), generator.variableDefinitions(), generator.directives(false), generator.selectionSet())
	case "fragment", "Fragment":
		return description + generator.join(keyword, generator.name(), generator.variableDefinitions(), "on", generator.name(), generator.directives(false), generator.selectionSet())
	case "extend", "EXTEND":
		// A description here is one graphql-js refuses, on an extension.
		if !generator.wild {
			description = ""
		}
		kind := generator.edge([]string{"schema", "scalar", "type", "interface", "union", "enum", "input", "directive", "thing"}, 8)
		tokens := []string{keyword, kind}
		switch kind {
		case "schema":
			tokens = append(tokens, generator.directives(true), generator.optional(50, generator.operationTypes))
		case "directive":
			tokens = append(tokens, "@", generator.name(), generator.directives(true))
		case "type", "interface":
			tokens = append(tokens, generator.name(), generator.implements(), generator.directives(true), generator.optional(50, generator.fields))
		case "union":
			tokens = append(tokens, generator.name(), generator.directives(true), generator.optional(50, func() string { return generator.join("=", generator.members("|")) }))
		case "enum":
			tokens = append(tokens, generator.name(), generator.directives(true), generator.optional(50, generator.enumValues))
		case "input":
			tokens = append(tokens, generator.name(), generator.directives(true), generator.optional(50, func() string { return generator.inputValues("{", "}") }))
		default:
			tokens = append(tokens, generator.name(), generator.directives(true))
		}
		return description + generator.join(tokens...)
	}
	return description + generator.join(keyword, generator.name(), generator.selectionSet())
}

// boundaries are the byte offsets where a character starts in a text, and its end.
func adamicBoundaries(text string) []int {
	var offsets []int
	for offset := range text {
		offsets = append(offsets, offset)
	}
	return append(offsets, len(text))
}

// document is one to three definitions; a wild one is then, more often than not, broken: cut short, or
// with a stray character, a piece of a string, or a separator put in at a character boundary.
func (generator *adamicGenerator) document() string {
	generator.wild = generator.chance(35)
	definitions := []string{}
	for count := 1 + generator.random.Intn(3); count > 0; count-- {
		definitions = append(definitions, generator.definition())
	}
	text := generator.separator() + generator.join(definitions...) + generator.separator()
	if generator.wildly(60) {
		boundaries := adamicBoundaries(text)
		at := boundaries[generator.random.Intn(len(boundaries))]
		switch generator.random.Intn(4) {
		case 0:
			text = text[:at]
		case 1:
			text = text[:at] + generator.pick(adamicStrays) + text[at:]
		case 2:
			text = text[:at] + generator.pick(adamicStringFaults) + text[at:]
		default:
			text = text[:at] + generator.pick(adamicSeparators) + text[at:]
		}
	}
	return text
}

// adamicDeepTexts nest each recursion of the parser hundreds deep, lists, objects, selection sets and
// list types, and each again broken at the bottom or with a closer missing, so a throw leaves through
// every frame of the recursion, each holding the nodes its level has made.
func adamicDeepTexts() []string {
	lists := "{ a(b: " + strings.Repeat("[", 500) + "1" + strings.Repeat("]", 500) + ") }"
	objects := "{ a(b: " + strings.Repeat("{c: ", 300) + "1" + strings.Repeat("}", 300) + ") }"
	selections := strings.Repeat("{ a ", 300) + strings.Repeat("}", 300)
	types := "query($a: " + strings.Repeat("[", 400) + "Int" + strings.Repeat("!]", 400) + ") { a }"
	return []string{
		lists, strings.Replace(lists, "]) }", ") }", 1), strings.Replace(lists, "1", `"never closed`, 1),
		objects, strings.Replace(objects, "}) }", ") }", 1), strings.Replace(objects, "1", `1.e`, 1),
		selections, selections[:len(selections)-1], strings.Replace(selections, "{ a }", "{ a(b: ?) }", 1),
		types, strings.Replace(types, "!]) {", "!) {", 1), strings.Replace(types, "Int", "Int @", 1),
	}
}

// adamicEscapedLine writes a text as one line of the cases file.
func adamicEscapedLine(text string) string {
	return ">" + strings.NewReplacer(`\`, `\\`, "\t", `\t`, "\n", `\n`, "\r", `\r`).Replace(text)
}

// adamicWritten is the output form: printable ASCII as itself but a backslash as two, and any other
// code point as \u{HEX}.
func adamicWritten(text string) string {
	var builder strings.Builder
	for _, character := range text {
		switch {
		case character == '\\':
			builder.WriteString(`\\`)
		case character >= 0x20 && character <= 0x7e:
			builder.WriteRune(character)
		default:
			fmt.Fprintf(&builder, `\u{%X}`, character)
		}
	}
	return builder.String()
}

// adamicUTF16Offsets maps every byte offset that starts a character, and the end, to its UTF-16 index.
func adamicUTF16Offsets(text string) []int {
	offsets := make([]int, len(text)+1)
	index := 0
	for position, character := range text {
		for byteIndex := position; byteIndex < position+utf8.RuneLen(character); byteIndex++ {
			offsets[byteIndex] = index
		}
		index += utf16.RuneLen(character)
	}
	offsets[len(text)] = index
	return offsets
}

// adamicOutline writes a tree as parser_test.go's outline does, with UTF-16 offsets and strings in the
// output form.
func adamicOutline(builder *strings.Builder, value any, offsets []int) {
	switch typed := value.(type) {
	case nil:
		builder.WriteString("undefined")
	case *estree.Node:
		fmt.Fprintf(builder, "%s[%d,%d]", typed.Type(), offsets[typed.Range[0]], offsets[typed.Range[1]])
		if keys := typed.Keys(); len(keys) > 0 {
			builder.WriteString("{")
			for index, key := range keys {
				if index > 0 {
					builder.WriteString(" ")
				}
				builder.WriteString(key + "=")
				adamicOutline(builder, typed.Get(key), offsets)
			}
			builder.WriteString("}")
		}
	case []*estree.Node:
		builder.WriteString("(")
		for index, child := range typed {
			if index > 0 {
				builder.WriteString(" ")
			}
			adamicOutline(builder, child, offsets)
		}
		builder.WriteString(")")
	case string:
		builder.WriteString(`"` + adamicWritten(typed) + `"`)
	case bool:
		builder.WriteString(strconv.FormatBool(typed))
	default:
		panic(fmt.Sprintf("an unexpected %T in a tree", typed))
	}
}

// adamicAnswer is what main.ts prints for one text.
func adamicAnswer(builder *strings.Builder, number int, text string) {
	fmt.Fprintf(builder, "case %d\n", number)
	document, comments, err := Parse(text)
	if err != nil {
		builder.WriteString("error " + adamicWritten(err.Error()) + "\n")
		return
	}
	offsets := adamicUTF16Offsets(text)
	adamicOutline(builder, document, offsets)
	builder.WriteString("\n")
	for _, comment := range comments {
		builder.WriteString("comment ")
		adamicOutline(builder, comment, offsets)
		builder.WriteString("\n")
	}
}

func TestAdamicPortCases(t *testing.T) {
	requestPath := os.Getenv("ADAMIC_PORT_REQUEST")
	if requestPath == "" {
		t.Skip("run by stage1/cohere/graphql/graphql_test.go")
	}
	encoded, err := os.ReadFile(requestPath)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicRequest
	if err := json.Unmarshal(encoded, &request); err != nil {
		t.Fatal(err)
	}

	texts := append(adamicCohereTexts(t), adamicDeepTexts()...)
	generator := &adamicGenerator{random: rand.New(rand.NewSource(request.Seed))}
	for count := 0; count < request.Generated; count++ {
		texts = append(texts, generator.document())
	}

	var cases, answers strings.Builder
	for number, text := range texts {
		if !utf8.ValidString(text) {
			t.Fatalf("case %d is not UTF-8, which the port reads only as Node decodes it: %q", number, text)
		}
		cases.WriteString(adamicEscapedLine(text) + "\n")
		adamicAnswer(&answers, number, text)
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0o644); err != nil {
		t.Fatal(err)
	}
}
