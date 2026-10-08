# Wave19 checker integration

Base: origin/area/stage1-lint 9b7547976. The outstanding codex/typeaware-wave-19 branch was merged with this area before port work. codex/lint-port-no-script-url is already an ancestor of area and needed no landing merge. No history was rewritten.

## Claimed rules

| Rule | Status | Evidence |
| --- | --- | --- |
| @typescript-eslint/consistent-generic-constructors | Blocked: live archive rejects isolated-declarations | blocked/typescript-consistent-generic-constructors.md |
| @typescript-eslint/dot-notation | Missing property declaration modifiers | blocked/typescript-dot-notation.md |
| @typescript-eslint/no-array-constructor | Missing shared resolvesToAGlobal helper | blocked/typescript-no-array-constructor.md |
| symbol-description | Missing shared resolvesToAGlobal helper | blocked/symbol-description.md |
| valid-typeof | Missing shared identifierIsShadowed helper | blocked/valid-typeof.md |
| nexus/correctness-no-uncleared-race-timeout | Missing declaration ancestry and global augmentation flags | blocked/nexus-correctness-no-uncleared-race-timeout.md |
| nexus/correctness-no-process-exit-after-output | Missing resolved signature declaration metadata and ancestry | blocked/nexus-correctness-no-process-exit-after-output.md |
| nexus/correctness-require-blocking-standard-streams | Missing program source/module graph | blocked/nexus-correctness-require-blocking-standard-streams.md |
| require-await | Missing declared generic signature target/substitutions | blocked/require-await.md |

Each blocker note gives an exact upstream symbol and file:line, a minimal source, and the absent shared API. These eight have no descriptor or private workaround. React reservations in the old claim were superseded by wave16 and remain excluded.

The proposed, unregistered rule uses checker.askFile for isolatedDeclarations, checker.ask for node-symbol-origin, shared Frames, shared forFile comments, shared reporting and multi-edit fixes. Both constructor and type-annotation option modes are implemented. The Go adapter passes the upstream options struct unchanged. The upstream prefix TestConsistentGenericConstructors includes every upstream test function. Raw witnesses cover both modes, a firing default, and Unicode comment preservation. The shared comment helper's documented uncached path is used; no private comment scanner or cache is added.

The compiling mutant removes the moved type arguments from the repair. It must be caught by byte comparison rather than compilation. No checker or shared-helper file is changed.

## Validation and environment

Initial setup raced a branch switch and failed reading declaration_ancestry.go. The first parity build also failed before any test with no space left on device. Removed only generated ELF binaries in the worker's old scratch, retaining all sources/evidence, freeing 3,232,261,993 bytes. Stable-checkout setup passed: go build 55.525s, cache warm 55.804s, total 55.854s, nproc 5 and quota 4 CPUs. Temporary test files use /tmp/adamic-gate, on the separate 8.8GB temporary filesystem.

A focused compilation found unsupported lastIndexOf with a position argument. The rule now uses Go cohere's own delimiter loops; no compiler patch was needed. Setup, registry generation, vet and test output are retained in evidence. The full package command sets all corpus, profile and optional throughput inputs. Any remaining named shared dependency skip is reported rather than bypassed.

## Live boundary found

No claimed rule is ready to register on area. The proposed generic module lowered and compiled sanitized native, but TestRulesAgree failed on its first typed case: unsupported checker question isolated-declarations, exit 70. Matched generic upstream cases: 0; mutant not run. The first three old wave19 question registrations exist only on the unlanded branch. The proposed generic descriptor is nested in generic-reference and excluded from discovery. This directory owns only documentation/reference evidence; no lint behavior is changed.

## All-input landing run

The merge checkpoint is not green. The entire lint package was attempted once on codex/typeaware-wave-19, with the TypeScript checkout pinned and clean, ADAMIC_LINT_BENCH=1, and both profile inputs pointing to one fresh directory. The command and exact environment are in evidence/inputs.json. Its cold caches use /tmp to avoid workspace disk exhaustion.

Result: package build FAIL (exit 1), wall 232.935866s. Package counts: 0 pass, 1 fail, 0 skip. Test counts: 0 pass, 0 fail, 0 skip, because no test ran. No check was skipped or relaxed. nproc 5; sampled one-minute load minimum/median/maximum are in evidence/summary.json. New rule upstream cases matched: 0. New mutants executed: 0; there is no mutant log line to claim.

The old branch's checker files fail against the current cohere shim. Exact examples: bridge/tsgo/checker/declaration_facts.go:22 calls source.Path(), which no longer exists; facts.go:290 passes RootedFilePath to out.text(string); wave19_program_modules.go:76 passes ResolvedFileName to GetSourceFileForResolvedModule, which now expects *module.ResolvedModule. wave19_resolved_callee.go:37 has the same file-name mismatch. evidence/lint.jsonl preserves every emitted compiler diagnostic. These files were not edited under the rule-directory-only scope.

No claimed rule is registered by this checkpoint. The prepared generic code remains reference only and is not a passing port. Exact per-rule blockers are in blocked/. The area-based live reproducer and this merged-branch build failure are distinct observations. The former proves a missing area question; the latter stops a green landing gate before tests.
