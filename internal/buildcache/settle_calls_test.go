package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// settleCallsSum is testdata/settle-calls.txt's sha256. Loom's planner/testdata/settle-calls.txt holds the same bytes,
// and loom's coverage tests hold every name in it to what its BPF tracer records (#9tr6c6j): a call Settle starts
// parsing changes this list, this sum and loom's copy together, or a read in it is invisible to every BPF trace.
const settleCallsSum = "3b1429cee2b83bd711bb4bbfc24a914c1defc5e39715fa987ece01b8a502eb63"

// pinnedSettleCalls is the pinned list's names.
func pinnedSettleCalls(t *testing.T) []string {
	t.Helper()
	text, err := os.ReadFile("testdata/settle-calls.txt")
	if err != nil {
		t.Fatal(err)
	}
	if sum := sha256.Sum256(text); hex.EncodeToString(sum[:]) != settleCallsSum {
		t.Fatalf("testdata/settle-calls.txt's sha256 is %x, pinned %s: loom's copy changes with it", sum, settleCallsSum)
	}
	var names []string
	for _, line := range strings.Split(string(text), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			names = append(names, line)
		}
	}
	return names
}

// settleParsedCalls is every call name this package matches, in each of its non-test Go files, by the rule loom's
// coverage tests use on its own reader: the string literals in the value of each var named *Calls, those a call is
// compared with by == or != (call, or x.call), the case lists of a switch on a call, and the prefixes
// strings.HasPrefix tests a call against. It reads what the code matches, never a comment, and a parse added in any
// file of the package is seen.
func settleParsedCalls(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob("*.go")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no Go files here: %v", err)
	}
	names := map[string]bool{}
	literal := func(expression ast.Expr) {
		if basic, ok := expression.(*ast.BasicLit); ok && basic.Kind == token.STRING {
			value, err := strconv.Unquote(basic.Value)
			if err != nil {
				t.Fatal(err)
			}
			names[value] = true
		}
	}
	isCall := func(expression ast.Expr) bool {
		switch node := expression.(type) {
		case *ast.Ident:
			return node.Name == "call"
		case *ast.SelectorExpr:
			return node.Sel.Name == "call"
		}
		return false
	}
	inspect := func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.GenDecl:
			if node.Tok != token.VAR {
				return true
			}
			for _, spec := range node.Specs {
				value := spec.(*ast.ValueSpec)
				for index, name := range value.Names {
					if strings.HasSuffix(name.Name, "Calls") && index < len(value.Values) {
						ast.Inspect(value.Values[index], func(inner ast.Node) bool {
							if expression, ok := inner.(ast.Expr); ok {
								literal(expression)
							}
							return true
						})
					}
				}
			}
		case *ast.SwitchStmt:
			if node.Tag != nil && isCall(node.Tag) {
				for _, clause := range node.Body.List {
					for _, expression := range clause.(*ast.CaseClause).List {
						literal(expression)
					}
				}
			}
		case *ast.BinaryExpr:
			if node.Op == token.EQL || node.Op == token.NEQ {
				if isCall(node.X) {
					literal(node.Y)
				}
				if isCall(node.Y) {
					literal(node.X)
				}
			}
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "HasPrefix" && len(node.Args) == 2 && isCall(node.Args[0]) {
				literal(node.Args[1])
			}
		}
		return true
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, inspect)
	}
	for _, calls := range []map[string]bool{atCalls, plainCalls, statCalls, writeCalls, forkCalls} {
		for name := range calls {
			if !names[name] {
				t.Fatalf("settle.go's sets hold %s and the package's syntax didn't yield it: the rule reads less than the code", name)
			}
		}
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)
	return sorted
}

// Settle parses exactly the pinned calls: one it gains is red until the list, its sum and loom's copy move with it, so
// loom's tracer is held to it; one it drops is red too, so the list never names a call no reader parses.
func TestSettleParsesExactlyThePinnedCalls(t *testing.T) {
	t.Parallel()
	parsed, pinned := settleParsedCalls(t), pinnedSettleCalls(t)
	for _, name := range parsed {
		if !slices.Contains(pinned, name) {
			t.Errorf("Settle parses %s, and testdata/settle-calls.txt doesn't list it: pin it here and in loom's copy, whose tests hold its tracer to it", name)
		}
	}
	for _, name := range pinned {
		if !slices.Contains(parsed, name) {
			t.Errorf("testdata/settle-calls.txt lists %s, and Settle doesn't parse it", name)
		}
	}
}

// A namedCallExclusion is a string literal in this package's non-test files that spells a call strace 6.8's x86_64
// table knows and is no parse of a trace line: how many times its file holds it, in the same enclosing expression, and
// what it is.
type namedCallExclusion struct {
	count  int
	reason string
}

// namedCallExclusions are those literals, by file, the call they name, and the source text of the expression around
// each (Tests' ruling on 873f35e1: what the literal is, never the line it sits on, so an edit above it moves nothing).
// Every other such literal is one Settle parses, so it's in testdata/settle-calls.txt.
var namedCallExclusions = map[string]namedCallExclusion{
	`buildcache.go "read" in os.Getenv("ADAMIC_BUILD_CACHE") == "read"`: {1, "ADAMIC_BUILD_CACHE=read, the cache mode that reads products and never builds them"},
	`reads.go "read" in os.Getenv("ADAMIC_BUILD_CACHE") == "read"`:      {1, "ADAMIC_BUILD_CACHE=read, readMode's test of the same mode"},
}

// Every string literal in this package's non-test files (import paths aside) that names a call strace 6.8's x86_64
// table knows, whole or as "name(" the way strace's line spells it (namedCalls), is either one the pinned list holds or
// a pinned exclusion, by file and enclosing expression, counted. It's the proof under the shape reader above (Tests'
// and Node's rule, Oct 17): a parse that names its call is seen whatever carries the name, a helper's parameter of any
// name, a map that isn't a *Calls, a switch, a table of its own. Its one blind spot is a call name the code builds
// rather than writes, "open" + "at", which no literal spells. An exclusion whose expression is gone, or held a different
// number of times, is red too, so the list stays exactly the literals there are.
func TestEveryCallNamedInTheReaderIsPinned(t *testing.T) {
	t.Parallel()
	table, err := os.ReadFile("cmd/traced/testdata/strace-6.8-x86_64.txt")
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, line := range strings.Split(string(table), "\n") {
		if fields := strings.Fields(line); len(fields) >= 2 && !strings.HasPrefix(line, "#") {
			known[fields[1]] = true
		}
	}
	pinned := pinnedSettleCalls(t)
	paths, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		imports := map[*ast.BasicLit]bool{}
		for _, spec := range file.Imports {
			imports[spec.Path] = true
		}
		// The nodes enclosing the one inspected, outermost first: a literal's enclosing expression is the nearest.
		var stack []ast.Node
		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			stack = append(stack, node)
			basic, ok := node.(*ast.BasicLit)
			if !ok || basic.Kind != token.STRING || imports[basic] {
				return true
			}
			value, err := strconv.Unquote(basic.Value)
			if err != nil {
				return true
			}
			enclosing := ast.Node(basic)
			if len(stack) > 1 {
				enclosing = stack[len(stack)-2]
			}
			text := string(source[files.Position(enclosing.Pos()).Offset:files.Position(enclosing.End()).Offset])
			for _, name := range namedCalls(value) {
				if !known[name] || slices.Contains(pinned, name) {
					continue
				}
				at := fmt.Sprintf("%s %q in %s", path, name, text)
				if _, excluded := namedCallExclusions[at]; !excluded {
					t.Errorf("%s:%d names a call strace traces (%s), and testdata/settle-calls.txt doesn't pin it: pin it (and loom's copy) if Settle parses it, or add it to namedCallExclusions with what it is", path, files.Position(basic.Pos()).Line, at)
				}
				seen[at]++
			}
			return true
		})
	}
	for at, exclusion := range namedCallExclusions {
		if seen[at] != exclusion.count {
			t.Errorf("namedCallExclusions holds %s %d times, and the package holds it %d: an exclusion is exactly the literals there are", at, exclusion.count, seen[at])
		}
	}
}

// callInLine is a call as strace's line spells it, a name directly followed by its opening parenthesis, at a word
// boundary: "fanotify_mark(" in a literal a parse matches line text with.
var callInLine = regexp.MustCompile(`(?:^|[^A-Za-z0-9_])([a-z0-9_]+)\(`)

// namedCalls are the call names a string literal spells: the whole literal, and every name in it followed by "(",
// as strace writes a call in its line (Tests' mutant on fb94405b4c: a parse by HasPrefix(call.arguments,
// "fanotify_mark(")).
func namedCalls(value string) []string {
	names := []string{value}
	for _, match := range callInLine.FindAllStringSubmatch(value, -1) {
		names = append(names, match[1])
	}
	return names
}
