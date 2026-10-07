Built: direct C calls for singleton CallTargets, exact allocation receivers, and exact interface receivers; JavaScript keeps equivalent dispatch.
Commits: implementation 6965ac23be66de5dc6c75a797eb13ca408109d73; prerequisite f64641f388f32d877522140e174dd0b1dcb6d13e; base merge ceb29f7fe684d254a4b5031dd87a45deb2314ed0.
Checks: native 77.045s, IR 0.534s, complete uncached oracle 71.532s, counts 13.826s, vet and formatting clean; visitor median 0.143476s to 0.103067s, instructions 1,565,590,554 to 1,350,550,540.
Mutants: two-target selection and ignored reassignment caught by Node stdout; disabled singleton, exact receiver, and interface optimizations caught by C assertions; all five exited 1.
Not covered: whole repository gate, actual stage 1 lint timing, cross-function receiver provenance, exact facts from joins or narrowing, accessor dispatch optimization, other processors.

The branch is `codex/devirtualize`. It began at current `origin/main`,
`e011f8f60899586d6373a5ccb07335ad82cfbf3c`, then merged the requested
`origin/codex/route-call-targets` prerequisite, exactly `f64641f`, rather than a newer
revision. The histories had diverged, so the merge could not fast-forward.
No cohere source was copied. This unit does not edit `emit.go`, `lower.go`,
`native.go`, or `oracle_test.go`; the prerequisite merge carries its own changes.

`callCode` asks `Program.CallTargets` and names its sole function directly.
With several conservative targets, an allocation or an uncaptured local with
one allocation declaration and no assignments can still prove the exact class.
The selected slot then names a direct function. The native statement walker checks
assignments in nested bodies, including try/catch/finally. A base constructor's
`this` is not assumed to have the base's exact identity.

Interface calls use a different ABI. An exact class receiver's named method now
calls the existing method adapter directly, without `adamic_object_callee` or a
function-pointer call. Keeping the adapter preserves its borrowed inputs and its
retains for consumed parameters. A single implementing class alone is insufficient:
an interface may also hold an object literal's own closure. The fixture includes
both a single implementer through an exact interface binding and an interface
parameter that also receives an object literal; the latter keeps dynamic lookup.

`CallTargets`, ownership and region plans, and exception summaries are unchanged.
The existing call-site exception tests and result cleanup remain in place. The
JavaScript backend is unchanged: selecting the sole possible target directly or
through its table has identical results and argument evaluation order, and the
complete oracle compares both backends to source Node.

The counts table changed only one existing row, `class_as_interface.a`: retains
360 to 349, releases 525 to 514. These are the eleven removed NULL closure
lookup temporaries. Allocations 372, frees 372, peak 60, and regions 0 are unchanged.
The new fixture has allocations 43, frees 39, retains 26, releases 66, peak 10,
and regions 4. Its baseline had retains 27 and releases 67, with all other columns
identical. Both printed identical results. Region allocations explain allocations
exceeding individually counted frees; the leak check passed.

Generated C from `internal/oracle/testdata/devirtualize.a`, with exact names:

```c
/* A Base parameter, one stable implementation and an overridden second method. */
adamic_string * adamic_temporary_3 = adamic_function_7_Base_stable(adamic_local_4_base, adamic_temporary_2);
double adamic_temporary_4 = ((double (*)(adamic_object *))adamic_virtual(adamic_local_4_base, 1))(adamic_local_4_base);

/* The exact Base allocation calls the overridden slot's Base implementation. */
double adamic_temporary_99 = adamic_function_8_Base_changing(adamic_temporary_98);

/* Interface single: Single = new Only(), directly through its existing adapter. */
adamic_value adamic_temporary_112 = adamic_method_23(adamic_temporary_109, (adamic_value[]){{.reference = adamic_temporary_111}});

/* Separate generic instantiations, both direct. */
double adamic_temporary_14 = adamic_function_11_Box_number_get(adamic_local_6_box);
adamic_string * adamic_temporary_16 = adamic_function_14_Box_string_get(adamic_local_7_box);

/* A direct throwing method retains its exceptional cleanup edge. */
adamic_string * adamic_temporary_21 = adamic_function_17_Thrower_fail(adamic_local_8_thrower, adamic_temporary_20);
if (adamic_thrown != NULL) {
    adamic_release(adamic_temporary_21);
    adamic_release(adamic_temporary_20);
    adamic_release(adamic_temporary_19);
    goto adamic_landing_18;
}
```

The hierarchy fixture declares `Later` after the call site in `visit`. Node and
both backends print `built7:stable:1`, then `built7:stable:2`. The exact call prints
`exact:1`; a temporary Later allocation prints `temporary:2`. Runtime string
results, generics of number and string, a throwing method with finally, an object
literal through the interface, and a reassigned receiver are also held by Node.

Every mutant started from the final implementation and was restored in `finally`.
Tests wrote directly to separate log files, with the uncached oracle enabled.
Neither stdout mutant was caught by compiler warnings or sanitizers: both native
programs exited 0 and disagreed with Node.

| Mutant | Check and observation |
|---|---|
| Change `len(targets) == 1` to `len(targets) >= 1` | Oracle stdout: subclass `built7:stable:2` becomes `built7:stable:1` |
| Emit singleton calls virtually | `TestDevirtualizedCalls`: `single target is not direct` |
| Disable exact receiver direct emission | `TestDevirtualizedCalls`: `exact receiver still dispatches` |
| Disable exact interface direct adapter emission | `TestDevirtualizedCalls`: `interface call still looks up its method` |
| Ignore assignments in the exact receiver proof | Oracle stdout: final `mutable:2` becomes `mutable:1` |

The mutant runner and logs are `/tmp/devirtualize-mutants.py`,
`/tmp/devirtualize-mutants.log`, and `/tmp/devirtualize-mutant-<name>.log`.

The visitor benchmark is `bench/devirtualize.a`. It repeatedly visits 128
syntax-shaped records, each with two child records, for 20,000 rounds. A visitor
routes by kind to identifier/literal rules and recursively visits children. An
inherited visitor class has the same implementations. This models the dispatch
shape of lint, not stage 1's complete parser, checker, or rule corpus.

Both release compilers used the same lowering and runtime at base merge
`ceb29f7`. The baseline restored only the two changed emitter functions to that
commit while building a separate compiler; the final source was then restored.
The new exact-receiver helper is unused by baseline emission. Its generated C has
four virtual call sites; final C has zero. Source Node, generated JavaScript,
baseline release, final release and final sanitized C all print `759040000`.

On Linux 6.18.44, AMD EPYC 9V74, clang 20.1.8, Go 1.27.1, Node 24.19.0,
`nproc` is 5 and the CPU quota is 4. Nine paired runs alternated which binary ran
first. Timing uses Python `time.perf_counter` around subprocess execution,
including startup. The final one-minute load average was 2.52. Callgrind 3.24.0
counted executed instructions separately from native timing.

| Release measurement | Before | After | Reduction |
|---|---:|---:|---:|
| Median wall seconds | 0.143475908 | 0.103066738 | 28.2% |
| Best wall seconds | 0.138880161 | 0.102305716 | 26.3% |
| Instructions, Callgrind Ir | 1,565,590,554 | 1,350,550,540 | 13.7% |

Observed: fewer instructions and faster paired fixture runs. Inference: direct
calls also enable clang's inlining and simplification; this is not a measurement
of dispatch overhead alone or a claim about actual stage 1 lint performance.

Both program and runtime release builds use these flags, with no LTO or `-march`:

```text
-std=c11 -Wall -Wextra -Werror -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign
-ffp-contract=off -fno-optimize-sibling-calls -O2
```

The executable links the same runtime archive and `-lm`. Sanitized verification
uses `-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all` instead of `-O2`.

`bash cloud/setup.sh` succeeded, logging Go ready 0s, clang ready 0s, Node ready
0s, submodules ready 0s, build cache warm 99s, and done 99s. Each toolchain shell
sourced `/workspace/adamic-tools/env.sh`. Neither perf nor Valgrind was installed.
System apt failed with exit 100 because this worker lacks root permissions; the
configured snapshot mirror also returned proxy HTTP 403. Valgrind was obtained
from the permitted Debian mirror and unpacked under `/tmp/devirtualize-valgrind`,
without changing the system installation.

Commands and observed outputs, each test run redirected to its named log:

```sh
go test ./internal/native ./internal/ir -count=1 -timeout 30m > /tmp/devirtualize-packages.log 2>&1
# ok native 77.045s; ok ir 0.534s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(devirtualize|class_inheritance|call_targets|borrow_element_virtual)' -count=1 -timeout 30m > /tmp/devirtualize-oracle-expanded.log 2>&1
# ok oracle 2.392s
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -count=1 -timeout 30m > /tmp/devirtualize-full-oracle.log 2>&1
# ok oracle 71.532s
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/devirtualize-counts.log 2>&1
# ok oracle 13.826s
go test ./internal/native -run 'TestDevirtualizedCalls|TestExactReceiverRejectsAssignments' -count=1 > /tmp/devirtualize-focused-final.log 2>&1
# ok native 0.095s
go vet ./... > /tmp/devirtualize-vet.log 2>&1
# exit 0, no diagnostics
gofmt -l cmd internal > /tmp/devirtualize-format.log
git diff --check > /tmp/devirtualize-diff-check-final.log
# both empty
node --disable-warning=ExperimentalWarning oracle/node.mjs bench/devirtualize.a > /tmp/devirtualize-benchmark-node.log 2>&1
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/devirtualize-visitor.mjs > /tmp/devirtualize-visitor-js.log 2>&1
/tmp/devirtualize-visitor-sanitized > /tmp/devirtualize-visitor-sanitize.log 2>&1
# all three: 759040000
VALGRIND_LIB=/tmp/devirtualize-valgrind/usr/libexec/valgrind /tmp/devirtualize-valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/tmp/devirtualize-before.callgrind /tmp/devirtualize-before > /tmp/devirtualize-before-instructions.log 2>&1
VALGRIND_LIB=/tmp/devirtualize-valgrind/usr/libexec/valgrind /tmp/devirtualize-valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/tmp/devirtualize-after.callgrind /tmp/devirtualize-after > /tmp/devirtualize-after-instructions.log 2>&1
# Ir: 1565590554 and 1350550540
```

All nine timing samples are in `/tmp/devirtualize-timings.log`; complete generated
benchmark C is `/tmp/devirtualize-before.c` and `/tmp/devirtualize-after.c`.
Fixture C is `/tmp/devirtualize-fixture.c`. No test output was piped.
The whole repository gate was not run: the native package, call-target package,
and complete oracle passed; stage 1 corpus packages were not run by this unit.

Landing validation, October 7, 2026: the standing rule now requires landing this
worker's pushed branch. `git fetch origin main:refs/remotes/origin/main` followed
by `git merge --no-edit origin/main` reported `Already up to date`; current main
remained `e011f8f60899586d6373a5ccb07335ad82cfbf3c`, including a second fetch after
validation. No rebase or force push was used.

The landing gate covers every compiler package changed relative to main, including
the inherited call-target prerequisite, not only the devirtualization emitter:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/devirtualize-landing-counts.log 2>&1
# ok oracle 11.348s; no further counts.md changes
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/ir ./internal/flow ./internal/fresh ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/devirtualize-landing-gate.log 2>&1
# ok native 157.059s; ir 1.354s; flow 102.290s; fresh 56.138s;
# lower 38.761s; oracle 147.351s; exit 0
```

The visitor was also regenerated from the branch and rechecked against source
Node, generated JavaScript through the oracle runtime resolver, and sanitized C.
Each printed `759040000`; `cmp` of both outputs against Node exited 0. Logs are
`/tmp/devirtualize-landing-visitor-{node,js,native}.log`.

All count-row differences relative to the landing base are below. The five
`call_targets_*` rows are new prerequisite regression fixtures, measuring the
analysis changes on virtual calls, bounded closures and sort callbacks. They have
no earlier rows to compare. `devirtualize.a` is this unit's new regression fixture.
The one changed preexisting row removes only NULL closure temporary count calls.
No other existing row changed, and landing regeneration changed no row further.

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
|---|---:|---:|---:|---:|---:|---:|
| call_targets_element.a, new | 45 | 43 | 36 | 64 | 7 | 2 |
| call_targets_region.a, new | 27 | 22 | 13 | 33 | 9 | 5 |
| call_targets_reuse.a, new | 15 | 12 | 9 | 21 | 6 | 3 |
| call_targets_closure.a, new | 25 | 25 | 8 | 31 | 6 | 0 |
| call_targets_sort.a, new | 11 | 11 | 9 | 15 | 7 | 0 |
| devirtualize.a, new | 43 | 39 | 26 | 66 | 10 | 4 |
| class_as_interface.a, before | 372 | 372 | 360 | 525 | 60 | 0 |
| class_as_interface.a, after | 372 | 372 | 349 | 514 | 60 | 0 |
