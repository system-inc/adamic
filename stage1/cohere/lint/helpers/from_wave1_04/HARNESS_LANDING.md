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
