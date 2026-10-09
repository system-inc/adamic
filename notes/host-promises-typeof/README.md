# Host promises and async typeof coverage

Base: origin/codex/host-promises-typeof at 961b302a25dfd8f02279ac63d96d39f08c872167.
This review covers the Adamic async/typeof surface. The existing C host bridge scenarios remain the native host tests; the source subset has no host creation or completion API.

## Cases and existing coverage

Only five existing executable oracle programs contained async syntax: async_plain.a, async_three.a, async_nested.a, async_throw.a, and async_typeof.a. The concurrency/refused async and await programs exercise rejection of async parallel work, not async execution.

| Code condition or source form | Existing executable fixture | New coverage |
| --- | --- | --- |
| Named function typeof, before declaration, parentheses | async_typeof | async_coverage_typeof |
| typeof used directly, in template, stored string | async_typeof | async_coverage_typeof |
| typeof compared with ===, !==, literal on either side | none | async_coverage_typeof |
| typeof inside async function before/after await | async_typeof | async_coverage_typeof |
| Same-named number parameter is not a function | async_typeof | unchanged |
| Same-named boolean parameter is not a function | none | async_coverage_typeof |
| Ordinary primitive typeof | async_typeof | async_coverage_typeof |
| No call or side effect from typeof | async_typeof | async_coverage_typeof |
| switch on named async typeof | none | unsupported_typeof_named_switch |
| Async arrow: comparison, template, switch, stored | none | four unsupported_typeof_arrow probes |
| Async class method: comparison, template, switch, stored | none | four unsupported_typeof_method probes |
| Compound typeof operand: conditional or identity | none (lowering unit tests only) | unsupported_typeof_conditional, unsupported_typeof_identity |
| Await primitive number, boolean, string | async_plain | async_coverage_discard |
| Await Promise.resolve(), number, true, dynamic string | async_plain, async_three | async_coverage_values, async_coverage_parameters |
| Named awaited calls, sequential calls, returned await | async_nested | async_coverage_values |
| Number arguments/results through named functions | async_nested | async_coverage_values, async_coverage_discard |
| String arguments to async functions | async_three, async_throw | async_coverage_parameters |
| String result received from a named function | async_three | async_coverage_values |
| String result forwarded by return await | none | async_coverage_values |
| Boolean arguments/results and return await | none | async_coverage_values |
| Void implicit result, returned await, discarded await | async_plain | async_coverage_discard |
| Never result, eager throw before first await | none | async_coverage_reject_eager |
| Mixed parameters, multiple declarations, locals live across suspension | none | async_coverage_parameters |
| Calling declaration before its source position | none | async_coverage_parameters |
| Parenthesized await callee, primitive string methods | async_plain (uppercase) | async_coverage_parameters (slice) |
| Empty and dynamic Unicode strings | none in async | async_coverage_values |
| Both booleans, signed zero, fraction, NaN, infinity | only true and ordinary numbers | async_coverage_values, async_coverage_parameters |
| Nonvoid result discarded, all primitive kinds | none | async_coverage_discard |
| stdout writes | all five | all new successful fixtures |
| stderr writes and ordering across await | none | async_coverage_parameters |
| Uncaught throw after await | async_throw | async_coverage_reject_nested |
| Rejection forwarded through several frames and return await | none | async_coverage_reject_nested |
| Empty Error message | none | async_coverage_reject_empty |
| Caught rejection | none | unsupported_caught |
| Explicit undefined returned | none | unsupported_undefined |
| Optional strings, including absent values, forwarded by named async functions | none | async_coverage_unions |
| Same-kind string and number literal unions | none | async_coverage_unions |
| Null, mixed-representation union, object, array, function, class, Map, Set, Error payloads | none | unsupported payload probes |

All new fixtures are explicitly registered in internal/oracle/oracle_test.go; the harness does not discover arbitrary files automatically. New count rows are in registry order. Native variants include release, ASan/UBSan with malloc, and ASan/UBSan with slabs, with leak checks for successful programs. Uncaught rejection fixtures exit 70; this harness normalizes source Node errors to Adamic panic text via oracle/adamic.mjs.

## Unsupported source surface

The checked diagnostics, Node stdout/stderr and exit codes are in observations.json. A rejected build has no native stdout or native exit code. These are compilation limits, not differing executable outputs.

* Methods are rejected by the unbound-method pass at internal/lower/refusals.go:137, even when only observed by typeof.
* Arrow declarations are rejected as function-valued locals at internal/lower/async.go:147; closures are also forbidden by asyncValue.
* Switch and try/catch are rejected by asyncBody's closed statement grammar. A caught rejection cannot currently be tested as an executable Adamic program.
* Compound function operands require async function values, rejected by functionValue in internal/lower/expression.go.
* Null, object, array, closure, class, Map, Set, Error and mixed-representation union payloads are rejected by asyncType at internal/lower/async.go:147. Explicit undefined is classified as void but then rejected as a value expression by asyncValue. Void is executable when returned implicitly or obtained through Promise.resolve(). Never is exercised by an unconditional throw. Same-kind literal unions and string | undefined share accepted primitive representations. number | undefined and boolean | undefined use pair representations instead and are rejected; their probes record that distinction. A conditional expression can produce an absent optional string even though a direct undefined expression is rejected; async_coverage_unions covers both present and absent paths.

Additional unsupported_*.a probes cover generic functions, synchronous helpers, anonymous/default declarations, optional/default/rest/destructured parameters, nested expression awaits, undefined await, void and uninitialized locals, destructuring, assignment, loops, finally, unawaited tasks, Promise executors/all/handles, console arity and numeric output, Error construction without a message, non-Error throws, and a recursive graph. Their observed diagnostics are in observations.json. The recursive function is deliberately never invoked on Node; observing its typeof terminates, while lowering still rejects the recursive graph. Several also have existing lowering tests in internal/lower/async_test.go. Async imports (the one-module restriction), thenable interop, user cancellation, and source host calls have no executable agreement witnesses because the implementation does not support them.

Host completion timing, late/duplicate completion, abandonment/cancellation, UTF-8 validation, hooks and thread affinity require the C host API. They cannot be invoked from this Adamic subset and were deliberately left to internal/native/testdata/async_host, as requested. Concurrent microtask ordering cannot be contrasted with inline resume from this closed straight-line source subset, since detached tasks and Promise handles are unavailable.

## Mutation proof

Changed exactly internal/lower/async.go:164 from l.constant("function") to l.constant("object"), ran the new async_coverage_typeof oracle uncached, and restored the file in a finally block.

Node stdout:

```text
true true true
boolean boolean
true function true
```

Native and JavaScript backend stdout under the mutant:

```text
false false false
boolean boolean
false object false
```

Both executions exited 0 with empty stderr. The oracle exited 1 on stdout disagreement, with no compiler refusal or clang warning. The restored compiler passes the async oracle again.

## Toolchain and commands

Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc: 5; cgroup cpu.max: 400000 100000.
Setup timing lines: go ready (0s); clang ready (1s); node ready (1s); submodules ready (1s); build cache warm (110s); done in 110s.
Environment: /workspace/adamic-tools/env.sh, sourced in each build/test shell.

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/async_' -count=1 -timeout 30m
python3 notes/host-promises-typeof/run.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go vet ./...
gofmt -l cmd internal
git diff --check
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
# With the one-line mutant, before restoring it:
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/async_coverage_typeof.a$' -count=1 -timeout 30m
```

run.py executes exactly these commands for every async fixture and unsupported probe, using a temporary output directory, and records all results:

```sh
go run ./cmd/adamic build <file> -o <out>
node --disable-warning=ExperimentalWarning oracle/node.mjs <file>
<out> # when the build succeeds
```

Test output was redirected to /tmp/host-oracle-final.log, /tmp/host-counts-final.log, /tmp/host-mutant.log, /tmp/host-vet.log, /tmp/host-builds-final.log and /tmp/host-gate.log and then read. Setup, initial checks, and the corrected standalone runner were also run before the final checks above. The standalone runner initially omitted the loader warning suppression; its final recorded run includes it.

## Final verification results

Eight new fixtures are registered, with eight new count rows. The final standalone observations contain 13 executable programs (the eight new programs and all five original async fixtures), all agreeing on stdout, stderr and exit code, plus 48 probes rejected at compile time. No differing executable program was found. Numeric/multiple-argument console probes are rejected by the TypeScript checker before the async console guards can run.

The final uncached focused async oracle passed in 4.941s; final full-table count regeneration passed in 27.653s. go vet, gofmt -l cmd internal and git diff --check passed. The original eight-fixture discovery sequence consisted of seven initial accepted programs and then the optional-string/literal-union program; the full repository gate started before that last fixture was registered. The final focused oracle, counts and standalone runs include all eight.

The full uncached repository gate exited 1. All packages other than stage1/cohere/markdownblocks passed, including internal/native (344.375s), internal/oracle (366.408s), and stage1/cohere/typeaware (1170.715s). TestMarkdownUnicodeWidths could not import /tmp/adamic-markdown-width/node_modules/emoji-regex/index.js. The same package reached its 30-minute timeout while leaf/list/root/table/whitespace checks were still running. This is a validation limit, not an observed async output disagreement. No claim is made that the full gate is green. Full output is preserved in gate.log; the semantic typeof mutant failure is preserved in mutant.log. The gate was allowed to finish and no remaining test or mutant processes were observed afterward.

The final static-check log is /tmp/host-vet-final.log. Before registration, the union program was also built and run independently as /tmp/async-unions.a with go run ./cmd/adamic build /tmp/async-unions.a -o /tmp/async-unions, node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/async-unions.a, and /tmp/async-unions. Both printed `undefined true value` then `b 2` and exited 0. Final observations.json records that same program under its committed fixture name.
