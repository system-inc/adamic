package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Exercise the proof directly: no unrelated subset refusal can mask an
// invalid flow fact or a cycle in the proposed helper summary.
func checkOverloadBodyObligation(t *testing.T, body string, want bool, extra ...string) {
	t.Helper()
	const prefix = `interface Result<T> { readonly value: T; }
interface Template { readonly kind: 'template'; readonly text: string; }
interface Numeric { readonly kind: 'numeric'; readonly text: string; }
type Input = Template | Numeric;
function evaluate(input: Template, flag: boolean): Result<string>;
function evaluate(input: Input, flag: boolean): Result<string | number>;
function evaluate(input: Input, flag: boolean): Result<string | number> | undefined {
`
	path := filepath.Join(t.TempDir(), "proof.a")
	if err := os.WriteFile(path, []byte(strings.Join(extra, "\n")+"\n"+prefix+body+"\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked, result: &ir.Program{}}
	var implementation, overload *ast.Node
	for _, declaration := range file.Statements.Nodes {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Name().Text() == "evaluate" {
			if declaration.Body() != nil {
				implementation = declaration
			} else if overload == nil {
				overload = declaration
			}
		}
	}
	got, failure := l.overloadBodyResult(implementation, overload)
	if got != want {
		t.Fatalf("proof=%t, want %t; %s", got, want, failure)
	}
}

func TestOverloadBodyObligationFallthrough(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `if (flag) return { value: 'yes' };`, false)
}

func TestOverloadBodyObligationReboundParameter(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `input = { kind: 'numeric', text: 'bad' };
return { value: 1 };`, false)
}

func TestOverloadBodyObligationCapturedWrite(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `let result: Result<string | number> = { value: 'yes' };
const mutate = (): void => { result = { value: 1 }; };
mutate(); return result;`, false)
}

func TestOverloadBodyObligationUnknownCondition(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `const test = (): boolean => true;
if (test()) return { value: 'yes' };
return { value: 1 };`, false)
}

func TestOverloadBodyObligationNumericBranch(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `if (flag) return { value: 1 };
return { value: 'yes' };`, false)
}

func TestOverloadBodyObligationSwitchFallthrough(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `switch (input.kind) {
case 'template':
case 'numeric': return { value: 1 };
}`, false)
}

func TestOverloadBodyObligationSwitchBreak(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `switch (input.kind) {
case 'template': break;
case 'numeric': return { value: 1 };
}
return { value: 'yes' };`, true)
}

func TestOverloadBodyObligationRecursiveHelper(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `function recurse(): boolean { return recurse(); }
if (recurse()) return { value: 'yes' };
return { value: 1 };`, false)
}

func TestOverloadBodyObligationUnknownArgument(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `const unknown = (): boolean => true;
function identity(value: boolean): boolean { return value; }
if (identity(unknown())) return { value: 'yes' };
return { value: 1 };`, false)
}

func TestOverloadBodyObligationHiddenAccessor(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `if (input.kind === 'template') return { value: 'yes' };
return { value: 1 };`, false, `class Hidden { get kind(): 'template' { return 'template'; } }`)
}

func TestOverloadBodyObligationEarlyThrow(t *testing.T) {
	t.Parallel()
	checkOverloadBodyObligation(t, `if (flag) {
throw new Error('stop');
return { value: 1 };
}
return { value: 'yes' };`, true)
}
