# Frequency-prioritized twenty-rule continuation

Baseline `e27d185a488c1c23c685ad0ed686b891dc2a8588`. The Go ranking runs
unmodified registered rules with default options, before any new verdicts are
ported. It visits 77 pinned TypeScript 6.0.3 compiler files and 321 tracked
repository `.ts` and `.a` files, including intentional probes: 398 parse-valid
files total. Type-checker and required-option exclusions, zero counts and any
rule panic are explicit in `volume_evidence/frequency.log`. Counts are raw rule
API findings, before project configuration or suppression. The manifest freezes
the input list at the baseline. Scratch compiler sources are not committed.

## Selection

Prioritize total compiler plus repository frequency, then completeness within
the existing parser and runtime capabilities. Three high-frequency candidates
are deferred: `no-inline-comments` (6,016; 31 of its 49 upstream cases need JSX,
plus configurable Go regular expressions), `id-length` (1,849; Unicode grapheme
segmentation and configurable regular expressions), and `consistent-return`
(546; its missing-return judgment calls Cohere's control-flow graph). They are
syntax-only, but need substantial dependencies beyond this rule slice. This
selection includes the highest-frequency rule, `one-var`, and introduces its
multiple-edit finding shape rather than restricting its fixes to one edit.

| Rule | Compiler | Repository | Total |
| --- | ---: | ---: | ---: |
| `one-var` | 6,731 | 1,420 | 8,151 |
| `nexus/consistency-no-ambiguous-identifier` | 5,469 | 303 | 5,772 |
| `nexus/consistency-no-abbreviated-identifier` | 4,827 | 52 | 4,879 |
| `@typescript-eslint/method-signature-style` | 1,508 | 2 | 1,510 |
| `nexus/consistency-require-type-suffix` | 1,170 | 180 | 1,350 |
| `@typescript-eslint/no-non-null-assertion` | 1,123 | 1 | 1,124 |
| `@typescript-eslint/prefer-literal-enum-member` | 1,043 | 0 | 1,043 |
| `base/consistency-no-console` | 1 | 1,041 | 1,042 |
| `no-plusplus` | 680 | 323 | 1,003 |
| `@typescript-eslint/no-wrapper-object-types` | 744 | 0 | 744 |
| `@typescript-eslint/prefer-enum-initializers` | 680 | 0 | 680 |
| `adamic/no-type-predicate` | 651 | 1 | 652 |
| `nexus/consistency-no-long-line-comment` | 540 | 99 | 639 |
| `nexus/consistency-no-multiline-arrow-function` | 540 | 72 | 612 |
| `prefer-destructuring` | 515 | 61 | 576 |
| `nexus/consistency-no-single-line-jsdoc` | 555 | 0 | 555 |
| `no-negated-condition` | 414 | 17 | 431 |
| `no-return-assign` | 276 | 0 | 276 |
| `nexus/consistency-no-shouting` | 118 | 53 | 171 |
| `nexus/consistency-no-enum` | 164 | 0 | 164 |

## Green step 1

Four rules implemented: type suffixes, console access, increment/decrement,
and written type predicates. Upstream capture grows from 873 to 944 unique
source/rule/options combinations. Final formatted source matches Go in Node
and sanitized native: 633,106 fixture bytes and 17,047,998 compiler/stage1 bytes
(163 files). Findings include ranges, exact messages and repair records; full
fixed source remains checked through the real Go edit engine.

Four executable mutants are caught on both Node and sanitized native: replace
the predicate node kind; invert the console receiver test; ignore the for-update
option; allow an interface's forbidden Type suffix. Package mutant run passed
in 66.082s. Cohere's 276-rule source/format gate passes. No new language gap in
this step. The other sixteen selected rules are pending, not delivered yet.

Setup: Go/clang/Node/submodules ready at 0s, cache warmed in 12s, total 12s;
`nproc` 5 and CPU quota four. Tests write logs directly to files.

```sh
source /workspace/adamic-tools/env.sh
cd cohere
go run -overlay=/workspace/scratch/lint-volume/frequency-overlay.json adamic_lint_frequency.go /workspace/scratch/lint-volume/frequency-manifest.txt > /workspace/scratch/lint-volume/frequency.log 2>&1
cd ..
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree)$' -count=1 -v -timeout 30m > /workspace/scratch/lint-volume/batch1-final-parity.log 2>&1
go test ./stage1/cohere/lint -run '^TestVolumeMutants$' -count=1 -v -timeout 30m > /workspace/scratch/lint-volume/batch1-mutants.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/cohere/lint/{lint,volume,volume_messages}.ts > /workspace/scratch/lint-volume/batch1-final-cohere.log 2>&1
```
