Built: per-overload readonly structural body proof for step 05 and step 16.
Commits: contract 086f5662; ruled contract 2851fc6f; implementation SHA is in the delivery report.
Validation: focused oracle/lower tests, counts, CallTargetReaders and package vet passed.
Mutants: full implementation types, existential annotated return, hidden accessor and throw fallthrough all caught.
Limits: reduced witnesses only; unsupported control flow, receivers and effects retain conservative stops.

The dependency remains origin/compiler/hidden-boundaries-main at 36e7fac4.
The refreshed dependency merge reported Already up to date. No unrelated worker
branch, protected file or cohere source was edited or copied.

The new proof seeds every implementation parameter from one overload's entire
admitted domain, including optional undefined. It tracks local declarations and
assignments separately on each path, follows readonly discriminants through a
stable boolean alias, walks switches with fallthrough and breaks, and treats
escaping throws as abrupt exits. Returned values use Adamic's relation and the
existing storage compatibility guard. Fixed ordinary checked helper returns
may suffice; preserving flow facts across a helper requires a conservative
no-write, no-capture, known-call summary. A matching accessor elsewhere in the
closed source tree refuses property flow facts because an interface can hide it.

The Block reduction preserves the non-arrow overload, saved isArrowFunction,
branch-assigned ConciseBody local, complete Block and shared statements identity.
The evaluate reduction preserves the TemplateExpression switch arm, checked
helper result, optional location, metadata, factory-returned overload set and
mutable metadata alias. Both run as .a with zero inserted overload checks.
The negative's template arm can return { value: 1 }; its refusal names
result.value and quotes that reachable numeric return with its source location.

The .ts field hatch is unchanged: all five existing kind/value tests pass,
including failing checks and the refused .a partners. Step 05's readonly hatch
cannot fall back to a fresh-record runtime check in .a. The older fresh writable
scalar-record checked conversion remains outside this change; its existing
writable evaluate fixture is still accepted and has unchanged counts.

Final commands (all test output went directly to files):

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 timeout 120 go test ./internal/oracle -run '^TestOverloadBodyProof|^TestOverloadFieldHatch' -count=1 -timeout 90s -v
# ok github.com/system-inc/adamic/internal/oracle 2.336s
# cache: native hits=0 misses=19; node hits=0 misses=15

timeout 120 go test ./internal/lower -run '^TestOverloadBodyObligation|^TestOverloadStructural|^TestCensusOverloadReturnProof$' -count=1 -timeout 90s -v
# ok github.com/system-inc/adamic/internal/lower 0.692s

timeout 300 python3 review/compiler/overload-proof/run-mutants.py
# all four caught by their intended assertions; every mutation restored

timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts
# ok github.com/system-inc/adamic/internal/oracle 51.234s

timeout 120 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s
# ok github.com/system-inc/adamic/internal/ir 16.299s

timeout 120 go vet ./internal/lower ./internal/oracle
# exit 0, no output

git diff --check
# exit 0, no output
```

Integration lane checks are run after the implementation commit before its push,
using the required fetched script and sourced toolchain. Their output is in
lane-checks-turn2.log and the final delivery report. The preceding doc push passed:
`lane checks 1.8 s: gofmt and tools on 47 Go files, t.Parallel on 5 test packages; vet 5 packages`.

Mutants are independent patches, never compilable Go evidence:

| Mutant | Catcher |
|---|---|
| full-implementation-types | Both positive oracle leaves fail with structural refusals at result.kind/result.value. |
| trust-overload-annotation | Skip narrowing and accept any statically annotated matching return without checking all reachable returns. The numeric negative test fails: program=true, error=<nil>. |
| hidden-accessor-is-data | Hidden-accessor proof control fails: proof=true, want false. |
| throw-falls-through | Early-throw proof control fails: proof=false, want true, because it sees an unreachable numeric return. |

No kill was a compiler error, backend build warning, timeout or unrelated refusal.
The negative source Node prints a number successfully; the test requires Adamic
to refuse before either backend receives IR. The positive tests use source Node,
backend Node, release native, sanitized native and leak checks.

The entire recorded counts diff is two added rows and one changed existing row:

| Fixture | Allocations/frees/retains/releases/peak/regions |
|---|---|
| overload_body_proof/block.a | 14/14/33/39/4/0 |
| overload_body_proof/evaluate.a | 14/14/36/43/7/0 |
| overload_results_transform.a | 8/8/9/11/5/0 -> 8/8/7/9/5/0 |

The older transform's result check is now statically proven, removing two
retain/release pairs. All other existing rows remain unchanged.

Every new top-level leaf calls t.Parallel first. Observed leaf seconds follow;
none approaches the one-instance 60-second limit:

| Test | Seconds |
|---|---:|
| TestOverloadBodyProofRefusesNumericReturn | 0.41 |
| TestOverloadBodyProofEvaluate | 1.19 |
| TestOverloadBodyProofBlock | 1.27 |
| TestOverloadBodyObligationCapturedWrite | 0.07 |
| TestOverloadBodyObligationSwitchBreak | 0.07 |
| TestOverloadBodyObligationSwitchFallthrough | 0.08 |
| TestOverloadBodyObligationHiddenAccessor | 0.06 |
| TestOverloadBodyObligationUnknownArgument | 0.06 |
| TestOverloadBodyObligationEarlyThrow | 0.06 |
| TestOverloadBodyObligationRecursiveHelper | 0.04 |
| TestOverloadBodyObligationReboundParameter | 0.05 |
| TestOverloadBodyObligationFallthrough | 0.04 |
| TestOverloadBodyObligationNumericBranch | 0.04 |
| TestOverloadBodyObligationUnknownCondition | 0.05 |

Setup completed with Go ready 0.024s, Node ready 0.025s, submodules ready
0.068s, Markdown dependencies ready 0.079s, clang ready 0.176s, build ready
10.619s, build cache warm 10.734s and done 10.758s. nproc is 5, with a
4-CPU quota. The printed environment was /workspace/adamic-tools/env.sh.
The first counts run found missing @types/node 25.3.3; installing the pinned
stage3/api dependencies with timeout 180 npm ci --prefix stage3/api fixed it.
That first run also caught an overbroad refusal of the legacy writable record
conversion; the final implementation confines the new mandatory body proof to
the ruled readonly covariance boundary.

Not covered: the full original tsc functions, general loops, try/catch/finally,
parameter rebinding, destructured/default/rest parameters, explicit or implicit
receiver proofs, general generic helper substitution, overloaded helper proof
dependencies, unknown callbacks or host effects, and escaped closure contracts.
These new proof forms conservatively stop; existing older mechanisms remain.
The whole packages and full gate were not run. No pull request was opened.
