Merged the integration c2 stack into checked wider writes toward step 10.
Parents: 38263aa9 and c79c75726375c937f0b18e18040e6769d916ed06.
Required validation outputs and count-row changes are recorded below after completion.
The existing contract mutants and the new reserved optional-field literal-set mutant must fail by observable behavior.
The waiting-on-views refusals remain recorded; this merge does not claim those slices.

All nine content conflicts and their resolutions are recorded in the merge commit message. The merge combines checked contracts with async IR, optional-field reservation, strict overload boundaries and runtime c2's packed slot cache. It retains both overload implementations: checked .ts object results use their field contracts; other overloads keep the stack's result and argument adapters. Writable callback inputs use the existing .ts actual-slot checks when covariant results are proven, while .a and other visitor paths keep their proof rules.

Runtime metadata reads use the slot returned by lookup. Dynamic optional shapes retain copied and reserved field contracts in aligned memory owned by the object. JavaScript records the same contracts without creating absent own properties. A new temporary-source oracle probe holds a fitting write and a pinned misfit to Node; removing the reserved field's literal set builds and stores 2, which the pinned exit-70 observation catches in both backends. The existing runtime shape from the async host explicitly has no contracts.

The counts run also exposed a stale packed-cache access in the integration reuse emitter. Presence is now indexed from the returned slot. This fix is independent of the checked-write shape metadata.

Stage3 found two stale statuses after the semantic merge. generic-optional-return and structural-method-statics now compile and agree with recorded and current Node under sanitizers. Only their stage0 status values change. The .a overload refusal remains refused, with the stack's more precise result.emitNode path and proof fix.

Commands send output directly to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk > /tmp/stack-setup-ready.log 2>&1
source /workspace/adamic-tools/env.sh
go build ./... > /tmp/stack-checked-delivery-build.log 2>&1
go vet ./internal/... > /tmp/stack-checked-delivery-vet.log 2>&1
python3 /tmp/stack-a-check.py --out /tmp/stack-checked-delivery-a-check > /tmp/stack-checked-delivery-a-check.log 2>&1
go test ./stage3/fixtures -count=1 -timeout 15m > /tmp/stack-checked-delivery-fixtures.log 2>&1
go test ./internal/oracle ./cmd/adamic ./internal/lower ./internal/native -run 'TestCheckedWider|TestCheckedViewsNext|TestCheckedSignature|TestOptionalFunctionValueRelation|TestOverloadedShorthandFunctionValueStaysNotYet|TestCensusOverload|TestCheckedNeverContractMutant|TestCheckedDiagnosticReferenceMutants|TestCheckedFlowContainerContractMutants|TestCheckedEmit|TestExplainCheckedWritesOutput|TestProvenRelations|TestObjectRefusalsExplainSoundness|TestObjectUnprovenShapesStayNotYet|TestRuntimeFieldLayoutsAreIncluded|TestUniformFieldsMatchNode|TestCEndsInNewline|TestOverload' -count=1 -timeout 15m > /tmp/stack-checked-delivery-focus.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -timeout 20m -args -update-counts > /tmp/stack-checked-delivery-counts.log 2>&1
```

A-check uses the previously reported gate policy 3e339bbf06e0695f1b1223065ee8d4eef4813ba8 and selects .a files changed against c79c7572. All 15 files retain their intended 11 refused / 4 proven split.

The first setup attempted to build while merge markers remained and failed on those markers in internal/ir/ir.go. Setup was rerun after resolving them and passed. Final setup timing lines: node 0.071s, Go 0.073s, markdown 0.204s, submodules 0.264s, clang 0.330s, WASI SDK 0.465s, go build 68.329s, test binaries deferred 70.183s, cache warm 70.193s, total 70.669s. nproc is 5; cpu.max is 400000 100000.

Initial validation failures are retained in /tmp/stack-checked-final-focus.log, /tmp/stack-checked-final-counts.log and /tmp/stack-checked-final-fixtures.log. They identified the merge repairs above; only the delivery logs are the final results. Obsolete validation descendants were stopped before final runs.

Final validation: build and vet exit 0; a-check exit 0 (11 refused, 4 proven); stage3 PASS (159.723s); focused oracle PASS (313.937s), CLI PASS (28.244s), lower PASS (31.841s), native PASS (6.117s); counts regeneration PASS (399.372s). All existing branch contract mutants and the added reserved-field literal-set mutant are caught. Exact added, removed and moved rows against both parents, with before/after allocation, free, retain, release, peak and region counters, are in stack-counts.json.
