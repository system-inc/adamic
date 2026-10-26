package worker

import "testing"

var wasmWorkerMutants = []workerMutation{
	{"original-js-body", "internal/javascript/javascript.go", "if external, ok := options.ExternalFunctions[index]; ok {", "if external, ok := options.ExternalFunctions[index]; ok && false {", "TestWasmWorkers/transform", "selected JavaScript body ran instead of Wasm"},
	{"reversed-crossing-arguments", "internal/javascript/javascript.go", `external, strings.Join(parameters, ", "))`, `external, strings.Join([]string{parameters[1], parameters[0]}, ", "))`, "TestWasmWorkers/transform", "Wasm/source oracle mismatch"},
	{"accept-console", "internal/worker/purity.go", "case ir.WriteLine:\n\t\t\t\treturn fmt.Errorf(\"calls console\")", "case ir.WriteLine:\n\t\t\t\t// Mutant ignores logging.", "TestWasmPurity/console", "expected refusal"},
	{"accept-host", "internal/worker/purity.go", "case ir.NodeFSFile:\n\t\t\t\treturn fmt.Errorf(\"impure: host operation %s\", node.Operation)", "case ir.NodeFSFile:\n\t\t\t\t// Mutant ignores host operations.", "TestWasmPurityRefusesHostOperations", "expected located host refusal before Wasm build"},
	{"result-not-released", "internal/worker/wasm/generate-crossing.mjs", "if (handle !== undefined) context.invoke('adamic_result_release', handle);", "if (false) context.invoke('adamic_result_release', handle);", "TestWasmWorkerMemory", "linear memory grew at request"},
}

func runWasmWorkerMutantsShard(t *testing.T, shard int) {
	requireMutants(t, true)
	for _, mutant := range wasmWorkerMutants {
		if familyOwns("TestWasmWorkerMutants", shard, mutant.name) {
			t.Run(mutant.name, func(t *testing.T) { runWorkerMutant(t, mutant) })
		}
	}
}
