Built: `react/jsx-no-useless-fragment`, directory registration, independent Go adapter, full repair-plan comparator and owned witnesses.
Commits: this directory is committed separately on `codex/lint-wave1-09`; see the branch log.
Commands/results: 75 upstream/owned combinations, 24599 identical protocol bytes; 247 frozen compiler/stage1 sources matched.
Mutant: `react-jsx-no-useless-fragment listener suppressed`; clean execution disagreed with Go on Node, JavaScript, sanitized native.
Not covered: full repository gate; production shared-driver repair serialization and dependency integration, as detailed below.

## Comparison

`python3 stage1/cohere/lint/rules/typescript-no-this-alias/validate.py --slug react-jsx-no-useless-fragment --work /tmp/wave109-react-final` with stdout/stderr redirected to `/tmp/wave109-react-final.log`. The source Node oracle uses `oracle/node.mjs`. Native builds use `native.Build` with `Sanitize: true`; the emitted `.mjs` is independently run through Node. The Go adapter invokes the actual upstream rule over a separately parsed original input. All successful processes exit zero with empty stderr. The mutant must also execute successfully: execution/type/lowering failure is never credited as a comparison catch.

The canonical stream preserves finding ids/messages and UTF-8 ranges, every suggestion id/message and ordered edit, automatic fixes and nonoverlapping fixed text. Suggestions remain unapplied. Fragment diagnostics can propose overlapping fixes; those edits are all compared, while application is explicitly refused rather than guessing which wins. This follows cohere's fixture overlap policy. Node-side offsets are independently converted from UTF-16 to bytes.

TypeScript compiler commit: `050880ce59e30b356b686bd3144efe24f875ebc8`. Stage1 sources include `.ts`, `.a` and gap fixtures; no gap exclusion. Inputs are frozen before running any backend. [Corpus hashes](evidence/corpus-hashes.json) preserve input and complete output digests without committing four duplicate copies of the compiler source. [Captured cases](evidence/cases.json), [upstream records](evidence/upstream.jsonl), raw case/mutant/build logs, and [validation counts](evidence/validation.json) are preserved here. The description suite's one unsupported JSX `<div>` case is excluded on the baseline parser and is not counted as passing.

## Throughput

| Backend | Findings | Seconds | Findings/s |
|---|---:|---:|---:|
| Sanitized native | 0 | 4.473705 | 0.000000 |
| Node | 0 | 0.782918 | 0.000000 |
| Go | 0 | 0.163501 | 0.000000 |

Rates cover TypeScript `src/compiler` only, best elapsed time of three whole-process count runs, including startup. Native uses sanitizers. A zero rate means the compiler corpus contains zero findings for this rule; the successful elapsed times are still reported. These are local observations, not release-mode steady-state performance claims.

## Integration limits

All new executable sources temporarily use `.ts`, as Ahra explicitly allowed while the local shared harness still discovers `rule.ts`. Only owned rule directories were edited. No shared generator, harness or reserved compiler file was changed. Rule-owned `suggestions(...)` / `fixes(...)` expose the full plans and the comparator checks them, but the branch's legacy `Finding` does not expose the new complete suggestion/edit API. The source must be connected to the API on `origin/codex/lint-harness-dot-a` (`f4d98cab50048692781da3599131317dc569d466`) during integration; the normal CLI's incomplete repair metadata is NOT certified as matching Go. Its published API is `Finding.suggestions`, `Finding.editStart/editEnd` and `context.report(...)` returning a finding.

Shared `TestRulesAgree` currently stops at dynamic RegExp in `typescript-no-empty-object-type/rule.ts:27`; see the aggregate report. Registration package tests pass. Full repository tests were not run. Setup succeeded: Go/clang/Node/submodules each 0s, cache 16s, total 16s, `nproc` 5.

The baseline parser rejects JSX before these listeners run. Comparisons above use an isolated checkout of `origin/codex/stage1-jsx-lint` at `a8a62d62ca49db7415e14c3887dd305022b17309`, plus unchanged registration foundation `context.ts`, `settings.ts`, `finding.ts` from this branch. That dependency was not merged into the working branch. The Go script kind follows each original filename. Unescaped-entity private option fields are read reflectively only in the temporary capture overlay, preserving their actual values; no upstream listener or decoder is modified.
