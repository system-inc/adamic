package lower

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A record on a graph cycle is held by one of two rules: ir.RecordLiteral (how a record literal lowers
// today) is planned, made a graph member when graphTypes selects its record type or graphFlows its
// site, and a record of a type it didn't select written into a graph slot stops (noteRecordSlot,
// #1vb0x5s); graphAdoptionBoundary refuses ir.ObjectLiteral{Record: true} (the older form). #a9e21hc was
// a record form the boundary's rule didn't see. So every IR expression whose Type method answers
// ir.Record outright, a record maker, is named here with the rule that covers it, and a new one is red
// until someone decides which rule holds it. No other expression makes a record, so no maker of a
// selected record type is one the plan can't see.
func TestGraphRecordFormsMeetARule(t *testing.T) {
	t.Parallel()
	covered := map[string]string{
		"RecordLiteral": "planned (graphFlowRules), a graph member by its selected record type or site, and noteRecordSlot's stop for a record type graphTypes didn't select (#a9e21hc, #1vb0x5s)",
	}
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var makers []string
	expressions := 0
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || function.Name.Name != "Type" || function.Body == nil {
				continue
			}
			receiver, ok := function.Recv.List[0].Type.(*ast.Ident)
			if !ok {
				continue
			}
			expressions++
			// Any return of the bare identifier Record makes it a record maker.
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if returned, ok := node.(*ast.ReturnStmt); ok && len(returned.Results) == 1 {
					if identifier, ok := returned.Results[0].(*ast.Ident); ok && identifier.Name == "Record" && !slices.Contains(makers, receiver.Name) {
						makers = append(makers, receiver.Name)
					}
				}
				return true
			})
		}
	}
	if expressions < 100 {
		t.Fatalf("found %d Type methods in ../ir: the enumeration no longer finds the IR", expressions)
	}
	for _, maker := range makers {
		if covered[maker] == "" {
			t.Errorf("ir.%s answers Record: decide which rule holds it on a graph cycle (planned as RecordLiteral is, with noteRecordSlot's stop, or graphAdoptionBoundary's refusal), then name it here", maker)
		}
	}
	for name := range covered {
		if !slices.Contains(makers, name) {
			t.Errorf("ir.%s is named here but no longer answers Record", name)
		}
	}
	// The older form is a flag on ObjectLiteral, not a type of its own: its arm in graphAdoptionBoundary
	// stays, read from the source so the arm can't go quietly.
	source, err := parser.ParseFile(files, "graph_types.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	flagged := false
	ast.Inspect(source, func(node ast.Node) bool {
		if function, ok := node.(*ast.FuncDecl); ok && function.Name.Name == "graphAdoptionBoundary" {
			ast.Inspect(function.Body, func(inner ast.Node) bool {
				if selector, ok := inner.(*ast.SelectorExpr); ok && selector.Sel.Name == "Record" {
					if identifier, ok := selector.X.(*ast.Ident); ok && identifier.Name == "value" {
						flagged = true
					}
				}
				return true
			})
		}
		return true
	})
	if !flagged {
		t.Error("graphAdoptionBoundary no longer reads ObjectLiteral.Record: the older record form has no rule")
	}
}

// recordWriterClass is how a record write is held on a graph cycle.
type recordWriterClass int

const (
	// recordWriterSeeded is a write graphFlows' collect, or members for a literal's entries, demands what
	// it stores at, by the record's value type: the node is a case of graphFlows.
	recordWriterSeeded recordWriterClass = iota + 1
	// recordWriterNotYet is a write no graph program reaches: its node is stopped or refused before.
	recordWriterNotYet
	// recordWriterHelper emits a write for its callers, each classed in its own right.
	recordWriterHelper
)

// recordWriter is one native emitter that stores into a record's table: the IR node it emits, its class
// and why.
type recordWriter struct {
	node  string
	class recordWriterClass
	why   string
}

// recordWriters is every native emitter that stores into a record's table, by the runtime writer it
// calls or a helper that does, each with how its write is held on a graph cycle (#1vb0x5s). The list is
// held to the emitters by recordWriterFindings, so a new writer is red until it's classed.
var recordWriters = map[string]recordWriter{
	"recordCall":        {"RecordCall", recordWriterSeeded, "set's and defineHidden's value, Arguments[2], demanded at the write by collect's RecordCall arm"},
	"recordCoalesce":    {"RecordCoalesce", recordWriterSeeded, "??='s Value, demanded at the write by collect's RecordCoalesce arm"},
	"dictionaryLiteral": {"RecordLiteral", recordWriterSeeded, "its entries, demanded by members' RecordLiteral arm where it's made; its spread's values were planned where they were stored"},
	"recordSet":         {"", recordWriterHelper, "the checked or plain set recordCall, recordWrite and writeProperty write through"},
	"recordLiteral":     {"ObjectLiteral", recordWriterNotYet, "the older record form, ObjectLiteral{Record}, which graphAdoptionBoundary refuses in a graph program"},
	"recordWrite":       {"SetProperty", recordWriterNotYet, "a SetProperty{Record}, a write into the older form's record, which graphAdoptionBoundary refuses in a graph program"},
	"writeProperty":     {"SetProperty", recordWriterNotYet, "its record branch, emitted only where the program holds the older form's record storage (hasRecordStorage), which graphAdoptionBoundary refuses in a graph program"},
}

// recordWriterCalls are the runtime's record writers, each storing a key and a value into a record's
// table (runtime record.c).
var recordWriterCalls = []string{"adamic_record_define(", "adamic_record_define_hidden(", "adamic_record_set(", "adamic_record_set_checked("}

// recordWriterFindings holds writers to the native emitters: every function that calls a runtime record
// writer, or a helper writers names, must be classed there, a seeded one's node a case of graphFlows
// (graph_flow.go), and every entry must still write. The runtime writes no record of its own outside
// record.c.
func recordWriterFindings(writers map[string]recordWriter) ([]string, error) {
	var findings []string
	flowSource, err := os.ReadFile("graph_flow.go")
	if err != nil {
		return nil, err
	}
	flowFiles := token.NewFileSet()
	flow, err := parser.ParseFile(flowFiles, "graph_flow.go", flowSource, 0)
	if err != nil {
		return nil, err
	}
	cases := map[string]bool{}
	ast.Inspect(flow, func(node ast.Node) bool {
		function, isFunction := node.(*ast.FuncDecl)
		if !isFunction || function.Name.Name != "graphFlows" {
			return true
		}
		ast.Inspect(function.Body, func(inner ast.Node) bool {
			if clause, isCase := inner.(*ast.CaseClause); isCase {
				for _, listed := range clause.List {
					if selector, isSelector := listed.(*ast.SelectorExpr); isSelector {
						if identifier, isIdentifier := selector.X.(*ast.Ident); isIdentifier && identifier.Name == "ir" {
							cases[selector.Sel.Name] = true
						}
					}
				}
			}
			return true
		})
		return false
	})
	if len(cases) == 0 {
		return nil, fmt.Errorf("found no case of graphFlows in graph_flow.go")
	}
	sources, err := filepath.Glob(filepath.Join("..", "native", "*.go"))
	if err != nil {
		return nil, err
	}
	found := map[string]bool{}
	read := 0
	for _, path := range sources {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		read++
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, source, 0)
		if err != nil {
			return nil, err
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			body := string(source[files.Position(function.Body.Pos()).Offset:files.Position(function.End()).Offset])
			writes := false
			for _, call := range recordWriterCalls {
				writes = writes || strings.Contains(body, call)
			}
			for name, writer := range writers {
				if writer.class == recordWriterHelper && name != function.Name.Name {
					writes = writes || strings.Contains(body, "."+name+"(")
				}
			}
			if !writes {
				continue
			}
			name := function.Name.Name
			found[name] = true
			writer, classed := writers[name]
			if !classed {
				findings = append(findings, fmt.Sprintf("%s: %s writes into a record's table and isn't classed in recordWriters: say whether graphFlows seeds what it stores, or no graph program reaches it, and why", path, name))
				continue
			}
			if strings.TrimSpace(writer.why) == "" {
				findings = append(findings, fmt.Sprintf("%s: %s is classed without saying why", path, name))
			}
			switch writer.class {
			case recordWriterSeeded:
				if !cases[writer.node] {
					findings = append(findings, fmt.Sprintf("%s: %s is classed seeded, and graphFlows has no case of ir.%s", path, name, writer.node))
				}
			case recordWriterNotYet, recordWriterHelper:
			default:
				findings = append(findings, fmt.Sprintf("%s: %s has no class: %d", path, name, writer.class))
			}
		}
	}
	if read == 0 {
		return nil, fmt.Errorf("read no emitter source")
	}
	for _, name := range slices.Sorted(maps.Keys(writers)) {
		if !found[name] {
			findings = append(findings, fmt.Sprintf("%s no longer writes into a record's table; drop it from recordWriters", name))
		}
	}
	runtime, err := filepath.Glob(filepath.Join("..", "native_runtime", "c", "*.c"))
	if err != nil {
		return nil, err
	}
	if len(runtime) == 0 {
		return nil, fmt.Errorf("read no runtime source")
	}
	for _, path := range runtime {
		if filepath.Base(path) == "record.c" {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, call := range recordWriterCalls {
			if strings.Contains(string(source), call) {
				findings = append(findings, fmt.Sprintf("%s calls %s, a record write no emitter's class holds", path, strings.TrimSuffix(call, "(")))
			}
		}
	}
	return findings, nil
}

// TestGraphRecordWritersAreClassed holds the IR nodes that write a record's slot to how a graph cycle
// holds each (Lower's C2 on #1vb0x5s): what it stores seeded at the write, or never in a graph program.
// It reads the writers off the native emitters rather than off a list, so a new writer kind is red until
// it's classed. Its mutant, a writer left unclassed, goes red here too.
func TestGraphRecordWritersAreClassed(t *testing.T) {
	t.Parallel()
	findings, err := recordWriterFindings(recordWriters)
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range findings {
		t.Error(finding)
	}
	// The mutant: recordCoalesce left out, as a new writer would be.
	unclassed := maps.Clone(recordWriters)
	delete(unclassed, "recordCoalesce")
	findings, err = recordWriterFindings(unclassed)
	if err != nil {
		t.Fatal(err)
	}
	named := false
	for _, finding := range findings {
		named = named || strings.Contains(finding, "recordCoalesce writes into a record's table and isn't classed")
	}
	if !named {
		t.Errorf("an unclassed writer, recordCoalesce, went unnoticed: %v", findings)
	}
}

// A record whose value type is loose, unknown or object, is never a graph type (Lower's C3 on
// #1vb0x5s): no graph node is unknown or object, so its values never select it, and on a cycle component
// only by a view related to it, as Record<string, unknown> is to every type literal, it's dropped. So
// its literal is made counted, wherever it's held.
func TestGraphLooseRecordsAreNeverSelected(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `type Node = { name: string; next: Node[] };
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
const loose: Record<string, unknown> = { k: a };
const objects: Record<string, object> = { k: b };
console.log(String(Object.keys(loose).length + Object.keys(objects).length));
`)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) == 0 {
		t.Fatal("no graph types: the program's cycle no longer seeds a graph region")
	}
	literals := 0
	walk(program.Main, func(node any) bool {
		if literal, isLiteral := node.(ir.RecordLiteral); isLiteral {
			literals++
			if program.IsGraph(literal.GraphTypes) {
				t.Errorf("a record of loose values is a graph member: types %v", literal.GraphTypes)
			}
		}
		return true
	})
	if literals != 2 {
		t.Fatalf("found %d record literals, want 2", literals)
	}
}
