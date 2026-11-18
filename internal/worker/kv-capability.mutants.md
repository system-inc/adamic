# Workers KV capability and its awaited get (#bxaz0p1 part 1, head 3b-1)

`TestWorkerKVCapabilityMutantsAreCaught` uses the existing Worker buildcache
mutation products and runs each unmutated control before its mutated child.
A KV binding reaches a built Worker through the bridge's KV check, and its one
admitted member road is `await <binding>.get(key)`.

| Mutation | Catching test | Catch line in kv_capability_test.go |
| --- | --- | --- |
| kv-slot-symbol: omit KV from the shared slot walk | TestWorkerKVCapabilityStops/let and /callback | 68 |
| kv-function-value: omit the direct-callee rule | TestWorkerKVCapabilityStops/callback | 68 |
| kv-registration: omit the host registration | TestWorkerKVCapabilityStops/object_literal | 68 |
| kv-ingress: read a KV field as a string field | TestWorkerKVIngressReachesBridgeCheck | 190 |
| kv-field-name: trust the program interface's spelling | TestWorkerKVFieldIdentityIsPreludeSymbol | 117 |
| kv-purity: drop the capability loop in CheckPure | TestWorkerKVPurityRefusesOpaqueBindings | 167 |
| kv-suspended: drop the frame rule | TestWorkerKVSuspendedResultsWait/held_promise | 269 |
| kv-suspended-callee: judge only a named callee, not the call | TestWorkerKVSuspendedResultsWait/held_iife | 269 |
| kv-get-await: admit any get, not only await's direct operand | TestWorkerKVGetRoadsStayStopped/get_stored | 303 |
| kv-duck: drop the bridge's KV namespace check | TestWorkerKVGetAwaitedUnderNode/empty_binding | 518 |
| kv-text-or-null: drop get's `.then(adamicWorkersTextOrNull)` | TestWorkerKVGetAwaitedUnderNode/number | 518 |

The let, callback and object-literal checks require ordinary Worker lowering to
complete after the mutation: another lowering stop is not the named witness.
The ingress check requires the bridge to list KV as a binding and a real
namespace to reach the handler; the retargeted mutant reads it as a string.

Native's refusal of a host call it doesn't emit (a workers: module, a Promise
form, or a module other than node:path and node:fs) is `native.HostCallRefusal`,
an ordinary error WASI and WASIExports return before emission, with the
emitter's panic kept behind it as the backstop. It lives in another package, so
its mutant uses that package's own harness, as internal/lower's realpath mutant
does: `TestProduct_NativeWorkersHostCallMutant` builds the mutated
internal/native test binary with HostCallRefusal's test made `isCall && false`,
and `TestNativeWorkersHostCallMutantIsCaught` requires
TestNodeHostCallRefusesWorkersAndPromises to fail at
internal/native/library_node_fs_directory_test.go:37.
