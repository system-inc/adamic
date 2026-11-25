package lower

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

func TestGenericVoidReadbackUsesOwnSignature(t *testing.T) {
	t.Parallel()
	const source = `
function read<T>(values: T | T[] | undefined): T | T[] | undefined { if (values === undefined) return undefined; return values; }
function overload<T>(values: T | T[] | undefined): T | T[] | undefined;
function overload<T>(values: T | T[] | undefined): T | T[] | undefined { if (values === undefined) return undefined; return values; }
const nonvoid = read<number>([3]);
const genuine = read<void>([]);
const foreign = overload<number>([3]);
`
	file := writeAgreementSource(t, "main.ts", source)
	program, err := load.Load([]string{file})
	if err != nil {
		t.Fatal(err)
	}
	entry := program.Files()[0]
	types, release := program.Checker(t.Context(), entry)
	defer release()
	l := &lowering{checker: types, program: program}
	statements := entry.Statements.Nodes
	call := func(index int) *ast.Node {
		return statements[index].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Initializer()
	}
	readCall := call(3)
	readTarget := types.GetSignatureFromDeclaration(statements[0])
	withoutMapper := *types.GetResolvedSignature(readCall)
	mapper := reflect.ValueOf(&withoutMapper).Elem().FieldByName("mapper")
	reflect.NewAt(mapper.Type(), unsafe.Pointer(mapper.UnsafeAddr())).Elem().SetZero()
	if withoutMapper.Target() != readTarget || !l.signatureReachesVoid(readTarget, &withoutMapper) {
		t.Fatal("the copied signature must reach fallback with read's exact target")
	}

	for _, probe := range []struct {
		name        string
		declaration *ast.Node
		call        *ast.Node
		resolved    *checker.Signature
		want        string
	}{
		{"nonvoid", statements[0], readCall, types.GetResolvedSignature(readCall), ""},
		{"genuine void", statements[0], call(4), types.GetResolvedSignature(call(4)), "read's T = void (read from the signature)"},
		{"foreign overload", statements[2], call(5), types.GetResolvedSignature(call(5)), "overload's T = void (read from the signature)"},
		{"exact target with nil mapper", statements[0], readCall, &withoutMapper, "read's T = void (read from the signature)"},
		// Production function calls reject missing signatures before this judgment (generic.go).
		{"nil resolved signature", statements[0], readCall, nil, "read's T = void (read from the signature)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			// These subtests share the parent's checker lease and diagnostic memo.
			l.genericVoidVerdicts = nil
			target := types.GetSignatureFromDeclaration(probe.declaration)
			err := l.refuseGenericVoidReads(probe.call, probe.declaration, nil, target, probe.resolved)
			if probe.want == "" {
				if err != nil {
					t.Fatalf("known nonvoid binder: %v", err)
				}
				return
			}
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.want) ||
				!strings.HasSuffix(refused.What, "as a returned value") {
				t.Fatalf("want %q preserved at the returned read, got %v", probe.want, err)
			}
		})
	}
}

func TestGenericVoidReadbackLeavesBodyJudgments(t *testing.T) {
	t.Parallel()
	const mixed = `
function combine<T extends {}>(xs: T | T[] | undefined, ys: T | T[] | undefined): T | T[] | undefined {
    if (xs === undefined) return ys;
    return xs;
}
const out = combine<number>([2, 3], 1);
if (out !== undefined && Array.isArray(out)) {
    for (const item of out) console.log(String(item));
}
`
	_, err := lowerSourceAs(t, mixed, ".ts")
	var pending *NotYet
	if !errors.As(err, &pending) || pending.What != "a function returning T | T[] | undefined" {
		t.Fatalf("want the existing return layout judgment after the nonvoid readback, got %v", err)
	}
	const genuine = `
function first<T>(items: T[] | undefined): unknown {
    if (items === undefined) return undefined;
    return items[0];
}
const hidden: () => void = () => 42;
const items = [1].map(hidden);
console.log(String(first<void>(items)));
`
	_, err = lowerSourceAs(t, genuine, ".ts")
	var refused *Refused
	if !errors.As(err, &refused) ||
		!strings.HasPrefix(refused.What, "a generic instantiated at void whose body reads the value: first's T = void, read at ") ||
		!strings.HasSuffix(refused.What, "through an element read whose value is read as a returned value") {
		t.Fatalf("want the existing genuine void refusal, got %v", err)
	}
}

func TestGenericVoidReadbackRequiresClosedBinding(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, binder, local string
		want                bool
	}{
		{"anonymous field", "{ payload: U }", "", false},
		{"call parameter", "(value: U) => void", "", false},
		{"call result", "() => U", "", false},
		{"index value", "{ [key: string]: U }", "", false},
		{"local interface capture", "Box", "interface Box { payload: U }", false},
		{"number", "number", "", true},
		{"closed field", "{ payload: number }", "", true},
		{"signature constraint captures outer", "<V extends U>(value: V) => V", "", false},
		{"signature default captures outer", "<V = U>() => V", "", false},
		{"signature receiver captures outer", "(this: U) => number", "", false},
		{"closed inner generic", "<V>(value: V) => V", "", true},
		{"recursive open", "Loop", "type Loop = { next?: Loop; payload: U }", false},
		{"recursive closed", "Loop", "type Loop = { next?: Loop; payload: number }", true},
		{"void field", "{ payload: void }", "", false},
		{"void array", "void[]", "", false},
		{"closed receiver", "(this: number) => number", "", true},
		{"closed constraint", "<V extends number>(value: V) => V", "", true},
		{"closed signature", "(value: number) => string", "", true},
		{"project reference", "ProjectReference", "", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := `
function read<T>(values: T | T[] | undefined): T | T[] | undefined {
    if (values === undefined) return undefined;
    return values;
}
interface ProjectReference { readonly path: string; }
function outer<U>() {
    ` + probe.local + `
    const result = read<` + probe.binder + `>([]);
}
`
			file := writeAgreementSource(t, "main.ts", source)
			program, err := load.Load([]string{file})
			if err != nil {
				t.Fatal(err)
			}
			entry := program.Files()[0]
			types, release := program.Checker(t.Context(), entry)
			defer release()
			declaration := entry.Statements.Nodes[0]
			body := entry.Statements.Nodes[2].Body().AsBlock().Statements.Nodes
			statement := body[len(body)-1].AsVariableStatement()
			call := statement.DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Initializer()
			target, resolved := types.GetSignatureFromDeclaration(declaration), types.GetResolvedSignature(call)
			l := &lowering{checker: types, program: program}
			if resolved == nil || resolved.Target() != target || !l.signatureReachesVoid(target, resolved) {
				t.Fatal("expected an exact target and active fallback")
			}
			binder := types.GetTypeAtLocation(declaration.TypeParameters()[0].Name())
			if actual := l.genericVoidProvenNonvoid(declaration, binder, target, resolved); actual != probe.want {
				t.Fatalf("binding %s: proven nonvoid = %t, want %t", probe.binder, actual, probe.want)
			}
		})
	}
}
