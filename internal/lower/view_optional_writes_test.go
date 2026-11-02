package lower

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"
	"testing"
)

// Language's v1 on f245bfe4: Element is a branded supertype with an optional name, and fresh is made
// without one, so its native shape has no slot for it. Element's count is the update's, and its
// required label the required pin's.
const viewOptionalWritePrelude = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
interface Element extends Declaration { _elementBrand: undefined; name?: string; count?: number; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
`

// The reason checkViewOptionalWrites gives.
const viewOptionalWriteReason = "a write to optional field name through a checked view, whose slot a fixed shape can't add yet"

// A write through a checked view to a field the view declares optional is NotYet at the write, whatever
// its form, and whether or not the value holds the field, since its slot may be absent and the checked
// write can't add one (#c7w6j0y; Adamic, Lower and Compiler on Language's v1, Nov 2). A required field
// keeps its checked write and agrees with Node. Mutant: checkViewOptionalWrites off; with
// ADAMIC_VIEW_OPTIONAL_WRITE_MUTANT=1 v1 must lower instead and its sanitized native must disagree with
// Node, which prints w.
func TestViewOptionalWriteIsNotYet(t *testing.T) {
	t.Parallel()
	const fresh = `const fresh: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined } as Node;
`
	const held = `const held: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, name: "a", count: 1 } as Node;
`
	for _, probe := range []struct{ name, source, want string }{
		{"v1 assigned", fresh + `function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
rename(fresh);`, viewOptionalWriteReason},
		{"v2 present", held + `function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
rename(held);`, viewOptionalWriteReason},
		{"v4 stored view", fresh + `function keep(n: Node): Element | undefined { return isElement(n) ? n : undefined; }
const e = keep(fresh); if (e) { e.name = "w"; console.log(String(e.name)); }`, viewOptionalWriteReason},
		{"compound", held + `function rename(n: Node): void { if (isElement(n)) { n.name += "x"; console.log(String(n.name)); } }
rename(held);`, viewOptionalWriteReason},
		{"nullish", fresh + `function rename(n: Node): void { if (isElement(n)) { n.name ??= "w"; console.log(String(n.name)); } }
rename(fresh);`, viewOptionalWriteReason},
		{"or", fresh + `function rename(n: Node): void { if (isElement(n)) { n.name ||= "w"; console.log(String(n.name)); } }
rename(fresh);`, viewOptionalWriteReason},
		{"increment", held + `function bump(n: Node): void { if (isElement(n)) { n.count!++; console.log(String(n.count)); } }
bump(held);`, viewOptionalWriteReason},
		{"delete", held + `function drop(n: Node): void { if (isElement(n)) { delete n.name; console.log(String(n.name)); } }
drop(held);`, ""},
		{"guarded", held + `function rename(n: Node): void { if (isElement(n)) { if (n.name !== undefined) { n.name = "w"; } console.log(String(n.name)); } }
rename(held);`, viewOptionalWriteReason},
		{"v3 cast", fresh + `function rename(n: Node): void { const e = n as Element; e.name = "w"; console.log(String(e.name)); }
rename(fresh);`, ""},
		// A class instance has a slot for each field the class declares, but none for an optional member
		// an interface merges into it. Tagged's required merged name makes name checked under its bare name
		// (mergedFieldRead), so Box's write to its merged optional name is a checked write too.
		{"merged class member", `interface Tagged { name: string; }
class Tagged { kind = 1; }
function show(t: Tagged): void { console.log(String(t.name)); }
interface Box { name?: string; }
class Box { kind = 2; }
const b = new Box(); b.name = "w"; console.log(String(b.name));`, viewOptionalWriteReason},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerTypeScriptAssertionSource(t, viewOptionalWritePrelude+probe.source)
			if os.Getenv("ADAMIC_VIEW_OPTIONAL_WRITE_MUTANT") == "1" {
				if probe.name != "v1 assigned" {
					t.Skip("the mutant is shown on v1")
				}
				if err != nil {
					t.Fatalf("the mutant must admit the program: %v", err)
				}
				code, out, errs := brandedNative(t, program)
				if code == 0 && out == "w\n" {
					t.Fatalf("the mutant's sanitized native agrees with Node: %q", out)
				}
				t.Logf("the mutant admits it: sanitized native exit %d, stdout %q, stderr %.300q where Node prints \"w\\n\"", code, out, errs)
				return
			}
			var pending *NotYet
			var refused *Refused
			if !errors.As(err, &pending) && !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.want) {
				t.Fatalf("want NotYet or Refused with %q, got %v", probe.want, err)
			}
			t.Logf("stops: %v", err)
		})
	}
	t.Run("required field", func(t *testing.T) {
		t.Parallel()
		lowersAndAgreesWithNodeAs(t, viewOptionalWritePrelude+`function relabel(n: Node): void { if (isElement(n)) { n.label = "z"; console.log(n.label); } }
const labelled: Node = { kind: 171, pos: 1, _declarationBrand: undefined, _elementBrand: undefined, label: "l" } as Node;
relabel(labelled);`, ".ts")
	})
}

// viewWriteProducers are the functions that make a SetProperty with a ViewReceiverTypeID, each through
// viewWriteReceiverTypeID, which the walk has to see or it proved nothing about them.
var viewWriteProducers = []string{"increment", "setProperty", "updateProperty", "logicalAssignment"}

// Every SetProperty a view may check takes its ViewReceiverTypeID from viewWriteReceiverTypeID, so
// checkViewOptionalWrites sees every write to an optional field (Lower's ruling, Nov 2). The test reads
// this package's source: an ir.SetProperty literal that sets ViewReceiverTypeID sets it to a call of the
// helper, so a sixth producer can't skip it. Mutant: updateProperty given l.viewReceiverTypeID directly,
// and it's named here.
func TestViewWriteReceiverHoldsEveryWrite(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var skipped []string
	producers := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				literal, isLiteral := node.(*ast.CompositeLit)
				if !isLiteral || !setPropertyLiteral(literal) {
					return true
				}
				for _, element := range literal.Elts {
					field, isField := element.(*ast.KeyValueExpr)
					if !isField {
						continue
					}
					if key, isKey := field.Key.(*ast.Ident); !isKey || key.Name != "ViewReceiverTypeID" {
						continue
					}
					producers[function.Name.Name] = true
					helper := false
					if call, isCall := field.Value.(*ast.CallExpr); isCall {
						selector, isSelector := call.Fun.(*ast.SelectorExpr)
						helper = isSelector && selector.Sel.Name == "viewWriteReceiverTypeID"
					}
					if !helper {
						skipped = append(skipped, files.Position(field.Pos()).String()+" in "+function.Name.Name)
					}
				}
				return true
			})
		}
	}
	sort.Strings(skipped)
	if len(skipped) > 0 {
		t.Errorf("a SetProperty takes its ViewReceiverTypeID from something other than viewWriteReceiverTypeID, so a write to an optional field through a checked view can lower unjudged; give it the helper's:\n%s", strings.Join(skipped, "\n"))
	}
	for _, name := range viewWriteProducers {
		if !producers[name] {
			t.Errorf("the walk never saw %s make a SetProperty with a ViewReceiverTypeID", name)
		}
	}
}

// setPropertyLiteral is whether literal is an ir.SetProperty{...}.
func setPropertyLiteral(literal *ast.CompositeLit) bool {
	selector, isSelector := literal.Type.(*ast.SelectorExpr)
	if !isSelector || selector.Sel.Name != "SetProperty" {
		return false
	}
	qualifier, isQualified := selector.X.(*ast.Ident)
	return isQualified && qualifier.Name == "ir"
}
