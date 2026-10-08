# Corrected area recheck and generic callback arguments

Built generic values with extra contextual callback arguments; added four Node fixtures.
Prior commits: receiver ruling 64246ed0, generic values b96e6356, unions aa7041d5, c41 merge 7f440e63.
Focused fixtures, counts, and complete lower/oracle package tests pass; commands below.
Seven generic mutants fail; the nine union rules are rechecked on the corrected branch.
Verified reason-site coverage: receivers 0/46, generic values 5/17, object unions 5/5, filesystem 0/1.

## Branch correction

The in-progress views-integration merge was aborted, never committed or pushed.
No views code remains. c41c0e062e99da37820f822968d4df1b48cdaee7 is retained
by 7f440e632d2a6ccec47327f01276229194be5b37. Its refusal of non-null assertions
in .a governs; the new reductions check indexed values explicitly instead.
`git merge-base --is-ancestor origin/codex/views-integration HEAD` exits 1.
No other unlanded worker branch was merged. Latest fetched area/compiler remains
b410340dc8f889b5799c3bc519117c63def3aa24. Push once after this whole unit passes.

## Change and oracle evidence

A concrete callback context may have more parameters than its generic source
function. JavaScript evaluates all supplied arguments and ignores extras.
The existing closure forwarder already reads only the source parameters, so
only genericFunctionValue's arity proof changes. Required source parameters
missing from the context, unresolved binders, mismatched instantiated types,
and multiple differently typed value specializations stay refused.
No backend, IR, other lower function, or runtime C change is needed for this unit.

The separately registered .a fixtures reduce contains' generic equality default,
canonical filename/prefix identity defaults, an object-position mapper's two
identity fields, and an indexed equality callback with an ignored argument.
They check repeated function identity and argument identity, early returns,
argument side effects, and a runtime-built ignored string's lifetime.
Node, JavaScript output, release native, ASan/UBSan native, and leak checks agree.

The canonical-name reduction binds identity to a typed local before selecting
it in a conditional. The direct conditional reduction encounters a separate
Refused from the generic invariance proof, although census measurement reports
no lowering findings. This is a limitation, not certification that the original
conditional passes the ordinary compiler. No other worker's proof was weakened.

## Commands and results

Commands source /workspace/adamic-tools/env.sh; every test writes to a log.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestGenericFunctionValueBoundaries|TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_census' -count=1 -timeout 10m > /tmp/notyet-night-census-fixtures.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 internal/oracle/testdata/run-generic-function-value-mutants.py > /tmp/notyet-night-generic-mutants-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/notyet-night-census-counts.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/notyet-night-generic-packages.log 2>&1
```

Focused tests: lower 0.358s, oracle 0.663s. Counts: 24.975s, four new rows,
all existing rows unchanged. Whole touched packages: lower 41.308s,
oracle 197.221s. All exit 0. No full gate run.

The committed mutant runner restores each source in finally. Its final seven
mutants all exit 1 without build errors or compiler panics:

| Mutant | Catcher |
| --- | --- |
| Remove specialization cache reuse | Node identity comparison, including mapper fields, becomes false |
| Allow multiple value specializations | boundaries/identity incorrectly lowers |
| Drop exact instantiated parameter proof | boundaries/parameter incorrectly lowers |
| Drop exact instantiated result proof | boundaries/result incorrectly lowers |
| Restore rejection of extra contextual parameters | extra-arguments Node fixture cannot lower |
| Invent a phantom binder from another binder | boundaries/phantom incorrectly lowers |
| Restore the original generic-value refusal | the four census fixtures and original fixture cannot lower |

Final logs: /tmp/adamic-gate/generic-function-value-mutants-on78amru/*.log.
The first old-refusal mutant had an unused Go variable and was an invalid kill;
it was corrected to preserve that binding and rerun. Only the final semantic
kills above count.

## Replays and raw CSV accounting

Table pin 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7, roots/raw.csv:
46 receiver rows/roots, 22 generic observations/17 roots, five union roots,
one filesystem root. Deduplication is (kind, where, reason, text).
All 69 distinct roots were replayed after c41. The 22 generic observations
were additionally replayed using their raw CSV unit, which preserves the
instantiated caller instead of selecting an uninstantiated dependency declaration.

```sh
go run ./stage3/census/latent/replay -project /tmp/notyet-this-adapted/src/tsc/tsc.ts -where /tmp/notyet-this-adapted/src/compiler/core.ts:232:1 -kind NotYet -reason 'a generic function as a value' > /tmp/notyet-night-generic-context-0.log 2>&1
```

A replay exits 1 when the requested old signature disappears, including when
there are no findings. Parse its findings rather than treating exit 1 as failure.
These are explicitly checker-rejected entry-root census measurements, not a
claim that the TypeScript project passes Adamic's ordinary checker.

| Generic root site | Verified selected-unit result |
| --- | --- |
| core.ts:220:112 | CSV caller indexOfAnyCharCode at core.ts:232:1 lowers; bare contains still stops on T |
| core.ts:2378:40 | No lowering findings; conditional reduction caveat above |
| core.ts:2442:107 | No lowering findings |
| utilities.ts:6084:62 | Advances to ElementAccessExpression at 6086:20 and 6088:68 |
| sourcemap.ts:821:24 | No lowering findings |

Conservative verified count is 5/17. core.ts:814:162 advances from the arity
stop to an unresolved type parameter in its uninstantiated declaration; the
new fixture certifies the concrete callback pattern, not a census-site unlock.
core.ts:780:53's CSV caller at utilities.ts:6094:5 retains a structural-method
stop at core.ts:776:9 and reading insertIndex at 781:9, plus refused casts.
It is not counted. Other sites retain unresolved T, unknown/object, unsupported
function-value union parameters, or absent concrete contexts. Function equality
on genuinely polymorphic values needs a shared source identity and representable
calling convention; the current specialization proof does not provide that.

Receiver replays: 33 reproduce the exact refusal, 13 stop earlier on __String.
checker.ts:1456:5 still refuses this; utilities.ts:8473:5 stops at 8471:51 on
__String. The language receiver/construction ruling in THIS-OUTSIDE.md is still
required. No receiver extension was built, and no coverage is claimed.

All five union reasons remain absent. checker.ts:37210:16 now retains only
core.ts:1750:25 unknown in its selected measurement, rather than the previous
non-null stop. tsbuildPublic.ts:265:50 lowers. The remaining three retain their
ResolvedConfigFileName/Path, rest-parameter, and array stops in UNION-KINDS.md.
The nine union semantic mutants are recertified separately below.

Filesystem: both the diagnostic-position replay and the raw CSV startTracing
unit at tracing.ts:59:5 stop on any at tracing.ts:41:9. The corrected base has
no node:fs.openSync bridge; optional_node_host.go is an unhandled stub.
The host bridge must land through area/compiler before this context can be
certified. No unbound-local acceptance and no filesystem coverage are claimed.
The prefix-value dependency (1b185548 on codex/notyet-statements-small) is named,
not merged; it and the node host bridge are unlanded dependencies.

Raw replay logs: /tmp/notyet-night-area-replay-{0..68}.log and summary JSON.
Final generic contexts: /tmp/notyet-night-generic-context-{0..21}.log and
/tmp/notyet-night-generic-contexts-final.json. Filesystem raw-unit log:
/tmp/notyet-night-fs-context.log. Setup timings and nproc=5 are in THIS-OUTSIDE.md.

## Corrected-branch union certification

```sh
ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-union-final-mutants.py > /tmp/notyet-night-area-union-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-union-null-mutant.py > /tmp/notyet-night-area-union-null-mutant.log 2>&1
```

All nine distinct semantic mutants again fail with exit 1: replace KindIs with
typeof, corrupt native array/map/object tags, corrupt JavaScript array/map/object
predicates, exclude the native map iterator tag, and remove native null protection.
The first runner also checks the inverted null guard; the second checks its
removal, so this is nine distinct rules, not ten. Catchers are the normal,
optional, iterator, and checked stale fixtures described in UNION-KINDS.md.
The full oracle package run on this corrected branch passes those fixtures.
Source restoration leaves no backend changes; git diff --check is clean.

Final restored-source check: ADAMIC_GATE_UNCACHED=1 go test ./internal/lower
./internal/oracle -run 'TestGenericFunctionValueBoundaries|TestNonNull|TestNarrowedUnionObjectTagLowers|TestNativeAgreesWithNode/internal/oracle/testdata/(generic_function_(value|census)|union_object_kind)'
-count=1 -timeout 10m > /tmp/notyet-night-area-final-focus.log 2>&1.
Passes: lower 0.378s, oracle 1.240s. Formatting reports no files.
