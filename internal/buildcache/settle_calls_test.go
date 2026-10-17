package buildcache

import (
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// settleCallsSum is testdata/settle-calls.txt's sha256. Loom's planner/testdata/settle-calls.txt holds the same bytes,
// and loom's coverage tests hold every name in it to what its BPF tracer records (#9tr6c6j): a call Settle starts
// parsing changes this list, this sum and loom's copy together, or a read in it is invisible to every BPF trace.
const settleCallsSum = "bcaba1c42667fb4aa6cc2af30b55105ee35e7895c32e5e167be2875af0174c0d"

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

// settleParsedCalls is every call name settle.go matches, by the rule loom's coverage tests use on its own reader: the
// string literals in the value of each var named *Calls, those a call is compared with by == or != (call, or x.call),
// and the prefixes strings.HasPrefix tests a call against. It reads what the code matches, never a comment.
func settleParsedCalls(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "settle.go", nil, 0)
	if err != nil {
		t.Fatal(err)
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
	ast.Inspect(file, func(node ast.Node) bool {
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
	})
	for _, calls := range []map[string]bool{atCalls, plainCalls, statCalls, writeCalls, forkCalls} {
		for name := range calls {
			if !names[name] {
				t.Fatalf("settle.go's sets hold %s and its syntax didn't yield it: the rule reads less than the code", name)
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
