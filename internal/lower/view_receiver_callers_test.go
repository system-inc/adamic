package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// viewReceiverCallers is every call of the receiver id functions in this package, by file, enclosing
// function and callee, with how many such calls it makes and why each is a read or a write. A read is
// checked under the id viewFieldReceiverTypeID decides, so the read lowering emits and every guard asking
// whether that read is checked ask the same id; a write takes viewWriteReceiverTypeID, which keeps
// viewReceiverTypeID's raw id (Lower's rule on 148e6df3, after the record guard asked the raw id of a routed
// read and failed open; #x0nd1mh's write helper, Nov 2).
var viewReceiverCallers = []struct {
	caller string
	calls  int
	role   string
}{
	{"object.go property viewFieldReceiverTypeID", 1, "read guard: a union or optional field read beside record storage is NotYet when its read is checked (checkRecordGuardedReads)"},
	{"object.go viewedField viewFieldReceiverTypeID", 1, "read: a property access's emitted field read, raw when it's an assignment target's read half"},
	{"object.go elementAccess viewFieldReceiverTypeID", 1, "read: a tuple element's emitted read, raw when it's an assignment target's read half"},
	{"collections.go destructureFrom viewFieldReceiverTypeID", 1, "read: a destructured binding's emitted field read"},
	{"assignments.go objectDestructuringAssignment viewFieldReceiverTypeID", 1, "read: an object pattern assignment's emitted field read, as a destructured binding's"},
	{"view_widened.go viewLie viewFieldReceiverTypeID", 1, "read guard: whether a view's source holds a field honestly, or only as a checked read"},
	{"view_copies.go checkedViewType viewFieldReceiverTypeID", 1, "read guard: a whole copy reads every field, NotYet when its type is a checked view's receiver"},
	{"interface_cast.go checkedViewReceiverField viewFieldReceiverTypeID", 1, "read guard: the id the asked access is checked under, read or write by its node"},
	{"interface_cast.go viewFieldReceiverTypeID viewReceiverTypeID", 1, "write: an assignment target keeps its write's raw id"},
	{"interface_cast.go viewFieldReceiverTypeID viewReadReceiverTypeID", 1, "read: the routed id, the only caller of viewReadReceiverTypeID"},
	{"interface_cast.go judgeViewAccess checkedViewReceiverField", 1, "read guard and write guard: an accessor alias refused at a checked read, a checked object field write refused, by the access's own node, the one judgment both of view's walks make (#q1hyhzc)"},
	{"interface_cast.go judgeViewKeyedWrite checkedViewReceiverField", 1, "write guard: a checked object field written by a computed key's literal names, by the write's own node (#13p3y0c), the one judgment both of view's walks make"},
	{"interface_cast.go legacyView checkedViewReceiverField", 1, "read guard: a checked field alias's conversion gap (checkedFieldAliasGap), by the access's own node"},
	{"interface_cast.go checkedViewReceiverField checkedViewReceiverField", 1, "read guard: a union receiver's member, under the same node"},
	{"destructure_view.go refuseIncompatibleDestructuredView checkedViewReceiverField", 1, "read guard: a destructured field read through an object-union view"},
	{"view_optional_writes.go viewWriteReceiverTypeID viewReceiverTypeID", 1, "write: every SetProperty's raw id, its optional fields kept for checkViewOptionalWrites"},
	{"logical_assignment.go logicalAssignment viewWriteReceiverTypeID", 1, "write: a logical assignment's keyed SetProperty"},
	{"class.go fieldStore viewWriteReceiverTypeID", 1, "write: a property assignment's SetProperty, and each arm of a computed one's"},
	{"class.go updateProperty viewWriteReceiverTypeID", 1, "write: a compound assignment's SetProperty"},
	{"assignments.go increment viewWriteReceiverTypeID", 1, "write: an increment's SetProperty"},
}

// Every call of viewReceiverTypeID, viewReadReceiverTypeID, viewFieldReceiverTypeID,
// viewWriteReceiverTypeID and checkedViewReceiverField is named in viewReceiverCallers as a read or a write, so a new caller has
// to say which id it asks, and a guard of a read can't ask the raw id the read isn't checked under.
// The test reads this package's source.
func TestViewReceiverCallersAreClassified(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	callees := map[string]bool{"viewReceiverTypeID": true, "viewReadReceiverTypeID": true, "viewFieldReceiverTypeID": true, "viewWriteReceiverTypeID": true, "checkedViewReceiverField": true}
	files := token.NewFileSet()
	found := map[string]int{}
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
				if call, isCall := node.(*ast.CallExpr); isCall {
					if selector, isSelector := call.Fun.(*ast.SelectorExpr); isSelector && callees[selector.Sel.Name] {
						found[entry.Name()+" "+function.Name.Name+" "+selector.Sel.Name]++
					}
				}
				return true
			})
		}
	}
	classified := map[string]int{}
	for _, caller := range viewReceiverCallers {
		classified[caller.caller] = caller.calls
	}
	var unclassified []string
	for key, calls := range found {
		if classified[key] != calls {
			unclassified = append(unclassified, key+": "+strconv.Itoa(calls)+" calls, classified "+strconv.Itoa(classified[key]))
		}
	}
	for key := range classified {
		if found[key] == 0 {
			unclassified = append(unclassified, key+": classified, never called")
		}
	}
	sort.Strings(unclassified)
	if len(unclassified) > 0 {
		t.Errorf("a receiver id call isn't classified as a read or a write; a read asks viewFieldReceiverTypeID, a write viewWriteReceiverTypeID, and viewReceiverCallers says which and why:\n%s", strings.Join(unclassified, "\n"))
	}
}

// checkedFieldWriters is every function in this package allowed to write a CheckedFields key, and why.
var checkedFieldWriters = map[string]string{
	"markCheckedField": "the one helper: it counts each new key under its name, which view's indexed walk reads to know what to ask again (#q1hyhzc)",
	"readiness":        "a record's field names, written once lowering is done, after the last view",
}

// Every `CheckedFields[...] =` outside the test files is in markCheckedField or readiness, so a key added
// mid-lowering is counted and view's indexed walk asks the accesses of its name again (view_sites.go). A
// write that skips the count leaves an access judged before it unchecked, where the whole walk would
// have checked it. Mutant: one of the five writes made directly again, and this names it.
func TestCheckedFieldsAreWrittenThroughMarkCheckedField(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var direct []string
	writers := map[string]bool{}
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
				assignment, isAssignment := node.(*ast.AssignStmt)
				if !isAssignment {
					return true
				}
				for _, written := range assignment.Lhs {
					index, isIndex := written.(*ast.IndexExpr)
					if !isIndex {
						continue
					}
					if selector, isSelector := index.X.(*ast.SelectorExpr); isSelector && selector.Sel.Name == "CheckedFields" {
						writers[function.Name.Name] = true
						if _, allowed := checkedFieldWriters[function.Name.Name]; !allowed {
							direct = append(direct, files.Position(assignment.Pos()).String()+" in "+function.Name.Name)
						}
					}
				}
				return true
			})
		}
	}
	// A renamed helper would leave this test reading nothing.
	for writer := range checkedFieldWriters {
		if !writers[writer] {
			direct = append(direct, writer+": allowed, writes no CheckedFields key")
		}
	}
	sort.Strings(direct)
	if len(direct) > 0 {
		t.Errorf("a CheckedFields key is written past markCheckedField, so view's indexed walk never asks its name's accesses again:\n%s", strings.Join(direct, "\n"))
	}
}
