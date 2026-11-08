package lower

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestProcessOptionalStdoutHandleSysInitializer(t *testing.T) {
	t.Parallel()
	source := "import 'node:process';\nconst handle = (process.stdout as any)?._handle as { setBlocking?: (value: boolean) => void; };\n"
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	handles := 0
	walk(program.Main, func(value any) bool {
		if call, ok := value.(ir.ProcessCall); ok && call.Operation == "handle" {
			handles++
		}
		return true
	})
	if handles != 1 {
		t.Fatalf("sys.ts initializer needs one stdout handle operation, got %d", handles)
	}
}

func TestProcessOptionalAbsentReceiverStopsByName(t *testing.T) {
	t.Parallel()
	const source = "import 'node:process';\nconst value = process.stdout.columns?.toFixed();\n"
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	func() {
		checked, release := program.Checker(context.Background(), file)
		defer release()
		l := &lowering{program: program, checker: checked}
		links := 0
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().QuestionDotToken != nil {
				links++
				receiver := node.AsPropertyAccessExpression().Expression
				declared := checked.GetTypeOfSymbol(l.symbol(receiver))
				if declared.Flags()&checker.TypeFlagsUnion == 0 || checked.GetNonNullableType(declared) == declared {
					t.Fatal("host axiom must leave columns optional")
				}
				if path := l.processPath(node); path != "" {
					t.Fatalf("absent receiver must stop processPath, got %q", path)
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if links != 1 {
			t.Fatalf("want one absent-receiver optional link, got %d", links)
		}
	}() // Return the borrowed checker before Lower acquires it again.
	_, err = Lower(context.Background(), program)
	var refused *NotYet
	const want = "main.ts:2:15: stage 0 can't lower a call through ?. (an optional call) yet"
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), "/"+want) {
		t.Fatalf("want absent receiver refusal %q, got %v", want, err)
	}
}

func processOptionalEverywhereMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "process-optional-everywhere.test", "./internal/lower", nil, []buildcache.Mutation{{
		File:   "internal/lower/process.go",
		Before: "func (l *lowering) processPathPresent(node *ast.Node, path string) bool {\n",
		After:  "func (l *lowering) processPathPresent(node *ast.Node, path string) bool {\n return true // mutant: every receiver is present\n",
	}})
}
func TestProduct_ProcessOptionalEverywhereMutant(t *testing.T) {
	t.Parallel()
	processOptionalEverywhereMutant(t)
}
func TestProcessOptionalEverywhereMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := processOptionalEverywhereMutant(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run", "^TestProcessOptionalAbsentReceiverStopsByName$", "-test.count=1", "-test.v")
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	const line = `absent receiver must stop processPath, got "process.stdout.columns.toFixed"`
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), line) {
		t.Fatalf("optional-everywhere mutant: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Log("TestProcessOptionalAbsentReceiverStopsByName goes red: " + line)
}

// The present host receiver is uncast process, and both reads use processPath.
func TestProcessOptionalPresentReadMatchesDot(t *testing.T) {
	t.Parallel()
	dot, err := lowerSourceAs(t, "import 'node:process';\nconst value = process.execArgv;\n", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	optional, err := lowerSourceAs(t, "import 'node:process';\nconst value = process?.execArgv;\n", ".ts")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dot.Main, optional.Main) {
		t.Fatalf("present optional read must have dot IR: dot=%#v optional=%#v", dot.Main, optional.Main)
	}
	if len(dot.Main) != 1 {
		t.Fatalf("want one declaration, got %#v", dot.Main)
	}
	declaration, ok := dot.Main[0].(ir.Declare)
	if !ok {
		t.Fatalf("want a declaration, got %#v", dot.Main[0])
	}
	call, ok := declaration.Value.(ir.ProcessCall)
	if !ok || call.Operation != "execArgv" || call.Of != ir.Array {
		t.Fatalf("want the real execArgv host read, got %#v", declaration.Value)
	}
}

func TestProcessOptionalNullableReceiverStopsByName(t *testing.T) {
	t.Parallel()
	const source = "import 'node:process';\nconst value = process.exitCode?.toString();\n"
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	func() {
		checked, release := program.Checker(t.Context(), file)
		defer release()
		l := &lowering{program: program, checker: checked}
		links := 0
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().QuestionDotToken != nil {
				links++
				symbol := l.symbol(node.AsPropertyAccessExpression().Expression)
				if symbol == nil || nodeLibraryName(symbol) != "node:process.Process.exitCode" || !implementedNodeMembers[nodeLibraryName(symbol)] {
					t.Fatal("exitCode must be the registered Node member")
				}
				if symbol.Flags&ast.SymbolFlagsOptional != 0 {
					t.Fatal("exitCode must have no optional flag")
				}
				declared := checked.GetTypeOfSymbol(symbol)
				if declared.Flags()&checker.TypeFlagsUnion == 0 || checked.GetNonNullableType(declared) == declared {
					t.Fatal("exitCode's required declaration must remain nullable")
				}
				if path := l.processPath(node); path != "" {
					t.Fatalf("nullable receiver must stop processPath, got %q", path)
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if links != 1 {
			t.Fatalf("want one nullable-receiver link, got %d", links)
		}
	}()
	_, err = Lower(t.Context(), program)
	var refused *NotYet
	const want = "main.ts:2:15: stage 0 can't lower toString on a union of differently held members (function source text and mixed value dispatch are not represented) yet"
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), "/"+want) {
		t.Fatalf("want nullable receiver refusal %q, got %v", want, err)
	}
}

// load.Load uses ExactOptionalPropertyTypes. GetTypeOfSymbol still includes the
// missing type of an optional member, and GetNonNullableType removes it. Walk
// through unregistered intermediates too, including indexed property values.
func TestProcessOptionalFlagIsRedundantForReachableMembers(t *testing.T) {
	t.Parallel()
	const source = "import 'node:process';\nconst root = process;\nconst columns = process.stdout.columns;\n"
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(t.Context(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	var root *ast.Node
	var columns *ast.Symbol
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindVariableDeclaration {
			switch node.Name().Text() {
			case "root":
				root = node.AsVariableDeclaration().Initializer
			case "columns":
				columns = l.symbol(node.AsVariableDeclaration().Initializer)
			}
		}
		return node.ForEachChild(visit)
	}
	file.AsNode().ForEachChild(visit)
	if root == nil || columns == nil || nodeLibraryName(columns) != "node:tty.WriteStream.columns" || columns.Flags&ast.SymbolFlagsOptional == 0 {
		t.Fatal("the corrected columns declaration must be optional")
	}
	optional := map[*ast.Symbol]bool{}
	check := func(symbol *ast.Symbol, path string) {
		if symbol.Flags&ast.SymbolFlagsOptional == 0 || optional[symbol] {
			return
		}
		optional[symbol] = true
		declared := checked.GetTypeOfSymbol(symbol)
		if checked.GetNonNullableType(declared) == declared {
			t.Fatalf("optional member %s (%s) is already non-nullable: %s", path, nodeLibraryName(symbol), checked.TypeToString(declared))
		}
		t.Logf("%s: declared %s is rejected by the non-nullable comparison", path, checked.TypeToString(declared))
	}
	// columns is bound by its dedicated adapter, not the implementation registry.
	// Keep it as an explicit non-vacuous host-axiom witness as well.
	check(columns, "process.stdout.columns")
	seen := map[*checker.Type]bool{}
	registered := map[string]bool{}
	var walk func(*checker.Type, string)
	walk = func(proven *checker.Type, path string) {
		proven = checked.GetNonNullableType(proven)
		if seen[proven] {
			return
		}
		seen[proven] = true
		for _, symbol := range checked.GetPropertiesOfType(proven) {
			memberPath := path + "." + symbol.Name
			if implementedNodeMembers[nodeLibraryName(symbol)] {
				registered[nodeLibraryName(symbol)] = true
				check(symbol, memberPath)
			}
			walk(checked.GetTypeOfSymbol(symbol), memberPath)
		}
		for _, info := range checked.GetIndexInfosOfType(proven) {
			walk(info.ValueType(), path+"[key]")
		}
	}
	walk(checked.GetTypeAtLocation(root), "process")
	if len(registered) == 0 || len(optional) == 0 {
		t.Fatal("the equivalence audit must check registered members and the optional columns witness")
	}
	t.Logf("checked %d registered member declarations and %d optional declarations including columns", len(registered), len(optional))
}

func processOptionalNullableMutant(t *testing.T) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "process-optional-nullable.test", "./internal/lower", nil, []buildcache.Mutation{{
		File:   "internal/lower/process.go",
		Before: " && l.checker.GetNonNullableType(declared) == declared",
		After:  "",
	}})
}
func TestProduct_ProcessOptionalNullableMutant(t *testing.T) {
	t.Parallel()
	processOptionalNullableMutant(t)
}
func TestProcessOptionalNullableMutantIsCaught(t *testing.T) {
	t.Parallel()
	binary := processOptionalNullableMutant(t)
	for _, pin := range []struct{ name, catch string }{
		{"TestProcessOptionalNullableReceiverStopsByName", `nullable receiver must stop processPath, got "process.exitCode.toString"`},
		{"TestProcessOptionalAbsentReceiverStopsByName", ""},
	} {
		ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
		command := exec.CommandContext(ctx, binary, "-test.run", "^"+pin.name+"$", "-test.count=1", "-test.v")
		stdout, stderr, err := childguard.Capture(ctx, command)
		cancel()
		if pin.catch != "" {
			var exited *exec.ExitError
			if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), pin.catch) {
				t.Fatalf("nullable mutant must red %s: err=%v stdout=%s stderr=%s", pin.name, err, stdout, stderr)
			}
			t.Log(pin.name + " goes red: " + pin.catch)
		} else {
			if err != nil || !strings.Contains(string(stdout), "--- PASS: "+pin.name) || len(stderr) != 0 {
				t.Fatalf("nullable mutant must spare %s: err=%v stdout=%s stderr=%s", pin.name, err, stdout, stderr)
			}
			t.Log("same nullable mutant binary spares " + pin.name + ": exit 0")
		}
	}
}
