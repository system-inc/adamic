Built TestViewMixedUnionSelectsNumber for successful JavaScript member selection.
Branch compiler/javascript-guards, base origin/main e2492670b06a4dce1deafe837158ce0366cf2bc4; test-only changes.
Focused existing oracles passed under M4; the new restored leaf passed in 0.06s.
Exact audit M4 failed the new leaf in 0.07s: exit 0, stdout "1\n", expected "0\n".
No compiler change, whole-package run, native fixture, counts row or additional union shape was added.

Task #5k0pffq. Audit evidence: test-audit/internal-javascript ef1f7e69, review/test-audit/internal-javascript/M4.diff and M4-witness-before.mjs. The mutation changes internal/javascript/view_unions_mixed.go's successful return from index to index + 1. The exact diff is saved as M4.diff here. Production was restored before committing.

First checked existing coverage with M4 applied:
ADAMIC_GATE_UNCACHED=1 timeout 100 go test ./internal/oracle -run '^Test(CheckedViewObjectPrimitiveSource|CheckedViewUntaggedSourceDispatch|CheckedViewV2MembershipMutant|NarrowedUnionMemberCheck)$' -count=1 -v -timeout 90s
All four top-level tests passed; package binary elapsed 3.884s, no skips. Full output: existing-oracles-M4.log.
Source search finds no successful production call through the selector outside its package tests. The only other selector call is the untagged runtime's fallback with an empty member list, so it cannot reach the mutated return. Oracle source dispatch uses the separate plain selector. This is the call-graph reason those oracles do not guard this return; it is not a claim that the entire oracle package was run.

The new leaf reproduces the saved witness: number 42, a sole number member, no contract adapter. Source Node's Array.prototype.findIndex independently gives the zero-based index 0. The test pins that Node result, then pins the actual production MixedUnionRuntime selector to the same stdout, empty stderr and exit 0. It uses the existing Node helper, no copied selector implementation or stub.

Commands, each with bounded execution and output in logs:
- timeout 100 go test ./internal/javascript -run '^TestViewMixedUnionSelectsNumber$' -count=1 -v -timeout 90s, M4 applied: exit 1, assertion shows stdout 1 rather than 0, no panic or build failure (M4-new-test.log).
- git apply -R review/compiler/javascript-guards/M4.diff, then the same focused command: exit 0, 0.060s (restored-new-test.log).
- timeout 30 node review/compiler/javascript-guards/number-member-node.mjs: exit 0, stdout 0 (number-member-node.log).
- timeout 90 go vet ./internal/javascript: exit 0, empty output (vet.log).
- Integration lane checks: output in lane-checks.log. The new test calls t.Parallel first and is gofmt formatted.
Counts were not regenerated: no registered oracle fixture or runtime behavior changed.

Setup: GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh passed. Go ready 0.023s, Node 0.022s, submodules 0.059s, markdown dependencies 0.066s, clang 0.155s, go build 39.409s, total 39.639s. nproc 5, CPU quota 4. Full setup output: setup.log.
