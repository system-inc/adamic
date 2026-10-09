# Split refinement: panic and JSON are Wasm-safe

Panic is a call abort, not an impurity for this boundary: the JavaScript bridge
handles AdamicPanic and the Wasm host handles exit 70 by discarding the instance.
Both answer 500 and preserve the panic message. This includes Coalesce.Panic and
compiler-inserted checks; recursive traversal still examines the panic message's
operands for I/O. JSONDecode was already recognized and remains pure; JSONEncode
is now recognized too. Their results depend on arguments, with no I/O. Unknown
operations remain impure and include their IR name in the reason. Signature,
parameter mutation, crossing-cost and entry-point rules are unchanged.

## Real compute Worker

Source: origin/codex/workers-compute-a at 9aa73a2, workers/compute/handler.a.

The starting area/platforms tip b7c75f5 contains split but no encoder; the compute
branch contains the encoder but no split. An isolated, uncommitted merge of the
two in /tmp/adamic-split-refine-compute supplies a complete integration witness.
Only the refinement's split implementation and fixtures are overlaid afterward.
No encoder/compiler changes are committed on the refinement branch. The final
branch is fast-forwarded onto the current platform base 5133ea7 before landing;
the split and schema builder are unchanged by that update.

Before, running go run ./cmd/adamic-split workers/compute/handler.a:

```text
json javascript signature: parameter 3 not crossable
error javascript impure: unknown operation JSONEncode
invalid javascript impure: calls error
integer javascript too small to cross
skuValid wasm entry
primes wasm entry
numberAt javascript impure: panic
summarize javascript impure: calls numberAt
closure javascript too small to cross
quote javascript signature: parameter 1 not crossable
asciiToken javascript too small to cross
topWords javascript signature: return not crossable
closure javascript too small to cross
handle javascript impure: unknown operation JSONEncode
```

After, running the same command and source:

```text
json javascript signature: parameter 3 not crossable
error javascript signature: parameter 3 not crossable
invalid javascript signature: return not crossable
integer javascript too small to cross
skuValid wasm entry
primes wasm entry
numberAt javascript too small to cross
summarize wasm entry
closure javascript too small to cross
quote javascript signature: parameter 1 not crossable
asciiToken javascript too small to cross
topWords javascript signature: return not crossable
closure javascript too small to cross
handle javascript signature: parameter 1 not crossable
```

summarize now qualifies as a crossing entry point alongside primes. numberAt
is pure but still too small to cross independently. The encoder callers stay
JavaScript because their signatures cannot cross, rather than because encoding
is classified as an unknown operation.

## Fixtures and observations

panic_decode.a covers explicit panic, a numberAt-style coalescing panic,
a summarize-style loop returning a readonly record, pure decoding, a noncrossable
decode-result union, and console I/O. compute_json.a adds schema-driven encoding
and mirrors the compute Worker's helper and summary shape.

All fixtures run as source on Node, emitted JavaScript on Node, and sanitized
native. The encoder fixture runs in the integration witness. On a compiler base
without the encoder, TestFixtures reports an explicit skip for compute_json.a;
the test enables automatically when the implementation is present.

Observed new fixture output, identical in all three executions:

```text
panic_decode.a:
summary
6 2 2 Ok

compute_json.a:
summary
{"count":3,"sum":6}
{"count":2,"sum":9}
{"count":2,"sum":13}
```

## Checks and mutants

Logs use /tmp/workers-split-refine- as the prefix. Commands:

    source /workspace/adamic-tools/env.sh
    go test -count=1 -v ./internal/split ./cmd/adamic-split > /tmp/workers-split-refine-base-tests.log 2>&1
    go vet ./internal/split ./cmd/adamic-split > /tmp/workers-split-refine-vet.log 2>&1
    ADAMIC_SPLIT_SWEEP=1 go test -count=1 -timeout 5m -run TestAnalyzeOracleFixtures -v ./internal/split > /tmp/workers-split-refine-sweep.log 2>&1

In the isolated integration witness:

    go test -count=1 -v ./internal/split ./cmd/adamic-split > /tmp/workers-split-refine-integration-tests.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -v ./internal/oracle -run 'TestJSONDecode|TestJSONEncode' > /tmp/workers-split-refine-json-oracle.log 2>&1

All passed. The sweep covered 344 compilable programs. The uncached oracle took
39.8s: 48 native misses, 85 Node misses, no hits. Each of these six mutations
was applied alone, produced an assertion failure (exit 1, no build failure),
and was restored. Summary: /tmp/workers-split-refine-mutants.log.

| Mutant | Check that caught it |
| --- | --- |
| Console accepted as pure | TestFixtures: logger/handler golden decisions |
| Unknown expression accepted as pure | TestUnknownOperationIsImpure: futureOperation must be named in reason |
| Panic marked impure again | TestFixtures/panic_decode.a: checked becomes impure |
| Coalescing panic marked impure again | TestFixtures/panic_decode.a: numberAt and summarize regress |
| JSONDecode removed from recognized operations | TestFixtures/panic_decode.a: decodeSummary regresses |
| JSONEncode removed from recognized operations | TestFixtures/compute_json.a in integration witness: encodeSummary regresses |

Setup: go ready 0s; clang/node/submodules ready 1s; cache warm and total
128s; nproc 5. Log: /tmp/workers-split-refine-setup.log.

The 600-request corpus was not rerun; the supplied corpus result motivates the
change but is not claimed as a new observation. No Wasm/bridge code, lowering,
native emission, or JavaScript emission changes. The full repository gate and
workerd were not run.
