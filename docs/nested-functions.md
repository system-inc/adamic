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
function declarations, first-class generic nested values, dynamic this, first-class sibling references inside another sibling, and
calls to a declaration in a different ancestor group remain loud NotYet cases.
The initial design retained optional/default/rest closure ABI gaps. The October 7
omission fix below closes optional/default parameters; rest remains NotYet. Plain
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
* Storing a generic inner<T> declaration as a value. Direct generic calls are supported.
* inner(...values: number[]). Optional/default parameters are covered by the omission follow-up.
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


## Real TypeScript fixtures follow-up

The fixture branch 7bf6f57 was merged with two parents in c08c41c before these
fixes. The eleven original sources and their historical status records remain
unchanged. Counts against Node progressed as follows:

| Fix | Original fixtures matching Node |
| --- | --- |
| Baseline b15216d | 3/11 |
| Own captured frame parameters, including read-only initialization | 6/11 |
| Return a local assignment with one RHS evaluation and write-through | 6/11; parser then exposes the conservative cycle refusal |
| Represent undefined-only values | 6/11; cleanup then exposes generic nested calls |
| Instantiate generic nested direct calls in their lexical frame | 7/11 |
| Prove closed literal inputs cannot reach the frame | 8/11 |

Passing originals are 02, 03, 04, 05, 06, 07, 10 and 11. Oracle registration
holds each original to Node, the JavaScript IR backend, sanitized native C,
release native C, and LeakSanitizer. Three additional executable regressions
cover generated-string parameter ownership and escape, generic captures across
frame invocations and number/string/undefined instantiations, and assignment
results with a side-effecting RHS. Ownership applies to any EnvironmentCell
parameter, while existing anonymous closure borrowing remains unchanged.

Generic direct calls instantiate with the lexical owner's locals, capture stack,
and outer substitutions. The instantiation cache includes the owner index.
Generic declarations have no first-class value representation; such references
remain NotYet. Undefined-only results use the existing null reference ABI.
Returned assignments save the RHS before writing the target and return that
saved value. Property, element and destructuring assignment returns remain loud
NotYet cases.

The parser's scanner parameter has a cycle-compatible callback type, but every
actual caller supplies an object literal containing only scalars or capture-free
closures. The closed-input exception checks every IR body and main, rejects
parameter reassignment, owner closure escape, virtual/opaque calls, nonliteral
arguments, and any expression or statement outside its explicit whitelist.
Object/element stores are outside that whitelist. This proves that this input
graph cannot acquire a reference back into the frame. It does not relax the
cycle rule for arbitrary fresh objects, mutable inputs, or escaping callbacks.
An IR field-store mutation must invalidate the proof. Weak remains the general
way to break a captured closure cycle.

The three unchanged originals retain typed Refused diagnostics:

* 01 scanner frame: first a non-null assertion. It also uses var, truthiness,
  comma expressions and generic helpers returned
  as values. Supporting all 24 helpers requires resolving those separate limits.
* 08 symbol recursion: object truthiness as a condition. An explicit undefined
  comparison is required by Adamic's current language policy.
* 09 constituent recursion: first a non-null assertion, also truthiness and a
  sibling reference through a reducer arrow. Explicit null checks/conditions
  plus anonymous callback forwarding of a sibling's environment are needed.

No permanent language policy is changed to accept these originals. There is no
claim of 11/11 or of compiling the complete scanner. AllocateEnvironment, its
single frame layout, and the native allocation implementation are unchanged.
Runtime's environment-placement work can still see one site and make one
placement decision. No escape proof or placement optimization is added here.


### Language decisions paused on original sources

The exact programs are the checked-in, unchanged
[01_scanner_frame.a](../stage3/fixtures/nested-functions/01_scanner_frame.a),
[08_checker_symbol_recursion.a](../stage3/fixtures/nested-functions/08_checker_symbol_recursion.a),
and [09_checker_constituent_recursion.a](../stage3/fixtures/nested-functions/09_checker_constituent_recursion.a).
The first diagnostic-producing expressions, verbatim, are:

```text
01: return s.codePointAt(i)!;
08: return symbol.parent ? `${getSymbolPath(symbol.parent)}.${symbol.escapedName}` : symbol.escapedName as string;
09: type.flags & TypeFlags.Union && (type as UnionType).origin ? getConstituentCount((type as UnionType).origin!) :
```

The choices are to keep Adamic's prohibition on non-null assertions and require
proven null checks (01 and 09), or change that language policy; and to require
boolean conditions (08 and 09), or specify/support TypeScript truthiness.
01 additionally needs a decision on var and definite-assignment assertions.
This unit stops on those decisions, as requested, without altering the originals
or counting normalized substitutes as successes. Signature and callback-forwarding
work can be designed separately, but cannot make these exact sources acceptable
under the existing language policy.

### Follow-up proof results

All logs below are saved files. Compiler overlays leave production sources intact.

| Mutant | Catch |
| --- | --- |
| Mark an EnvironmentCell parameter borrowed | TestNestedCapturedParametersAreOwned rejects the borrowed slot |
| Omit the write in a returned assignment | Node comparison of nested_assignment_return.a differs |
| Evaluate assignment RHS twice | Node comparison of the same fixture differs |
| Allow a field store in the closed-input proof | TestClosedFrameInputRejectsMutation fails on the IR store probe |
| Lose declared storage for an undefined-narrowed field | maybe_number_slots.a exposes an incompatible native call ABI |
| Treat an undefined-narrowed Weak read as present | weak_parent.a panics where Node continues |

Logs are /tmp/adamic-real-mutant-{borrowed-parameter,dropped-return-write,
double-return-evaluation,closed-input-mutation,undefined-field-storage,
undefined-weak-presence}.log. The first field-store mutant attempt failed to
compile because it was also inserted in the expression switch. That invalid
attempt is not proof; the corrected, statement-only overlay compiles and fails
on the intended mutable-graph assertion.

The earlier broad Node run caught the field storage and Weak regressions; both
were repaired and their focused rerun passes, including all 17 synthetic nested
fixtures (2.403s, /tmp/adamic-real-regression-repair.log). Source audit reports
41 unchanged upstream helper bodies (/tmp/adamic-real-source-audit.log).
The count update adds eleven rows and changes no existing rows (12.361s,
/tmp/adamic-real-counts-final.log). The final complete package gate and vet/format
logs are /tmp/adamic-real-final-{gate,vet,format}.log. No complete stage1 gate,
whole tsc compilation, placement proof, or 11/11 claim is made.


Final follow-up gate: ADAMIC_GATE_UNCACHED=1 go test ./internal/lower
./internal/native ./internal/flow ./internal/fresh ./internal/oracle -count=1
-timeout 20m passed: lower 32.423s, native 175.564s, flow 116.661s,
fresh 54.390s, and the complete oracle 174.070s. Vet of lower/oracle and
format checks of both directories passed with no diagnostics. The final focused
original-fixture run passed in 3.114s (/tmp/adamic-real-eight-final.log).
Published code is 990eb119f44cfb1db1a4f6f2814cb7f9e91b2b29, with merge
c08c41c0e023104a228e5795c24c59291bdd9297. The branch reports 8/11;
the three language decisions above remain paused.


## Feature branch prepared for integration, October 7

Only codex/nested-functions is pushed. No push or merge into main or an area/
branch was performed. Integration owns those branches and will merge this
feature branch itself. The earlier landing instruction was clarified before
any such write took place.

Current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965 is merged,
without rebase or force, in ee3f647f0f971e19b3b8371e10f2b6613ae4d622.
The earlier main e011f8f was first merged in f7ab977. Its gate passed, then
a fresh fetch found the call-target and devirtualization integration on main.
The second merge therefore received its own complete affected-package gate.
Both merges were clean, with no manual compiler source edits. AllocateEnvironment
and the eight-original-fixture behavior remain unchanged by this preparation.

Setup succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 82s, done 82s. nproc is 5, cgroup quota 4 CPUs. Every build/test
shell sources /workspace/adamic-tools/env.sh.

### Counts after both main merges

The whole table was regenerated with go test ./internal/oracle -run
'^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts.
The first regeneration passed in 29.349s; the final one passed in 12.721s.
Neither changed the merged table. Compared with current origin/main, this branch
adds exactly 25 rows, removes none, and changes no existing row:

* Seventeen synthetic nested-function fixtures measure the new frame/closure
  programs, including generated-string ownership, generic captures and returned
  assignments. Three TDZ programs terminate exceptionally, so their rows are
  not assertions that allocations equal frees on normal completion.
* Eight original TypeScript fixtures measure the unchanged upstream bodies that
  now compile and match Node. These are new observations, not performance deltas
  against an earlier native implementation of those fixtures.

Main's existing rows are preserved, including its class_as_interface.a change
from 360/525 retains/releases to 349/514 and its borrowed-element lifetime rows.
Those are inherited main measurements, not improvements attributed to nested
functions. The six new call-target/devirtualization rows from main are also
preserved. There is no additional count movement from integrating these units.

### Final uncached gate

ADAMIC_GATE_UNCACHED=1 go test ./internal/ir ./internal/flow ./internal/fresh
./internal/lower ./internal/native ./internal/javascript ./internal/oracle
-count=1 -timeout 30m passed. Results: ir 13.560s, flow 110.036s, fresh 63.580s,
lower 40.225s, native 180.614s, complete oracle 174.385s. JavaScript has no
standalone tests; its emitted output is checked by the complete oracle.
This covers every code package changed by this feature branch against main,
including the refusal and sibling-cycle mutant tests. The original fixture count
remains 8/11; the three language-policy decisions above remain paused.

go vet ./... and gofmt -l cmd internal passed with no diagnostics. Logs are
/tmp/adamic-nested-land-targets-{counts,gate,vet,format}.log. The first gate is
preserved in /tmp/adamic-nested-land-{counts,gate,vet,format}.log. No complete
repository stage1 gate was run for this integration preparation. The earlier
individual compiler overlay mutants were not rerun; the permanent oracle mutant
and refusal checks ran in both complete oracle gates.


## Reduced borrowed-parameter witnesses

Three exact reduced function bodies supplied from adamic-reduce, based on full
fixtures 02, 03 and 04 at b15216d, are now executable oracle fixtures beside the
existing synthetic programs:

| Fixture | Driver output |
| --- | --- |
| nested_reduced_parameter_unused.a | scanner created followed by a newline |
| nested_reduced_parameter_unreachable.a | read: followed by a newline |
| nested_reduced_parameter_escaped.a | scanner 4 followed by a newline |

Each driver builds its text at runtime with concatenation and String(number).
The first two preserve their unused/unreachable nested declarations exactly.
The third saves createScanner's returned function and invokes it in the next
statement, after both the enclosing frame and argument statement have ended.
No compiler implementation was changed for these additions.

The focused uncached Node/JavaScript/native/release/sanitizer/leak oracle passed
in 2.156s (/tmp/adamic-nested-witness-oracle.log). Four compiler overlay mutants
were run independently, with production files unchanged:

* Remove the EnvironmentCell exclusion from borrowed parameters. Each of the
  three witnesses, run separately, fails with native: a store into the borrowed
  parameter text. Logs are /tmp/adamic-nested-witness-mutant-borrow-{unused,
  unreachable,escaped}.log.
* Omit the retain when storing a reference into an environment cell. The escaped
  witness compiles, then ASan reports heap-use-after-free in adamic_retain called
  by getText, after the text argument was freed at the end of the createScanner
  statement. This is a runtime failure, not a clang warning. The log is
  /tmp/adamic-nested-witness-mutant-no-cell-retain.log.

The full count regeneration passed in 15.433s and adds only the following rows.
All existing rows are unchanged. Columns are allocations, frees, retains,
releases, peak live and regions:

| Reduced witness | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| unused | 4 | 4 | 3 | 6 | 4 | 0 |
| unreachable | 5 | 5 | 4 | 8 | 4 | 0 |
| escaped | 4 | 4 | 6 | 10 | 4 | 0 |

These new rows measure new driver programs, including runtime-built strings and
the counted frame/closure; they are not changes to any full fixture's counts.
The original TypeScript fixture count remains 8/11. Main is still e8ba3d5, and
there are no new language decisions on main that clear the three paused originals.
The narrow parser proof is retained. No graph-regions work is built before that
unit reaches main, and none of the remaining originals is blocked only by cycles.

Setup: all tool readiness lines 0s, cache warm and total 77s, nproc 5 with a
4-CPU quota. Test shells source /workspace/adamic-tools/env.sh. Complete oracle,
vet and format logs are /tmp/adamic-nested-witness-{complete-oracle,vet,format}.log.
No complete repository stage1 gate or other new compiler feature is claimed.

The complete uncached oracle passed in 76.919s, including all twenty synthetic
nested fixtures, the eight supported originals, the three typed refusal probes
and the permanent cycle mutant. Vet and formatting passed without diagnostics.


## Omitted optional arguments, October 7

This branch alone reproduces the scanner omission bug, without any of the other
compiler tips. The unchanged probe at scanner-proof faae0e9 is Refused first at
start!, and also uses truthiness. Replacing only (newText || '') with
(newText ?? '') and start! with (start ?? 0) gives a policy-compatible witness:
Node prints 11, release native prints 0 with exit 0, and ASan reports a
stack-buffer-overflow. The interface call supplied one argument slot; the nested
implementation read three. The two source normalizations are equivalent for
this driver. The normalized witness is not claimed as compilation of the exact
original. The exact original needs the non-null and taste policy work as well.

Every closure and interface-method call now carries its supplied argument count.
The callee guards reads of optional scalar/reference slots, producing absence
rather than accessing beyond the argument array. This includes actual optional
parameters hidden by a narrower static function signature. Direct typed calls
pad with the same explicit absent helper. CallClosure lowering fits each argument
to the resolved signature, pads omitted signature parameters with typed undefined,
and treats a defaulted parameter as optional at entry. Defaults are still
evaluated inside the callee. Sibling direct calls retain their CallClosure
effects and use the same count-aware ABI. AllocateEnvironment is unchanged.

Numbers retain their reserved packed undefined NaN. References use NULL,
including strings, objects, arrays, functions, maps and Weak handles. Optional
booleans use a three-state uint8_t slot: 0 false, 1 true, 2 absent. The pair
representation remains unchanged outside slots. General union function arguments
remain NotYet; optional boolean collection fields/captured cells retain their
existing conservative guards. This does not open unsupported collection layouts.

The call audit covers direct named calls, interface object functions, class
interface thunks, dynamic interface dispatch, function values, extracted object
method values, named siblings, escaping nested declarations, arrows, accessor
closure adapters, Array.from/map/visits/reduce/sort and Map/Set callbacks.
Runtime callbacks pass their actual supplied count, so extra optional/default
parameters also work. Module functions with optional/default parameters can now
be read as values and used in shorthand objects. Unbound class methods requiring
dynamic this retain their existing restrictions.

Nine new oracle fixtures cover the scanner witness and required/explicit-undefined
controls; optional number/string/boolean/object parameters through direct,
shorthand field, extracted field, function value, arrow and escaped nested calls;
narrower signatures; defaults; class/interface dispatch; and callback defaults.
Present zero, false, empty string and object controls distinguish absence from
valid present inputs. The boolean fixture also checks optional boolean results.

The permanent zero-padding IR mutant changes the scanner's two absent numeric
slots into present zeros. It compiles, exits 0, has empty stderr and passes the
leak check; only Node output comparison catches native 0 against Node 11.
The proof log is /tmp/adamic-omitted-mutant.log. Initial mutant harness runs used
a relative path incompatible with the observation cache; those harness failures
are not counted as proof. The corrected run uses an absolute path.

Setup succeeded: go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s,
cache warm 27s, total 27s, nproc 5 with a four-core quota. Tests source
/workspace/adamic-tools/env.sh and write output to logs. The original real-source
fixture count remains 8/11; 01, 08 and 09 still pause on the language decisions
listed above. No graph-regions changes, escape proof, main/area write, complete
repository stage1 gate or full native tsc compilation is claimed.


### Census follow-up after optional/default support

The independent pinned inventory still reports 5,574 nested declaration sites
on 5,574 lines in 57 files. It is a syntax inventory, so that number is unchanged.
The supplementary structural audit now removes optional/default parameters from
its restriction list. Blocked declaration sites fall from 643 to 184; blocked
reference sites remain 4,563. Ten of the 57 files clear these structural
restrictions, up from seven. The additional three are factory/emitHelpers.ts,
transformers/declarations/diagnostics.ts and utilitiesPublic.ts. This is not
evidence that any of these complete upstream files passes checking, representation
or cycle analysis. Generic declarations remain conservatively classified in
this structural audit; direct generic calls have narrower implementation support.

The current audit is /tmp/adamic-omitted-census.json, checked against the pinned
independent sites.json. The binder-coordinate deletion mutant was rerun against
the updated audit and still fails with nested census site mismatch.

### Exact probe and integration preparation

The unchanged probe is preserved at
internal/oracle/refusals/omitted_scanner_original.ts. Its dedicated oracle test
holds Node to stdout 11, exit 0 and empty stderr, and holds lowering to the
existing typed non-null assertion refusal. Both backends are checked on the
normalized executable witness, not misreported as compiling the unchanged one.

The urgent fix was pushed as bdab8c71cfa6eab3fbbf09ee323ff13a235ca216.
Origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 is merged into this feature
branch in 9fbfea162dd1525d99f3dedf08d2e430e03d230d. No push or merge into main
or any area/ branch occurred. The only merge conflict was in counts.md; both
sets of fixtures were retained and the table was regenerated. Regeneration
passed in 22.353s. Compared with that main tip there are 37 added rows, zero
removed rows and zero changed existing rows. The nine omission fixtures account
for nine additions; twenty synthetic nested fixtures and eight real-source
fixtures account for the remaining twenty-eight. Row order follows the registry.

Before merging main, the frozen urgent-fixture uncached oracle passed in 4.094s.
The initial broad gate started before the former optional-boolean NotYet
expectations and the mutant path were corrected. Its native, IR, flow and fresh
packages passed; its lower package failed the obsolete negative expectations,
and its oracle failed only the relative-path mutant harness. Those initial
failures are recorded in /tmp/adamic-omitted-gate.log and are not a green gate.
The corrected lower run passed in 25.770s. The corrected uncached exact-probe
and zero-padding mutant tests passed in 0.966s after the merge.


### Final merged validation and additional mutants

The merged uncached gate passed:

ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/ir
./internal/flow ./internal/fresh ./internal/javascript ./internal/oracle
-count=1 -timeout 30m

Results: lower 41.674s, native 197.799s, IR 12.547s, flow 128.223s,
fresh 57.167s, complete oracle 185.637s. JavaScript has no standalone tests;
the oracle holds its emitted programs to Node. The log is
/tmp/adamic-omitted-merged-gate.log. go vet ./... and gofmt -l cmd internal
passed without diagnostics, in /tmp/adamic-omitted-merged-{vet,format}.log.
The entire repository stage1 gate was not run.

Three further independent Go overlays mutate only absent(valueType), leaving
production sources intact: absent numbers become present zero, absent booleans
become present false, and absent strings become the empty string. Each corresponding
omitted_{number,boolean,string}.a fixture compiles and exits 0 with empty stderr,
no sanitizer failure and no leak. Node output comparison alone rejects it,
including direct calls and function values called through narrower signatures.
Logs are /tmp/adamic-omitted-fallback-{number,boolean,string}.log; each test
command exits 1 on stdout differs. The permanent scanner zero-padding mutant,
these three fallback mutants, and the updated census coordinate mutant are the
five mutation proofs run for this follow-up. The earlier named-function cycle
mutant and existing refusal checks also ran in the full merged oracle gate.

The original real-source fixture count is still 8/11. This follow-up closes
optional/default signatures in 01 but does not change its earlier policy refusal.
01, 08 and 09 stay paused on the exact programs and choices already recorded.
The shared AllocateEnvironment construct, cycle proof and heap placement are
unchanged. Nothing was built against unmerged graph-regions.

## Reader omission and callback forwarding follow-up

The exact Reader interface witness, its direct class call, its subclass override
through the base type, and its string variant all print `true undefined` in Node,
the native release build, the sanitized native build, and the JavaScript backend.
They pass the leak check. The existing omission fix covers all four; no new
argument representation or padding path was introduced. Their fixtures are
`omitted_reader{,_direct,_override,_string}.a`. A permanent IR mutant replaces
the exact Reader call's absent number with present zero. It exits successfully
and without a leak; Node comparison rejects `false 0` instead of `true undefined`.

An anonymous callback can directly call a named helper in its enclosing named
group. Its IR function records ForwardedNestedParent. Completing that group's
frame appends the whole shared environment layout to each forwarding callback,
including captures discovered in later sibling bodies. Intermediate anonymous
closures carry the layout too. The original AllocateEnvironment remains the one
allocation site for the enclosing frame's captured locals. No sibling function
binding becomes a captured cell.

A callback can also capture locals belonging to its immediate helper. Consequently
its layout need not equal the target helper's layout. Its direct call constructs
a temporary counted target-layout closure carrier, invokes the target code,
and releases the carrier. Both emitters use the carrier expression in CallClosure,
rather than assuming ClosureSelf has the target layout. This costs one closure
allocation per such direct call. It does not allocate another environment record.
Region analysis still sees the original AllocateEnvironment site and ordinary
MakeClosure captures; no escape proof, region rule, borrow rule, or reuse rule is
disabled. Calling across another named group and first-class sibling values remain
refused. Only direct calls through anonymous closures in the same group are added.

The cycle finder sees the full forwarded layout. A saved callback that calls a
helper reading that same saved callback is still cycle-capable and refused. Four
fixtures cover recursion with callback-local captures, a callback escaping both
frames, two levels of anonymous callbacks, and a capture discovered in a later
sibling body. All match Node in both backends with sanitizers and leak checks.
A permanent wrong-carrier mutant replaces MakeClosure with ClosureSelf in the
escaped fixture. It exits successfully without a leak; Node output alone catches
it. The count table adds four callback rows (29, 16, 9, and 9 allocations) and
four Reader rows. No pre-existing row changes.

Nested declaration rebinding keeps the exact NotYet message
`rebinding a nested function declaration`, including generic declarations and
array destructuring targets. The public `refusals/nested_rebinding.a` witness
stops earlier at TypeScript TS2630. A lower package test suppresses that diagnostic
in a read-only overlay and bypasses only the suppression-directive gate to pin
the lowering message. Removing the guard in a Go overlay makes that test fail
because lowering succeeds. The public compiler still refuses suppression directives.

An unpushed scratch merge used nested-functions 4d49d5b plus this follow-up,
non-null-check a02613eff851e07f567ab9934adfc8a2dc98eeca,
taste-not-soundness d2c05df34443d9c0ae4fe6639beef2c5aee8a4ad, and
interface-downcasts 21558749e9b18ed0416e465ece1f164c00821259.
Ten of eleven original real-source fixtures match Node in both backends, under
sanitizers and leak checks. Fixture 08 clears with truthiness; fixture 09 clears
with anonymous sibling-call forwarding. Fixture 01 stops at its exact line 72,
`var pos: number;`, with `Adamic 0.1 refuses var; use const or let`. That remaining
blocker belongs to non-null-check. Its asserted initializer `var text = textInitial!`
is accepted, but an ordinary uninitialized var is still refused. The feature branch
alone retains 8/11 because the three policy branches are not merged here.
Scratch merge resolutions retained the omission fix's byte representation for
optional boolean slots instead of mixing the two branches' encodings. Nothing
from graph-regions was built or merged.

The developer-tools generator was built from a detached read-only checkout of
8a04d2a3be4d1644a13c73ee8964280e5e42847e, with GOFLAGS=-buildvcs=false to avoid
unavailable VCS stamping. The command was `/tmp/adamic-reader-fuzz -root
/workspace/adamic -with interface-omitted-optional -seed 1 -count 200 -parallel 4
-work /tmp/adamic-reader-fuzz-run -v`. In 3m10s it reported 187 agreed, 13 checked,
0 findings, 0 NotYet, 0 invalid, and 0 unfit. Checked programs stop at matching
inserted checks in both compiled backends; they are not counted as agreeing with
Node. Replaying seed 113 showed an array-bounds panic, `index 1 is outside an array
of length 1`, in both backends. The scene is opt-in and does not force every seed
to contain an interface omission. The full log is /tmp/adamic-reader-fuzz-200.log.

This turn's setup log is /tmp/adamic-reader-setup.log: go 0s, clang 1s, node 1s,
submodules 1s, warm 80s, total 80s; nproc 5, quota 4. Reader oracle and mutant
proofs are in /tmp/adamic-reader-proof.log. Callback, rebinding, and combined
measurement logs are /tmp/adamic-callback-rebinding.log,
/tmp/adamic-callback-mutant.log, /tmp/adamic-callback-cycle.log,
/tmp/adamic-rebinding-final.log, /tmp/adamic-rebinding-mutant.log, and
/tmp/adamic-combined-after-callback-final.log. The first combined measurement was
9/11; the callback fix gains fixture 09 and reaches 10/11. No main or area branch
was pushed to or merged into.

### Current main merge and landing gate

The feature branch merged origin/main c7991b900362796aefd111474e65eb5398e91953
with merge commit 59f0d0aaf132969f5f9b55bac4b1e91ffd2a4ba1. The only conflict was
counts.md: both independent sets of added rows were retained. Regeneration then
moved main's five proven-predicate rows before the real nested fixtures to match
registration order. No measured values in an existing row changed. The pre-merge
whole oracle run overlapped this merge and failed solely on table ordering; it
is not the landing gate. The fresh post-merge gate is recorded separately in
/tmp/adamic-nested-merged-gate.log.

The post-merge focused uncached Node gate covering all nested and omitted-argument
fixtures and their oracle mutants passed in 44.623s. Count regeneration passed
in 70.944s. Vet and whitespace checks passed with empty logs at
/tmp/adamic-nested-merged-{vet,format}.log. The unpushed scratch merge was also
updated with this feature merge and remeasured: still ten matches and the same
fixture 01 var refusal, in /tmp/adamic-combined-current-main.log. Scratch conflict
resolutions removed the old type-predicate refusal because main now proves these,
retained the non-null branch's asserted initializers, and kept both sets of count
rows. No new policy was implemented in the feature branch.

The landing gate passed: `ADAMIC_GATE_UNCACHED=1 go test ./internal/lower
./internal/native ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh
./internal/oracle -count=1`. Output: lower 56.861s, native 275.718s, IR 3.462s,
JavaScript no standalone tests (covered by the whole oracle), flow 173.988s,
fresh 99.593s, whole oracle 247.308s. The full repository stage1 gate was not run.
The whole oracle includes the Reader zero-padding, callback wrong-carrier, and
existing sibling-cycle leak mutants. The independent rebinding-guard overlay
also failed as intended, with lowering succeeding when the refusal was removed.
