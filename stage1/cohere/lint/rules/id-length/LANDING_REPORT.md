Rebased wave1-02 onto main, preserving every existing main rule; twelve entry modules temporarily use .ts.
Commits: original pushed aba098eb; rebased history e048f593; reconciliation 4d1fbd48; base origin/main e8ba3d5d; final landing SHA is reported with the push.
Checks: main oracle PASS 54.748s, 1,272 upstream cases, 751,320 equal bytes; registry PASS 0.043s; owned-rule gate FAIL 2.086s.
Mutants: eleven descriptor rejection controls and the root-only import rewrite mutant were caught; no owned-rule semantic mutant was rerun successfully after rebase.
Not covered: owned-rule integration, emitted JavaScript after rebase, full gate, regex options, CFG and JSX gaps; branch is NOT landing-ready and no new helper is claimed.

Landing branch: `codex/lint-wave1-02-land`. Original branch is retained unchanged. The 38 commits rebased without deleting or replacing any main rule. During conflicts main's complete versions were restored for the shared README, runner, harness, oracle and settings. Later unconflicted foundation edits to main.ts, CLAUDE.md and .gitignore were also restored to main. All original implementations, messages and option decoders in main remain intact.

Only twelve owned rule entry modules were renamed from rule.a to rule.ts, following Ahra's temporary allowance until the shared harness loads .a. Their bytes are unchanged. Each owned mutant's file path was updated. The font-policy driver import was updated to the renamed restricted-types entry. Supporting Adamic modules remain .a.

Every shared-file hunk applied during this reconciliation is in [landing-shared-hunks.patch](evidence/landing-shared-hunks.patch). There are exactly four hunks, all in lint_test.go:

- Import regexp for the existing nested-module import rewrite helper.
- Import the existing registry package for descriptor validation.
- Guard mutant replacement with from != "", permitting the existing tests to request a plain scratch copy without manufacturing an empty-string mutation.
- Append prepareRegistry, portImport and rewritePortImports from the original registration foundation. No existing main test, fixture list, decoder, dispatch or oracle logic is replaced.

The shared runner, Go oracle, README, entry point, settings, CLAUDE.md and .gitignore have zero final hunks against the pinned main. The inherited context, registration tests, generator, generator tests, command and registration documentation are unchanged foundation additions from the original branch, not new edits in this reconciliation; their complete diff is recorded separately in [landing-inherited-foundation.patch](evidence/landing-inherited-foundation.patch). Other inherited registration evidence stays historical.

Commands and observations:

- bash cloud/setup.sh initially failed the compile-only cache-warm step because prepareRegistry and rewritePortImports were undefined after keeping main's harness. Initial tool and submodule timing lines were all 0s. The final retry passed: Go, clang, Node and submodules 0s each; cache warm 15s; total 15s. Go 1.27.1, clang 20.1.8, Node 24.19.0. nproc=5; cpu.max=400000 100000. Environment sourced from /workspace/adamic-tools/env.sh.
- go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -timeout=15m -v passed after the allowed entry renames, capturing 1,272 upstream cases. Main's existing recovery exclusions/refusals were retained. Go, source Node and sanitized native matched for 751,320 output bytes. This proves the preserved main slice, not the twelve new rules.
- go test ./stage1/cohere/lint/registry -count=1 -v passed in 0.043s, including deterministic regeneration, ten descriptor rejection controls and duplicate oracle adapter rejection. These are metadata validation checks, not successful semantic mutants of the owned rules.
- go test ./stage1/cohere/lint -run '^TestNestedOutsideModuleCopy$' -count=1 -v passed in 0.358s after adding the empty-copy guard. Original nested-module import rewriting works on Node and native. Restoring the old root-only rewrite was caught by ERR_MODULE_NOT_FOUND on Node and native module loading. This is an import-resolution mutant, not a semantic output mutant.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -timeout=15m -v failed in 2.086s: @eslint-community/eslint-comments/require-description witness reports no findings. Main's preserved static Go oracle does not yet select the new directory adapters; the preserved runner also does not yet call the generated listeners. A green main-only result must not be presented as owned-rule parity.

Every run wrote directly to a log. Initial failures and successful retries are retained under evidence/landing-*.log. No full or uncached integration gate was run, and no fresh findings-per-second result is claimed.

A proposed additive bridge would load the owned Go adapters, append their listener selection while excluding legacy duplicates, and connect generated listeners to main's preserved traversal. It would also need the original raw option text for owned decoders. Automatic approval review rejected that command before execution, stating that broad post-rebase edits to the shared runner, settings, Go oracle and harness change dispatch and rule selection beyond the explicit permission for conflict resolution. No part of that rejected command was applied. The risk identified was expanding shared behavior beyond the permitted rebase scope.

The branch is a reviewable rebased snapshot, not green or landing-ready. The WIP cap remains in force. Further shared registration integration needs explicit approval; regex options, CFG and JSX/native adapter boundaries also remain documented in the original reports. No helper or rule claim is added.

Landing refresh: fetched all origin heads and rebased cleanly onto origin/main f8013f0baac41ddc340d76f83bddde38536a8f07. No conflicts and no new shared-file hunks. Rebased implementation snapshot ea59d14bf3ee771e6845dcab282ae4d279aef0c5. The post-rebase TestOwnedWitnesses gate failed again in 2.091s: require-description witness reports no findings. Raw output is evidence/landing-f8013f0b-owned.log. The runner and Go oracle still match main exactly. No successful full-rule oracle or new semantic mutant is asserted; the landing gate remains red and no helper is claimed. The original branch remains historical and is not landing-ready. Only the own -land branch is updated; main and area branches are untouched.
