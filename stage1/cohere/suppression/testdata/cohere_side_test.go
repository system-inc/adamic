// This file is not built here. suppression_test.go lays it into cohere's own package,
// cohere/internal/lint/suppression, with `go test -overlay`, so that it can call cohere's Build and its
// scanner exactly as cohere's tests do, without a byte of the submodule changing.
//
// It writes every case the port is asked, as the cases file main.ts reads, and Go cohere's answer to
// each, in the words main.ts prints. The sources are cohere's own: every source its tests in this
// package build an index of, read out of those tests' Go with go/ast (a strings.Join of string
// literals, or a string literal holding a comment or a newline), so the cases follow cohere's tests as
// they change. Then generated sources, from a seeded generator that writes lines of code with
// directive comments in every spelling, scope and shape, near misses, and the strings, templates and
// regular expressions a directive must not be read inside.

package suppression

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/lint/ecmascript/directives"
)

// adamicRequest is what suppression_test.go asks.
type adamicRequest struct {
	Seed      int64  `json:"seed"`
	Generated int    `json:"generated"`
	Cases     string `json:"cases"`
	Answers   string `json:"answers"`
}

// adamicSubjects are the rules registered as reporting on directives: the probe cohere's own
// directive_subject_test.go registers from init.
var adamicSubjects = []string{"probe/reports-on-directives"}

// adamicRules are the rule names every source is asked about, besides the names its directives write:
// the probe that reports on directives, an ordinary rule, a plugin-qualified name and its bare form, a
// suffix that doesn't fall on a `/`, a core rule a twin spelling names, and the empty name.
var adamicRules = []string{"probe/reports-on-directives", "probe/ordinary-rule", "nexus/consistency-no-enum", "consistency-no-enum", "no-enum", "no-use-before-define", ""}

// adamicCohereSources reads every source cohere's tests in this package build an index of out of
// their Go: each strings.Join of string literals, and each string literal holding a comment marker or
// a newline.
func adamicCohereSources(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	seen := map[string]bool{}
	add := func(source string) {
		if !seen[source] {
			seen[source] = true
			sources = append(sources, source)
		}
	}
	unquote := func(expression ast.Expr) (string, bool) {
		literal, isLiteral := expression.(*ast.BasicLit)
		if !isLiteral || literal.Kind != token.STRING {
			return "", false
		}
		text, err := strconv.Unquote(literal.Value)
		return text, err == nil
	}
	files := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "adamic_") {
			continue
		}
		file, err := parser.ParseFile(files, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.ImportSpec:
				return false
			case *ast.CallExpr:
				selector, isSelector := typed.Fun.(*ast.SelectorExpr)
				if !isSelector || selector.Sel.Name != "Join" || len(typed.Args) != 2 {
					return true
				}
				list, isList := typed.Args[0].(*ast.CompositeLit)
				separator, isString := unquote(typed.Args[1])
				if !isList || !isString {
					return true
				}
				var parts []string
				for _, element := range list.Elts {
					part, isPart := unquote(element)
					if !isPart {
						return true
					}
					parts = append(parts, part)
				}
				add(strings.Join(parts, separator))
			case *ast.BasicLit:
				if text, isString := unquote(typed); isString && (strings.Contains(text, "/") || strings.Contains(text, "\n")) {
					add(text)
				}
			}
			return true
		})
	}
	return sources
}

// adamicEdgeCode is code at the edges of the scanner's classes: every character after which a `/` opens
// a pattern, and characters after which it divides, each before `/x/`, some with a space the scanner
// skips (space, tab) or doesn't (vertical tab, form feed, no-break space) between them; and strings and
// templates that end at their line's end or at an escaped quote. A pattern `/ //x/` is read as a
// pattern on one side of a class's edge and as a division and a line comment on the other, so the
// comments found differ, where `/x/` lexes alike either way.
var adamicEdgeCode = func() []string {
	var pieces []string
	for _, before := range []string{"", "(", ",", "=", ":", "[", "!", "&", "|", "?", "{", "}", ";", "+", "-", "*", "~", "^", "<", ">", "%",
		")", "]", "a", "0", "\"\"", "''", "``", "."} {
		for _, between := range []string{"", " ", "\t", "\v", "\f", "\u00a0"} {
			pieces = append(pieces, before+between+"/x/ ", before+between+"/ //x/ ")
		}
	}
	return append(pieces, "'a\\'", "\"a\\\"", "`a\\``", "'\n", "/[\\]/]/", "/a\\/b/")
}()

// adamicGenerated writes a source of a few lines, each some code and perhaps a comment, most of them
// directives.
func adamicGenerated(random *rand.Rand) string {
	pick := func(choices ...string) string {
		return choices[random.Intn(len(choices))]
	}
	code := []string{
		"const a = 1;", "x", "x;", "a / b", "a / b / c", "(", ")", "{", "}", "= ", ", ", ";", "return ", "enum E {}", " ", "\t",
		`"/* eslint-disable */"`, `'// eslint-disable-line'`, `"a\"b"`, `'it\'s'`, `"unterminated`, `'\\'`,
		"`a ${", "}`", "`x`", "`${ {a: 1}.a }`", "`${'}'}`", "`${`${b}`}`", "`// eslint-disable`", "`${ /* eslint-disable-line */ a }`",
		"/[/*]/", "/a\\/b/", "= /x/g", "(/=/)", "x = /* c */ /re/", "a++ / 2",
		"\u00e9", "\U0001f600", "\u0085", "\ufeff", "\u00a0", "\u2028", "\u3000",
	}
	var lines []string
	for range 1 + random.Intn(7) {
		var line strings.Builder
		for range random.Intn(4) {
			if random.Intn(5) == 0 {
				line.WriteString(adamicEdgeCode[random.Intn(len(adamicEdgeCode))])
			} else {
				line.WriteString(code[random.Intn(len(code))])
			}
		}
		// Most lines carry a comment, and some two, so a block and its enable can share a line.
		comments := 0
		if random.Intn(10) < 7 {
			comments = 1 + random.Intn(2)
		}
		for range comments {
			opener := pick("// ", "//", "/* ", "/*", "{/* ", "/** ")
			line.WriteString(opener)
			line.WriteString(pick("", "", " ", "\t", "\u00a0", "\u0085", "\ufeff", "\u2028", "\u3000"))
			line.WriteString(pick("eslint-disable", "eslint-disable-next-line", "eslint-disable-line", "eslint-enable",
				"cohere-disable", "cohere-disable-next-line", "cohere-enable", "verify-disable-line", "oxlint-disable-next-line",
				"oxlint-enable", "eslint-disabled", "eslint-disable-lines", "eslint-disable-next-line,", "eslint-enabled", "not a directive", "Eslint-disable",
				"ESLINT-DISABLE-LINE", "eslint-DISABLE", "eslint-disable-NEXT-LINE", "Cohere-enable"))
			// JavaScript whitespace ends the directive word; U+0085 does not.
			if random.Intn(4) == 0 {
				line.WriteString(pick("\v", "\f", "\u00a0", "\u2003", "\u3000", "\u0085"))
			}
			for range random.Intn(3) {
				line.WriteString(pick(" no-console", " nexus/consistency-no-enum", " consistency-no-enum", ",", ", ", " probe/ordinary-rule",
					" probe/reports-on-directives", " no-use-before-define", "\tno-enum", " ,", " a,,b"))
			}
			if random.Intn(2) == 0 {
				line.WriteString(pick(" -- why", " --", "-- because", " -- a -- b", " -- reason\u0085", " -- \ufeffreason\ufeff", " --\u3000why"))
			}
			if strings.HasPrefix(opener, "/*") || strings.HasPrefix(opener, "{/*") {
				if random.Intn(4) == 0 {
					line.WriteString(pick("\n * continued", "\n *   more reason", "\n no-console\n"))
				}
				closer := pick(" */", "*/", " */", "")
				if opener == "{/* " && closer != "" {
					closer += "}"
				}
				line.WriteString(closer)
			}
			line.WriteString(pick(" ", "", "x; "))
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, pick("\n", "\n", "\r\n"))
}

// TestAdamicPortCases writes the cases file and Go cohere's answers.
// Not parallel: registers subjects in the shared directives registry and writes environment-request output paths.
func TestAdamicPortCases(t *testing.T) {
	path := os.Getenv("ADAMIC_PORT_REQUEST")
	if path == "" {
		t.Skip("run by stage1/cohere/suppression/suppression_test.go in the adamic repository")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var request adamicRequest
	if err := json.Unmarshal(contents, &request); err != nil {
		t.Fatal(err)
	}

	sources := adamicCohereSources(t)
	cohereCount := len(sources)
	if cohereCount < 40 {
		t.Fatalf("read %d sources out of cohere's tests, where there are over 40: the reading has stopped working", cohereCount)
	}
	// Sources that don't rest on the generator's draw: a block and its enable on one line, before code;
	// a directive word followed by whitespace other than a space or a tab; and a pattern after a
	// vertical tab, which the scanner doesn't skip.
	sources = append(sources,
		"x; // eslint-disable-line\u00a0no-x\ny; // eslint-disable-line\vno-x\nz; // eslint-disable-line\tno-x",
		"(\v/ // x/ // eslint-disable-line\n(/ // x/ // eslint-disable-line",
		"/* eslint-disable */ /* eslint-enable */\nenum A {}",
		"enum A {}\n/* eslint-disable no-enum */ x; /* eslint-enable no-enum */ /* eslint-disable-line -- why */\nenum B {}",
	)
	random := rand.New(rand.NewSource(request.Seed))
	for range request.Generated {
		sources = append(sources, adamicGenerated(random))
	}

	escape := strings.NewReplacer("\\", "\\\\", "\t", "\\t", "\n", "\\n", "\r", "\\r")
	var cases, answers strings.Builder
	for _, subject := range adamicSubjects {
		directives.RegisterSubject(subject)
		fmt.Fprintf(&cases, "subject\t%s\n", escape.Replace(subject))
	}
	for number, source := range sources {
		if !utf8.ValidString(source) {
			t.Fatalf("%q is not UTF-8, which the port reads its cases as", source)
		}
		// toUnits is a byte offset's UTF-16 index; past the end, one unit a byte.
		units := map[int]int{}
		unit := 0
		for offset, character := range source {
			units[offset] = unit
			if character >= 0x10000 {
				unit += 2
			} else {
				unit++
			}
		}
		toUnits := func(offset int) int {
			if index, isStart := units[offset]; isStart {
				return index
			}
			if offset < 0 {
				return offset
			}
			if offset >= len(source) {
				return unit + offset - len(source)
			}
			t.Fatalf("%d is inside a character of %q", offset, source)
			return 0
		}

		fmt.Fprintf(&cases, "source\t%s\n", escape.Replace(source))
		fmt.Fprintf(&answers, "case %d\n", number)
		index := Build(source)
		comments := scanComments(source)
		for _, comment := range comments {
			scopeWord, honored := directives.Recognize(source[comment.pos:comment.end])
			if !honored {
				scopeWord = "-"
			}
			fmt.Fprintf(&answers, "comment %d %d %s\n", toUnits(comment.pos), toUnits(comment.end), scopeWord)
		}
		for position, directive := range index.Directives() {
			rules := make([]string, len(directive.Rules))
			for ruleIndex, rule := range directive.Rules {
				rules[ruleIndex] = adamicQuote(rule)
			}
			fmt.Fprintf(&answers, "directive %d %s rules=%s reason=%s line=%d start=%d end=%d pos=%d end=%d\n", position, directive.Kind,
				strings.Join(rules, ","), adamicQuote(directive.Reason), directive.Line, directive.StartLine, directive.EndLine,
				toUnits(directive.Pos), toUnits(directive.End))
		}
		rules := slices.Clone(adamicRules)
		for _, reference := range index.RuleReferences() {
			fmt.Fprintf(&answers, "reference %s %d %d\n", adamicQuote(reference.Name), toUnits(reference.Pos), toUnits(reference.End))
			if !slices.Contains(rules, reference.Name) && len(rules) < len(adamicRules)+3 && !strings.ContainsAny(reference.Name, "\t\n\r\\") {
				rules = append(rules, reference.Name)
			}
		}

		// The questions, in order: each rule at the start of every line, at every comment, and past the
		// end; and the line of a few offsets between them.
		offsets := []int{0}
		for position := range len(source) {
			if source[position] == '\n' {
				offsets = append(offsets, position+1)
			}
		}
		for _, comment := range comments {
			offsets = append(offsets, comment.pos, comment.end)
		}
		offsets = append(offsets, len(source), len(source)+3)
		for _, offset := range offsets {
			for _, rule := range rules {
				fmt.Fprintf(&cases, "query\t%s\t%d\n", escape.Replace(rule), toUnits(offset))
				suppressed := "0"
				if index.Suppresses(rule, offset) {
					suppressed = "1"
				}
				fmt.Fprintf(&answers, "query %s %d %s\n", rule, toUnits(offset), suppressed)
			}
		}
		for _, offset := range []int{-1, 0, len(source) / 2, len(source), len(source) + 3} {
			for offset > 0 && offset < len(source) && !utf8.RuneStart(source[offset]) {
				offset--
			}
			fmt.Fprintf(&cases, "lineof\t%d\n", toUnits(offset))
			fmt.Fprintf(&answers, "lineof %d %d\n", toUnits(offset), index.LineOf(offset))
		}

		counts := make([]string, len(index.Directives()))
		for position := range index.Directives() {
			counts[position] = strconv.Itoa(index.AppliedCount(position))
		}
		fmt.Fprintf(&answers, "applied %s\n", strings.Join(counts, " "))
		fmt.Fprintf(&answers, "total %d\n", index.TotalApplied())
		fmt.Fprintf(&answers, "unused %s\n", adamicIndexes(index.Directives(), index.Unused()))
		fmt.Fprintf(&answers, "without-reason %s\n", adamicIndexes(index.Directives(), index.WithoutReason()))
	}
	if err := os.WriteFile(request.Cases, []byte(cases.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(request.Answers, []byte(answers.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d sources read out of cohere's tests, %d generated", cohereCount, request.Generated)
}

// adamicIndexes is the indexes, in all, of the directives in some, joined by spaces.
func adamicIndexes(all []*Directive, some []*Directive) string {
	indexes := make([]string, len(some))
	for position, directive := range some {
		indexes[position] = strconv.Itoa(slices.Index(all, directive))
	}
	return strings.Join(indexes, " ")
}

// adamicQuote is main.ts's quote.
func adamicQuote(text string) string {
	var builder strings.Builder
	builder.WriteByte('"')
	for _, character := range text {
		switch {
		case character == '\\':
			builder.WriteString("\\\\")
		case character == '"':
			builder.WriteString("\\\"")
		case character == '\n':
			builder.WriteString("\\n")
		case character == '\r':
			builder.WriteString("\\r")
		case character == '\t':
			builder.WriteString("\\t")
		case character < 0x20 || character == 0x7f:
			fmt.Fprintf(&builder, "\\u%04x", character)
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}
