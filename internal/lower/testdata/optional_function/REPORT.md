# Optional callable presence

Built on codex/method-presence-test, retaining method-presence commit 722a1c5afb5c070354af2f149030ae65feec7f32.
Implementation commit: 49347bcea2517632ee69535d158b1b5e0788b8fd.
Merged origin/main ce0750f2 through merge commit 73534167707b3c4f2ad91d88e3326c9572d2baaf.

A condition reading a closure whose checker type includes undefined lowers to !IsUndefined.
A nullable callable uses the closure representation and lowers its condition to !IsNull.
Both operations evaluate their source once. A required function condition remains refused.
Nullable representation is admitted only when every present member already has closure representation.

Conservative assumption: null and undefined are supported as separate absence types.
A union holding both remains refused because the existing pointer representation cannot distinguish them.
The lower refusal pins preserve that boundary and the existing nullable-object boundary.

## Observations

Node prints true false for locals, parameters and fields, and true false false for captures
whose callback is cleared between reads. The same fixture is compared with both Adamic backends,
including sanitized and release native builds and the leak check.

The two original memoize probes are copied intact from d0e9a373, apart from a provenance comment.
Node prints 42 42 1 for both. Adamic reaches its independent adamic/cycle-capable refusal for
callback (a) and pending (b). TestOptionalFunctionMemoizeBoundary records this limitation;
it does not count those probes as native successes. A concrete memoize attempt also hit this guard.
No cycle analysis was changed. The named evidence/p2-repeat/results.json was absent from both
origin/codex/stage3-adapt-memoize and origin/codex/stage3-parser-proof.

## Verification

Every command's output was redirected to a log file.

- export GOPROXY='https://proxy.golang.org|direct'; ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
  passed; cumulative timing: Node .022s, Go .023s, markdown .072s, submodules .087s,
  clang .220s, build 49.211s, cache 49.559s, done 49.647s. nproc=5, cgroup quota=4 CPUs.
  Sourced /workspace/adamic-tools/env.sh. Go 1.27.1, Node 24.19.0, clang 20.1.8.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/lower
  passed in 36.472s; /tmp/optional-function-lower-final.log.
- go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded -args -update-counts
  passed in 38.584s; /tmp/optional-function-counts.log.
  New row: allocations 28, frees 28, retains 25, releases 51, peak 13, regions 0.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle
  passed in 130.023s; /tmp/optional-function-oracle-final.log.
- go vet ./..., gofmt -l cmd internal, git diff --check: clean.

The complete lowering and oracle packages were run; the whole-repository test gate was not run.

## Mutants

Run source /workspace/adamic-tools/env.sh then python3 internal/lower/testdata/optional_function/run-mutants.py.
The runner restores each compiler source in finally and verifies the intended failure, not a build failure.

- present-undefined-false: Node stdout disagreement catches the inverted presence test.
- present-null-false: Node stdout disagreement catches the inverted nullable test.
- required-allowed: TestOptionalFunctionKeepsRequiredRefusal catches a nil error.
- nullable-refused: the Node fixture catches the lost nullable representation at lowering.
- nullable-object-allowed: TestOptionalFunctionNullableObjectIsNotYet catches a nil error.
- mixed-absence-allowed: TestOptionalFunctionMixedAbsenceIsNotYet catches a nil error.

An initial present-as-false run failed at an unrelated cycle refusal in an exploratory concrete
memoize fixture. That fixture was removed before the meaningful rerun. An initial mixed-absence
mutant was not caught because a second representation guard still rejected undefined; after the
nullable representation was restricted to all-callable present members, the final mutant removed
the actual mixed-absence barrier and the refusal pin caught it.

All six final mutants exited 1 as required. Logs: /tmp/adamic-optional-function-mutants and
/tmp/optional-function-mutants.log. The earlier five method-presence mutants remain documented
in the preceding unit report and have not been rerun for this extension.

## Limits

The exact generic memoize probes cannot be native successes without a separate sound cycle-analysis
change. Combined null and undefined unions need a distinct absence representation. This extension
covers branch and loop conditions through condition lowering; it does not expand arbitrary function
truthiness through ! or value-producing logical operators. The earlier host process.nextTick descriptor
gap remains. No prohibited compiler files or cohere sources were edited.

Front-3 integration at pinned 16cb9b105127f28eb5a6d1d0af7757c6083e2d27 composes the census ToBoolean proof with optional/nullable callable presence and the required-callable condition refusal. Primitive/boolean conditions and unary callable observations retain their existing census behavior. The generic representation-erasing cast guard runs before methodObservation. Detached-own-method support and observation-only method exemptions coexist; all escaping-method refusals remain covered.

A separate method_presence_wide_union.a fixture exercises presence for a number|string method signature that still has no closure slot on front-3, where optional boolean signatures now do. Both original method-presence and optional-callable sources remain intact. Existing status bytes outside stage0 are verified unchanged; one existing Refused diagnostic changes in assertions/06_memoize_clear.a, with no Compiles downgrade.

The initial six-package integration gate passed lower/native/IR/oracle/fixtures but exposed a test inventory mismatch in flow's SSA, path, mutation-range and liveness checks: its glob assumed the two original memoize refusal probes produced IR. Classify only those two sources with the existing refusal probes; their independent Node output and cycle-capable refusal pins remain active. The corrected flow package passes uncached in 130.246s. Vet and gofmt are rerun after this test-only fix. No compiler source changed for the inventory repair. Raw first failure and corrected flow evidence are in stage3/front-3/item20.
