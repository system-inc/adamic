package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// sourceSlotKindClasses is how the source-slot certificate (#2msw6t8) treats each IR expression or
// statement that holds an operand and names a field or a key, found in internal/ir's source by
// TestEveryNamedReadIsClassedForSourceSlots, so a kind added to the IR fails it until someone says how
// the certificate judges it. judged is certifySourceSlots' own walk, held to W; named is
// sourceSlotNamedKinds, which reads no fixed object's slot at a type's claim; stops is a kind the walk
// gives the write's stop for, as it does any kind naming a written field that sourceSlotNamedKinds
// doesn't list; elsewhere is judged by another pass or reads no field by the name it holds.
var sourceSlotKindClasses = map[string]struct{ class, why string }{
	"Property":             {"judged", "a field read, optional, checked, a callee or destructuring's, held to W at its declared type; a callee's keeps the stop"},
	"BoundMethod":          {"stops", "bind resolves the slot as a closure with no check of what it holds (road_bound_method.ts)"},
	"DynamicProperty":      {"named", "reads by the allocation's own runtime shape into a tagged union, claiming nothing"},
	"HasProperty":          {"named", "tests presence, reading no value"},
	"HasAccessor":          {"named", "tests whether a class owns an accessor, reading no value"},
	"OwnFunction":          {"named", "tests whether a function is there, reading no value"},
	"PhantomMember":        {"named", "reads a checker-only member of a primitive, never an object's slot"},
	"RegExpProperty":       {"named", "reads a regexp result array's metadata, never a fixed object"},
	"RegExpGroup":          {"named", "reads a regexp match's groups, never a fixed object"},
	"SetProperty":          {"named", "a write, judged at view (recordSourceSlotWrite)"},
	"Labeled":              {"named", "names a label"},
	"CheckedCast":          {"elsewhere", "tests its tag field against literals at run time"},
	"DynamicKeyedProperty": {"elsewhere", "a run-time key read by the allocation's own runtime shape into a tagged union, claiming nothing"},
	"HasOwn":               {"elsewhere", "a run-time key's presence test, reading no value"},
	"LogicalAssignment":    {"elsewhere", "its read is an ir.Property the walk judges, and a logical write through a view keeps the stop"},
	"CollectionIterator":   {"elsewhere", "Key is which of a collection's sequences it walks, no field"},
	"ForOf":                {"elsewhere", "Key binds an entry's key, no field"},
	"MakeError":            {"elsewhere", "Name is the error's name, a value it's given, no field"},
	"RecordCoalesce":       {"elsewhere", "a record's key, and no fixed object is a record"},
	"MapNew":               {"elsewhere", "a Map's keys, never a field"},
	"MapKeys":              {"elsewhere", "a Map's keys, never a field"},
	"MapForEach":           {"elsewhere", "a Map's keys, never a field"},
	"MapGet":               {"elsewhere", "a Map's key, never a field"},
	"MapSet":               {"elsewhere", "a Map's key, never a field"},
	"MapHas":               {"elsewhere", "a Map's key, never a field"},
	"MapDelete":            {"elsewhere", "a Map's key, never a field"},
}

// Every IR expression and statement that holds an operand and has a field named Name, Key, Field,
// Keys, KeyNames or Names is classed in sourceSlotKindClasses, and each class agrees with what
// certifySourceSlots does: a named kind is in sourceSlotNamedKinds, a stopped one isn't and names its
// field by Name, so the walk's fail-closed guard reaches it. No operand hides behind a pointer, which the
// walk doesn't follow. Mutant: a kind taken out of the table, and this goes red naming it.
func TestEveryNamedReadIsClassedForSourceSlots(t *testing.T) {
	t.Parallel()
	// Reads: internal/ir/*.go
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	type field struct {
		name string
		kind ast.Expr
	}
	structs := map[string][]field{}
	nodes := map[string]bool{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.TypeSpec:
				if structure, ok := node.Type.(*ast.StructType); ok {
					for _, declared := range structure.Fields.List {
						for _, name := range declared.Names {
							structs[node.Name.Name] = append(structs[node.Name.Name], field{name.Name, declared.Type})
						}
					}
				}
			case *ast.FuncDecl:
				if node.Recv == nil || len(node.Recv.List) != 1 || node.Type.Params.NumFields() != 0 {
					return false
				}
				if node.Name.Name != "Type" && node.Name.Name != "statement" {
					return false
				}
				if identifier, ok := node.Recv.List[0].Type.(*ast.Ident); ok {
					nodes[identifier.Name] = true
				}
				return false
			}
			return true
		})
	}
	if len(nodes) < 120 {
		t.Fatalf("found %d expressions and statements in ../ir: the enumeration no longer finds the IR", len(nodes))
	}
	// holds is whether a field's type can hold an operand, and through a pointer when pointed is set.
	var holds func(ast.Expr, bool, map[string]bool) bool
	holds = func(held ast.Expr, pointed bool, seen map[string]bool) bool {
		switch held := held.(type) {
		case *ast.ArrayType:
			return holds(held.Elt, pointed, seen)
		case *ast.StarExpr:
			return pointed && holds(held.X, pointed, seen)
		case *ast.MapType:
			return holds(held.Key, pointed, seen) || holds(held.Value, pointed, seen)
		case *ast.Ident:
			if held.Name == "Expression" || held.Name == "Statement" {
				return true
			}
			if seen[held.Name] {
				return false
			}
			seen[held.Name] = true
			for _, inner := range structs[held.Name] {
				if holds(inner.kind, pointed, seen) {
					return true
				}
			}
		}
		return false
	}
	keyNames := map[string]bool{"Name": true, "Key": true, "Field": true, "Keys": true, "KeyNames": true, "Names": true}
	found := map[string]bool{}
	for _, kind := range slices.Sorted(maps.Keys(nodes)) {
		operand, named, byName := false, false, false
		for _, declared := range structs[kind] {
			operand = operand || holds(declared.kind, false, map[string]bool{kind: true})
			named = named || keyNames[declared.name]
			byName = byName || declared.name == "Name"
			if _, isPointer := declared.kind.(*ast.StarExpr); isPointer && holds(declared.kind, true, map[string]bool{kind: true}) {
				t.Errorf("ir.%s holds an operand behind its pointer field %s, which certifySourceSlots' walk doesn't follow", kind, declared.name)
			}
		}
		if !operand || !named {
			continue
		}
		found[kind] = true
		classed, ok := sourceSlotKindClasses[kind]
		if !ok {
			t.Errorf("ir.%s holds an operand and names a field or a key, and has no class for the source-slot certificate: say how certifySourceSlots judges it in sourceSlotKindClasses", kind)
			continue
		}
		inNamed := false
		for listed := range sourceSlotNamedKinds {
			inNamed = inNamed || listed.Name() == kind
		}
		switch classed.class {
		case "named":
			if !inNamed {
				t.Errorf("ir.%s is classed named but sourceSlotNamedKinds doesn't list it", kind)
			}
		case "stops":
			if inNamed || !byName {
				t.Errorf("ir.%s is classed stops, so it must name its field by Name and stay out of sourceSlotNamedKinds", kind)
			}
		case "judged", "elsewhere":
			if inNamed {
				t.Errorf("ir.%s is classed %s but sourceSlotNamedKinds lists it", kind, classed.class)
			}
		default:
			t.Errorf("ir.%s has an unknown class %q", kind, classed.class)
		}
	}
	for kind := range sourceSlotKindClasses {
		if !found[kind] {
			t.Errorf("ir.%s is classed but isn't an IR kind holding an operand and naming a field", kind)
		}
	}
	for listed := range sourceSlotNamedKinds {
		if listed.PkgPath() != reflect.TypeOf(ir.Property{}).PkgPath() || sourceSlotKindClasses[listed.Name()].class != "named" {
			t.Errorf("sourceSlotNamedKinds lists %s, which sourceSlotKindClasses doesn't class named", listed)
		}
	}
}
