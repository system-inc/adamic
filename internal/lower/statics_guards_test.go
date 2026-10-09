package lower

import (
	"os"
	"path/filepath"
	"testing"

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

func TestPrivateAndPublicStaticsAgreeWithNode(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "class Box { static field = 11; static #secret = 29; static read(): number { return this.#secret; } } class Child extends Box {} console.log(`${Box.read()} ${Box.field} ${Child.field}`);")
}

func TestNamespaceFactoryBindingsAgreeWithNode(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "namespace Parser { const factory = { createNodeArray: (n: number): number => n + 1, createNumericLiteral: (n: number): number => n + 2 }; var { createNodeArray: factoryCreateNodeArray, createNumericLiteral: factoryCreateNumericLiteral } = factory; export function run(): void { console.log(`${factoryCreateNodeArray(10)} ${factoryCreateNumericLiteral(20)}`); } } Parser.run();")
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
	result := runAgreementNode(t, path)
	const want = "adamic: panic: read before assignment: field 'n' in new Box().n\n"
	if result.code != 70 || string(result.stderr) != want {
		t.Fatalf("read-before-assignment: exit=%d stdout=%q stderr=%q, want panic naming new Box().n", result.code, result.stdout, result.stderr)
	}
}
