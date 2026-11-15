package worker

import "testing"

// Each producer builds only one mutated test executable. Keep products top-level:
// independent overlays must never share a product's subtest trace.
var asyncWorkerMutants = []workerMutation{
	{"async-drop-await", "internal/worker/worker.go", `"output = await handle(input);"`, `"output = handle(input);"`, "TestAsyncBridgeAwaits", "async bridge:"},
	{"async-await-outside-try", "internal/worker/worker.go", `return strings.Replace(bridge, needle, "output = await handle(input);", 1), nil`, `return strings.Replace(strings.Replace(bridge, needle, "", 1), "try {", "output = await handle(input);\n\t\ttry {", 1), nil`, "TestAsyncBridgeCatchesRejectedPanic", "async bridge:"},
	{"async-accept-non-async", "internal/worker/worker.go", `if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) {`, `if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAsync) || workerLibraryPromise(result) {`, "TestWorkerAsyncContract/non-async-promise", "async contract accepted=false:"},
	{"async-remove-wasm-refusal", "internal/worker/worker.go", `if err := refuseAsyncWasm(program, asyncHandle, options.Wasm); err != nil {`, `if err := refuseAsyncWasm(program, asyncHandle, options.Wasm); false {`, "TestWorkerAsyncWasmRefusals", "async Wasm refusal"},
	{"async-remove-needle-guard", "internal/worker/worker.go", `if count := strings.Count(bridge, needle); count != 1 {`, `if count := strings.Count(bridge, needle); false {`, "TestAsyncBridgeRequiresExactlyOneCall/missing", "async bridge needle must fail closed:"},
}

func TestProduct_WorkerAsyncAwaitMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, asyncWorkerMutants[0], false)
}
func TestProduct_WorkerAsyncTryMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, asyncWorkerMutants[1], false)
}
func TestProduct_WorkerAsyncContractMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, asyncWorkerMutants[2], false)
}
func TestProduct_WorkerAsyncWasmMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, asyncWorkerMutants[3], false)
}
func TestWorkerAsyncMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, mutant := range asyncWorkerMutants {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			runWorkerMutant(t, mutant)
		})
	}
}

func TestProduct_WorkerAsyncNeedleMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, asyncWorkerMutants[4], false)
}
