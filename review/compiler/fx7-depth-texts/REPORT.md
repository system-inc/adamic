Built: sound deep untagged views, matching failure kinds and declared type text, plus nine fixtures toward views native and the step 24 stack guard.
Commits: core 7ace5f465a6bb9df2719fe94d3f03096a87eca9b; current-main merge a58836e8dd1782bf0f429a4a609cc101cbf3d8cd; declared-shell control 97979aaff4d272efc4de1a1375e17d5f2738066e.
Checks: targeted views, reader guard, fixtures, counts refresh and integration lane checks passed.
Mutants: all ten source mutants failed their intended semantic assertion; patches and logs accompany this report.
Not covered: other operating systems, WASI, the full gate, or the final JavaScript platform stack threshold beyond 2,000 links.

Native keeps the recursive walker and its existing (contract, reference) path cycle detection. Runtime C can call ADAMIC_CHECK_STACK through adamic.h and the existing stack.c linkage. The old membership-depth cutoff is replaced with that macro, so exhaustion uses the existing exit-70 RangeError panic. No heap-stack conversion was needed. A forced guard and an actual 20,000-link walk under a 512 KiB stack both produced the exact stack-guard panic in release and sanitized native builds. A cyclic runtime object still returns successfully in both native builds; the JavaScript cycle control also terminates.

JavaScript removes its fixed membership bound. Its branching union walk and object field walk use loops instead of callbacks. A defined optional union resolves its only non-undefined arm in the current frame, subject to the original unsupported/nominal checks. Logical depth bookkeeping is retained so restoring the old 128 cutoff is caught at 65 links. Recursion still uses the platform stack. Before optional-frame reduction, the 2,000-link JavaScript fixture hit the platform stack guard: this was a failed comparison, preserved in before-optional-frame-reduction.log. After the reduction, 65, 200 and 2,000 links all match source Node, release native, sanitized native and backend JavaScript. No per-backend divergence remains at these depths.

Native changes only diagnostic classification: tuple storage reports array and static class storage reports function. Membership classification remains unchanged. The existing mixed-union test supplied an int pointer where the runtime contract requires an object; it now supplies an actual adamic_object. Lowering carries the checker's printed declared field type in SetProperty.DeclaredType. Both emitters use it for checked-write diagnostics. JavaScript needed this correction too: its p32/p33 text previously said undefined, while native printed an empty name. Both now name string | number. Representation fallbacks also never produce an empty type name.

The p09, p32, p33, p67 and p72 fixtures are copies of the corresponding existing fxspptb review witnesses, conservatively treated as the requested attachments. Their tests pin complete stderr lines and exit 70 independently, then assert both Adamic backends agree. Source Node decides the successful deep-chain outputs; it is not the text oracle for intentionally wrong views. A further scalar-tuple witness covers the scalar classifier. Cycle controls use runtime objects because Adamic does not admit owning reference cycles. The counts diff adds exactly nine rows and changes no existing row.

Runtime ownership clearance for @system_adamic_runtime, exact final source lines:
- internal/native/runtime/view_unions_untagged.c:94 replaces the depth cutoff with ADAMIC_CHECK_STACK().
- internal/native/runtime/view_unions_mixed.c:69 through 73 adds diagnostic tuple/static-side classification.
- internal/native/runtime/object.c:229 through 233 adds the same classification for scalar view failures.
No other runtime file changed. These are one replacement and ten added lines.

Setup and commands actually run:
- export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh. Go ready 0.072s, Node ready 0.092s, clang ready 0.465s, markdown dependencies ready 1.353s, submodules ready 18.637s, Go build ready 221.091s, build cache warm 221.195s, done 221.226s. nproc=5, cgroup cpu.max=400000 100000. Sourced /workspace/adamic-tools/env.sh in build/test shells.
- timeout 100 go test ./internal/ir -run '^TestCallTargetReaders$' -count=1 -timeout 90s: passed, 15.435s.
- timeout 100 go test ./internal/lower ./internal/native ./internal/javascript -run 'View|Untagged' -count=1 -timeout 90s: passed, 4.240s / 0.346s / 0.183s.
- timeout 100 go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run '^TestCallTargetReaders$|View|Untagged|^TestFX7' -count=1 -timeout 90s -v: passed before and after merging current main. The merged-main log is main-tests.log; all five packages passed.
- timeout 180 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 150s -args -update-counts: passed, 83.032s. The first attempt failed solely because @types/node 25.3.3 was absent. timeout 120 npm ci --prefix stage3/api installed the pinned declarations, exit 0. The initial failure is preserved separately.
- timeout 400 python3 review/compiler/fx7-depth-texts/run-mutants.py: exit 0, all ten caught. Each go test invoked by the runner uses -timeout 90s and an external 100s subprocess bound; sources are restored in finally.
- timeout 100 go test ./internal/native ./internal/ir -run '^TestFX7ViewCycleAndStackGuard$|^TestCallTargetReaders$' -count=1 -timeout 90s -v: passed after using sh, 0.285s / 0.679s.
- git fetch -q origin main devtools/fast-gate cloud/merge-tree: fetched, but origin's configured refspec tracks only main. Explicit remote-tracking refspecs then made origin/devtools/fast-gate and origin/cloud/merge-tree available.
- git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -: passed. Output: lane checks 11.2 s: gofmt and tools on 10 Go files, t.Parallel on 3 test packages; a-check 9 .a files; vet 3 packages. Its first run identified undeclared bash; the test now uses declared sh, with no tool-inventory change.
- git diff --check: passed.

Every added top-level test uses t.Parallel. New leaf durations observed in final-tests.log, with the counts refresh concurrently loading the four-CPU quota:
- TestFX7Depth65: 1.18s; TestFX7Depth200: 1.96s; TestFX7Depth2000: 2.02s.
- TestFX7TextP09: 1.55s; P32: 1.25s; P33: 1.09s; P67: 1.37s; P72: 1.13s; ScalarTuple: 1.29s.
- TestFX7ViewCycleAndStackGuard: 1.61s; TestFX7ViewCycle: 0.16s.

Mutant receipts:
- native-depth-128.patch: TestFX7Depth65 fails with exit 70 instead of the Node result.
- javascript-depth-128.patch: TestFX7Depth65 fails with exit 70 instead of the Node result.
- native-stack-guard.patch: TestFX7ViewCycleAndStackGuard fails because a crossed guard returns normally instead of exit 70.
- union-found-kind.patch: TestFX7TextP67 and TestFX7TextP72 fail the pinned function/array stderr text.
- scalar-found-kind.patch: TestFX7TextScalarTuple fails because native reports object instead of array.
- native-declared-name.patch: TestFX7TextP32 and TestFX7TextP33 fail the pinned declared type; native prints an empty name again.
- javascript-declared-name.patch: TestFX7TextP32 and TestFX7TextP33 fail the pinned declared type; JavaScript prints undefined again.
- javascript-optional-frames.patch: TestFX7Depth2000 fails because JavaScript reaches its platform stack limit.
- native-cycle.patch: TestFX7ViewCycleAndStackGuard fails because the cyclic control is rejected.
- javascript-cycle.patch: TestFX7ViewCycle fails because the cyclic control is rejected.
These are runtime behavior/assertion failures, not compiler warnings or build failures.

Scope limits: Linux with the setup's Go 1.27.1, clang 20.1.8 and Node 24.19.0. No whole-package/full-gate run, no PR, no other worker branch merged, and no copied cohere implementation. The code changes preserve existing checked-write acceptance behavior; this unit repairs its diagnostic text. The final evidence commit and pushed branch head are named in the delivery response.
