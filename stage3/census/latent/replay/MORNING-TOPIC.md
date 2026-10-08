# Morning topic-only enumNeverValue unit

Built string-result widening for concrete generic callbacks without checked views.
Base 6998ebc24ae353193cb1495d3d51308131a4b5c7; four owned commits carried, no topic merge commits.
Node fixtures, eight generic mutants, nine union rules, counts and all touched packages pass.
Newly cleared morning root: parser.ts:2413:92; next refusal is a string as a condition at 2414:13.
Conservative morning coverage: generic 5/21, object unions 5/5, receivers 0/35; remaining cases below.

## Branch and change

Fresh branch: codex/notyet-enum-values-topic, based on the newest fetched main.
Only owned non-merge commits were cherry-picked:
64246ed0 -> b98e5b46, b96e6356 -> 42391987, aa7041d5 -> fff894b6,
5cdfbf71 -> 1159559c. Neither c41 nor views nor census replay commits were carried.
Historical reports describe the old branch; this report governs this topic branch.
`git rev-list --merges origin/main..HEAD` produces no output.

The generic source returns a string literal union; its context accepts string or
string | undefined. Both use the same string pointer/reference ABI, with identical
ownership. The existing closure's reference return can therefore be read directly
in either context. No adapter or backend change is needed for the new rule.
Exact instantiated parameters, represented concrete binders and shared function
identity remain required. Optional numeric results need a different ABI and stay
refused; object result widening is skipped because it needs checked views.

The separately registered generic_function_value_string_result.a fixture passes
a runtime-built string through literal narrowing, calls the same generic identity
through both result contexts, invokes it as a spelling callback, tests an empty
input and repeats source function identity. Source Node, JavaScript backend,
release native, ASan/UBSan native and LeakSanitizer agree on exit 0 and stdout:

```text
else
else
if
empty
true
```

Carrying the union helper exposed its reference to adamic_kind_typed_array_iterator,
which current main does not define. That unlanded-runtime dependency was removed
from the owned helper. Current main's object and map-iterator tags remain checked.
No runtime C file changed and no runtime helper was added.

## Validation

All commands source /workspace/adamic-tools/env.sh; tests write to logs.
Setup used GOPROXY=https://proxy.golang.org|direct and bash cloud/setup.sh:
node 0.027s, Go 0.027s, submodules 0.087s, markdown 0.094s, clang 0.208s,
build 42.327s, deferred tests 42.515s, cache 42.516s, done 42.546s.
nproc=5; cpu.max=400000 100000. Setup log: /tmp/notyet-topic-setup.log.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestGenericFunctionValueBoundaries|TestNativeAgreesWithNode/internal/oracle/testdata/(generic_function_(value|census)|union_object_kind)' -count=1 -timeout 10m > /tmp/notyet-topic-new-focus-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 internal/oracle/testdata/run-generic-function-value-mutants.py > /tmp/notyet-topic-generic-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-union-final-mutants.py > /tmp/notyet-topic-union-mutants.log 2>&1
ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-union-null-mutant.py > /tmp/notyet-topic-union-null-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/notyet-topic-counts.log 2>&1
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/fresh ./internal/oracle -count=1 -timeout 30m > /tmp/notyet-topic-packages.log 2>&1
```

Focused: lower 0.645s, oracle 1.168s. Counts: 20.557s, one new row and one
carried names fixture's retains changing 38 -> 36 on main; no other rows move.
Whole packages: IR 31.819s, lower 45.422s, native 235.985s, freshness 65.160s,
oracle 228.821s; JavaScript has no standalone tests and runs through the oracle.
All final checks exit 0. No full repository gate run.

The eight generic mutants compile and fail their intended checks:

| Mutant | Catcher |
| --- | --- |
| Remove specialization reuse | Node function identity disagrees |
| Permit multiple value specializations | identity boundary incorrectly lowers |
| Remove exact parameter proof | parameter boundary incorrectly lowers |
| Remove result compatibility proof | numeric optional-result boundary incorrectly lowers |
| Reject extra contextual arguments | extra-arguments fixture cannot lower |
| Invent a phantom binder | phantom boundary incorrectly lowers |
| Restore exact-result-only rejection | new string-result fixture cannot lower |
| Restore generic-value refusal | Node fixtures cannot lower |

Logs: /tmp/adamic-gate/generic-function-value-mutants-cfkmoujo/*.log.
Nine union rules also fail semantically: replace KindIs with typeof (stale fixture
misses its checked panic and ASan catches bad layout); corrupt native array/map/
object tags (normal fixtures falsely panic); corrupt JavaScript array/map/object
predicates (normal or stale output disagrees); exclude map iterator (iterator
read falsely panics); remove null protection (optional fixture trips UBSan).
The inverted null guard is also killed, but not counted as a tenth rule.
All temporary mutations restore their files. The initial carry's missing runtime
tag was a compile failure, not a semantic mutant; only the final checks count.

## Morning replay evidence and limits

Pinned census: e8c283b5, stage3/notyet-table/rerun-0730/after/roots.csv.
Exact rows: 35 receiver roots, 21 generic-value roots, five object-union roots.
The old filesystem tracePath reason is absent from this morning list.
The morning census overlay was read into /tmp/notyet-topic-census-tools and built
against current topic sources using a scratch Go overlay. None of its commits
or source files were added to the topic branch. Its no-output guards remain active.

All 26 generic/union roots were replayed before the new rule; after-log for the
parser uses the rebuilt /tmp/notyet-topic-replay-after binary:

```sh
/tmp/notyet-topic-replay-after -project /tmp/notyet-this-adapted/src/tsc/tsc.ts -where /tmp/notyet-this-adapted/src/compiler/parser.ts:2413:92 -kind NotYet -reason 'a generic function as a value' > /tmp/notyet-topic-parser-after.log 2>&1
```

Before: parser.ts:2413:92 reports incompatible instantiated result. After: that
stop disappears and 2414:13 refuses a string as a condition. Earlier parameter,
prefix-update and Iterable for-of stops remain. Replay exit 1 means the requested
old signature disappeared; it does not prove the selected whole unit lowers.
These are checker-rejected entry-root measurements, distinct from the ordinary
compiler acceptance and runtime oracle proof of the .a reduction.

Verified generic root count is five: core.ts:2378:40, core.ts:2442:107 and
sourcemap.ts:821:24 have no lowering findings; utilities.ts:6084:62 advances to
its element-access stops; parser.ts:2413:92 advances as described above.
The new rule adds one of the 21 morning roots to the four verified carried sites.
Uninstantiated T and genuinely polymorphic contexts remain NotYet. The bare
contains declaration is not counted from a replay that stops earlier on T.
A direct conditional identity reduction still meets an unrelated invariance
refusal; its fixture uses a typed local, as documented in NIGHT-AREA-RECHECK.md.
classFields.ts:2071:33 expects void from a value-returning callback; disposing
that result needs a separate calling-convention proof and is not added here.
Object widening would need checked views and was skipped.

All five original union signatures disappear on the main-based replays. The
selected tsbuildPublic.ts:265:50 unit has no findings; others retain unrelated
unknown, non-null, prefix-update, rest-parameter and branded-type stops.
The unchanged union fixture/mutant proofs certify the reason-site change.

Two receiver morning examples still reproduce the exact stop: checker.ts:1456:5
and utilities.ts:8491:5. docs/0.1.md limits this to methods and constructors;
docs/library_function_expressions_for_in.md refuses explicit and dynamic
receivers. No extension was made without that language ruling. Other unbound
reads were not accepted: the larger reading performance group (57 roots) uses
namespace imports, outside the module contract in docs/0.1.md, and needs a module
ruling rather than fabricated bindings.

Logs: /tmp/notyet-topic-replay-{0..25}.log, /tmp/notyet-topic-replays.json,
/tmp/notyet-topic-parser-after.log, /tmp/notyet-topic-this-{checker,utilities}.log.
