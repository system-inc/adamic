package worker

import "testing"

var environmentMutants = []workerMutation{
	{"environment-drop-validation", "internal/worker/environment.go", `if (typeof value !== 'string') {`, `if (false) {`, "TestEnvironmentBridgeRejectsBindingsPrivately", "environment bridge:"},
	{"environment-drop-projection", "internal/worker/environment.go", `"handle(input, environment)"`, `"handle(input, env)"`, "TestEnvironmentBridgeProjectsDeclaredNames", "environment bridge:"},
	{"environment-accept-number", "internal/worker/environment.go", `if types.GetTypeOfSymbol(field) != types.GetStringType() {`, `if false {`, "TestWorkerEnvironmentContract/number", "environment contract accepted=false:"},
	{"environment-log-value", "internal/worker/environment.go", `console.error('environment binding ' + JSON.stringify(name) + ' is missing or not a string');`, `console.error('environment binding ' + JSON.stringify(name) + ' is missing or not a string: ' + String(value));`, "TestEnvironmentBridgeRejectsBindingsPrivately", "environment bridge:"},
	{"environment-remove-wasm-needle-guard", "internal/worker/wasm.go", `if count := strings.Count(bridge, needle); count != 1 {`, `if count := strings.Count(bridge, needle); false {`, "TestInitializeWasmBridgeRequiresExactlyOneExport/missing", "Wasm bridge needle must fail closed:"},
}

func TestProduct_WorkerEnvironmentValidationMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentMutants[0], false)
}
func TestProduct_WorkerEnvironmentProjectionMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentMutants[1], false)
}
func TestProduct_WorkerEnvironmentContractMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentMutants[2], false)
}
func TestProduct_WorkerEnvironmentPrivacyMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentMutants[3], false)
}
func TestWorkerEnvironmentMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, m := range environmentMutants {
		t.Run(m.name, func(t *testing.T) { t.Parallel(); runWorkerMutant(t, m) })
	}
}

func TestProduct_WorkerEnvironmentWasmNeedleMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, environmentMutants[4], false)
}
