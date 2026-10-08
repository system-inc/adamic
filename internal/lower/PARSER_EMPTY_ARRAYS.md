Built: parser discovery row 3's exact probe changes from refusal to Node agreement; generic number/object empty fallbacks also agree.
Commit: implementation 297000c043a41cea8916d450c6f46220bb6e3e26 on codex/parser-generic-empty-array, based on current main 71d7e491.
Commands: focused uncached oracle passes (1.079s), counts update passes (51.084s), vet and formatting pass; final package gate recorded below.
Mutants: bottom view, readonly destination, object layout, mutable bottom, scalar guard, contextual scope, falsy empty array and shared fresh array all caught.
Not covered: full native parser acceptance, earlier parser blockers, full repository test suite and a fresh census recount.

The request follows row 3 in origin/codex/stage3-parser-proof 2179dd8,
stage3/drivers/parser/BLOCKERS.md. That row was discovered behind explicit
scratch-only stubs; passing this minimal does not prove the validated parser
slice reaches or passes it. Its original diagnostic is invariant-mutable: a
never[] fallback seen as U[] even though the declared result is readonly U[].
The exact native-generic-empty-array.a probe is copied byte-for-byte into the
oracle fixtures; its baseline refusal is /tmp/empty-array-before-probe.log.
The probe's comment names upstream core.ts:414; the discovery list names the
adapted core.ts:230. Neither source was rewritten to bypass lowering.

The lowering now treats never as having no observable value in a readonly
view, while retaining the reverse element relation for mutable containers.
Fallback branches use their actual contextual array destination, so returning
a shared empty array through readonly U[] does not invent a mutable U[] view.
This contextual override is limited to arrays: the existing gaps.a fixture
caught a first attempt that also used synthetic object-destructuring types.
Mutable never[] aliases and inferred mutable fallback locals remain refused.
Nonempty generic literal casts remain refused as well.

An empty never[] literal has a real array identity and no element slots;
lowering can therefore give it an unused placeholder element representation.
An empty literal cast to T[] instead uses the existing monomorphized element
representation. TestEmptyArrayInstantiations checks number and object IR element
layouts separately. Arrays and array-or-undefined unions may use lazy coalescing
for || because all arrays are truthy, including []. This proof does not extend
to numbers or strings; tests hold their unsupported || shapes at NotYet.
The existing coalescing operation evaluates the left side once and the fallback
only when missing. No native emitter or runtime changes were needed.

The expanded generic_empty_array.a fixture checks readonly module-level
never[], original input identity, [] truthiness, fresh literal identity, writes
and reads through number/object instantiations, a writable object constraint,
and lazy single left-side evaluation. Strings are built at runtime so object
reference ownership is exercised. Original-source Node prints:

```text
0 0 true true
true true
true false
9 freshfresh
0 0 false
true true
true 0 2 1
0
```

Recorded counts, columns allocations/frees/retains/releases/peak/regions:
exact probe 3/3/1/5/3/0; expanded fixture 32/32/24/60/17/0.
Both finish with all allocations freed. The differential oracle compares the
source on Node, sanitized native, counted native and emitted JavaScript.

Mutant evidence:

- Remove the never bottom rule: expanded fixture's writable constraint is
  refused by Lower. An initial {} constraint masked this mutant; the writable
  constraint fixture was added and the mutant rerun, now caught.
- Ignore the readonly contextual destination: the exact probe again refuses
  in Lower with invariant-mutable.
- Force object array element layout to number: TestEmptyArrayInstantiations
  fails its number/object layout assertion.
- Skip the reverse mutable relation for never elements:
  TestEmptyArrayViewsStillRefuse catches accepted unsafe mutable aliases.
- Make the array || guard accept scalars:
  TestEmptyArrayScalarFallbacksStayNotYet fails on a compiled scalar fallback.
- Apply contextual fallback destinations to all types again: the existing
  gaps.a oracle fixture refuses its object destructuring. This isolates the
  array-only restriction that repaired the initial gate regression.
- Treat [] as falsy in lowered IR: the sanitized mutant finishes cleanly;
  Node catches false true instead of true false for present-empty identity.
- Replace fresh<number>'s empty literal by the shared module array: the
  sanitized mutant finishes cleanly; Node catches 1 0 true instead of 0 0 false
  after writing a prior fresh result. Object instantiations remain fresh in
  this mutant so only output disagreement catches it.

Four production mutants were rerun using Go source overlays, not repository
edits: /tmp/empty-array-production-mutants-final.log and per-mutant logs.
Scalar and contextual overlay logs are /tmp/empty-array-scalar-mutant.log and
/tmp/empty-array-context-mutant.log. Runtime mutants are permanent tests in
internal/oracle/empty_arrays_test.go. Each production mutant exits 1 from the
stated test; both runtime-mutant tests pass only when stdout alone disagrees,
with exit 0 and empty sanitizer stderr. No mutant was killed by clang -Werror.

Toolchain setup for this working session: Go ready 0.030s, Node ready 0.036s,
submodules ready 0.082s, clang ready 0.301s, Markdown dependencies ready 1.158s,
go build ready 37.562s, test binaries deferred, cache warm 37.680s,
done 37.714s. nproc 5; cpu.max 400000 100000. Go 1.27.1, clang 20.1.8,
Node 24.19.0. Environment sourced from /workspace/adamic-tools/env.sh.
Setup log /tmp/empty-array-setup.log.

Commands (all test output redirected directly to logs):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(native-generic-empty-array|generic_empty_array|gaps)' -count=1 -v -timeout 30m > /tmp/empty-array-focused-scoped.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/empty-array-counts-final.log 2>&1
python3 /tmp/empty-array-mutants.py > /tmp/empty-array-production-mutants-final.log 2>&1
go test -overlay=/tmp/empty-array-scalar-overlay.json ./internal/lower -run TestEmptyArrayScalarFallbacksStayNotYet -count=1 > /tmp/empty-array-scalar-mutant.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -overlay=/tmp/empty-array-context-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/gaps' -count=1 > /tmp/empty-array-context-mutant.log 2>&1
go vet ./... > /tmp/empty-array-vet-final.log 2>&1
gofmt -l cmd internal > /tmp/empty-array-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -count=1 -timeout 30m > /tmp/empty-array-gate-final.log 2>&1
```

The initial broad contextual attempt failed counts and the full oracle only at
gaps.a's destructuring refusal; retained in /tmp/empty-array-gate.log and
/tmp/empty-array-counts.log. The narrowed implementation was then retested.
Vet and formatting logs are empty; git diff --check passes. No code was copied
from cohere and no protected compiler orchestration or oracle file was edited.

Final touched-package gate exits 0: internal/lower 41.208s and the complete
uncached internal/oracle suite 144.963s. This includes both permanent runtime
mutants, the new negative checks and recorded-count comparison. The complete
repository test suite was not run; the touched lowering package and complete
oracle were run. A final fetch leaves origin/main at 71d7e491, already the
branch base, so no merge commit is needed. No full native parser or scanner
acceptance is inferred from either isolated fix.
