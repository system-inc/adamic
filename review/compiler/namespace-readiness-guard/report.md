Built task #agwccbw: a behavioral guard for namespace readiness survivor u038 M14, with no IR-only fallback.
Base: origin/compiler/after-chain a7edbde53b4ae2dfe93e8e1c721946d89d373155; delivery commit reported with the push.
The shared Node agreement test passes; independent source Node prints TypeError, ready:7, continued and exits 0.
Audit M14 makes backend stdout ReferenceError, ready:7, continued; the new test fails by stdout disagreement and the mutant is restored.
Not covered: native execution, full packages, or the full gate; no oracle fixture or counts.md change.

A mutable function alias (`let indirect = read`) is an unresolved edge in namespaceCallable. Preflight cannot follow it to refuse the early call, so the runtime namespace forwarder must enforce readiness. Before Pending's namespace body executes, read() attempts Pending.value; after the catch, namespace initialization runs and the same alias returns 7. The program prints the caught error's name and demonstrates continued execution.

Node's early error is TypeError, not ReferenceError: the TypeScript namespace's runtime object is still undefined. In the unmutated lowered program, the namespace readiness check throws that TypeError before the member's own temporal-dead-zone check. M14 starts the namespace readiness flag true, bypassing the outer check and exposing the member's ReferenceError instead. Both runs exit 0 after catching the error, so stdout comparison is what kills M14. This is an observable counterexample, not an inference that the flag is dead.

The original audit patch is copied verbatim from test-audit/internal-lower-namespaces d0f3ad16, review/test-audit/internal-lower-namespaces/M14.diff. Its line-400 path is line 426 on the after-chain base. The new test carries no extra IR assertions: it calls lowersAndAgreesWithNode directly. The source witness is saved under review, not added to the oracle corpus.

Commands and observations:

- Fetch explicit remote-tracking refs for compiler/after-chain and test-audit/internal-lower-namespaces, then checkout -b compiler/namespace-readiness-guard origin/compiler/after-chain. This checkout otherwise fetched branch tips into FETCH_HEAD only.
- `source /workspace/adamic-tools/env.sh`: reuse the established toolchain; GOPROXY is https://proxy.golang.org|direct, nproc 5 with four-CPU quota.
- `timeout 100 go test ./internal/lower -run '^TestNamespaceForwarderReadinessBeforeInitialization$' -count=1 -timeout 90s -v`: initial PASS, leaf 0.14s; final PASS, leaf 0.24s, package 0.280s.
- `timeout 125 python3 review/compiler/namespace-readiness-guard/replay.py`: apply M14, focused test exit 1, leaf 0.13s, backend stdout ReferenceError versus source TypeError; finally reverse the patch. Full command/status in result.json, output in M14.log.
- `timeout 15 node --disable-warning=ExperimentalWarning oracle/node.mjs review/compiler/namespace-readiness-guard/witness.a`: stdout TypeError then ready:7 then continued, exit 0, recorded in source-node.log.
- `timeout 100 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s -v`: PASS before push; complete output in call-targets.log.

The assigned after-chain dependency is intentionally retained; this unit does not merge or rebase onto main, whose panic contract cannot support the caught-error witness. Integration lane checks are scoped to the unit against its prescribed a7edbde5 base, rather than treating thousands of inherited after-chain changes as this test-only unit.
