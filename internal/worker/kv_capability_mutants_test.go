package worker

import "testing"

var kvCapabilityMutants = []workerMutation{
	{"kv-slot-symbol", "internal/lower/worker_environment.go", `(!kvOnly && r.exact(t)) || r.kv(t)`, `(!kvOnly && r.exact(t))`, "TestWorkerKVCapabilityStops/let", "KV capability boundary let"},
	{"kv-function-value", "internal/lower/worker_kv.go", `if r.kvFunction(symbol) {`, `if false && r.kvFunction(symbol) {`, "TestWorkerKVCapabilityStops/callback", "KV capability boundary callback"},
	{"kv-registration", "internal/lower/library_node.go", `if load.WorkerKV(l.checker, proven) {`, `if false && load.WorkerKV(l.checker, proven) {`, "TestWorkerKVCapabilityStops/object_literal", "KV capability boundary object_literal"},
	{"kv-ingress", "internal/worker/environment.go", `fields = append(fields, environmentField{name: field.Name, kv: true})`, `fields = append(fields, environmentField{name: field.Name})`, "TestWorkerKVIngressReachesBridgeCheck", "KV ingress contract"},
	{"kv-field-name", "internal/worker/environment.go", `load.WorkerKV(types, types.GetTypeOfSymbol(field))`, `types.GetTypeOfSymbol(field).Symbol()!=nil && types.GetTypeOfSymbol(field).Symbol().Name=="KvNamespace"`, "TestWorkerKVFieldIdentityIsPreludeSymbol", "KV field identity"},
	{"kv-suspended", "internal/lower/worker_kv.go", `if err := r.kvSuspended(n); err != nil {`, `if err := r.kvSuspended(n); false && err != nil {`, "TestWorkerKVSuspendedResultsWait/held_promise", "KV suspended result held_promise"},
	{"kv-suspended-callee", "internal/lower/worker_kv.go", `if !handsBinding || !r.suspendedResult(r.l.checker.GetTypeAtLocation(n)) {`, `if !handsBinding || !r.suspendedResult(r.l.checker.GetTypeAtLocation(n)) || ast.SkipParentheses(n.Expression()).Kind != ast.KindIdentifier {`, "TestWorkerKVSuspendedResultsWait/held_iife", "KV suspended result held_iife"},
	{"kv-purity", "internal/worker/purity.go", `if capabilities[program.Functions[index].Position] {`, `if false && capabilities[program.Functions[index].Position] {`, "TestWorkerKVPurityRefusesOpaqueBindings", "KV purity"},
	{"kv-get-await", "internal/lower/library_workers.go", `return site.Parent != nil && site.Parent.Kind == ast.KindAwaitExpression && site.Parent.Expression() == site`, `return true`, "TestWorkerKVGetRoadsStayStopped/get_stored", "KV get road get_stored"},
	{"kv-duck", "internal/worker/environment.go", `if (value === null || typeof value !== 'object' || typeof value.get !== 'function' || typeof value.put !== 'function' || typeof value.delete !== 'function') {`, `if (false) {`, "TestWorkerKVGetAwaitedUnderNode/empty_binding", "KV get under Node empty_binding"},
	{"kv-text-or-null", "internal/javascript/library_workers.go", `+ ").then(adamicWorkersTextOrNull)"`, `+ ")"`, "TestWorkerKVGetAwaitedUnderNode/number", "KV get under Node number"},
}

func TestProduct_WorkerKVSlotMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[0], false)
}
func TestProduct_WorkerKVFunctionMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[1], false)
}
func TestProduct_WorkerKVRegistrationMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[2], false)
}
func TestProduct_WorkerKVIngressMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[3], false)
}
func TestProduct_WorkerKVFieldMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[4], false)
}
func TestProduct_WorkerKVSuspendedMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[5], false)
}
func TestProduct_WorkerKVSuspendedCalleeMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[6], false)
}
func TestProduct_WorkerKVPurityMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[7], false)
}
func TestProduct_WorkerKVGetAwaitMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[8], false)
}
func TestProduct_WorkerKVDuckMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[9], false)
}
func TestProduct_WorkerKVTextOrNullMutant(t *testing.T) {
	t.Parallel()
	buildWorkerMutant(t, kvCapabilityMutants[10], false)
}
func TestWorkerKVCapabilityMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, i := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10} {
		m := kvCapabilityMutants[i]
		t.Run(m.name, func(t *testing.T) { t.Parallel(); runWorkerMutant(t, m) })
	}
	t.Run("slot-callback", func(t *testing.T) {
		t.Parallel()
		m := kvCapabilityMutants[0]
		m.check = "TestWorkerKVCapabilityStops/callback"
		m.witness = "KV capability boundary callback"
		runWorkerMutant(t, m)
	})
}
