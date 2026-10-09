Built representation helpers for self-comparison toward step 04, task #7c6b4pq.
Base: origin/main 031a1259; delivery SHA accompanies the push report.
Validation: source Node, both backends, release, sanitized, counted and wasm32-wasi pass; a-check and stage3 pass.
Mutant: replace numeric self-equality with true; sweep.a catches NaN while the mutant compiles and exits cleanly.
Not covered: source == and != remain refused by the existing coercion rule; no new coercion semantics.

Number and boolean equality and reference identity go through static inline helpers with distinct parameters. Numeric equality uses C's IEEE equality, so NaN is unequal to itself and signed zeros are equal. Reference equality compares object pointers, including canonical closures and receiver-returning array operations. String value comparison, optional scalar comparisons and boxed union comparisons keep their existing helpers. Both equality and inequality use this path; inequality negates equality.

The common native binary emitter serves expression comparisons, switch tests and view-field tests. Closure operands already use their canonical closure representation on this path. The JavaScript backend retains === and !==. It needs no change.

Three new .a fixtures pin Node's observed output. closure.a is the nested-function witness; array-fill.a checks the returning receiver and stored elements. sweep.a compares a closure, object, array, string, boolean, NaN, negative zero and one with themselves, and checks signed-zero equality, separately constructed string value equality and distinct object identity.

TestSelfCompare builds each fixture in all four native modes, compares every output with source Node and the JavaScript backend, and checks LeakSanitizer and counted allocation balance. TestSelfCompareConstantMutant changes the emitted numeric self-equality to true. The sanitized mutant builds and exits 0 without stderr. Node says `NaN: false true`; the mutant says `NaN: true true`. The test requires that exact observable error, not a compiler or sanitizer failure.

Commands (output goes directly to the named logs):

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh --wasi-sdk > /tmp/self-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestSelfCompare' -count=1 -v -timeout 10m > /tmp/self-focused-ready.log 2>&1
python3 /tmp/self-a-check.py --tree /workspace/adamic --out /tmp/self-a-check > /tmp/self-a-check.log 2>&1
go test ./stage3/fixtures -count=1 -timeout 15m > /tmp/self-stage3.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(closure_convention_|closures.a|method_closures.a|unions.a|numbers.a)' -count=1 -timeout 10m > /tmp/self-convention.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 15m -args -update-counts > /tmp/self-counts.log 2>&1
```

The a-check uses gate policy 3e339bbf06e0695f1b1223065ee8d4eef4813ba8 from Git because that runner is not present on this main checkout. It selects only .a files added or changed against origin/main: all three compile, zero failures. No whole package oracle or full gate was run.

Setup passed. Its timing lines were node 0.041s, go 0.046s, markdown dependencies 0.135s, submodules 0.135s, clang 0.307s, WASI SDK 5.116s, go build 123.899s, test binaries deferred 124.343s, cache warm 124.345s, total 124.398s. nproc is 5; cgroup cpu.max is 400000 100000. Node is v24.19.0, Go is go1.27.1 and native/WASI clang is 20.1.8.

An initial focused test failed because it passed relative paths into the oracle cache's source identity routine; the corrected test uses absolute paths. A preliminary CLI reproduction used incorrect flag ordering and printed usage, so it is not claimed as an observed compiler regression. The final all-mode runs above are the validation evidence.

Final outputs: focused oracle PASS (45.082s), stage3 fixtures PASS (69.190s), closure convention and numeric oracle PASS (31.339s), TestCountsAreRecorded PASS (106.022s). The counts diff adds exactly three rows and changes no existing row. All three changed .a files passed a-check.
