Built: all seven owned helpers rebased onto the landed harness area; no new helper claimed.
Commits: origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e includes main 39638d9e278d38bb5aeae887f46d55a70e47aaad; rebased helper tip f7374c9c629f389888c411577338684475c49e65 before this evidence commit.
Commands and outputs: complete owned helper gate PASS 191.671s, owned vet PASS; setup PASS 92s, nproc 5.
Mutants: twenty compiling semantic mutants caught by actual Go comparisons on Node, emitted JavaScript and ASan/UBSan native; names and witnesses in the seven reports and evidence/landed-helper-tests.log.
Not covered: rule-branch shared certification is blocked by runtime option RegExp lowering and the separate Tailwind provider; no new rule/helper claimed and no shared harness files changed.

The area rebase dropped seven already-integrated helper/inventory foundation commits
and applied all nineteen owned helper commits cleanly. The owned helper sources
are unchanged. All result and callback-trace comparisons reran on the current
compiler. The area includes current main and the new shared runtime profile.
The expected allocator leak-test integration was accepted as inherited area code;
no shared test or compiler diff was reverted.

source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m
> /tmp/wave104-landed-helper-tests.log 2>&1: exit zero, PASS 191.671s.
go vet ./stage1/cohere/lint/helpers/from_wave1_04: exit zero, empty log.
bash cloud/setup.sh: Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s,
build cache warm 92s, done 92s on 5 processors; cgroup quota 4 cores.

Seven helpers remove 32 listed prerequisite edges across six distinct Tailwind
rules. The five collapse helpers serve class-order, shorthand, conflicting and
unknown-class rules. Project root and design-system wrapper also serve canonical
classes and variant order. No additional rule becomes completely helper-ready.
Existing report boundaries (opaque loaders, explicit dependencies and consumer
fixture input-domain replays) remain unchanged. This is helper parity evidence,
not a whole CSS-engine or whole-rule native certification.

Landing refresh, 2026-10-07: both origin heads advanced. This branch rebased
cleanly onto area b84a9d9314b65d3d0261ee017e233287b4f071da, including main
c7991b900362796aefd111474e65eb5398e91953. The ledger has no changes.
The full owned seven-helper gate passes again in 207.960s, with all twenty
compiling semantic mutants caught by actual Go comparisons on source Node,
emitted JavaScript and ASan/UBSan native. Owned helper vet passes.
Commands: source /workspace/adamic-tools/env.sh; go test
./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m; go vet
./stage1/cohere/lint/helpers/from_wave1_04. Evidence: newarea-helper-tests.log
and newarea-helper-vet.log. The temporary worktree used the initialized pinned
cohere submodule from the primary worktree; its gitlink is unchanged.
The related rule branch is pushed at de97588de1fd8370cf0d9252d1df0d51f23a2b3f.
Its core oracle passes on this base, but dynamic option RegExp lowering still
refuses the Tailwind surface. No additional helper or rule claimed. The full
repository gate and its seventeen required correctness checks were not run.

Registry-only harness refresh onto area b46914832d70e00847d82d5d221ab7bb24040c53:
Main remains c7991b900. Ledger unchanged, helper rebase clean. Full owned
seven-helper actual-Go gate PASS 204.771s; twenty compiling semantic mutants
caught on source Node, emitted JavaScript and ASan/UBSan native. Helper vet
passes. Commands: go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1
-v -timeout=10m; go vet ./stage1/cohere/lint/helpers/from_wave1_04, with the
existing toolchain environment. Evidence b469-helper-tests.log and
b469-helper-vet.log. Helpers still remove 32 prerequisite edges across six
Tailwind rules; no additional fully helper-ready rule is claimed.
Related rules pushed at 2b3f09b3e4623da8c4cbcb94bf7069b8857dcde2: core green,
Tailwind blocked by runtime RegExp lowering. No new claim and no full gate.
