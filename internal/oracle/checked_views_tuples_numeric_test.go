package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCheckedViewTupleOriginalNumericBrandReadRefuses(t *testing.T) {
	root, _ := tupleOriginalInputs(t)
	path := filepath.Join(t.TempDir(), "numeric-brand-read.a")
	source := "import type { IncrementalBuildInfoFileId } from '" + filepath.ToSlash(filepath.Join(root, "compiler/builder.d.ts")) + "';\ninterface Base { readonly kind: string; }\ninterface Target extends Base { readonly id: IncrementalBuildInfoFileId; }\nconst raw = {kind: 'probe', id: 7};\nconst base: Base = raw;\nconst viewed = base as Target;\nconst id = viewed.id;\nconst brand = id.__incrementalBuildInfoFileIdBrand;\nconsole.log(typeof brand);\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte("undefined\n")}, onNode(t, path)); difference != "" {
		t.Fatal("Node: " + difference)
	}
	_, err := lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "a value of type any") {
		t.Fatalf("expected any read refusal, got %v", err)
	}
}

type originalTupleCase struct{ name, node, prefix, diagnostic string }

func verifyOriginalTupleCases(t *testing.T, cases []originalTupleCase) {
	t.Helper()
	root, _ := tupleOriginalInputs(t)
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			path, program := tupleOriginalProgram(t, root, test.name)
			if difference := disagreement(run{stdout: []byte(test.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if position := os.Getenv("ADAMIC_TUPLE_NUMERIC_MUTANT"); position != "" {
				if changeTupleOriginalRead(program, func(p ir.Property) bool { return p.Name == position }, func(p ir.Property) ir.Property { p.View = ""; return p }) == 0 {
					t.Fatal("mutant changed no position check")
				}
			}
			want := run{stdout: []byte(test.node)}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stdout: []byte(test.prefix), stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			js := onJavaScriptBackend(t, program)
			if difference := disagreement(want, js); difference != "" {
				t.Fatalf("JavaScript: %s; stdout %q stderr %q exit %d", difference, js.stdout, js.stderr, js.exitCode)
			}
			actual, binary := nativelyUncached(t, program)
			if difference := disagreement(want, actual); difference != "" {
				t.Fatalf("sanitized: %s; got %#v", difference, actual)
			}
			if difference := disagreement(want, releasedUncached(t, program)); difference != "" {
				t.Fatal("release: " + difference)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewTupleOriginalReferencedMap(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"references-good", "7\n9\n", "", ""},
		{"references-key-wrong", "false\n9\n", "", "field read failed: entry[0] is not a IncrementalBuildInfoFileId; expected IncrementalBuildInfoFileId, found boolean"},
		{"references-value-wrong", "7\nfalse\n", "7\n", "field read failed: entry[1] is not a IncrementalBuildInfoFileIdListId; expected IncrementalBuildInfoFileIdListId, found boolean"},
		{"references-length-wrong", "7\n9\n", "", "field read failed: references[element] is not a [fileId: IncrementalBuildInfoFileId, fileIdListId: IncrementalBuildInfoFileIdListId]; expected [fileId: IncrementalBuildInfoFileId, fileIdListId: IncrementalBuildInfoFileIdListId], found array"},
		{"references-record-wrong", "7\n9\n", "", "field read failed: references[element] is not a [fileId: IncrementalBuildInfoFileId, fileIdListId: IncrementalBuildInfoFileIdListId]; expected [fileId: IncrementalBuildInfoFileId, fileIdListId: IncrementalBuildInfoFileIdListId], found object"},
	})
}

func TestCheckedViewTupleOriginalSignaturePositions(t *testing.T) {
	verifyOriginalTupleCases(t, []originalTupleCase{
		{"signature-position-good", "7\nstring\n", "", ""},
		{"signature-position-empty", "7\nobject\n", "", ""},
		{"signature-position-single", "7\nobject\n", "", ""},
		{"signature-position-id-wrong", "false\nstring\n", "", "field read failed: value[0] is not a IncrementalBuildInfoFileId; expected IncrementalBuildInfoFileId, found boolean"},
		{"signature-position-value-wrong", "7\nboolean\n", "7\n", "field read failed: value[1] matches no member of [] | EmitSignature; expected [] | EmitSignature, found boolean"},
	})
}
