package lower

import (
	"context"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// sameRecordStorage's verdict on an intersection on either side (#mft3zc6). One of plain shapes is held
// as a fixed object, so it's judged as one: a record seen as it, a finite Partial<Record<K, V>> or an
// index signature, isn't the same storage, nor is it seen as one of those, nor is a fixed object whose
// field is a record where the intersection's is fixed, while a fixed object with the same fields is, seen
// either way. An intersection with a dictionary member isn't one objectIntersection holds, and keeps the
// answer present's nil gives.
func TestSameRecordStorageFixedIntersection(t *testing.T) {
	t.Parallel()
	const prelude = "type Entry = { count?: number } & { count?: number };\n" +
		"type Nested = { inner: { count?: number } } & { label?: string };\n" +
		"type Mixed = Record<string, number> & { readonly id: number };\n"
	for _, probe := range []struct {
		name, from, to string
		same           bool
	}{
		{"partial_record", `Partial<Record<"count", number>>`, "Entry", false},
		{"index_signature", "{ [key: string]: number }", "Entry", false},
		{"fixed_object", "{ count?: number; label: string }", "Entry", true},
		{"nested_record_field", "{ inner: Record<string, number> }", "Nested", false},
		{"nested_fixed_field", "{ inner: { count?: number; extra: string } }", "Nested", true},
		{"dictionary_member", "Record<string, number>", "Mixed", true},
		{"source_partial_record", "Entry", `Partial<Record<"count", number>>`, false},
		{"source_index_signature", "Entry", "{ [key: string]: number | undefined }", false},
		{"source_fixed_object", "Entry", "{ count?: number }", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			l, from, to := sameRecordStorageParameters(t, prelude+"function view(from: "+probe.from+", to: "+probe.to+"): void {}\n")
			if got := l.sameRecordStorage(from, to, newRecordStorageWalk()); got != probe.same {
				t.Errorf("sameRecordStorage(%s, %s) = %t, want %t", probe.from, probe.to, got, probe.same)
			}
		})
	}
}

// sameRecordStorageParameters loads source as a .ts program and gives back a lowering over its checker
// and the types of its last function's parameters from and to.
func sameRecordStorageParameters(t *testing.T, source string) (*lowering, *checker.Type, *checker.Type) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	statement := file.Statements.Nodes[len(file.Statements.Nodes)-1]
	if statement.Kind != ast.KindFunctionDeclaration {
		t.Fatalf("last statement is %v, want the function view", statement.Kind)
	}
	parameters := statement.AsFunctionDeclaration().Parameters.Nodes
	return &lowering{checker: checked, program: program, result: &ir.Program{}}, checked.GetTypeAtLocation(parameters[0].Name()), checked.GetTypeAtLocation(parameters[1].Name())
}
