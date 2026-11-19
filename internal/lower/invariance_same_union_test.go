package lower

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

const sameUnionArms = `interface Start { flags: number; items: undefined; readonly mutation: undefined }
interface Label { flags: number; items: string[] | undefined; readonly mutation: undefined }
interface Mutation { flags: number; items: string[]; readonly mutation: boolean }
type Flow = Start | Label | Mutation;
`

func TestInvarianceSameUnionViewLowers(t *testing.T) {
	t.Parallel()
	const source = sameUnionArms + `function hold(flow: Flow): Flow | undefined { return flow; }
const view = hold({ flags: 1, items: undefined, mutation: undefined });
console.log(String(view?.flags));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, source, extension)
			if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "1\n" || stderr != "" {
				t.Fatalf("native exit %d, stdout %q, stderr %q; want stdout 1", code, stdout, stderr)
			}
		})
	}
}

func TestInvarianceSameUnionViewDependsOnNarrowedArm(t *testing.T) {
	t.Parallel()
	const source = sameUnionArms + `function starts(): Flow { return { flags: 1, items: undefined, mutation: undefined }; }
function use(start: Flow): number {
	let flow: Flow = start;
	const replace = (): void => { flow = starts(); };
	const view: Flow | undefined = flow;
	if (flow.mutation === true) {
		replace();
		return flow.items.length;
	}
	return 0;
}
console.log(String(use({ flags: 2, items: ["a"], mutation: true })));
`
	const want = "a read of flow narrowed to Mutation where a call may have rewritten it (flow is assigned inside a function)"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != want || !strings.HasSuffix(stop.Where, "main"+extension+":12:10") {
				t.Fatalf("got %v, want %q at the read on line 12", err, want)
			}
		})
	}
}

// TypeScript rejects removing undefined at a normal assignment before widened
// can judge it. Ask widened directly with types obtained from the checker.
func TestInvarianceSameUnionViewStillJudgesNullishRemoval(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		for _, absent := range []string{"undefined", "null"} {
			t.Run(absent+extension, func(t *testing.T) {
				t.Parallel()
				l, types := sameUnionTypes(t, sameUnionArms+"type Source = Flow | "+absent+";\n", extension)
				if found := l.widened(types["Source"], types["Flow"], map[[2]*checker.Type]bool{}); found == nil {
					t.Fatalf("removing %s bypassed the arm-by-arm judgment", absent)
				}
			})
		}
	}
}

func TestInvarianceSameUnionViewStillJudgesDifferentIdentity(t *testing.T) {
	t.Parallel()
	const source = sameUnionArms + `type Other = Start | Label | Mutation;
function hold(flow: Flow): Other | undefined { return flow; }
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			_, types := sameUnionTypes(t, source, extension)
			flow, other := types["Flow"], types["Other"]
			if flow == other {
				t.Fatal("checker interned the two named unions to one identity")
			}
			if !slices.Equal(flow.Types(), other.Types()) {
				t.Fatal("the two unions must hold the same arms")
			}
			_, err := lowerSourceAs(t, source, extension)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "which can write") {
				t.Fatalf("got %v, want the mutable view refusal", err)
			}
		})
	}
}

// A mutable container's element is a slot: widenedArguments asks it back to the source before it
// recurses into widened, so the skip never reaches it, and undefined written through ys can't land
// where xs reads a Flow (Lower's should on #hjkfhws).
func TestInvarianceSameUnionViewStillJudgesContainerElement(t *testing.T) {
	t.Parallel()
	const source = sameUnionArms + `function hold(xs: Flow[]): number {
	const ys: (Flow | undefined)[] = xs;
	return ys.length;
}
console.log(String(hold([])));
`
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension[1:], func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(refusal.What, "which can write") {
				t.Fatalf("got %v, want the mutable element refusal", err)
			}
		})
	}
}

func sameUnionTypes(t *testing.T, source, extension string) (*lowering, map[string]*checker.Type) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main"+extension, source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	types := map[string]*checker.Type{}
	for _, statement := range file.Statements.Nodes {
		if statement.Kind == ast.KindTypeAliasDeclaration {
			types[statement.Name().Text()] = checked.GetTypeAtLocation(statement.Name())
		}
	}
	return &lowering{checker: checked, program: program}, types
}
