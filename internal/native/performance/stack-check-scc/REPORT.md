Built: scratch inline measurement and an inactive SCC classifier; production stack-check placement is unchanged.
Commit: classifier 89f451b, based on origin/main c01907a; benchmark inputs 4fdb5af and 0a3e651.
Commands and outputs: parse 9,259,807,920 to 8,738,361,433 Ir, 5.631288%; uncached native and oracle PASS.
Mutants: four classifier mutants failed assertions; a profile summary +1 failed instruction reconciliation.
Not covered: frame headroom proof, emitter activation, new deep-recursion fixtures or their runtime mutants, and Wasm Workers measurements.

## Cache and user-time follow-up

The [cache comparison](cache/REPORT.md) adds modeled L1 instruction misses and
ordinary user time for every measured variant. Forced inlining saves 5.63% Ir,
raises I1 misses 0.18%, expands .text 7.27%, and has a higher median user time in
seven rounds. Removing the same checks with clang-selected inlining saves 2.96%
Ir and 6.08% I1 misses with only 0.31% .text growth. Neither establishes a timing
speedup or changes the incomplete production status. The follow-up uses the
retained c01907a binaries, independently of the newer main merged for delivery.

## Sibling-call follow-on

The [sibling-call comparison](tail-calls/REPORT.md) replaces only the global
flag in scratch: instructions increase 0.118%, best-of-ten user time increases
0.816%, and modeled L1 instruction misses decrease 0.628%. It fails the 2%
gate, so the follow-on stops at measurement with no production flag changes.

## Measurement first

Reconstructed the exact SPLIT.md parse-only driver from batch 8 commit
4189abd3490757e8abe13722ceb365c451293e92. Used prepare.py from the named
4fdb5af snapshot, with its repository and Go-harness paths adapted to scratch.
The only driver change remains deleting `visit(context, root)`. Current main's
compiler and runtime are c01907a. The TypeScript corpus is the same 77 files,
pinned to 050880ce59e30b356b686bd3144efe24f875ebc8; hashes are retained.

The scratch modification removes exactly one ADAMIC_CHECK_STACK from each of
Scanner_code, Scanner_advance and isIdentifierStart, and marks their prototypes
and definitions `static inline __attribute__((always_inline))`. No scanner,
parser, runtime or production emitter file was changed. The exact C delta is
[scratch-inline.diff](scratch-inline.diff). Both generated C files are archived.

| Whole-process Callgrind Ir | Before | Scratch experiment | Reduction |
|---|---:|---:|---:|
| Parse, same harness and corpus | 9,259,807,920 | 8,738,361,433 | 521,446,487 (5.631288%) |

Both Callgrind processes exit 0 and print `0\n`. The same harness emitted as
JavaScript runs on Node and exits 0, prints `0\n`, and has empty stderr.
This count-mode comparison does not establish full AST parity. Both profiles
have the nonfatal brk-segment warning also present in SPLIT.md; both completed.
The repository's stage1/typescript/scanner/profile.py reconciles every self cost
with each summary. A scratch mutant increasing only summary by one is rejected
with `callgrind self costs do not sum to summary`.

Disassembly has 90 direct calls to these helpers before and zero after.
Address-taken scanner method bodies remain, as their tables require.
The fresh baseline differs from SPLIT.md's historical 9,259,261,285 by 546,635
instructions. This report compares the newly rebuilt pair, not a new binary to
an old profile. No wall-time parse speedup is claimed.

Exact build flags for both versions, from the repository root:

```sh
source /workspace/adamic-tools/env.sh
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/stack-check-scc/parse-before.c internal/native/runtime/*.c -lm -o /workspace/scratch/stack-check-scc/native-before > /tmp/stack-check-clang-before.log 2>&1
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O2 -g -I internal/native/runtime /workspace/scratch/stack-check-scc/parse-after.c internal/native/runtime/*.c -lm -o /workspace/scratch/stack-check-scc/native-after > /tmp/stack-check-clang-after.log 2>&1
VALGRIND_LIB=/workspace/scratch/stack-check-scc/valgrind/usr/libexec/valgrind /workspace/scratch/stack-check-scc/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/stack-check-scc/parse-before.callgrind /workspace/scratch/stack-check-scc/native-before --manifest /workspace/scratch/stack-check-scc/compiler.txt --count > /workspace/scratch/stack-check-scc/parse-before.stdout 2> /workspace/scratch/stack-check-scc/parse-before.stderr
VALGRIND_LIB=/workspace/scratch/stack-check-scc/valgrind/usr/libexec/valgrind /workspace/scratch/stack-check-scc/valgrind/usr/bin/valgrind --tool=callgrind --callgrind-out-file=/workspace/scratch/stack-check-scc/parse-after.callgrind /workspace/scratch/stack-check-scc/native-after --manifest /workspace/scratch/stack-check-scc/compiler.txt --count > /workspace/scratch/stack-check-scc/parse-after.stdout 2> /workspace/scratch/stack-check-scc/parse-after.stderr
```

Callgrind is 3.24.0, extracted from Ubuntu's valgrind_3.24.0-0ubuntu2_amd64.deb
in scratch. `apt-get download valgrind` first failed with exit 100: no candidate
and an unreadable apt configuration warning. The public archive worked.
The initial shallow checkout lacked 4189abd; fetching codex/stage1-lint-batch8
supplied the exact commit. Neither workaround changes the benchmark.

## Unit stopped: what would have to change

Stopped at the user's request after the negative timing result. The classifier
is retained as an inactive analysis in 89f451b; it has no emitter call site.
All production stack checks and the global -fno-optimize-sibling-calls remain.
The later approval for a small named native.go frame-bound hook is not exercised:
no such hook was built, and no further implementation or benchmark is planned
for this unit. The 5.63% scratch instruction reduction came with 2.53% higher
median user time in seven ordinary runs, not a demonstrated parse speedup.

The concrete fallback trigger is the structural interface stored in
Statements.parser, declared readonly StatementContextInterface in
stage1/typescript/parser/statements.ts. For example, Statements.make calls
this.parser.make (source line 55; retained C line 16237), and Statements.type
calls this.parser.type (source line 58; C line 16266). Statements.block calls
parser.expect and parser.kind; Statements.semicolon calls parser.kind and
parser.next. Each is an ir.CallClosure on a method property whose receiver is
a field read. exactReceiverClass only proves allocations and singly bound
allocation locals; it does not prove field reads or constructor parameters.
Therefore exactReceiverMethod cannot bound these targets, and ClosureTargets
returns Unknown for the property. Any one activates the classifier's
program-wide fallback, marking all 227 functions.

A read-only inventory of the already retained parse-before.c finds **425**
adamic_object_callee sites in **20** Statements methods. The exact method
names, counts, generated function names and C lines are retained in
[unresolved-dispatch-sites.json](unresolved-dispatch-sites.json). It includes
make, type, expect, kind, next, node, bindingName, rootAssignment, token,
decorator, identifier, propertyName, entityName, peek, getAwait, setAwait,
typeLiteral, getIn, getYield, parameters, returnType, setIn, setYield,
typeParameters, depth, methodBody, primary, suffix, typeArguments,
nextIdentifierSameLine and rootExpression. The emitter uses the same
exactReceiverMethod proof before emitting these dynamic sites. This inventory
identifies actual fallback-causing interface calls, not a new IR probe or a
claim that every indirect C call is unknown. Inline callback literals elsewhere
can be bounded by ClosureTargets despite indirect C emission.

For a future attempt to pay, it would need all of the following:

- Proven target sets through constructor arguments and stored fields for these
  structural calls, or a sound restricted target-set fallback. A readonly
  interface signature alone cannot exclude an object with an own closure;
  assuming the one apparent Parser implementation would be unsound. One
  remaining Unknown still marks the entire program in this classifier.
- A measured smaller set that actually leaves scanner leaves unchecked, followed
  by a compact emitted shape whose ordinary user-time result improves. The
  forced-inline scratch result is evidence against assuming instruction savings
  produce a speedup. The clang-selected alternative is documented separately
  and also does not establish a timing win.
- Proven build-specific bounds for unchecked machine-frame paths, including
  runtime/adapters, sanitizer variants and panic headroom, then the emitter
  hook and the requested recursion/placement mutants. The frame-bound scope
  approval removes an ownership restriction; it does not supply that proof.

This final update only inspects retained artifacts and adds documentation.
No compiler build, test gate, new measurement, runtime edit or emitter change
was performed after the stop instruction. The preceding tests and mutants are
reported below with their original coverage limits.

## Classifier and the remaining proof

internal/native/stack_checks.go builds direct and virtual edges with CallTargets,
closure and built-in callback edges with ClosureTargets, and resolves exact
interface receivers using the existing devirtualization proof. Tarjan components
mark every member of a recursive component, including self recursion. An
unresolved function-value call marks every function, covering its possible targets
and their descendants. Top-level unresolved calls count too. This is conservative
and deliberately does not infer a bounded target from a static signature.

The classifier is not hooked into functionBody or signatures. Rebuilding the
compiler with it produces byte-identical parse C (cmp exit 0). A temporary probe,
removed after use, observes **227 candidates out of 227 functions** for this
parse harness because an unresolved call activates the conservative fallback.
Thus even an emitter hook alone would not reproduce the scratch gain. A more
precise sound indirect-target analysis remains necessary for this program.

The existing runtime reserves min(256 KiB, stack size / 8), independently of the
program's call graph. Neither this margin nor an IR function count establishes
an upper bound on unchecked machine frames: the program can contain an arbitrarily
long acyclic chain and large argument-pack arrays; clang controls inlining,
spills, ABI frames and sanitizer instrumentation. ASan and release builds have
different frame sizes. This is a limitation of the proposed proof, not an observed
crash or a claim that the current fixtures fail.

A build-specific solution would obtain actual clang frame data, include runtime
and adapter edges, bound paths between checks after inlining, retain panic room,
and reject unbounded/dynamic frame reports. It must also account for small stacks
and unchecked paths from main. That requires a build-pipeline integration, normally
in internal/native/native.go, which the unit explicitly excludes. No fixed margin
is asserted proven, no partial estimate is substituted, and production checks
remain in place. This delivery is incomplete; it does not implement the requested
production optimization.

## Tests and mutants

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/native -run 'TestRecursiveStackComponents|TestStackCheck' -count=1 > /tmp/stack-check-classifier.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/native ./internal/oracle -count=1 -timeout 30m > /tmp/stack-check-native-oracle.log 2>&1
python3 /workspace/scratch/stack-check-scc/mutants.py > /tmp/stack-check-mutants.log 2>&1
go test ./internal/native -run 'TestRecursiveStackComponents|TestStackCheck' -count=1 > /tmp/stack-check-classifier-restored.log 2>&1
gofmt -l internal/native/stack_checks.go internal/native/stack_checks_test.go > /tmp/stack-check-gofmt.log
go vet ./internal/native > /tmp/stack-check-vet.log 2>&1
```

Native passed in 155.236s and oracle in 159.337s, uncached, with existing
release, ASan/UBSan and leak lanes. Focused classifier tests pass after source
restoration. The exact-interface test was added after the full package run and
passed in the focused run. Formatting, touched-package vet and diff checks pass.
No full repository gate is claimed.

All 512 three-vertex directed graphs are compared against transitive closure,
independent of Tarjan's low links. Target tests cover mutual recursion, acyclic
descendants, virtual implementation sets, constant and literal closures, unknown
top-level closure calls, nested exceptional bodies, exact and assigned interface
receivers, and map, visit, reduce, Array.from, Map/Set forEach and sort callbacks.

| Independently run mutant | What caught it |
|---|---|
| Clear function 1's recursive-component decision | TestRecursiveStackComponentsAgainstReachability: graph 000001010, function 1 recursive but unmarked |
| Ignore unknown target sets | unresolved top-level target assertion |
| Use only Call.Function for virtual calls | virtual implementation cycle assertion |
| Omit ArrayMap callback edges | built-in callback cycle assertion |
| Increase Callgrind summary only by 1 | self-cost reconciliation |

Every classifier mutant compiled and failed its intended test with exit 1.
These are **classifier** mutants, not runtime check-removal mutants. Removing one
check in a mutual-recursion component may leave another check that still panics;
byte-for-byte panic fixtures alone do not necessarily detect that missing member.
Placement assertions would be needed as well. New deep-recursion fixtures through
closures, interfaces, map callbacks and two modules have not been added or proven.

## Workers sieve

Used the original checksum driver, sieve.a, sieve.txt, statistics.mjs and wait4.c
from 0a3e651. The extraction and measurement adapter are scratch-only. Before and
after refer to the unchanged compiler and the compiler containing the inactive
classifier, respectively. Generated sieve C compares equal; both binaries have
SHA-256 e87342cacd50d14ab99e37e194e39edbfd1057a06e1b7b6da99be6f728ed79be.

Native uses the standard native.Flags(Options{}) release flags: C11 warnings,
-ffp-contract=off, -fno-optimize-sibling-calls and -O2, without -g, sanitizers or
allocation counting. wait4 uses -O2 -std=c11 -Wall -Wextra -Werror -pedantic.
Node 24.19.0 runs oracle/node.mjs with --disable-warning=ExperimentalWarning.
K=1000, three corpus lines, 3000 requests per measured batch, three interleaved
rounds, independent K=0 startup subtraction, no warmup. All native preflight and
measured response checksums agree with Node-source.

| Native sieve, ms/request | Before | Inactive classifier |
|---|---:|---:|
| Median wall, startup subtracted | 0.724771 | 0.719452 |
| Best wall, startup subtracted | 0.715630 | 0.708477 |
| Median CPU, startup subtracted | 0.723948 | 0.718666 |
| Best CPU, startup subtracted | 0.714980 | 0.707641 |

The artifacts are identical; these differences are noise, not an optimization.
[sieve-measurements.json](sieve-measurements.json) retains all child rusage,
checksums, loads and pairs. This is the native lane only, using the original
instruments via sieve-measure.mjs, not the complete four-column Workers benchmark.
Current main lacks the sibling Wasm service/host checkout required by that harness;
no Wasm, workerd or HTTP results are claimed.

## Delivery and setup

Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5, cgroup quota 4 CPUs.
`bash cloud/setup.sh`: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready
0s, build cache warm 115s, done 115s. Environment: /workspace/adamic-tools/env.sh.
Every test/program output was sent to a file. No cohere code was copied.

Before delivery, `git fetch --no-recurse-submodules origin main` and
`git merge --no-edit origin/main` reported Already up to date (c01907a).
No emitter sibling was merged: no production emitter file was touched, and this
is an inactive helper pending the missing proof. Only codex/stack-check-scc is pushed.
