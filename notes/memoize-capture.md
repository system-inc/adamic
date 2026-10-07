Built: restored opt-in cycle tracing and pinned the assertion stop and real self-capture refusal; the finder gate is unchanged.
Commits: implementation 0aaf96d95d980f8b1ab9b47e52b0732de5332eb6; base origin/main 48c05d091f0a43c31cbe051b1d6578d99eeedf19.
Checks: complete internal/lower passed (26.234s), filtered closure oracle passed (6.626s), repository vet and formatting passed.
Mutants: actual callback = result self-capture refused; disabling assignable closure edges made its regression test fail with got <nil>.
Limits: no full repository gate or graph-region implementation; host undefined! clears the pointer but its readiness guard panics before the next condition can evaluate false.

## Observations and interpretation

Pinned main stops before cycle finding on the exact one-line reproducer:

```
adamic: /tmp/memoize-original.a:1:127: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case
exit status 1
```

Node prints `cwd`. I fetched the named host branch and reproduced its refusal in a detached scratch worktree at 08b5b2c430e41f72728d2fd4458169044d08f733, adding only the same tracing helper. Its cycles.go was identical to pinned main before instrumentation. With ADAMIC_TRACE_CYCLES=1 the path is:

```
cycle reach: cell callback -- slot contents --> () => string
  () => string -- closure capture at /tmp/memoize-original.a:1:72 --> cell callback
  cell callback -- same captured cell --> cell callback
adamic: /tmp/memoize-original.a:1:21: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function callback() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)
exit status 1
```

The captured callback cell holds a function of type () => string. The returned closure is assignable to exactly that type and captures that cell. This is a real cycle-capability path under the documented closed-world, type-based rule: a permitted value of the cell's type can lead back to the cell. It is not a fabricated signature match. A corresponding executed store can close the cycle:

```a
function tie(callback: () => string): () => string {
    const result = () => callback();
    callback = result;
    return result;
}
const get = tie(() => 'cwd');
console.log('made');
```

The supplied memoize invocation itself has no actual heap cycle: its callback is a leaf closure, and its only callback write clears it. That distinction matters. Accepting this program precisely would require a value-flow or fresh-write relaxation for captured cells; docs/memory.md expressly leaves captured cells outside the existing relaxation. I have not removed a valid edge or claimed the concrete call creates a cycle. The new regression test keeps the actual self-capture refused with the callback diagnostic and adamic/cycle-capable code pinned.

Runtime graph regions are the appropriate ownership direction for admitting potentially cyclic strong closure/cell graphs. They would need to account for both closure-to-cell capture edges and cell-to-closure value edges, with external owners controlling lifetime. This unit does not implement or validate that runtime work. Such ownership does not resolve the separate readiness behavior below.

## Native undefined behavior

To isolate the write from the cycle gate on the named host branch, I used:

```a
function clear<T>(callback: () => T): void {
    if (callback) { console.log('before:true'); callback(); } else { console.log('before:false'); }
    callback = undefined!;
    if (callback) { console.log('after:true'); callback(); } else { console.log('after:false'); }
}
clear(() => 'cwd');
```

Node prints before:true then after:false. The sanitized native binary prints before:true then exits 70 with:

```
adamic: panic: read before assignment: variable 'callback' in callback
```

Generated C contains the following sequence (temporaries abbreviated only by omission of surrounding code):

```c
adamic_closure * adamic_temporary_4 = adamic_local_0_callback;
adamic_local_0_callback = NULL;
adamic_release(adamic_temporary_4);
adamic_ready_0 = false;
if (!adamic_ready_0) {
    static const char message[] = "read before assignment: variable 'callback' in callback";
    adamic_panic(message, sizeof message - 1);
}
adamic_closure * adamic_temporary_5 = adamic_local_0_callback;
if (adamic_temporary_5 != NULL) {
```

The old callable is released and the stored value is NULL. It does not retain a callable value. The pointer predicate would be false, but the readiness guard prevents reaching it. Therefore the requested false-after-undefined! demonstration fails on this host revision; this is an observed native behavior mismatch, not a successful proof of that requirement. Pinned main does not support the assertion at all. I did not edit native lowering for this separate host feature.

An optional-typed control demonstrates ordinary undefined clearing:

```a
function clear(callback: (() => string) | undefined): void {
    if (callback) { console.log('before:true'); callback(); } else { console.log('before:false'); }
    callback = undefined;
    if (callback) { console.log('after:true'); } else { console.log('after:false'); }
}
clear(() => 'cwd');
```

Both Node and the sanitized host native executable print before:true then after:false; native exit is 0.

## Validation and setup

All test outputs were redirected to log files, without piping. Environment: nproc=5; CPU quota 4; Go 1.27.1, Node 24.19, clang 20.1.8. GOPROXY was https://proxy.golang.org|direct. The first setup overlapped my import edit and failed with `internal/lower/cycles.go:5:2: could not import os (open : no such file or directory)`; I reran setup after stabilizing the source. Successful bash cloud/setup.sh timings in /tmp/memoize-setup-retry.log:

```
setup: go ready (0.183s)
setup: node ready (0.230s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.026s
setup: markdown dependencies ready (0.495s)
setup: submodules ready (0.623s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1.803s)
setup: go build ready (100.398s)
setup: test binaries deferred (use --warm-tests) (100.662s)
setup: build cache warm (100.669s)
setup: done (100.926s)
```

Sourced /workspace/adamic-tools/env.sh for builds and checks.

- ADAMIC_TRACE_CYCLES=1 go test ./internal/lower -run 'TestMemoize' -count=1 -v: passed, /tmp/memoize-focused.log.
- Temporarily disabled all assignable closure edges, then go test ./internal/lower -run TestMemoizeSelfCaptureIsCycleCapable -count=1 -v: failed as required, `want the self-capturing callback cell refused, got <nil>`, /tmp/memoize-cycle-mutant.log. Restored before subsequent checks.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestMemoize|TestNativeAgreesWithNode/internal/oracle/testdata/closures' -count=1 -timeout 30m: lower passed 0.328s and oracle passed 6.626s, /tmp/memoize-filtered.log.
- go test ./internal/lower -count=1 -timeout 30m: passed 26.234s, /tmp/memoize-lower.log.
- go vet ./... and gofmt -l cmd internal: empty successful logs /tmp/memoize-vet-all.log and /tmp/memoize-gofmt-all.log. git diff --check passed.
- Exact main and host traces: /tmp/memoize-main-trace.log and /tmp/memoize-host-trace.log; Node original output /tmp/memoize-node.log.
- Host native probes: go run ./cmd/adamic build /tmp/memoize-clear.a -o /tmp/memoize-clear-native --sanitize and corresponding optional control. Outputs in /tmp/memoize-clear-{node,native,build,exit}.log and /tmp/memoize-clear-optional-{node,native,build,exit}.log; generated C /tmp/memoize-clear.c.

Fetched current origin/main and merged it into this branch: Already up to date. No protected native files, runtime files, oracle_test.go, or cohere source were edited. No PR was opened. The previous arena never[] fix was deliberately not ported to this pinned-base unit; only its trace instrumentation was restored.
