package worker

import "testing"

var kvCapabilityMutants = []workerMutation{
	{"kv-slot-symbol", "internal/lower/worker_environment.go", `(len(kvOnly) == 0 && r.exact(t)) || r.kv(t)`, `(len(kvOnly) == 0 && r.exact(t))`, "TestWorkerKVCapabilityStops/let", "KV capability boundary let"},
	{"kv-function-value", "internal/lower/worker_kv.go", `if r.kvFunction(symbol) {`, `if false && r.kvFunction(symbol) {`, "TestWorkerKVCapabilityStops/callback", "KV capability boundary callback"},
	{"kv-registration", "internal/lower/library_node.go", `if load.WorkerKV(l.checker, proven) {`, `if false && load.WorkerKV(l.checker, proven) {`, "TestWorkerKVCapabilityStops/object_literal", "KV capability boundary object_literal"},
	{"kv-ingress", "internal/worker/environment.go", `return nil, &lower.NotYet{Where: "worker", What: "environment field " + strconv.Quote(field.Name) + ": a Workers binding in the environment waits on the bridge's KV check"}`, `_ = (*lower.NotYet)(nil); fields = append(fields, field.Name); continue`, "TestWorkerKVIngressWaitsForBridge", "KV ingress contract"},
	{"kv-field-name", "internal/worker/environment.go", `load.WorkerKV(types, types.GetTypeOfSymbol(field))`, `types.GetTypeOfSymbol(field).Symbol()!=nil && types.GetTypeOfSymbol(field).Symbol().Name=="KvNamespace"`, "TestWorkerKVFieldIdentityIsPreludeSymbol", "KV field identity"},
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
func TestWorkerKVCapabilityMutantsAreCaught(t *testing.T) {
	t.Parallel()
	for _, i := range []int{0, 1, 2, 3, 4} {
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
