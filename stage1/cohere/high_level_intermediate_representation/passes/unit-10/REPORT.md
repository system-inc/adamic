# Unit 10 stopped: preservation input cannot replay

Base: `9addb0e8a48b1cf60621ae24ef2196c860ff33ce`. Changes stay in this directory. No shared graph, replay, arena, census, compiler or Go production source is changed.

Implemented `CompareManualMemoDependencies`, disagreement merge and result messages, with messages verbatim from Go. The tagged adapter overlays beside pinned Go HIR and writes state immediately before and after comparison. The input runner imports `../../replay/index.ts`, uses its checked PayloadIndex parser and the shared function-owned IdentifierIndex handles. No analysis implementation is copied. SSA is referenced through the existing HIR types; mutation_aliasing is not needed by this comparison and is not invented.

## Certificate

| Coverage | Sanitized native | Node source | Emitted JavaScript |
| --- | ---: | ---: | ---: |
| Preservation census | 0 / 1,465 | 0 / 1,465 | 0 / 1,465 |
| Flow census | 0 / 23 | 0 / 23 | 0 / 23 |
| Comparison probes | 46 / 46 | 46 / 46 | 46 / 46 |

The 46 probes are 20 root/path/optional/ref cases, all 25 result merges, and the unknown-result message. These are development probes, not a sample counted as the full census. The validator, exact diagnostics, all pipeline checkpoints and composition are not certified or implemented here. No census case or error path was counted as a success. See CERTIFICATE.json and validation/tests.txt.

`accept-missing-manual-dependency` changes the final subpath disagreement to success. It builds and exits successfully; only byte comparison catches the wrong `shallower` finding result. Caught on Node, emitted JavaScript and ASan/UBSan native. No expected output is used by the implementation.

## Shared input blocker and request to hir-01

`TestScopeReplayBlocked`, comparison_test.go:154: actual Go `BuildReactiveScopeTerminals` emits a `Scope` terminal for `function f(){return[]}`. The literal graph uses no checker. `testdata/scope.graph.txt` is its complete Go-produced graph, not a fabricated scope witness. Node, emitted JavaScript and sanitized native all stop at `unknown terminal Scope` in `../../replay/decode.ts:134`. `../../core.ts:89` excludes Scope from TerminalType.

Request: add the central Scope terminal, function-local checked ScopeIndex identity mapping, its block/fallthrough/order fields, and matching replay decode/encode/dump/copy/visitor support. Go's record is `cohere/internal/lint/ecmascript/high_level_intermediate_representation/terminal.go:324`; its producer is `scope_terminals.go:335`. Unit 10's source graph still contains that terminal immediately before `ValidatePreservedManualMemoizationWithPruned` (`preserve_manual_memoization.go:107`), so dropping it or replaying construction instead would falsify the input.

The validator additionally needs the public schemas owned by units 6, 7 and 8: ReactiveScopes/ScopeIdentity, ScopeDependencies including its temporaries registry, and the complete ReactiveFunction tree and sequence values. Please relay those schema exports to their owners; unit 10 will write its own before/after sidecars. Full `AnalyzePreservedManualMemoization` (`preserve_manual_memoization.go:439`) needs the actual public pass entry points for units 3–9 and the mutation_aliasing module imported by reference. Those modules are absent on this base. No names or algorithms are guessed. Final root entry/cache exports remain hir-01's responsibility.

No compiler language gap was observed in the comparison. The Scope selector test is intentionally expected to fail when the owner closes that gap, prompting the full validator census rather than preserving a skip. The stopped state is not a green preservation lane.

## Commands and setup

`source /workspace/adamic-tools/env.sh`; `go test -json -count=1 -timeout=20m ./stage1/cohere/high_level_intermediate_representation/passes/unit-10`: 2 pass, 0 fail, 0 skip, package time 56.045s. Both top-level tests and both tagged Go tests call t.Parallel. Owned `go vet` and `gofmt -l` pass with empty output. Full repository and full HIR census gates were not run because the required validator input fails in the shared decoder.

Initial `bash cloud/setup.sh` failed: `fatal: cannot create directory at 'tsc/testdata/baselines/reference/fourslash': No space left on device`; the pinned TypeScript submodule could not finish checkout. Cleared disposable Go cache with `go clean -cache`, completed only that fresh submodule's pinned checkout, and retried setup. Success: Go ready 0.021s, Node 0.025s, submodules 0.073s, markdown 0.074s, clang 0.188s, build 217.203s, complete 217.333s; nproc 5 (CPU quota 4). Timing lines are in validation/setup.txt; the initial log is /tmp/hir-unit-10-setup.log.

The final fetch moved origin/stage1-hir/wip to ancestor 48c45f879 (zero commits ahead of this assigned base), not to a new framing commit. This lane retains the explicitly assigned 9addb0e8 base; no rollback or rebase was performed.
