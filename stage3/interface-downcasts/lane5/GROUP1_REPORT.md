Built: lazy callable reads now check producer arity, parameter and result representations in native and JavaScript; two Program directory contracts have source fixture coverage.
Commits: this checkpoint follows integration ba59427ccc7afecae29a305c41e6e9c7867e5610 and lane5 adapter 9c4d09044643050174409da3f8fde8303739cf6d; publish only codex/views-callables.
Commands: source, focused package, broader oracle, allocation counts and vet evidence is recorded below; the full repository gate is not claimed.
Mutants: eight source mutations and two non-callable component mutations are semantically caught; individual logs name each assertion.
Not covered: 306 candidate pairs / 1,492 static candidate reads remain; exact corpus reachability and whole-program certification are unmeasured.

The ranked inventory remains 308 candidate pairs / 1,503 static reads. The first
fixed ordinary scalar-returning contracts are Program.getCurrentDirectory (rank33,
8 reads) and Program.getCommonSourceDirectory (rank79, 3 reads). Each has nine
source fixtures: valid read/call, wrong value, wrong arity, wrong result, generic
helper, callback, storage, direct call and wrong direct-call arity. Unread generic
members remain unadmitted, demonstrating lazy admission. Optional callback absence
returns undefined. Positive fixtures agree with Node in native release, sanitizer
and JavaScript and pass leak checks. Negative runtime checks pin exit70 and the
field, complete expected function type and found kind/shape. Node's wrong-arity
and wrong-result controls run successfully; its wrong-value TypeError is normalized
to exit70 by the existing Node harness. No source type assertion certifies producer
metadata: actual generated closure code identity selects final IR signatures.

These are candidate contract fixtures, not a claim that all 11 original tsc sites
have been lowered or proven reachable. candidate-progress.json and the ranked
ledger preserve that distinction. Lazy REPORT.md retains the static lane2 table;
ADAPTED-CENSUS.md reports actual runtime reachability unmeasured because checker
diagnostics prevent production IR. The old Unknown union of 11,648 explicit
witnesses remains an upper bound, not the denominator for this checkpoint.

Working completion estimate for the full 308-candidate inventory: October 12,
2026, 18:00 UTC. This replaces the October14 Unknown-fallback estimate now that
minimal hooks are authorized. It is an estimate, not an unconditional promise;
remeasure demand when the checker-clean ledger and exact table become available.

Named hooks are listed under Lane5 in docs/checked-views-plan.md. Ordinary field
reads, native dynamic method lookup and resolved method dispatch receive signature
certificates. Dynamic viewed methods have valid/wrong-arity/wrong-result source
controls; source arity0 method bypass is caught before arguments can reach the
method. Receiver closures and stored class-method aliases remain unclaimed.
The current reconciled branch uses its existing closure code(self,args,count)
convention; no alternative ABI was introduced. The closure-convention integration
is not yet claimed. Higher-ranked intrinsics, optional/rest parameters, overloaded
and generic signatures, void contracts, aggregate parameters/results and implicit
callable-element reads remain unfinished. Scalar source activation deliberately
excludes aggregate types rather than discarding transitive result provenance.

Parser decision, from origin/codex/stage3-parser-proof 5d777de3:
The reduced native-function-any-view.a makes a fresh cache containing a copied
function value; it creates no alias to an original boolean-returning slot. That
is sound. The namespace-free reduction compiles and prints "function view loaded"
in all backends and Node. The original namespace syntax remains outside this lane.
The result of a function value is not a writable slot. Reverse writable-container
relations still refuse an AnyFunction alias that could replace an original boolean
callback with a void callback. The refusal is pinned by its invariant-mutable code
and "which can write AnyFunction where () => boolean is read" diagnostic. Node
shows the actual write-back consequence: undefined.

An immediate discarded (viewed.shouldLog as AnyFunction)() preserves its known
zero-argument scalar source signature. Its original viewed field read runs the
callable certificate before the call; a wrong-arity producer stops at debug.shouldLog.
The valid fixture prints producer/called and agrees with Node. Required source
arguments cannot be erased: that case and calls through stored erased markers
retain the named "call through an erased never-rest callable marker" refusal.
Arbitrary stored marker recovery to a narrower callable type is not implemented
by this checkpoint. No erased marker is trusted just because its result is unused.

Mutants rerun in this checkpoint:
- lower-skip-read-contract: removing the read adapter permits the wrong-arity
  callback to print7/exit0 in native release; the pinned source assertion catches it.
- native-skip-shape-dispatch: bypassing the emitted certificate has the same
  defined forbidden-call witness and catcher.
- native-accept-wrong-arity and javascript-accept-wrong-arity: each permits the
  zero-argument producer where one number argument is declared; source exit pins fail.
- javascript-skip-shape: bypassing the shape helper permits that source call.
- native-skip-method-signature: bypassing the actual method pointer certificate
  permits the arity0 method to print7; method source exit/message pins fail.
- lower-drop-marker-arity-proof: erasing the required-parameter guard admits
  the forbidden marker call; discarded-required's compile-refusal assertion fails.
- lower-allow-write-back: dropping reverse mutable-slot relation admits the
  alias; write-back's named invariant refusal assertion fails.
- native-accept-number: a valid stack closure-shaped witness tagged number
  performs a defined C call when the kind check is dropped; exit70 pin catches it.
- js-call-number: bypassed kind guard reaches a non-callable; named field/expected/
  found pin catches the resulting wrong failure. These last two are component
  witnesses, not source dispatch certification. No mutant is credited for clang
  errors or sanitizer crashes. Every mutated production file is restored.

Validation evidence is in logs/group1 and logs/source-mutants. A full-package
attempt was interrupted by the environment restart after IR passed19.260s; it is
not a completed gate. Focused retries and final source controls are recorded in
this checkpoint. The two optional-write failures were reproduced on an unmodified
ba59427 detached worktree: nominal-slot refuses the nominal class label contract;
nominal-subclass-slot lacks a reifiable source-slot write certificate. Only these
two are excluded from the broader oracle retry. Test output is always logged.

Setup was not repeated: the installed toolchain remains pinned. Original setup
seconds: Node .052, Go .054, clang .408, markdown3.595, submodules461.962,
build696.006, cache696.105, done696.130. nproc5, CPU quota4. GOPROXY remains
https://proxy.golang.org|direct; shells source /workspace/adamic-tools/env.sh.


Final completed commands and observed results:

- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallable'
  -count=1 -timeout10m: PASS19.047s, source-final.log, explicit shell exit0.
- go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript
  -run 'Test.*View|TestPrepareViewCallableRead|TestSharedArrayContractAdapter'
  -count=1 -timeout15m: lower PASS3.582s, native PASS5.170s, JavaScript PASS.800s;
  IR has no matches in this focused run. focused-packages.log, exit0.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
  -run 'Test.*View|TestPrepareViewCallableRead|TestCheckedViewCallable|TestSharedArrayContractAdapter|TestDefaultTaggedInterface|TestOptional|TestMixedUnion|TestPhantomOverload'
  -skip '^TestOptionalCheckedWrites$/^nominal-(subclass-)?slot$'
  -count=1 -timeout20m: PASS68.150s, oracle-broad-final.log, exit0.
- go test ./internal/oracle -run '^TestCheckedViewCallableCounts$' -count=1
  -timeout10m -args -update-counts: PASS3.121s. Only 28 new lane5 measured
  rows are appended to counts.md; unrelated row order and values are preserved.
- go vet ./internal/ir ./internal/lower ./internal/native ./internal/javascript
  ./internal/oracle: exit0, empty vet.log. gofmt on changed Go files: empty
  format.log. git diff --check: exit0.
- python3 lane5/run-source-mutants.py: eight semantic kills, source-mutants.log
  and logs/source-mutants/*.log. python3 lane5/run-shape-mutants.py --only
  native-accept-number and --only js-call-number: both semantic kills, copied
  component failure logs. The earlier 21 guards retain their prior evidence.

The intermediate broader retry found only a Node control expectation error:
the existing harness converts TypeError to exit70, not raw Node exit1. The pin
was corrected using observed Node output; the final uncached run above is green.
After this push: 2 candidate contracts / 11 static candidate reads covered in
source fixtures, 306 candidate pairs / 1,492 reads remaining. Exact reaching-view
production count remains unmeasured. Send the published tip to the integrator
through the user; no direct lane merge, external message or pull request is made.
