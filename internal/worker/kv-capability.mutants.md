# Workers KV capability, closed runtime ingress (#bxaz0p1 part 1)

`TestWorkerKVCapabilityMutantsAreCaught` uses the existing Worker buildcache
mutation products and runs each unmutated control before its mutated child.
No KV binding reaches a built Worker in this head.

| Mutation | Catching test | Catch line in kv_capability_test.go |
| --- | --- | --- |
| kv-slot-symbol: omit KV from the shared slot walk | TestWorkerKVCapabilityStops/let and /callback | 62 |
| kv-function-value: omit the direct-callee rule | TestWorkerKVCapabilityStops/callback | 62 |
| kv-registration: omit the host registration | TestWorkerKVCapabilityStops/object_literal | 62 |
| kv-ingress: admit the field without the bridge check | TestWorkerKVIngressWaitsForBridge | 166 |
| kv-field-name: trust the program interface's spelling | TestWorkerKVFieldIdentityIsPreludeSymbol | 103 |
| kv-purity: drop the capability loop in CheckPure | TestWorkerKVPurityRefusesOpaqueBindings | 151 |

The let, callback and object-literal checks require ordinary Worker lowering to
complete after the mutation: another lowering stop is not the named witness.
The ingress check requires the field-named NotYet and no output directory.
