# Named nested function declarations

Design written before implementation, against origin/main ef3d907.

## Existing implementation

Arrows lower to an IR function with Closure set and a MakeClosure expression.
Ordinary function expressions use the same convention, reject dynamic this, and
bind a named expression to ClosureSelf. Each captured local moves into one counted
cell. Every intervening closure carries that cell; writes update it and reads
snapshot its value before another call can change it. A closure owns its cells,
its caller owns a returned closure, and a containing array owns its elements.
The cycle finder follows function types into every compatible closure's captures.
Captured cells are never relaxed by the fresh-write proof.

## Representation and scope

Allocate one shared environment record per enclosing frame, at one explicit
IR allocation site. Its stable layout lists captured slots in local-index order,
with each slot holding its value, reference-kind flag, and initialization-ready
bit. Siblings carry code plus access to this same record. Captured parameters
start ready; body locals start unready. Forwarded ancestor slots keep their
ancestor environment alive and are not copied as independent values.

The record is counted on the heap until an escape proof selects another
placement. Captured cells are interior views into the record, not separate heap
allocations. Existing arrow/function-expression captures of those views must
retain the record, so a mixed named/anonymous closure cannot outlive its storage.
The environment destructor releases every reference-valued slot once. Cells must
never individually free interior storage. Closure identity remains a separate
counted code/environment value, as the existing ABI requires.

Represent the allocation as an explicit AllocateEnvironment IR construct whose
ordered Cells are the complete frame layout. The function records that same
layout, and the construct occurs once at entry, before hoisted declaration
initialization and captured parameter initialization. Keep it visible through
flow and native statement walking rather than synthesizing unrelated cell_new
calls in emission. An empty capture layout needs no environment allocation.

internal/native/region.go currently discovers allocations through statement
walking and conservatively treats closure parameters/captures as escaping. A
future environment escape fixed point can key a placement map by this exact IR
statement and follow MakeClosure edges through Function.Environment. If no
closure retaining any slot escapes, region.go may select frame stack storage
or the enclosing call's region for the entire record at that site. If any such
closure escapes, the single site retains heap placement. The closure records
must also receive compatible placement/lifetime treatment before count elision.
Reference-valued contents still need destruction at scope/region end; allocation
placement does not prove those contents disposable. This unit does not implement
the escape proof, stack placement, region placement, or count elision.

The allocation profile at codex/tsc-allocation-profile 1e7a52b,
docs/stage3-tsc-profile.md, observes 978.0 MiB at checkTypeRelatedTo and 763.9 MiB
at getFlowTypeOfReference, about 48% of the 3,618.6 MiB sampled compiler allocation.
Those functions contain nested helpers capturing scratch locals. Non-escape and
closure/context attribution are profile inferences, not proven facts. This layout
makes the proposed lifetime optimization one decision at one allocation site
without relying on either inference for correctness today.

Every sibling has exactly the same environment layout. A direct sibling call
calls the sibling's code with the current environment carrier. Neither function
captures a variable holding the other function. Mutual recursion adds no strong
edge between function values. Returning a declaration from its enclosing frame,
passing it, or storing it in an array keeps its counted closure and cells alive.
Every invocation of the enclosing function creates new cells. Three lexical
levels forward the same ancestor cells through the intermediate closure.

Initially accept declarations directly in a function body. Block-scoped
function declarations, generic nested declarations, dynamic this, first-class sibling references inside another sibling, and
calls to a declaration in a different ancestor group remain loud NotYet cases.
Optional/default/rest parameters retain existing closure ABI gaps. Plain
destructured parameters retain the existing lowering support. Rebinding is rejected
by the TypeScript checker (TS2630), with a defensive lowering guard.
These are implementation gaps, not permanent language refusals. A sibling call
is supported; returning a sibling value from inside another sibling is not yet
supported, because preserving declaration identity needs a separate binding
strategy. Never substitute a freshly allocated closure and change ===.

## Cycles

A captured variable holding a nested function that reaches that variable closes
a cycle. Register every nested function in the existing cycle finder, including
the complete shared environment, and refuse this with adamic/cycle-capable.
No fresh-write relaxation for cells. Shared sibling environments can retain more
cells than a function reads, so the finder must use the complete retained vector,
not merely syntactic reads. This deliberately makes refusal conservative.

For example, refuse:

```a
function make(): () => number {
    let saved: (() => number) | undefined = undefined;
    function read(): number { return saved === undefined ? 0 : saved(); }
    saved = read;
    return read;
}
```

Restructure so saved is not captured, or declare saved Weak<() => number> and
keep read in a separate strong owner. The same rule applies when the captured
slot is an array or an object containing the closure. Never silently leak.
A mutant that captures sibling declaration bindings as ordinary closure cells
must pass output comparison and fail the leak check, proving that output alone
cannot certify this representation.

## Hoisting and the temporal dead zone

Prebind all direct let/const names and nested function signatures before lowering
any sibling body. Allocate captured cells at block entry, empty and not ready.
Then initialize nested declaration values before source statements. Original
let/const initializers still execute at their original lines. A cell's ready bit
becomes true only after initialization completes. Captured parameters start ready.

Extend the existing Checked read/write and ready-check machinery to captured
cells. Reads check before fetching the value. Assignments evaluate their right
side before checking readiness, as JavaScript does. A call before a nested
function's declaration is valid. A call before a captured let/const declaration
panics with ReferenceError: Cannot access 'name' before initialization in both
backends. A not-ready reference cell contains NULL so unwinding never releases
uninitialized storage. Initialization is distinct from assignment.

The JavaScript backend must use explicit cells and readiness as native does;
Node executing the original source is the independent TDZ oracle. Preserve the
existing limitation that inserted panics do not become catchable exceptions.

## Analyses and ownership

Keep captured variables excluded from SSA and reuse sources. A nested call can
write captured state, so direct code dispatch must retain the conservative
CallClosure effects for freshness, mutable ranges, array-element borrowing and
lent reads. Do not promote it to a pure named call just because its target is
known. Nested parameters keep the closure convention's owned counts. Captured
values and closure environments stay on the heap; regions must not allocate
anything that a closure can keep. No global disabling of regions or Perceus.

Preallocation adds empty declarations before initialization. Analyses must see
the real initialization and every subsequent write. The ready bit is not a
proof of freshness. Captures still escape even when the body only calls another
sibling. Throw and return cleanup release frame-owned cells and closure values
exactly once; escaped closures keep their own cell references.

## Evidence required before completion

Run minimal/r01, read/write captures, hoisting, even/odd, returned closure,
array-stored closure and three-level capture fixtures against original-source
Node, both backends, ASan/UBSan and LeakSanitizer. Add refusal probes for every
listed gap and a captured function cycle. Record counts. Run mutants for sibling
capture cycles, lost write-through, lost hoisting and omitted readiness checks.
Report actual observations separately from predicted results.

The pinned census is 429c117 and its filename is stage3/census/REPORT.md.
Its 5,574 sites in 57 files are a source inventory, not 57 successful checker
runs reaching NotYet. The corpus stops at checking. Rerun the inventory and
actual minimal probes separately; never describe a removed diagnostic branch
as proof that tsc compiles. Count files clearing this specific source-form gap
only after classifying retained NotYet cases, rather than claiming all 57 clear.

## Implementation observations

The first design was published as b2034215240ebaf36b6ca86d2c288da82638bc7f before
implementation began. The profile-driven single-record design was published as
dd43be0c3976a0bce48877f4dbe05a71053fae67 before implementing that revision. The initial implementation used the closure ABI, hoisted
bindings, shared captured cells, equal sibling layouts, and direct sibling code
dispatch. No protected emitter, lowering entry point, or oracle registry file was
edited. No cohere code was copied. Freshness, borrows, regions, and Perceus remain
enabled. Direct calls still have CallClosure effects; there is no purity shortcut.

The executable fixtures are nested_minimal (byte-identical to census r01),
captures, hoisting, mutual, returned, array, three_levels, tdz, tdz_write, weak,
destructured, destructured_tdz, mixed named/arrow captures, and captured
destructured parameters. Their original source runs under Node;
the oracle compares native and JavaScript backend output and exit status, runs
ASan/UBSan, and checks native successful executions for leaks. TDZ aborts exit
70, so their counts intentionally include allocations not freed at process abort.
The negative cycle fixture lives outside the executable corpus, in
internal/oracle/refusals/nested_cycle.a, and has its own refusal test.

Counts are recorded in internal/oracle/counts.md. Existing fixture rows do not
change. Minimal allocates/frees two heap values. Mutual recursion allocates/frees
ten, retains 46 times and releases 52 times: the current carrier is retained on
each sibling call. With the single record, read/write captures allocate/free 22
instead of 24, returned captures 21 instead of 23, and array-stored captures 17
instead of 19. Peak live values fall from 8 to 7, 13 to 11, and 11 to 9 respectively.
These measurements include fixture strings and arrays, not solely environments.
The ABI still carries cell pointers and retains the owning record once per
captured pointer. This is correct but does not minimize reference-count traffic.
On this amd64 build sizeof(adamic_cell) is 40 and the environment header is 24
bytes. The prior standalone cell with readiness but without owner was 32 bytes.
Fewer allocations do not establish lower total bytes or faster execution.
No optimization was disabled, so there is no disabled-optimization delta.

The implementation emits AllocateEnvironment once per named-declaration frame
with captures. FrameEnvironment identifies its ordered slot layout; EnvironmentCell
marks interior views. heap.c redirects retain/release of an interior view to its
owner, and environment destruction releases every reference-valued slot. Empty
layouts allocate nothing. Captured destructured parameters belong to this layout.
Anonymous closures retaining any slot have their full retained layout expanded
for cycle analysis. A function reading only count can therefore be refused when
a disjoint saved slot in the same record holds that function.

Block-local anonymous captures keep their existing per-execution cells; those
lexical environments can have different instances per iteration. This unit's
single frame record covers function-body bindings and parameters accessible to
the supported direct-body named declarations. It does not merge distinct block
instances or convert the entire anonymous-closure implementation into regions.

## Refused programs and mutation evidence

The cycle example above is refused by adamic/cycle-capable. Weak is independently
tested as an accepted, leak-free alternative. These remaining gaps have named
refusal probes in internal/lower/nested_functions_test.go:

* A declaration inside an if block.
* A generic inner<T> declaration.
* inner(value?: number), inner(value = 1), or inner(...values: number[]).
* Returning sibling a from b, or storing inner as a value inside inner itself.
* A third-level function calling a declaration from its grandparent's sibling group.
* A declaration with a dynamic this parameter.

These are conservative implementation limits. Slotless parameter/results and
other unsupported representations also retain their existing NotYet guards.
Nested functions in methods/classes and every possible interaction with existing
unsupported syntax have not received dedicated fixtures.

Every requested mutant was run. Production sources were restored afterward.

| Mutant | Observation and catcher |
| --- | --- |
| Siblings capture both closure bindings instead of direct dispatch | Node output still agrees; LeakSanitizer detects 512 bytes in 10 allocations. Permanent IR mutant test. |
| Drop numeric write-through to a captured cell | Node prints updated counters; native prints stale counters. Output comparison fails without a sanitizer or leak failure. |
| Initialize declarations at their source line | Call-before-declaration fails: native UBSan reports a NULL closure access; JavaScript backend throws TypeError; Node prints 7. |
| Omit captured-read readiness check | Node exits 70 with ReferenceError; native prints 0 and exits 0. |
| Omit captured-write readiness check | Node exits 70 with ReferenceError; native prints 5 and exits 0. |
| Omit nested closure registration in cycle finder | Refusal test fails; accepted program agrees with Node but LeakSanitizer finds 104 bytes in two allocations. |
| Drop an interior cell's owner retain | Returned-closure fixture fails under ASan with heap-use-after-free. |
| Omit AllocateEnvironment | The IR one-site invariant test reports zero allocations instead of one. |
| Hide disjoint slots retained through the same record | A dedicated cycle refusal test fails, accepting the cycle. |
| Delete binder.ts:567 from the independent census inventory | Census coordinate audit throws nested census site mismatch. |

Logs are /tmp/adamic-nested-mutant-lost-write.log,
/tmp/adamic-nested-mutant-no-hoist.log, /tmp/adamic-nested-mutant-no-tdz.log,
/tmp/adamic-nested-mutant-no-write-tdz.log,
/tmp/adamic-nested-mutant-unregistered-cycle-final.log,
/tmp/adamic-nested-mutant-no-env-retain.log, /tmp/adamic-nested-mutant-no-env-site.log,
/tmp/adamic-nested-mutant-disjoint-slots.log, and
/tmp/adamic-nested-census-mutant.log. The first call-only cycle probe did not
produce a leak report; a surviving conservative root is a possible explanation,
not an observation. The final global-owner probe clears that owner at exit and
produces the leak above. This records the unsuccessful probe rather than
pretending every source shape exposes a cycle to LeakSanitizer.

## Census measurement

Both runs use the pinned stage3/census/inventory.cjs at 429c117, TypeScript
v6.0.3 sources at 050880ce59e30b356b686bd3144efe24f875ebc8, and TypeScript 6.0.3
as the inventory parser. Before and after each report 77 files and 18,867 total
sites. The exact old nested-function reason remains 5,574 sites on 5,574 lines
in 57 files, first binder.ts:567: that tool hardcodes syntax counting and cannot
measure a lowering improvement. Actual minimal/r01 now lowers and runs.

Run the supplementary audit with:

```sh
CENSUS_TYPESCRIPT=/path/to/typescript node docs/nested-functions-census.cjs \
  /path/to/TypeScript /path/to/census/before/sites.json
```

It checks all declaration coordinates against the independent inventory, then
classifies declaration shape and references by upstream checker symbol identity.
It excludes type-only references and checks shorthand property values. It does
not certify representations, cycles, upstream checking, or whole-file compilation.
Seven of the 57 files have no remaining restrictions measured by this audit:
factory/baseNodeFactory.ts, factory/utilities.ts, performance.ts, tracing.ts,
transformer.ts, transformers/destructuring.ts, and transformers/utilities.ts.
The audit finds 643 blocked declaration sites and 4,563 unsupported reference
sites. These counts overlap by file and are not additive to the old inventory.
The other 50 retain unsupported declarations or cross-group/value references.
No claim that seven full tsc files compile is supported by this experiment.

## Toolchain and validation

cloud/setup.sh succeeded: go ready 0s, clang ready 1s, node ready 1s, submodules
ready 1s, cache warm 98s, done 98s. nproc is 5; cgroup CPU quota is 4 cores.
Tool versions are Go 1.27.1, clang 20.1.8, Node 24.19.0. Shell commands source
/workspace/adamic-tools/env.sh. Test output is saved to logs, never piped.

The initial full uncached go test -count=1 -timeout 30m ./... gate exposed the
negative-fixture placement issue in flow corpus scans. Moving that source into
refusals fixes the layout without suppressing a test or changing flow analysis.
The initial full gate was stopped after 21 minutes because unrelated stage1
suites were still running after that known failure. The final scope is all
changed packages and the complete oracle package, uncached, plus vet and format.
No full post-change stage1 gate or native tsc compilation is claimed.


| Final command | Result and log |
| --- | --- |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/flow ./internal/fresh ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -count=1 -timeout 20m | flow 124.817s, fresh 55.249s, lower 36.812s, native 186.688s pass; ir/javascript have no standalone tests. Oracle behavior checks pass, but its count-table check fails because the running binary had 13 fixtures while the table gained fixture 14. /tmp/adamic-nested-environment-gate.log |
| go test ./internal/lower ./internal/oracle -count=1 -timeout 15m | Final frozen 14-fixture registry: lower 19.744s and complete oracle 83.965s pass. Ordinary oracle cache enabled. /tmp/adamic-nested-final-complete-oracle.log |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNested\|TestNativeAgreesWithNode/internal/oracle/testdata/nested_' -count=1 -timeout 10m | All 14 fixtures plus refusal and sibling-cycle mutant tests pass, 3.610s. /tmp/adamic-nested-final-fourteen.log |
| go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts | Pass, 32.953s; fourteen added rows, zero existing rows changed. /tmp/adamic-nested-environment-counts-final.log |
| go vet ./... | Exit 0, no diagnostics. /tmp/adamic-nested-final-vet.log |
| gofmt -l cmd internal | Exit 0, no files. /tmp/adamic-nested-final-format.log |

The complete post-change stage1 gate remains uncovered. No escape proof, stack or
region placement, count elision, whole tsc native compilation, or performance
claim is included. The census's syntax counter is unchanged; seven-file structural
clearance is the limited supplementary measurement, not successful compiler runs.
