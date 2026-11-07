package lower

import (
	"context"
	"errors"
	"os/exec"
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
	var refused *Refused
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
