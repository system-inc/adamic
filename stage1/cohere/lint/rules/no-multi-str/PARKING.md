# Rule branch parking on current main

Initial target main: f8013f0baac41ddc340d76f83bddde38536a8f07. Previous pushed rule tip: 6a66e2d3e42854d4385801d032f17cff314f55b7. Replayed all 33 owned commits with `git rebase --onto origin/main 615fe315 codex/lint-wave1-08`. Implementation tip before this evidence commit: dfae1de4b. Shared registration/helper foundation commits are deliberately excluded from this replay; no conflict resolution or production edit was made in shared files.

## Named parking blocker

The branch still needs the shared directory registration, .a module loading, emitted-JavaScript comparison, profile compilation and shared finding model from the harness unification tracked as #zmh9v36. Main's shared lint files do not discover these eighteen directory ports. This names the shared blocker required by Ahra's parking authorization. It does not claim integration through main's default driver. Full requested fixture coverage also retains the parser gaps below, so this report does not assert that the shared harness is the only remaining obstacle. No new helper is claimed while that stricter landing cap remains unresolved. Once the integration SHA is supplied, rebase and recheck this branch before taking further work.

Fresh comparisons use current-main compiler, runtime and parser plus all eighteen rebased rule directories. The historical registration/harness from 6a66e2d3 is restored only in scratch; the existing owned compatibility patch and comparison adapter are applied through a Go overlay. They serialize Go's actual findings, fixes and suggestions. None of those scratch shared files are committed on this branch. `parking.py` reproduces this arrangement; the historical Git object must remain available, and cohere and the pinned TypeScript checkout must already be initialized.

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/no-multi-str/parking.py --scratch /tmp/wave08-parking-reproduction --typescript /tmp/lint-wave1-08-typescript > /tmp/wave08-parking-reproduction.log 2>&1
```

This landing run assembled the same scratch and ran its component gates directly. Detailed fresh logs are in parking_evidence/. All observations below describe this scratch validation, not an integrated shared harness.

## Scope and exclusions

The original eighteen ports and rule-specific fixtures, supported upstream cases, findings/fixes/suggestions, comparison mutants and throughput measurements remain documented under their six batch evidence directories. This landing check does not repeat throughput measurement and makes no new speed claim. Known JSX and malformed legacy-number parser exclusions remain explicit in upstream logs. No full repository test gate, incoming shared Diagnostic adapter or kind-indexed shared driver is certified here. No shared harness, compiler, lowering or native emission source was changed. No main or area branch was pushed.

## Observed parser failure in the landing run

The initial combined run included the existing `TestWave08Shapes/TSX` gap probe. Its source is `const x = <div className="ml-4" />;`. Source Node exited 70 with `adamic: panic: parser slice expected GreaterThanToken, got Identifier at 15`. That run exited 1 after all six supported upstream batches, corpus comparisons and other shapes passed. The supported gate is re-run with `/^NonTSX$`, retaining the failing log rather than hiding it. No parser or shared harness fix was attempted outside ownership. The older upstream exclusions also remain: 14 JSX and two computed-key recovery cases in batch one, one JSX case in batch two, and six JSX plus three malformed legacy-number cases in batch six. These prevent a claim of byte-for-byte parity over every upstream fixture.

The all-origin helper audit refreshed 540 references and read 19 claim occurrences. Legacy comments ownership in HELPERS.md reserves the apparent 24-consumer candidates; all larger concrete helpers are reserved, leaving four-consumer candidates such as scanNumber, numberWithSuffix and isAngle. Those candidates serve enforce-consistent-class-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. None was claimed or implemented in this landing unit.

## Fresh observed gates

Witnesses PASS in 119.755s: 29,360 identical bytes. Compiler/stage1 corpus: 358 files and 13,862,188 identical bytes. All six supported upstream batches passed: 208 + 183 + 133 + 212 + 76 + 74 = 886 source/rule/options cases. Initial combined run exited 1 in 619.806s solely because the explicit TSX shape gap was included.

All eighteen owned semantic mutants PASS in 436.021s; each compiled and exited zero with empty stderr before output comparison caught it on source Node, emitted JavaScript and ASan/UBSan native:

- `default_clause_suppressed`
- `core_default_suppressed`
- `loop_direction_reversed`
- `module_declaration_suppressed`
- `constructor_return_suppressed`
- `delete_identifier_suppressed`
- `null_inequality_only`
- `multiline_suppressed`
- `decimal_suggestion_wrong`
- `octal_suppressed`
- `computed_replacement_wrong`
- `foreign_read_suppressed`
- `gating_invalid_suppressed`
- `direction_escape_removed`
- `optional_parameter_ignored`
- `constraint_suggestion_wrong`
- `const_append_wrong`
- `enum_second_suggestion_wrong`

Directory registry PASS in 0.114s, including rejection mutants. Whole-repository `go vet ./...` on the actual rebased rule tree exited zero with empty log after linking the existing cohere checkout into that isolated worktree. Filtered uncached external `TestTheOracleCatchesOneByte` PASS in 0.495s, zero native/Node cache hits. The first vet/oracle attempt lacked the worktree submodule and failed before testing; the submodule link resolved that environment issue, with no tracked source edit.

Supported combined uncached gate exited zero. Command: `go test -overlay=/tmp/wave08-park-overlay/overlay.json ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestOriginalUpstream|TestWave08UpstreamSupported|TestThirdUpstream|TestFourthUpstream|TestFifthUpstream|TestSixthUpstream|TestWave08Shapes|TestThirdShapes|TestFifthShapes|TestCompilerAndStage1Agree)$/^NonTSX$' -count=1 -v -timeout=20m`, with ADAMIC_GATE_UNCACHED=1 and ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript. See supported.log for total timing and every equal-byte count.

## Refreshed current main c01907a70

Current target: c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Main advanced while tests ran. Rebasing the owned rule history again replayed cleanly; pre-evidence tip 9f373b306f29cd473639310d37ae9acd5c18d499. `git diff --name-only f8013f0ba origin/main -- cmd internal stage1 oracle go.mod go.sum go.work` lists only internal/oracle/stage3_hook_test.go. Other additions are Stage 3 and documentation. Compiler, runtime, parser, owned rule programs and Stage 1 corpus are identical to the supported 399.189s gate and eighteen-mutant run above; those broader results retain that evidence boundary rather than being called new post-rebase runs.

Refreshed the scratch with current-main additions, then reran uncached `go test -overlay=/tmp/wave08-park-overlay/overlay.json ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m`: PASS in 24.246s, 29,308 identical bytes on Go, Node, emitted JavaScript and sanitized native. Temporary filenames account for differing byte totals between runs. Actual rebased tree `go vet ./...` exited zero with empty c019_tree-vet.log. Uncached filtered external oracle PASS in 0.387s with zero native/Node cache hits. Unoverlaid scratch vet failed on historical profile_test.go's portFiles variable/function mismatch; this is precisely why the reproducible runner uses the owned compatibility overlay. That scratch failure is retained as c019_vet.log, without changing shared source.

The owned helper branch also rebased to this main and freshly passed all four helpers and mutants in 19.617s; vet exited zero and external oracle passed in 0.501s. Its pushed evidence tip is 6982efe8d5993dd549c42a1651b22b919bdc75e8. No fresh claim or implementation. Both branches are pushed only to their own names; parser coverage remains the reason not to assert the harness-only parking exception.
