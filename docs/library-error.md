# Error constructor observations and stack capture

`Error.captureStackTrace` is a truthy function value. Its identity is stable across
reads, and calling it with a target and an optional constructor function is a
defined no-op on both backends. Arguments are still evaluated in source order.
The JavaScript backend emits Adamic's empty function and never calls V8's real
`Error.captureStackTrace`. Native has no JavaScript frames to trim.

Every `.stack` or literal `['stack']` read remains NotYet, including destructuring and reads from a
plain object passed to captureStackTrace. It is never silently undefined. If
`.stack` is ever lowered, this slice is the place to define what it holds and
ensure that both backends implement that definition.

The implementation is in `internal/lower/library_error.go`. It synthesizes an
ordinary IR closure, so no new C runtime translation unit or V8 algorithm port is
needed. Error instances, subclasses, causes and catchability use the compiler
unit's nominal class representation from `codex/error-classes-counts`, merged at
`66766ab`. This slice does not edit `error_classes.go` or `exceptions.go`.

`typeof Error` yields `function`. Other reads of Error as a constructor value,
including aliases and constructor arguments, stay NotYet with a reason:
overloaded calls and constructor aliases are not lowered. The census-pinned
TypeScript v6.0.3 commit is `050880ce59e30b356b686bd3144efe24f875ebc8`.
An audit of its compiler sources found Error construction, Debug.fail's capture
feature test and call, and sys.ts's stackTraceLimit reads/writes; it found no
passing of Error as a constructor argument. stackTraceLimit remains unsupported.

The pinned loader dependency `69c71d5` is merged. The official @types/node 25.3.3
signature is `captureStackTrace(targetObject: object, constructorOpt?: Function):
void`. A type-only node:fs import in these fixtures activates those declarations.
The original scanner probe and upstream Debug.fail spell the member access
through `Error as any`. This exact member receiver is recognized as an intrinsic
using the underlying library Error identity. No general any value is lowered.
The unmodified scanner probe is an oracle fixture alongside a typed Node fixture.
The typed probe also calls capture inside its feature branch, avoiding the
checker's TS2774 diagnostic for testing an always-defined function without using
it there.

The Debug.fail fixture constructs an Error, selects `stackCrawlMark || fail`, calls
captureStackTrace, throws, and prints message and name from the caller's catch.
It runs once with undefined and once with a passed function. The crawl-marker
parameter is written as a required function-or-undefined parameter because
ordinary functions with optional parameters cannot yet be read as values.
Detached capture calls, function identity, omitted constructorOpt and typeof are
also held to Node. The post-capture stack fixture must fail at compile time, and
lowering tests assert its specific reason, including plain-object and bracket
views.

Linux measurements use test262 commit
`c8c798898646638cd0c24879f8e0374e847e7d74`, with adaptation enabled.

Before the slice, after merging the compiler and loader dependencies:

| Directory | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/Error | 0 | 0 | 7 | 0 | 9 |
| built-ins/Error/isError | 0 | 0 | 5 | 0 | 7 |
| built-ins/Error/prototype | 0 | 0 | 6 | 0 | 3 |
| built-ins/Error/prototype/constructor | 0 | 0 | 1 | 0 | 1 |
| built-ins/Error/prototype/message | 0 | 0 | 0 | 0 | 1 |
| built-ins/Error/prototype/name | 0 | 0 | 0 | 0 | 1 |
| built-ins/Error/prototype/stack | 0 | 0 | 2 | 0 | 33 |
| built-ins/Error/prototype/toString | 8 | 0 | 2 | 0 | 7 |
| **Total** | **8** | **0** | **23** | **0** | **62** |

After the slice:

| Directory | Pass | Disagreement | Refused | Crashed | Skipped |
|---|---:|---:|---:|---:|---:|
| built-ins/Error | 0 | 0 | 7 | 0 | 9 |
| built-ins/Error/isError | 0 | 0 | 5 | 0 | 7 |
| built-ins/Error/prototype | 0 | 0 | 6 | 0 | 3 |
| built-ins/Error/prototype/constructor | 0 | 0 | 1 | 0 | 1 |
| built-ins/Error/prototype/message | 0 | 0 | 0 | 0 | 1 |
| built-ins/Error/prototype/name | 0 | 0 | 0 | 0 | 1 |
| built-ins/Error/prototype/stack | 0 | 0 | 2 | 0 | 33 |
| built-ins/Error/prototype/toString | 8 | 0 | 2 | 0 | 7 |
| **Total** | **8** | **0** | **23** | **0** | **62** |

No test262 test newly passes. captureStackTrace is Node-specific; its new oracle
fixtures agree with Node on both backends, including native ASan, UBSan and leak
checks. Refusal tests distinguish unsupported stack observations from an absent
value.

Commands write all test output to logs:

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic-test262 -adapt -test262 /tmp/library-error-test262 built-ins/Error > /tmp/library-error-before.log 2>&1
go run ./cmd/adamic-test262 -adapt -test262 /tmp/library-error-test262 built-ins/Error > /tmp/library-error-after.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestLibraryErrorRefusals|TestLibraryErrorCounts|TestNativeAgreesWithNode/internal/oracle/testdata/library_error' -count=1 -timeout 30m > /tmp/library-error-complete-focused.log 2>&1
go vet ./... > /tmp/library-error-vet.log 2>&1
```

Setup completed on the stable merged tree: Go ready 0s, clang ready 0s, Node ready
0s, submodules ready 0s, build cache warm 82s, done 82s, nproc 5, cgroup quota 4.
The first setup's cache warm overlapped dependency merging and failed on the
inconsistent intermediate files; the stable rerun succeeded.

The required mutants were run independently and restored:

| Mutant | Check that caught it |
|---|---|
| Feature test uses IsUndefined without negation | Probe stdout differs on both backends: unavailable instead of available |
| Capture closure throws Error("capture mutant") | Debug.fail fixture's caught messages and final exit differ on both backends |
| Remove the stack-read refusal hook | Plain-object post-capture stack test unexpectedly lowers |
| Remove the destructured stack refusal | Renamed stack binding unexpectedly lowers |
| typeof Error yields object | typeof fixture stdout differs on both backends |
| Allocate a new capture function on each read | Identity observation prints false on both backends |

All six mutant tests exited 1; neither depended on clang warnings or a refusal.
Logs are `/tmp/library-error-mutant-falsy_feature.log` and
`/tmp/library-error-mutant-capture_throws.log`.

Full counts regeneration was attempted with
`go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args
-update-counts`, logging to `/tmp/library-error-counts.log`. It is blocked by an
integration failure in the compiler dependency: guardRuntimeRanges calls
finishClassCalls after structural accessors have registered their targets,
clearing them, and CallTargets panics on class_features_accessors.a. The new slice
counts are measured and checked separately by TestLibraryErrorCounts. The full
repository gate is not claimed green while that dependency failure remains.

The accessor-target panic also reproduces on the dependency-only merge `89b04e0`,
before this slice. Its log is `/tmp/library-error-dependency-blocker.log`.
`go test ./internal/lower ./internal/load -count=1 -timeout 30m` was attempted:
loading passed (1.360s); lowering panicked in TestClassFeaturesAccessorRefusals
with the same missing-target failure. `/tmp/library-error-packages.log` retains
the full output. Fixing that shared compiler integration remains with its owner
unless explicitly authorized. No edit to lower.go was made.

The slice's shared hooks are three lines each at expression.go:382-384,
control.go:98-100 and object.go:598-600, plus the named capture-read exemption in
refusals.go:139. A four-line binding refusal hook at refusals.go:79-82 also
keeps destructured stack fields behind the same boundary. The oracle registration lives in a new slice test file;
oracle_test.go is unchanged.
