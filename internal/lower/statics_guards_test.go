package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

// Keep acceptance checks from passing when lowering returns no program or body.
func requireLoweredOutput(t *testing.T, program *ir.Program) {
	t.Helper()
	if program == nil || len(program.Main) == 0 {
		t.Fatal("lowering returned no executable output")
	}
	writes := 0
	walk(program.Main, func(node any) bool {
		if _, ok := node.(ir.WriteLine); ok {
			writes++
		}
		return true
	})
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if _, ok := node.(ir.WriteLine); ok {
				writes++
			}
			return true
		})
	}
	if writes == 0 {
		t.Fatal("lowering lost every console write")
	}
}

type staticsNodeResult struct {
	stdout, stderr []byte
	exit           int
}

func runStaticsNode(t *testing.T, path string) staticsNodeResult {
	t.Helper()
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", runner, path)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	result := staticsNodeResult{stdout: stdout.Bytes(), stderr: stderr.Bytes()}
	if err != nil {
		var exited *exec.ExitError
		if !errors.As(err, &exited) || ctx.Err() != nil {
			t.Fatalf("Node execution: %v: %s", err, result.stderr)
		}
		result.exit = exited.ExitCode()
	}
	return result
}

// compiler/lower-agree was not pushed when this unit checked for it. This is
// the minimal source-versus-backend runner, following class_static_guard_test.go.
func staticsAgreeWithNode(t *testing.T, source, want string) {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "source.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	expected := runStaticsNode(t, path)
	if expected.exit != 0 || string(expected.stdout) != want {
		t.Fatalf("source Node: exit=%d stdout=%q stderr=%q", expected.exit, expected.stdout, expected.stderr)
	}
	program, err := lowerSource(t, source)
	if err != nil {
		t.Fatal(err)
	}
	requireLoweredOutput(t, program)
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	actual := runStaticsNode(t, generated)
	if actual.exit != expected.exit || !bytes.Equal(actual.stdout, expected.stdout) {
		t.Fatalf("lowered Node: exit=%d stdout=%q stderr=%q; source Node: exit=%d stdout=%q", actual.exit, actual.stdout, actual.stderr, expected.exit, expected.stdout)
	}
}

func TestPrivateAndPublicStaticsAgreeWithNode(t *testing.T) {
	t.Parallel()
	staticsAgreeWithNode(t, "class Box { static field = 11; static #secret = 29; static read(): number { return this.#secret; } } class Child extends Box {} console.log(`${Box.read()} ${Box.field} ${Child.field}`);", "29 11 11\n")
}

func TestNamespaceFactoryBindingsAgreeWithNode(t *testing.T) {
	t.Parallel()
	staticsAgreeWithNode(t, "namespace Parser { const factory = { createNodeArray: (n: number): number => n + 1, createNumericLiteral: (n: number): number => n + 2 }; var { createNodeArray: factoryCreateNodeArray, createNumericLiteral: factoryCreateNumericLiteral } = factory; export function run(): void { console.log(`${factoryCreateNodeArray(10)} ${factoryCreateNumericLiteral(20)}`); } } Parser.run();", "11 22\n")
}

// The definite-assignment assertion inserts an Adamic-owned panic. Source Node
// reads undefined here; run the generated program on Node to check its message.
func TestReadinessErrorIncludesReceiverExpression(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, "class Box { n!: number; } console.log(`${new Box().n + 1}`);")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	result := runStaticsNode(t, path)
	const want = "adamic: panic: read before assignment: field 'n' in new Box().n\n"
	if result.exit != 70 || string(result.stderr) != want {
		t.Fatalf("read-before-assignment: exit=%d stdout=%q stderr=%q, want panic naming new Box().n", result.exit, result.stdout, result.stderr)
	}
}
