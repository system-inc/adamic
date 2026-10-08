package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type tupleOriginalManifest struct {
	Commit       string              `json:"upstream_commit"`
	Declarations map[string]string   `json:"declarations"`
	Fields       map[string][]string `json:"fields"`
	Pairs        []struct {
		Reads int `json:"read_count"`
	} `json:"pairs"`
}

func tupleOriginalInputs(t *testing.T) (string, tupleOriginalManifest) {
	t.Helper()
	root := os.Getenv("ADAMIC_TUPLE_ORIGINAL_DECLS")
	if root == "" {
		t.Skip("set ADAMIC_TUPLE_ORIGINAL_DECLS to pinned tuple declaration output")
	}
	data, err := os.ReadFile(filepath.Join(root, "tuple-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest tupleOriginalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	reads := 0
	for _, pair := range manifest.Pairs {
		reads += pair.Reads
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Declarations) != 78 || len(manifest.Pairs) != 6 || reads != 9 {
		t.Fatal("original tuple provenance changed")
	}
	for name, digest := range manifest.Declarations {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("declaration drift: " + name)
		}
	}
	return root, manifest
}

func tupleOriginalProgram(t *testing.T, root, name string) (string, *ir.Program) {
	t.Helper()
	input, err := os.ReadFile("../../stage3/interface-downcasts/tuples/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.ReplaceAll(string(input), "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(root, "compiler/builder.d.ts"))))
	bound = strings.ReplaceAll(bound, "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(root, "compiler/types.d.ts"))))
	path := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(path, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return path, program
}

func TestCheckedViewTupleOriginalOutSignature(t *testing.T) {
	root, manifest := tupleOriginalInputs(t)
	for _, test := range []struct{ name, node, diagnostic string }{
		{"emit-tuple-good", "signature\n", ""}, {"emit-string-good", "signature\n", ""}, {"emit-undefined", "absent\n", ""}, {"emit-helper-good", "signature\n", ""},
		{"emit-missing", "absent\n", "field read failed: viewed.outSignature is not initialized; expected EmitSignature | undefined, found missing"},
		{"emit-wrong-kind-noread", "boolean\n", "field read failed: viewed.outSignature matches no member of EmitSignature | undefined; expected EmitSignature | undefined, found boolean"},
		{"emit-wrong-kind", "undefined\n", "field read failed: viewed.outSignature matches no member of EmitSignature | undefined; expected EmitSignature | undefined, found boolean"},
		{"emit-wrong-position", "false\n", "field read failed: signature[0] is not a string; expected string, found boolean"},
		{"emit-helper-wrong", "false\n", "field read failed: signature[0] is not a string; expected string, found boolean"},
		{"emit-wrong-length", "signature\n", "field read failed: viewed.outSignature is not a [signature: string]; expected [signature: string], found array"},
		{"emit-record-noread", "object\n", "field read failed: viewed.outSignature is not a [signature: string]; expected [signature: string], found object"},
		{"emit-record", "signature\n", "field read failed: viewed.outSignature is not a [signature: string]; expected [signature: string], found object"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, program := tupleOriginalProgram(t, root, test.name)
			if difference := disagreement(run{stdout: []byte(test.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			complete := false
			for _, c := range program.ViewContracts {
				if c.Name != "IncrementalBundleEmitBuildInfo" {
					continue
				}
				fields := []string{}
				for _, f := range c.Fields {
					fields = append(fields, f.Name)
				}
				slices.Sort(fields)
				complete = complete || slices.Equal(fields, manifest.Fields[c.Name])
			}
			if !complete {
				t.Fatal("original receiver declaration was reduced")
			}
			if mutant := os.Getenv("ADAMIC_TUPLE_ORIGINAL_MUTANT"); mutant != "" {
				tupleOriginalMutant(t, program, mutant)
			}
			want := run{stdout: []byte(test.node)}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			// Check the semantic mutant first in JavaScript; unsafe native reads cannot
			// mask the missing named refusal with a sanitizer fault.
			js := onJavaScriptBackend(t, program)
			if difference := disagreement(want, js); difference != "" {
				t.Fatalf("JavaScript: %s; stdout %q stderr %q exit %d", difference, js.stdout, js.stderr, js.exitCode)
			}
			actual, binary := nativelyUncached(t, program)
			if difference := disagreement(want, actual); difference != "" {
				t.Fatalf("sanitized: %s; got %#v", difference, actual)
			}
			if difference := disagreement(want, releasedUncached(t, program)); difference != "" {
				t.Fatalf("release: %s", difference)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func tupleOriginalMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	changed := 0
	switch kind {
	case "skip":
		changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.Name == "outSignature" }, func(p ir.Property) ir.Property { p.View = ""; return p })
	case "nested":
		changed = changeTupleOriginalRead(program, func(p ir.Property) bool { return p.Name == "0" }, func(p ir.Property) ir.Property { p.View = ""; return p })
	case "shape":
		for i := range program.ViewContracts {
			c := &program.ViewContracts[i]
			if c.FixedTuple {
				c.FixedTuple = false
				c.Kind = ir.ViewObject
				changed++
			}
		}
	default:
		t.Fatal("unknown tuple mutant: " + kind)
	}
	if changed == 0 {
		t.Fatal("tuple mutant changed no check")
	}
}

func TestCheckedViewTupleOriginalTrackedSymbols(t *testing.T) {
	root, manifest := tupleOriginalInputs(t)
	for _, test := range []struct{ name, node, diagnostic string }{
		{"tracked-good", "1\n4\n", ""},
		{"tracked-undefined", "", ""},
		{"tracked-evaluation-good", "receiver\ncallback\n1\n4\n", ""},
		{"tracked-evaluation-undefined", "receiver\n", ""},
		{"tracked-meaning-wrong", "1\nfalse\n", "field read failed: item[2] is not a SymbolFlags; expected SymbolFlags, found boolean"},
		{"tracked-symbol-wrong", "false\n4\n", "field read failed: item[0].flags is not a SymbolFlags; expected SymbolFlags, found boolean"},
		{"tracked-length-wrong", "1\n4\n", "field read failed: trackedSymbols[element] is not a TrackedSymbol; expected TrackedSymbol, found array"},
		{"tracked-record-wrong", "1\n4\n", "field read failed: trackedSymbols[element] is not a TrackedSymbol; expected TrackedSymbol, found object"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path, program := tupleOriginalProgram(t, root, test.name)
			complete := false
			for _, contract := range program.ViewContracts {
				if contract.Name != "Symbol" {
					continue
				}
				var fields []string
				for _, field := range contract.Fields {
					fields = append(fields, field.Name)
				}
				slices.Sort(fields)
				complete = complete || slices.Equal(fields, manifest.Fields["Symbol"])
			}
			if !complete {
				t.Fatal("original Symbol declaration was reduced")
			}
			if difference := disagreement(run{stdout: []byte(test.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if os.Getenv("ADAMIC_TUPLE_TRACKED_MUTANT") == "meaning" {
				if changeTupleOriginalRead(program, func(p ir.Property) bool { return p.Name == "2" }, func(p ir.Property) ir.Property { p.View = ""; return p }) == 0 {
					t.Fatal("mutant changed no tuple position check")
				}
			}
			if os.Getenv("ADAMIC_TUPLE_TRACKED_MUTANT") == "guard" {
				changed := 0
				for i, statement := range program.Main {
					branch, ok := statement.(ir.If)
					if !ok || len(branch.Then) != 1 {
						continue
					}
					evaluate, ok := branch.Then[0].(ir.Evaluate)
					if !ok {
						continue
					}
					if _, ok := evaluate.Value.(ir.ArrayVisit); !ok {
						continue
					}
					branch.Condition = ir.BooleanConstant{Value: true}
					program.Main[i] = branch
					changed++
				}
				if changed != 1 {
					t.Fatalf("guard mutant changed %d checks", changed)
				}
			}
			want := run{stdout: []byte(test.node)}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			if test.name == "tracked-meaning-wrong" {
				want.stdout = []byte("1\n")
			}
			js := onJavaScriptBackend(t, program)
			if difference := disagreement(want, js); difference != "" {
				t.Fatalf("JavaScript: %s; stdout %q stderr %q exit %d", difference, js.stdout, js.stderr, js.exitCode)
			}
			actual, binary := nativelyUncached(t, program)
			if difference := disagreement(want, actual); difference != "" {
				t.Fatalf("sanitized: %s; stdout %q stderr %q", difference, actual.stdout, actual.stderr)
			}
			if difference := disagreement(want, releasedUncached(t, program)); difference != "" {
				t.Fatalf("release: %s", difference)
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// The existing union mutant walker covers function bodies. Include module reads
// without changing function identities or the compiled program's final graph.
func changeTupleOriginalRead(program *ir.Program, matches func(ir.Property) bool, edit func(ir.Property) ir.Property) int {
	program.Functions = append(program.Functions, ir.Function{Body: program.Main})
	changed := changeObjectPrimitiveRead(program, matches, edit)
	program.Main = program.Functions[len(program.Functions)-1].Body
	program.Functions = program.Functions[:len(program.Functions)-1]
	return changed
}
