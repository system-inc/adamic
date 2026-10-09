# Frequency-prioritized twenty-rule continuation

This is the dated record of how the ten rules below were selected and first verified. Since
2026-10-07 they live on the registry like every other rule, one directory each under `rules/`
(`no-plusplus`, `no-negated-condition`, `no-return-assign`, `base-consistency-no-console`,
`nexus-consistency-require-type-suffix`, `nexus-consistency-no-enum`, `adamic-no-type-predicate`,
`typescript-method-signature-style`, `typescript-no-wrapper-object-types`,
`typescript-prefer-literal-enum-member`). The files this record names have moved:

- `volume.ts` is gone; each rule's code is its directory's `rule.ts`.
- `volume_messages.ts` and `testdata/generate_volume_messages.py`, which generated it, are gone; each
  rule's exact descriptions are its directory's `messages.ts`. `testdata/volume_rules.json`, the
  generator's input, is gone; the selection table below is the same data.
- `TestVolumeMutants` and `volume_test.go` are gone; each of the ten mutants in the table below is its
  rule's `mutant.json`, run by `TestMutants`. The volume source and its four option rows are part of
  `generated()` in `lint_test.go`, and each rule has its own witness.
- `testdata/frequency.go` and `volume_evidence/` remain, as the ranking tool and the logs of the runs
  described here. The commands below are the ones those runs used.

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
selection targets the highest-frequency rule, `one-var`; its multiple-edit
finding shape is required work, not implemented by the first batch.

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

## Checkpoint 2: nine fully covered additions and one limited addition

The request for twenty more rules is **not complete**. Delivered beyond the
baseline twenty: type suffix, console, plusplus, type predicate, wrapper object
types, literal enum members, no-enum, negated condition, and return assignment.
Method-signature-style additionally handles valid sources, both styles,
overload rewrites, module/`this` fix declines and readonly suggestions, but has
the parser-recovery limit below. These are ten implementations, not ten fully
verified rules. The other ten names in the selection table remain unimplemented:
`one-var`, ambiguous identifiers, abbreviated identifiers, non-null assertions,
enum initializers, long line comments, multiline arrows, prefer-destructuring,
single-line JSDoc and shouting. Unimplemented work is not a language refusal.

Final formatted-source comparison: 1,272 upstream source/rule/options
combinations captured, 1,264 compared, eight separately bounded failure checks.
Seven admitted malformed combinations are findings-only (five inherited
no-div-regex cases and two bare-arrow method-signature styles). Supported
fixtures and generated input produce **748,598 identical bytes**. The pinned
compiler and all stage1 `.ts` sources are 165 files and produce **18,914,195
identical bytes**. Total: **19,662,793 bytes**, Go versus the same source on Node
and native with ASan/UBSan/leak checks. Proposed repairs, rejection records and
converged fixed source are compared with Cohere's actual edit engine, not a
second implementation used as its oracle.

### Recovery dependency and gap evidence

Four actual method-signature test sources, each in `method` and `property`:

```ts
interface I
interface I { m(a: string): void;
interface I { m<(a: string): void; }
interface I { m<T(a: T): T; }
```

Go completes on every one. Its property-style output reports a method and
proposed repair for the last three. Native and Node panic on missing body,
stray angle and half generic. Missing closing brace runs past the two-second
bound on both ports. The existing parser's `typeLiteral` loop has no EOF exit;
`gaps/5_parser_recovery.ts` reduces it. This is a parser coverage defect, not an
Adamic subset limitation. The eight combinations retain explicit log records
and are excluded from the identical-byte total. Fixing the parser's missing-token
recovery is a dependency for full method-signature fixture parity; no parser,
compiler or runtime code is changed in this lint unit.

A separate actual language gap: positioned `lastIndexOf`. The program in
`gaps/4_last_index_position.ts` prints `1` on Node; native lowering refuses
`lastIndexOf with these arguments`. The fixer uses a prefix slice with the
supported one-argument search. The gap test requires both observations.

### New semantic mutants

All ten execute and give wrong answers on Node and sanitized native:

| Family | Mutation | Comparison that catches it |
| --- | --- | --- |
| Type predicate | listen to NeverKeyword | missing finding/message/range |
| Console member | invert receiver name | missing and spurious findings |
| Update option | invert for-afterthought exemption | wrong loop update findings |
| Method style | omit method listener | missing findings and fixed member rewrite |
| Wrapper type | omit Number | missing finding and lowercase fixed output |
| Enum literal option | invert allow-bitwise decision | wrong finding/message under decoded option |
| Enum declaration | disable listener | missing name finding |
| Negated condition | invert condition polarity | missing negated if/ternary findings |
| Return assignment | invert parentheses option | wrong arrow/return findings |
| Interface suffix | allow forbidden Type suffix | missing interface finding |

The original 19 rule mutants and the two earlier performance mutants are also
rerun by the full package (31 mutants total including this continuation). A discarded trial
method-style mutant caused a fix loop and was not counted: its replacement
omits a listener and is caught by wrong output. A discarded condition mutant
was ineffective and was replaced by the polarity mutant. Formatting required
shorter mutation anchors; this changes no mutant's semantic purpose.

### Verification commands

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -count=1 -v -timeout 30m > /workspace/scratch/lint-volume/batch2-full-package.log 2>&1
go test ./internal/oracle -run '^(TestNativeAgreesWithNode|TestTheOracleCatchesOneByte)$/^internal$/^oracle$/^testdata$/^(strings|maps_and_text|indexing|visits|sorting)\.a$' -count=1 -v -timeout 10m > /workspace/scratch/lint-volume/batch2-filtered-oracle.log 2>&1
go vet ./stage1/cohere/lint > /workspace/scratch/lint-volume/batch2-vet.log 2>&1
/workspace/scratch/cohere --no-fix --no-cache stage1/cohere/lint/{lint,volume,volume_messages,settings}.ts stage1/cohere/lint/gaps/{4_last_index_position,5_parser_recovery}.ts > /workspace/scratch/lint-volume/batch2-final-cohere.log 2>&1
```

Full lint package: PASS, 591.243s, including all 31 mutants, byte comparisons,
bounded parser failure checks and proving gap tests. Profiling artifact and
historical snapshot jobs are opt-in and skipped; throughput is run separately.

Filtered core oracle: PASS, 13.283s, five fixtures plus its one-byte negative
control. Go vet: exit 0, no output. Cohere source/format gate: PASS, 276 rules,
seven checked, 100% Adamic-ready (five of five selected files). The whole
repository `go test ./...` gate is not claimed; validation is the full touched
package and the exact filtered core oracle above.

### Throughput

Thirty implemented rules enabled, count-only mode, all 77 pinned compiler
files, 14,866 findings identical in each run. Five interleaved fresh runs,
including file reads, parsing, visiting, messages, findings and sorting; fixes
are not applied in count-only mode.

| Runner | Best seconds | Findings per second |
| --- | ---: | ---: |
| Go | 0.673702 | 22,066.14 |
| Native Adamic release | 3.023437 | 4,916.92 |
| Node running the same source | 1.599961 | 9,291.48 |

Native takes 4.49 times Go's elapsed time and 1.89 times Node's for this work.
This measures thirty implementations, including the valid-source method rule;
it is not a same-rule-count comparison with the previous twenty-rule benchmark.
Machine: x86_64 Linux 6.18.44, AMD EPYC 9V74 80-Core, `nproc` five, CPU quota
four. Load before `1.49 1.20 0.99`, after `1.49 1.23 1.01`. Go 1.27.1,
Node 24.19.0, clang 20.1.8; native is the normal unsanitized optimized build.

```sh
ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^TestThroughput$' -count=1 -v -timeout 20m > /workspace/scratch/lint-volume/batch2-throughput.log 2>&1
```

PASS, 36.075s. Every round's count is checked against Go. The full package's
count-only mutant separately proves that a wrong count is caught even when
ordinary findings/fixed output remains identical. Raw validation, core-oracle,
source-gate and throughput output is in `volume_evidence/batch2-*.log`.
