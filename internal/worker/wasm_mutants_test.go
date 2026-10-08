package worker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmWorkerMutants(t *testing.T) {
	if os.Getenv("ADAMIC_WORKER_MUTANTS") != "1" {
		t.Skip("set ADAMIC_WORKER_MUTANTS=1 to run Wasm Worker mutants")
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		t.Skip("Wasm Worker mutants require WASI_SYSROOT")
	}
	// Not parallel: scratch compiler builds bound the opt-in resource cost.
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	baseline := mutantScratch(t, root)
	if report, err := mutantCheck(t, baseline, "TestWasm(Workers|WorkerMemory|Purity)"); err != nil {
		t.Fatalf("Wasm scratch baseline failed: %v\n%s", err, report)
	}
	mutants := []struct{ name, file, before, after, check, witness string }{
		{"original-js-body", "internal/javascript/javascript.go", "if external, ok := options.ExternalFunctions[index]; ok {", "if external, ok := options.ExternalFunctions[index]; ok && false {", "TestWasmWorkers/transform", "selected JavaScript body ran instead of Wasm"},
		{"reversed-crossing-arguments", "internal/javascript/javascript.go", `external, strings.Join(parameters, ", "))`, `external, strings.Join([]string{parameters[1], parameters[0]}, ", "))`, "TestWasmWorkers/transform", "Wasm/source oracle mismatch"},
		{"accept-console", "internal/worker/purity.go", "case ir.WriteLine:\n\t\t\t\treturn fmt.Errorf(\"calls console\")", "case ir.WriteLine:\n\t\t\t\t// Mutant ignores logging.", "TestWasmPurity/console", "expected refusal"},
		{"result-not-released", "internal/worker/wasm/generate-crossing.mjs", "if (handle !== undefined) context.invoke('adamic_result_release', handle);", "if (false) context.invoke('adamic_result_release', handle);", "TestWasmWorkerMemory", "linear memory grew at request"},
	}
	for _, mutant := range mutants {
		t.Run(mutant.name, func(t *testing.T) {
			scratch := mutantScratch(t, root)
			path := filepath.Join(scratch, mutant.file)
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(original), mutant.before) != 1 {
				t.Fatalf("mutation anchor is not unique: %s", mutant.before)
			}
			write(t, path, strings.Replace(string(original), mutant.before, mutant.after, 1))
			report, err := mutantCheck(t, scratch, mutant.check)
			assertMutantCaught(t, report, err, mutant.check, mutant.witness)
		})
	}
}
