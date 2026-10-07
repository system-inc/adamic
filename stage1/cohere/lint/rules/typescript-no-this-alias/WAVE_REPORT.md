Built: ten already-reserved rule candidates, nine with four-way repair comparisons, plus Node-complete empty-object-type.
Commits: one rule per commit on `codex/lint-wave1-09`; exact shas are recorded below after commits.
Commands/results: 757 Go/Node case combinations; 682 also matched emitted JavaScript and sanitized native; registration tests passed.
Mutants: one per rule; nine caught solely by Go comparison on all three execution backends, empty-object-type on Node only.
Not covered: full repository gate, dynamic-RegExp native lowering, shared repair API integration and baseline JSX parser integration; no new claims.

## Observations

The original React reservations are implemented and compared against actual Go cohere listeners. Their source Node, emitted JavaScript and sanitized native use an isolated checkout of `origin/codex/stage1-jsx-lint` at `a8a62d62ca49db7415e14c3887dd305022b17309`, combined with unchanged registration-foundation context/settings/finding files. No parser file was changed in this working branch. Earlier constructor, multi-push and truthiness refusals were resolved by changes to the owned listeners. None of those failed executions counts as a mutant catch. Fragment overlapping fix records are compared in full; overlapping application is explicitly refused.

Seven non-JSX rules were compared over 236 frozen compiler/stage1 source files, and React over 247 files including the dependency source. Compiler checkout: `050880ce59e30b356b686bd3144efe24f875ebc8`. These totals include stage1 `.ts`, `.a` and gap sources; unsupported JSX is excluded only from the baseline description upstream cases (one `<div>` case, documented in its report). Per-rule evidence contains original source/options combinations, complete canonical logs, empty successful stderr logs, digests of corpus inputs/outputs, timings and mutant witnesses.

All sources temporarily remain `.ts`, under Ahra's explicit fallback. Shared registration/harness/compiler files were not edited. Shared helpers could not be cleanly merged earlier; description uses a private, attributed copy of the six comment modules, with only import paths adjusted. Initial claim commits remain unchanged; new work is entirely in owned rule directories.

## Rates

| Rule | Compared cases | Findings/s: sanitized native / Node / Go |
|---|---:|---|
| [@typescript-eslint/no-non-null-asserted-optional-chain](../typescript-no-non-null-asserted-optional-chain/REPORT.md) | 25 | 1.815403 / 9.829085 / 41.312429 |
| [@typescript-eslint/no-non-null-assertion](../typescript-no-non-null-assertion/REPORT.md) | 37 | 274.503400 / 1480.970938 / 6567.943132 |
| [@typescript-eslint/no-this-alias](../typescript-no-this-alias/REPORT.md) | 45 | 0.000000 / 0.000000 / 0.000000 |
| [@typescript-eslint/no-dupe-class-members](../typescript-no-dupe-class-members/REPORT.md) | 31 | 0.000000 / 0.000000 / 0.000000 |
| [@typescript-eslint/no-empty-object-type](../typescript-no-empty-object-type/REPORT.md) | 75 | blocked / 35.329300 / 164.601660 |
| [@typescript-eslint/no-import-type-side-effects](../typescript-no-import-type-side-effects/REPORT.md) | 27 | 0.000000 / 0.000000 / 0.000000 |
| [@eslint-community/eslint-comments/require-description](../eslint-comments-require-description/REPORT.md) | 109 | 4.945473 / 33.824820 / 465.029397 |
| [react/no-invalid-html-attribute](../react-no-invalid-html-attribute/REPORT.md) | 294 | 0.000000 / 0.000000 / 0.000000 |
| [react/jsx-no-useless-fragment](../react-jsx-no-useless-fragment/REPORT.md) | 75 | 0.000000 / 0.000000 / 0.000000 |
| [react/no-unescaped-entities](../react-no-unescaped-entities/REPORT.md) | 39 | 0.000000 / 0.000000 / 0.000000 |

Rates are compiler-corpus counts, best whole-process elapsed time of three runs, including startup. A measured zero means zero findings, with the successful timings available in the linked reports. Sanitized native performance is not release performance. Setup log: Go, clang, Node and submodules 0s; cache 16s; total 16s; `nproc` 5.

## Exact blockers

`go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=15m` fails in 6.789s: `typescript-no-empty-object-type/rule.ts:27:58: stage 0 can't lower RegExp with a nonconstant pattern yet`. Node matches all 75 empty-object-type cases, full suggestions and compiler/stage1 corpus; Node mutant caught. Native/emitted-JS results or native rates are not claimed for that rule. The compiler files reserved by Ahra were not edited to work around it.

The baseline parser rejects JSX before a listener can run; [baseline JSX log](evidence/baseline-jsx.log) records the exact refusal. React comparisons are conditional on the independently published JSX dependency described above, not a claim that the baseline CLI parses JSX.

The legacy shared `Finding` has only one string repair/suggestion channel. Owned `suggestions(...)`/`fixes(...)` retain and compare every Go plan, but do not yet populate the new shared `Finding.suggestions` / edit-range API. `origin/codex/lint-harness-dot-a` became available during the final all-branch fetch at `f4d98cab50048692781da3599131317dc569d466`; integrating that shared change and connecting the owned plans remains required. The shared CLI's missing repair metadata is not certified. No shared generator or harness was edited. No full repository gate was run.

## Other worker overlap

Skipped `@next/next/google-font-display` and the duplicate Tailwind candidate: both already exist on `origin/codex/lint-wave1-01` (`7dedddb30b8a59ce40646903fdd75f7ba3dedae7`) and `origin/codex/lint-wave1-11` (`d316a491394beb55c40198e15553762611f20e8c`). The Tailwind scratch candidate and its four-way evidence remain in `/tmp/wave109-skipped-tailwind-port` and `/tmp/wave109-tailwind`; it is not registered or committed here.

The final all-origin audit also found description and the three non-null/this-alias candidates on other workers' branches. These were already reserved and were preserved here under the request to push the work; integration must select one tested implementation. No further claim was made. Plain `git fetch origin` had a restricted fetch mapping; the audit refreshed every branch with `git fetch origin '+refs/heads/*:refs/remotes/origin/*'`.

## Reproduction

Source environment: `source /workspace/adamic-tools/env.sh`. Run the owned `validate.py` with repeated `--slug` selections, `--work` and the pinned `--compiler` checkout. Redirect output to a `.log` file. Non-JSX command used all seven slugs and `/tmp/wave109-final-all`; React command used all three React slugs in the isolated dependency checkout and `/tmp/wave109-react-final`. Go output is independently produced from the unchanged real rule, never from Adamic. The capture overlay changes only case recording and lossless private-option serialization; it never changes a Go listener or decoder. Snapshotting prevents generated-module changes between backend runs from appearing as language divergences.

`go test ./stage1/cohere/lint/registry -count=1 -v` passed in 0.196s. Its descriptor mutants are in [registry.log](evidence/registry.log); the unit-specific semantic mutants are in each rule's `mutant.json` and evidence/run.log. No execution/type/lowering failure was counted as semantic detection.

## Rule commits

| SHA | Change |
|---|---|
| `dbb16ceb` | Port react/jsx-no-useless-fragment |
| `b6f7fc7a` | Port react/no-invalid-html-attribute |
| `f9dba2df` | Port react/no-unescaped-entities |
| `d6ae307a` | Port @typescript-eslint/no-dupe-class-members |
| `f0175dd1` | Port @typescript-eslint/no-empty-object-type |
| `ac41c084` | Port @typescript-eslint/no-import-type-side-effects |
| `0cff8ac0` | Port @typescript-eslint/no-non-null-asserted-optional-chain |
| `2f7caea6` | Port @typescript-eslint/no-non-null-assertion |
| `cc7e091a` | Port @typescript-eslint/no-this-alias |
| `ec7b05fa` | Port @eslint-community/eslint-comments/require-description |
