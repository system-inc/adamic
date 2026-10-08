Built step 05 checks that follow a known overloaded closure's static result promise, toward roadmap step 16.
Compiler fdb63c7bc8bf42976cfd5c0edd9186c78bf2a272, on main 54cbc125 (already merged).
Focused lower, Node/JavaScript/native sanitizer and release oracles, CLI checked sites, local a-check and records pass.
Dropping the indirect check and skipping single-signature checks both let the planted wrong result through and fail the pinned exit-70 test.
Unknown targets, mixed implementations, mutable callable storage, rest/observable argument counts and generic function values stay stopped; 06 is separate.

The original closure carries its public overload set through factory returns,
unchanged aliases and parameters of closed consumers. The compiler tracks its
known implementation separately from the checker-resolved promise. A call
through a single-signature variable or parameter matching the narrow overload
therefore retains the result.value check, after invoking that original closure
once. The check names the overload, invocation site, field and both result types,
and contributes to the existing checked-site count. No replacement callable is
exposed. Closure carriers precede arguments in the boundary entry to preserve
callee-before-argument evaluation. The current lexical implementation ABI is
selected; ambiguous instantiations retain a stop.

Four TypeScript pairs cover returned values, aliases, module values, narrow
variables, narrow parameters, separate factory captures, identity and evaluation
order. The valid order observation is word then ca. The captures observation is
true, word, word, false, other. Node accepts every planted numeric result and
prints 1; the inserted check instead exits 70 in both backends. Node observations
are in node.json. Eight Adamic refusal pairs name result.value; one positive
Adamic factory has a per-overload return proof and emits no result check. All
successful executions match Node, including sanitized native and release builds,
and pass the leak check. The liar witnesses pin the deliberate stop rather than
compare it with Node's acceptance.

The narrow fixture has two checked source invocations: the narrow local and the
narrow parameter. CLI c, js and build --explain-checks all report both. Each
mutant preserves that report but removes its actual guard, so it is caught by the
runtime exit assertion, not a compile error, count change or sanitizer failure.
Existing overload, structural-result, callback and field-hatch witnesses remain
green. Negative target tests retain stops for mixed parameters and opaque or
mutable storage. No backend emitter, runtime, protected orchestration file or
checker option was changed, and no cohere source was copied.

The census verifies all 82 adapted hashes at source pin 388096e6 and compares
against ../field-hatch/after.jsonl.gz. Its scratch-only overlay is regenerated
from the existing measurement tool and seven-interval scope patch. It emits no
executable. This is a boundary-intersection measurement on a checker-rejected
project, not a claim that the whole TypeScript compiler emits.

See table.md and next-stops.json for the seven regions. 14 reveals another
1,195 bytes, leaving 5,907 hidden. The next boundary inside the region is
utilities.ts:11338:9, reaching skipParentheses at utilities.ts:5041:1. Its
implementation declares Node while that overload promises Expression; the
required result._expressionBrand field is absent. censusOverload refuses that
relation. The previous 8,048-byte reveal in 06 and 324 signature bytes in 13
remain. Total assigned hidden bytes are 48,591. 01 and 13 remain checker stops,
05 waits on checked views, and 06 still requires the ruled visitor group.

Commands, with output saved directly to the accompanying logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/overload-escaped-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
# 5
go test ./internal/lower -run '^(TestOverloadValueStops|TestOverloadStructural|TestOverloadResults|TestOverloadCallback|TestCensusOverload|TestOverloadedShorthand|TestIndirectPredicateOverload)' -count=1 > /tmp/overload-values-final-lower.log 2>&1
# ok, 3.083s
go test ./internal/oracle -run '^(TestOverloadValues|TestOverloadValueProof|TestOverloadFieldHatch|TestPredicateDirectionCountsAreRecorded)$' -count=1 > /tmp/overload-values-final-oracle.log 2>&1
# ok, 1.722s
go test ./cmd/adamic -run '^TestExplainOverloadValueChecks$' -count=1 > /tmp/overload-values-cli.log 2>&1
# ok, 0.300s
python3 docs/overload-results/groups/values/a-check.py /tmp/overload-hatch-fast-gate.py > /tmp/overload-values-acheck.log 2>&1
# PASS: eight refused .a and one proven .a
python3 docs/overload-results/groups/values/run-mutants.py > /tmp/overload-values-mutants.log 2>&1
# both caught, mutant tests exit 1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -args -update-counts > /tmp/overload-values-counts.log 2>&1
# ok, 40.212s; five new rows, no existing numeric counts changed
```

The a-check executor is the previously inspected tooling pin 89cbe74a. Setup
cumulative timings: Go 0.021s, Node 0.022s, submodules 0.069s, markdown 0.072s,
clang 0.174s, Go build 34.857s, deferred tests 35.087s, cache 35.089s, done
35.118s. nproc is 5 and cgroup cpu.max is 400000 100000. A final rebuild initially
failed with no space left on device: the generated Go build cache occupied 30 GB
of the 32 GB filesystem. go clean -cache recovered 28 GB; the final focused
verification, a-check, mutants and records were then rerun successfully. No
source or evidence was removed. No whole package or full gate was run.
