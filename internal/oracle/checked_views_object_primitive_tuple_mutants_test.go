package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Every omission must execute successfully in every backend. A crash, another
// refusal, or a compile failure cannot certify a killed read-check mutant.
func TestCheckedViewObjectPrimitiveTupleMutants(t *testing.T) {
	declarations, _ := tupleOriginalInputs(t)
	for _, test := range []struct{ name, fixture, node, diagnostic string }{
		{"signature-id", "signature-position-id-wrong", "false\nstring\n", "field read failed: value[0] is not a IncrementalBuildInfoFileId; expected IncrementalBuildInfoFileId, found boolean"},
		{"signature-value", "signature-position-value-wrong", "7\nboolean\n", "field read failed: value[1] matches no member of [] | EmitSignature; expected [] | EmitSignature, found boolean"},
		{"root-member", "root-index-scalar", "boolean\n", "field read failed: values[0] matches no member of IncrementalBuildInfoRoot; expected IncrementalBuildInfoRoot, found boolean"},
		{"foreach-arity", "signature-foreach-wrong-arity", "object\ndone\n", "field read failed: selected[element] matches no member of IncrementalBuildInfoEmitSignature; expected IncrementalBuildInfoEmitSignature, found object"},
		{"out-outer", "lane4b-out-signature-wrong", "boolean\n", "field read failed: node.outSignature matches no member of EmitSignature | undefined; expected EmitSignature | undefined, found boolean"},
		{"out-shape", "emit-record-noread", "object\n", "field read failed: viewed.outSignature is not a [signature: string]; expected [signature: string], found object"},
		{"out-nested", "emit-wrong-position", "boolean\n", "field read failed: signature[0] is not a string; expected string, found boolean"},
	} {
		t.Run(test.name, func(t *testing.T) {
			input, err := os.ReadFile("../../stage3/interface-downcasts/tuples/" + test.fixture + ".a")
			if err != nil {
				t.Fatal(err)
			}
			source := string(input)
			if test.name == "root-member" {
				source = strings.Replace(source, "const roots = [7];", "const roots = [false];", 1)
			}
			if test.name == "out-nested" {
				source = strings.Replace(source, "console.log(typeof signature === 'string' ? signature : signature[0]);", "console.log(typeof signature === 'string' ? 'string' : typeof signature[0]);", 1)
			}
			source = strings.ReplaceAll(source, "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts"))))
			path := filepath.Join(t.TempDir(), test.name+".a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			if diff := disagreement(run{stdout: []byte(test.node)}, onNode(t, path)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			objectPrimitiveOriginalCount(t, program, "tuple-mutants/"+test.name+".a")
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			if test.name == "signature-value" {
				want.stdout = []byte("7\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
					t.Fatalf("original guard: %s; stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			changed := 0
			switch test.name {
			case "signature-id", "signature-value", "out-nested":
				position := "0"
				if test.name == "signature-value" {
					position = "1"
				}
				changed = changeTupleOriginalRead(program, func(read ir.Property) bool { return read.Name == position }, func(read ir.Property) ir.Property { read.View = ""; return read })
			case "root-member":
				changed = changeTupleArrayRead(program, func(read ir.ArrayIndex) ir.ArrayIndex { read.View = ""; read.TupleUnion = false; return read })
			case "foreach-arity":
				for index, contract := range program.ViewContracts {
					if contract.FixedTuple && len(contract.Tuple) == 2 {
						program.ViewContracts[index].Tuple = contract.Tuple[:1]
						changed++
					}
				}
			case "out-outer":
				changed = changeTupleOriginalRead(program, func(read ir.Property) bool { return read.Name == "outSignature" }, func(read ir.Property) ir.Property { read.View = ""; return read })
			case "out-shape":
				for index, contract := range program.ViewContracts {
					if contract.FixedTuple {
						program.ViewContracts[index].FixedTuple = false
						program.ViewContracts[index].Kind = ir.ViewObject
						changed++
					}
				}
			}
			if changed == 0 {
				t.Fatal("mutant found no original obligation")
			}
			mutated, _ := nativelyUncached(t, program)
			for backend, got := range []run{mutated, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if got.exitCode != 0 || len(got.stderr) != 0 || len(got.stdout) == 0 {
					t.Errorf("mutant must execute: backend %d stdout %q stderr %q exit %d", backend, got.stdout, got.stderr, got.exitCode)
					continue
				}
				t.Logf("%s caught backend %d: exit 0 stdout %q", test.name, backend, got.stdout)
			}
		})
	}
}
