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

Parked under system_adamic's explicit shared-harness exception. On current main f8013f0b, the preserved TestRulesAgree gate passes in 33.891s (evidence/parking-main.log). Owned directory registration remains the separately named red TestOwnedWitnesses gate, awaiting the single harness owner on #zmh9v36. No shared bridge or option-text edits are made. Numeric declaration parity and its twelve clean-executing mutants remain recorded in SPEED_REPORT.md; rule behavior validation remains bounded as in the original reports. This parking status does not claim whole-rule integration parity. Resume landing immediately when the harness commit is named, before taking further work.

Parking refresh on origin/main c01907a70: clean rebase, no new shared-file hunks. Preserved main TestRulesAgree PASS 30.010s, 751,320 equal bytes. Named shared registration blocker remains: TestOwnedWitnesses FAIL 2.550s with require-description witness missing findings. Both logs are retained under evidence/parking-c01907a7-*.log. No shared bridge or option-text edit. Incoming developer-tools changes remain intact. Continue parked under the explicit exception, awaiting the named harness SHA on #zmh9v36. No fresh whole-owned-rule parity is asserted.


## Named harness refresh, 2026-10-07

Replayed owned commits onto ab70f38d47de1d4974082b38f84a56af2368b7af,
then merged current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06.
Both commits are ancestors. Only codex/lint-wave1-02-land is pushed.
The original wave branch remains historical; helper delivery stays at
38392a1ca1e8a567173ea5965b85dff496b8f94a. No new helper claim.

Shared-file conflict hunks changed: zero relative to the named harness.
Superseded registration foundation commits were skipped. At the old
reconciliation commit, the owner's exact versions of .gitignore, CLAUDE.md,
context.ts, settings.ts, main.ts, lint_test.go, registration_test.go,
registry/registry.go, registry/registry_test.go and docs/lint-registration.md
were restored. Subsequent main changes merged cleanly. No main rule was
removed. The incoming allocator and compiler changes were retained.

Owned compatibility edits: Fix now extends shared SuggestionEdit; owned
Suggestion extends shared Suggestion and retains its private fixes alias;
SuggestedFinding fills the inherited suggestions array instead of redeclaring
it. The private driver reads edits. Twelve entry modules declare syntaxKinds
rather than using an unsupported forwarding export. The owned verifier now
preserves that declaration and is idempotent. Existing .ts entries use the
explicit temporary allowance; all new Adamic modules remain .a.

Observed gates, after sourcing /workspace/adamic-tools/env.sh:

- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1:
  FAIL 10.272s. Restricted-types witness has zero Go findings because this
  rule deliberately bans nothing by default and the witness gate supplies no
  configured Types. No runtime comparison was reached.
- go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1:
  final FAIL 47.290s, after capturing 2,760 upstream combinations and compiling
  sanitized native. Go oracle panics decoding an all-row: require-description
  receives another rule's allowLoop option and rejects that unknown field.
  Shared oracle option isolation is needed before comparison can run.
- python3 stage1/cohere/lint/rules/id-length/verify-numeric-listeners.py:
  final PASS, 136 equal bytes against actual Go kinds on source Node, emitted
  JavaScript and sanitized native. Each of twelve first-kind-plus-one mutants
  compiled and ran successfully; only byte comparison caught all twelve on
  all three backends. This proves declarations, not whole-rule behavior.
- git diff --check passes. nproc=5. Toolchain setup was already completed for
  this main at 100s total/cache warm; its log remains in the helper evidence.

Earlier failures are retained: initial suggestion field incompatibility;
unsupported forwarding export; and verifier interference that temporarily
restored forwarding exports during a concurrent baseline run. The verifier
was fixed, then both checks were rerun sequentially. None counts as a mutant.

Remaining owned integration boundaries also include the explicit private
CFG-end adapter for array-callback-return, multi-edit automatic fixes and the
old detailed-reporting guard. These have not been certified with the unified
shared driver. No silent fallback was added. No fresh whole-rule parity,
throughput, semantic-mutant green, full repository gate or landing-ready green
is asserted. The landing cap prevents further helper work until the named
shared gates run successfully. Shared harness files remain untouched.


Current-main refresh: origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 integrated with a clean merge, retaining the named harness ancestry. Only incoming inherited-static-field compiler changes were added; no shared lint hunks changed. TestOwnedWitnesses rerun FAIL 3.168s: restricted-types default witness still has zero findings. Shared option-aware witness support remains needed. No new claim, full-rule green, semantic mutant or throughput is asserted.


## Deduplication ledger refresh, 2026-10-07

Integrated named harness 41eb6eab2b6de45ede0a40250765be295ee25fbd by clean
merge, retaining current main b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Read the ledger before reconciliation. The original wave1-02 winning lineage
is retained on its landing successor, not treated as a new competing port.
Historical original branch is not an additional integration request.

Retained six registered owned ports: array-callback-return,
base/consistency-no-bare-throw, dot-notation, grouped-accessor-pairs, id-length
and @typescript-eslint/no-restricted-types. Retired six losing registrations,
entry implementations, Go adapters and mutant descriptors:

- require-description: winner wave1-05
- Tailwind physical direction: winner wave1-05
- non-null assertion and asserted optional chain: winner wave1-05
- no-this-alias: winner wave1-07
- arrow-body-style: winner wave1-10

Private utilities needed by winning rules, archived witnesses and historical
reports now live beneath id-length/retired-support/. Only owned relative
imports changed to preserve those dependencies. No winning rule from another
worker was deleted or copied over. Historical standalone validation scripts
for retired ports are archival and are not current commands. Named rule.json
kinds remain the registry subscriptions; historical numeric evidence is not
an assertion about current dispatch. No shared lint file hunk changed relative
to 41eb6eab2, including finding, context, main, registry, oracle and comparison.
Incoming batch8 retirement and developer-tools changes were retained.

Commands and observations:

- go test ./stage1/cohere/lint/registry -count=1: PASS 0.067s after moving
  archived support directories beneath an owned descriptor. The initial run
  explicitly rejected leftover descriptorless top-level directories and is
  preserved, not counted as green. Registry negative metadata controls ran;
  no new rule semantic mutant is claimed.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1:
  FAIL 6.146s, still at restricted-types' zero-finding default witness. This
  rule bans nothing by default; shared option-aware witness support is needed.
- go run ./cmd/adamic build stage1/cohere/lint/main.ts
  -o /tmp/wave02-ledger-driver --sanitize: exit 0, empty compiler log.
  This checks compilation after relocation, not runtime findings parity.
- git diff --check: PASS.

All output logs are evidence/harness-41eb6eab-*.log. No whole-owned-rule
runtime comparison, fresh semantic mutant or throughput is asserted. The
remaining CFG-end adapter and automatic-fix integration boundaries remain
uncertified. The helper branch dfa823098 remains green on current main with
its four helpers and twelve semantic mutants; it did not need a new rebase.
No helper or batch-only rule is claimed. The ledger names source batches for
35 batch-only rules but does not assign any specifically to wave1-02. New
work remains behind the named landing blocker. Main and area branches were
not pushed.


Landed-area refresh: rebased all forty own commits cleanly onto origin/area/stage1-lint 7481e0324e34a2537aafa9db7eeacda50405611b, including landed harness 50a5f105 and current main 39638d9e278d38bb5aeae887f46d55a70e47aaad. Ledger is byte-identical to the previously reconciled 41eb6eab2 version; the same six winning registered ports remain and six losing ports stay retired. Zero shared lint hunks against area. Registry PASS 0.077s; TestOwnedWitnesses FAIL 2.797s at restricted-types default witness with zero Go findings. Option-aware shared witness support remains absent after landing. Logs: evidence/area-7481e032-*.log. No new semantic mutant or full-rule runtime parity is asserted, no new helper or rule claimed.


Latest-area refresh: rebased forty-one own commits cleanly onto origin/area/stage1-lint d65a8f931c98655936ae04c6899f38f14862b73e. Incoming runtime search, equality and release performance changes retained unchanged. Ledger and shared witness harness are unchanged; the six winning ports remain, losing ports stay retired. Zero shared lint hunks. Registry PASS 0.065s; TestOwnedWitnesses FAIL 2.860s at restricted-types default witness with zero Go findings. Logs: evidence/area-d65a8f93-*.log. No runtime parity or semantic-mutant green is inferred from the metadata gate. Existing helper branch 0e346e5b4 remains based on unchanged origin/main 39638d9e2 with its last successful twelve-mutant verification; no new helper is claimed while the rule landing gate remains red.


Current-area refresh: forty-two owned commits rebased cleanly onto origin/area/stage1-lint b84a9d9314b65d3d0261ee017e233287b4f071da, containing origin/main c7991b900362796aefd111474e65eb5398e91953. Ledger and witness harness unchanged; six winners retained, six losers remain retired. Zero shared lint hunks. Registry PASS 0.066s; TestOwnedWitnesses FAIL 3.034s at restricted-types default witness with zero Go findings. Logs: evidence/area-b84a9d93-*.log. No new rule runtime comparison or semantic mutant is asserted. No new claims or relaxed correctness checks.


Legacy-registry refresh: rebased forty-three owned commits cleanly onto origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53. Incoming legacy rule migrations and handed-node runner changes retained exactly. Ledger and shared witness test unchanged; six winning owned ports remain and six losing ports stay retired. Zero shared lint hunks. Registry PASS 0.108s; TestOwnedWitnesses FAIL 6.176s at restricted-types default witness with zero Go findings. Logs: evidence/area-b4691483-*.log. No successful owned runtime parity or fresh rule semantic mutant is asserted. Helper delivery de010f220 remains green on unchanged main c7991b900; no helper rebase or additional test was necessary. No new claims, skipped correctness checks or relaxed gates; full required-input gate remains unverified.


Typeof/null main refresh: clean rebase onto origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, containing origin/main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Ledger and witness gate unchanged; six winners retained and six losing copies retired. Zero shared lint hunks. Registry PASS 0.091s; TestOwnedWitnesses FAIL 2.882s at restricted-types default witness with zero Go findings. Logs: evidence/area-d3a37422-*.log. Incoming typeof/null compiler fixes retained. No new rule runtime comparison or semantic mutant asserted; no new claim or relaxed correctness check.
