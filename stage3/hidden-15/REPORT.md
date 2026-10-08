Built: the requested oldSourceFiles closure witness, a capture mutant and a pinned structural/statics stop; no compiler fix.
Commits: base compiler/area-next-fixtures 4885cec50290686df487b62aac47c85d871ed40c; source census pin 388096e6; delivery SHA is supplied with the push report.
Commands and outputs: focused oracle PASS, counts refresh PASS on retry, exact replay reproduced; old/new region hidden intersections 6,260/6,260 bytes, difference 0.
Mutants: substituting an empty array for the captured oldSourceFiles read fails Node stdout comparison in both backends, with clean execution and no leaks.
Not covered: closing the hidden boundary, structural/statics dispatch implementation, whole-corpus recount, whole-package tests or the full gate.

## Roadmap step 30 result and dependency

This is a blocked implementation unit, not a claimed hidden-site fix. It preserves
semantic regression coverage and identifies the prerequisite for burning down
hidden-15, measured with step 05's hidden census. **Revealed bytes: 0.**
No production lowering or emitter file was edited. No worker branch was merged.

Assumption: a binding removed by census rollback must stay absent; it is not sound
to fabricate an empty array merely to make an isolated read lower. The supplied
closure witness already works on the requested area base. Its success does not
reproduce or fix this census stop.

The exact replay on area base 4885cec5 exits 0 and reproduces:

```text
/tmp/hidden-adapted/src/compiler/program.ts:2390:37: stage 0 can't lower reading oldSourceFiles yet
```

The selected attempted owner is program.ts:2345:5,
tryReuseStructureFromOldProgram. The earlier const declaration is
oldSourceFiles = oldProgram.getSourceFiles(), not an ancestor closure capture.
Its initializer fails first at program.ts:2383:32:

```text
a method call through a structural signature in a program with statics; use typeof the declaring class
```

The declaration's failure boundary is [116980,117043). The full-mode census
restores the pre-statement state, including locals, so the later loop over
oldSourceFiles has no binding and records [117206,123466). These observations
are present in both the exact replay and the program.ts census record.
Production compilation likewise cannot pass the preceding unsupported call.
Changing enumNeverValue to manufacture a value would falsify the observation
and leave the prerequisite unresolved.

**Dependency: hidden brief 03, structural signatures in programs with statics.**
The raising guard is callOrMethod in internal/lower/class.go, which checks a
MethodSignature while staticGlobals is nonempty. The brief lists the unlanded
codex/notyet-statics implementation as an existing candidate, but this unit
neither imports it nor claims it verified. Its owner must land a sound dispatch
fix into the area base. Removing that predecessor may expose further stops;
no downstream byte gain is inferred from its reason alone. The unit's named
expression.go territory is not an authorization to bypass that dispatch guard.

## Fixtures, oracle and mutant

The new sound fixture is
internal/oracle/testdata/hidden_boundary_old_source_files.a, the supplied
nested-call witness. Source Node prints 7 followed by a newline, exits 0 and
has empty stderr. Its focused registered oracle passes release native,
ASan/UBSan native, Linux leak checks and the JavaScript backend against Node.
The dedicated capture test independently pins the source observation.

The separate negative fixture is
internal/oracle/refusals/hidden_boundary_old_source_files_statics.a. It retains
an interface method returning an array and an unrelated static class field.
Source Node prints 1:7, exits 0 with empty stderr; lowering remains NotYet with
the exact structural/statics reason. It demonstrates the actual prerequisite,
not an unsafe JS program or the full census source.

The independently invoked opt-in mutant replaces the one captured IR read of
oldSourceFiles with ir.ArrayLiteral{Element: ir.Number}. It changes the nested
function's supplied array, leaving the source Node observation intact. Both
compiled backends exit 0 with empty stderr and print 0 instead of 7. The native
mutant builds, runs under sanitizers and passes the leak check; clang warnings,
a compilation refusal or sanitizer errors do not catch it. The Node stdout
comparison fails the test in both backends, exit 1. No production source mutation
needed restoration; the environment switch applies only to that invocation.

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_old_source_files|TestHiddenBoundaryOldSourceFiles' -count=1 -timeout 30m -v > /tmp/hidden-15-focused-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT_HIDDEN_15_CAPTURE=1 go test ./internal/oracle -run '^TestHiddenBoundaryOldSourceFilesCapture$' -count=1 -timeout 30m -v > /tmp/hidden-15-capture-mutant.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/hidden-15-counts-retry.log 2>&1
```

Focused oracle: PASS, 13.298s. Capture mutant: FAIL on native and JavaScript stdout,
0.428s. Counts retry: PASS, 52.753s. Every output goes to a log; no test run was
piped. No whole-package test or full gate ran. The preceding user-required setup
is a build, not an oracle gate.

## Counts changes

The first exact counts-refresh command failed because stage3/api/node_modules
lacked the required @types/node 25.3.3 for host fixtures. Installed the repository's
locked API dependencies using npm ci --prefix stage3/api --ignore-scripts
--no-audit --no-fund, then reran the same selector successfully. No dependency
manifest or lockfile changed.

Every changed counts row relative to the area base:

| Fixture | Change | Explanation |
|---|---|---|
| hidden_boundary_old_source_files.a | Added: allocations 5, frees 5, retains 1, releases 4, peak 5, regions 0 | New closure witness, counted once |
| logical_and_reference_maybe.a | Moved; unchanged 8 / 8 / 11 / 22 / 4 / 0 | Generated fixture-registration order, no numerical change |
| stage3/fixtures/taste/17_binder_flow.a | Removed stale 51 / 51 / 0 / 51 / 7 / 0 row | The area base already registers it lowers=false in taste_stage3_test.go; the refresh excludes its unsupported optional-field write. This unit did not change that fixture or registration |

No other row changed. The negative fixture is tested directly as NotYet and is
not counted as a natively executable program.

## Pinned region measurement

Adapted source was prepared using apply.sh from detached census-pin worktree
388096e6. All 82 file hashes match that pin's RESULT.json. program.ts SHA-256 is
7aaf5804093401db924e8278b142db40fcb2b14cef0566081cfaf870477b5284. No source reduction,
ancestor binding deletion or adaptation edit was used. Hash comparisons and
raw region arithmetic are in evidence/source-hashes.json and region-result.json.

The existing guarded full census overlay is built from area base 4885cec5.
For the region recount, program_scope.py adds only a scratch file-selection
condition to the census's outer file loop. All compiler-directory roots are still
loaded and registered, and every program.ts candidate is processed by the same
full-mode attempt, continuation and rollback machinery. Production Load/Lower
output guards stay enabled. The run records 247 units: 237 independent attempts,
10 checker-skipped bodies and 529 Boundary events.

A whole-corpus census was initially started, then stopped with exit 143 while
processing checker.ts. That partial run is not used for the region result. The
replacement program.ts census exits 0; this is a complete target-file recount,
not a whole-corpus count. Cross-file skipped dependencies cannot reveal bytes.
The independent stock catalogue lists only two declaration units overlapping
this interval: createProgram and tryReuseStructureFromOldProgram. The former
is split_checker_body; the latter is attempted and retains its own boundary
covering the entire original interval. No independently attempted nested unit
can expose a subrange. measure_region.py asserts these facts in addition to
calling the pinned hidden.py union/subtraction implementation.

| Original half-open UTF-8 interval | Old hidden intersection | New hidden intersection | Difference |
|---|---:|---:|---:|
| program.ts [117206,123466) | 6,260 | 6,260 | 0 bytes revealed |

The region's first boundary remains reading oldSourceFiles at 2390:37; it has
not advanced. The next recorded boundary after the interval in this attempt is
program.ts:2502:13, a Map whose keys aren't strings, numbers, booleans, objects,
arrays, maps or functions, boundary [123862,123955). It is outside the selected
interval and contributes no revealed-byte claim. A future structural/statics fix
must replay and recount again to identify the first newly reached stop inside
the loop. The assigned 6,260 bytes are not a whole-corpus total.

Reproduce the region recount after preparing the pinned source:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-15-recount-overlay > /tmp/hidden-15-recount-overlay.log 2>&1
python3 stage3/hidden-15/program_scope.py /tmp/hidden-15-recount-overlay /tmp/hidden-adapted/src/compiler > /tmp/hidden-15-recount-scope.log 2>&1
go build -buildvcs=false -overlay=/tmp/hidden-15-recount-overlay/program-overlay.json -o /tmp/hidden-15-recount ./stage3/census/latent/tool > /tmp/hidden-15-recount-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/hidden-15-recount /tmp/hidden-adapted/src/compiler /tmp/hidden-15-recount.jsonl > /tmp/hidden-15-recount.log 2>&1
python3 stage3/hidden-15/measure_region.py /tmp/hidden-15-source-pin /tmp/hidden-adapted/src/compiler /tmp/hidden-15-recount.jsonl /tmp/hidden-15-recount-region.json > /tmp/hidden-15-recount-region.log 2>&1
```

For the exact initial replay, the already built area worker used:

```sh
/tmp/hidden-15-area-replay -project /tmp/hidden-adapted/src/compiler -where /tmp/hidden-adapted/src/compiler/program.ts:2390:37 -kind NotYet -reason 'reading oldSourceFiles' > /tmp/hidden-15-area-replay.json 2> /tmp/hidden-15-area-replay.log
```

The worker is built with make_overlay.py and the normal replay/worker target.
It shares the inherited replay command's guarded attempt path. No replay or
census compiler source was committed.

## Setup and delivery scope

The first setup ran on the previous checkout after the default fetch failed to
include the required area ref. Explicitly fetched the area ref, created the
requested branch, then reran setup to align its checker submodules. Both complete
setup logs are retained. On the actual area base, the cumulative timing lines were:
Node ready 0.021s, Go ready 0.022s, clang ready 0.159s, markdown dependencies
installed step-duration 1.782s and ready 1.910s, submodules ready 5.011s,
Go build ready 211.269s, test binaries deferred 211.653s, build cache warm
211.661s, done 211.793s. nproc 5; cpu.max 400000 100000; Go 1.27.1,
Node 24.19.0, clang 20.1.8. Sourced /workspace/adamic-tools/env.sh for every
build and test shell. The GOPROXY fallback was set before setup; no module 403
occurred in this unit.

Delivery consists only of this unit's fixture, tests, counts refresh and evidence
on the requested area tip. No origin/main or other worker changes were merged.
No cohere code was copied. git diff --check and named-file gofmt checks pass.
The delivery branch is codex/hidden-15-old-source-files; no PR is opened.
