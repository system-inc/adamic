Built pushJump, popJump and makeUnreachable in separate .a files: twelve dependency edges across four rules.
Commits: claim e4b23dd4 pushed before source; rebasing completed implementation onto advanced main b8fb957a before publication.
Commands: focused Go/source Node/emitted JavaScript/sanitized native PASS 59.403s, 6,997 snapshots; current-main full regression pending.
Mutants: eleven transition mutants and an empty-stack guard mutant compile and are caught; current-main regression pending.
Not covered: complete native CFG, dependency wiring, whole-rule findings/fixes/suggestions or full repository gate.

## Claims and landing

Only codex/lint-helpers-03 is pushed by this worker. All forty-two earlier helpers were complete, oracle-green and pushed at a2519dfd on c01907a before claiming. Every claims file across twenty origin/codex/lint-helpers* branches was checked. The comments bundle remains reserved by shared HELPERS.md. These concrete CFG helpers tie the highest available count at four. Regex-engine internals are not selected for a handwritten matcher. Claim e4b23dd4 precedes source.

A refreshed scan finds later duplicate pushJump/popJump claims on slot 05. Slot 03 e4b23dd4 at 06:00:46 UTC precedes slot 05 85af6690 at 06:02:56 UTC, so slot 03 retains them. makeUnreachable remains unique. During validation main advances from c01907a to b8fb957a, adding inherited static-field reads. The old-base full gate is stopped for rebase; it is not counted as a pass. The user explicitly authorizes rebase onto current main. No main or area branch is pushed.

## Behavior and independent evidence

README.md documents the three contracts. Node/block identities are opaque numeric handles (-1 for nil), not numeric AST kinds. These helpers do no kind dispatch. LabelsOf, block allocation and successor storage remain explicit separately owned dependencies. The private builder stack does not expose Go slice-header aliases; replacement arrays preserve target-object identities, but arbitrary Go backing-store aliases are outside this interface. Barrier presence and its values are separate fields because stage 0 refuses a boolean-array/null field; no compiler change is made.

The oracle uses actual Go CFG Build on every parser root of the captured complete runtime sources. Temporary overlays rename the three helper method definitions only to install before/after wrappers; helper bodies remain unchanged. Actual Go supplies label lists, block handles and before/after snapshots. Native/source/emitted adapters replay each operation from its observed pre-state using explicit dependencies. This proves these state transitions, not native graph traversal or the separately owned dependencies.

The source capture covers all four consumers and 2,119 runtime inputs. Upstream core/react package capture gate exits 0. Controls add nested labels, do/for/switch/break/continue/return/throw/finally roots, stack depths 0..7, successor counts 0..9, both reachability/incoming flags and absent/present barrier arrays. There are 6,997 snapshot lines. Empty-pop rejection is observed separately in real Go; source Node, emitted JavaScript and native panic with exit 70 and the owned invariant message. Its guard mutant compiles, exits 0 without stderr, and prints survived. Invalid ASTs beyond parser-produced consumer roots and arbitrary cyclic graph representations are not covered.

Initial oracle failure: relative fixture filenames caused the Go parser's absolute-path invariant panic. The owned adapter now normalizes filenames. A second failed run recorded stage 0's unsupported boolean[] | null field. The port's explicit presence flag resolves it. evidence/focused.log and focused-retry.log retain these failures; focused-supported.log is the successful transition gate, and empty-guard.log is the successful refusal mutant gate.

## Consumers

Each helper removes one dependency from each of array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Twelve edges, four distinct rules; readiness.json subtracts only this worker's forty-five delivered helpers, never other workers' claims. It lists the remaining dependencies and any final-helper readiness. No native whole-rule parity is claimed.

## Commands

Setup `bash cloud/setup.sh`: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 44s, done 44s. nproc=5, cpu.max=400000 100000, 17.6GB. Go1.27.1, clang20.1.8, Node24.19.0. Source /workspace/adamic-tools/env.sh in each toolchain shell.

Focused transition gate: `go test ./stage1/cohere/lint/helpers/slot03 -run TestBatch15 -count=1 -v -timeout=20m` -> focused-supported.log, PASS 59.403s. Separate `-run TestBatch15EmptyStackRefusal` -> empty-guard.log, PASS 0.743s. Full touched-package gate: `go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m`. Repository vet `go vet ./...`; format `gofmt -l cmd internal stage1/cohere/lint/helpers/slot03`; six uncached input probes `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m`. Each command writes directly to an owned evidence log. Corpus regeneration runs twice; hashes and current-main results are recorded before publication.

## Every new transition mutant

- TestBatch15Mutants/push_jump.a: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:0:false:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#01: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:0:true:true:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#02: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:0:true:false:/labels:1;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#03: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/1:2:0:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/push_jump.a#04: batch15_test.go:125: compiled semantic mutant caught at line 90: got "1/2:1:-1:true:false:/labels:0;" Go "1/2:1:0:true:false:/labels:0;"
- TestBatch15Mutants/pop_jump.a: batch15_test.go:125: compiled semantic mutant caught at line 125: got "0/" Go "1/2:1:6:true:false:/"
- TestBatch15Mutants/pop_jump.a#01: batch15_test.go:125: compiled semantic mutant caught at line 92: got "1/2:1:0:true:false:/" Go "0/"
- TestBatch15Mutants/make_unreachable.a: batch15_test.go:125: compiled semantic mutant caught at line 6770: got "0:false:false/1,/true,|1:false:false//nil" Go "0:false:false/1,/false,|1:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#01: batch15_test.go:125: compiled semantic mutant caught at line 1: got "2:true:true/2,/nil|3:false:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#02: batch15_test.go:125: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|2:true:true/3,/nil" Go "2:true:true/3,/nil|3:false:false//nil"
- TestBatch15Mutants/make_unreachable.a#03: batch15_test.go:125: compiled semantic mutant caught at line 1: got "2:true:true/3,/nil|3:true:false//nil" Go "2:true:true/3,/nil|3:false:false//nil"

Empty-stack guard deletion is caught by the exit-70 refusal comparison; its compiled mutant exits 0 and prints survived. All transition variants must compile and execute with empty stderr before changed snapshots count. Full current-main regression records earlier mutants too.
