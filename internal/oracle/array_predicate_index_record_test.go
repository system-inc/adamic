package oracle

import (
	"path/filepath"
	"testing"
)

// An index-only record arm is in Array.isArray's domain (rulings.md:260's amendment, Nov 26, @system_adamic on
// #gcec0bx): AdamicJsonRecoveryObject's one string index holds the JSON union, which admits no callable, so no array
// reaches the arm without a hatch, and the brand decides `AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject` both
// ways. Each program is tsc's readJson test over an object, an array, a string, a number, true, null and undefined, and
// agrees with Node on the sanitized native build, the release build and the JavaScript backend:
//   - json_record: readJson's own spelling, `... && !Array.isArray(value) ? value : {}`, printing each record's key
//     count and its read of a; `? value` is a read narrowed to the record, checked as an object, not null and not an
//     array, the array members excluded by the brand since the domain admits the record (locals.go);
//   - json_record_test: the test alone, printing its answer, so the arm's admission is held apart from the read of the
//     narrowed record after it;
//   - json_recovery_record and json_recovery_record_test: the same two over the record as stage3's adaptation 52
//     publishes it, `{ [key: string]: AdamicJsonRecoveryValue | undefined }`, parser recovery's undefined values in it;
//   - narrowed_union_object_tag: `number[] | { [key: string]: number }` assigned an array and read as one, the
//     checked narrowed read excluding the record arm by the brand (locals.go). It was pinned refused in reland_refused
//     until the amendment admitted its arm, and its refused twin there now holds unknown values
//     (TestNarrowedUnionObjectTagRefusal).
//
// Mutant (array_predicate_index_record.mutants.md, M2): the brand answering false at a site whose arm is index-only, the
// array read as the record. Each row goes red on its array line (a record printing 2 keys, a test printing true).
//
// The fixtures sit outside testdata's top level, so the flow and fresh corpora don't take them as their own.
func TestArrayPredicateIndexArmAgrees(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, output string }{
		{"json_record", "object 1 number\narray 0 absent\nstring 0 absent\nnumber 0 absent\nboolean 0 absent\nnull 0 absent\nundefined 0 absent\n"},
		{"json_record_test", "object true\narray false\nstring false\nnumber false\nboolean false\nnull false\nundefined false\n"},
		{"json_recovery_record", "object 1 number\narray 0 absent\nstring 0 absent\nnumber 0 absent\nboolean 0 absent\nnull 0 absent\nundefined 0 absent\n"},
		{"narrowed_union_object_tag", "1\n"},
		{"json_recovery_record_test", "object true\narray false\nstring false\nnumber false\nboolean false\nnull false\nundefined false\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_predicate_index_arm", row.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if disagreement(run{stdout: []byte(row.output)}, want) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", want.exitCode, want.stdout, want.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			machine, _ := natively(t, program)
			for backend, got := range map[string]run{"native": machine, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", backend, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}
