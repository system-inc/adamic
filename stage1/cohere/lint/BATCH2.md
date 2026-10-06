# Syntax-only lint batch 2

Branch `codex/stage1-lint-batch2`, cut from
`origin/codex/typescript-scanner` at
`0090256e607c3f2de7d5b67cef680ec010f95c1d`.
Implementation and harness commit: `bd6431476b78146dcf2adaa901c2aaaff47ecde2`.

Twenty additional rules, each in its own TypeScript file, with one-line
registrations in `batch2_registry.ts`. Eighteen core rules and two TypeScript
rules, all without a checker or binder dependency. `rule_context.ts` provides
syntax indexes, source spans, token signatures and report construction. The
existing traversal gets one registry call. No parser, compiler or runtime
source is changed.

The other worker's complete twenty-rule selection, including unimplemented
choices, was excluded. Its branch was fetched before each rule; the observed
tips are in [batch2_tip_checks.json](batch2_tip_checks.json). A final fetch still returned `0090256e`;
all twenty remained outside its selected list and implemented TS rules
([batch2_evidence/tip-final.log](batch2_evidence/tip-final.log)). Qualified TypeScript names are selected by
their unqualified n-to-z rule names.

## Findings, fixes and upstream fixtures

The oracle is unmodified Go cohere rules, loaded through the existing build
and capture overlays. Cohere pin:
`715ba94f3608a6500086b1076ce5cb7e51b836db`.
The same TS implementation runs under Node and native C compiled by clang.
Comparisons use complete output bytes: displayed findings, byte positions,
message IDs and text, fix/suggestion kind, replacement and edit range,
suggestion text, and the iteratively fixed source. Suggestions remain
suggestions. The capture checks every selected rule has upstream fixtures.
Filename extensions are retained where the constraint rule's TSX/MTS/CTS
suggestion requires a comma.

Observed passing comparisons:

- Twenty positive controls plus two option cases: 23,860 identical bytes.
  Controls include clean boundaries, quoted/static/generator constructors,
  regex token equality, empty bracket literals and decoded options.
- This batch: 616 captured upstream cases; 615 compare successfully, 176,812
  identical bytes. One malformed recovery case is explicitly refused below.
- Combined rules: 1,888 unique captured source/rule/options/extension cases,
  with unsupported recovery cases asserted separately; generated controls
  also included. 933,123 identical bytes.
- All 77 pinned TypeScript compiler files plus 110 current stage1 TS files:
  187 files, 19,162,042 identical bytes with ASan/UBSan native code.
- Release compiler findings and fixes before timing: 18,028,800 identical
  bytes on Go, Node and the `-O2` native build.

TypeScript 6.0.3 corpus pin:
`050880ce59e30b356b686bd3144efe24f875ebc8`.

## Mutants

Every row below has a separate native sanitized rebuild. Its selected-rule
activation name is changed from `ctx.enabled('<actual rule>')` to
`ctx.enabled('omitted-batch2-rule')`. The positive control selects the actual
rule, so its findings disappear. Both Node and native must finish successfully,
then their complete output must differ from Go. A compile failure, panic or
sanitizer failure does not count as a caught mutant. This is a mutation of
selected-rule dispatch; it does not disable the `all` selection path.

| Rule | Captured upstream cases | Dispatch mutant caught by |
| --- | ---: | --- |
| `no-caller` | 8 | Node and sanitized native |
| `no-eq-null` | 14 | Node and sanitized native |
| `no-empty-static-block` | 9 | Node and sanitized native |
| `no-proto` | 14 | Node and sanitized native |
| `no-script-url` | 29 | Node and sanitized native |
| `no-self-compare` | 25 | Node and sanitized native |
| `no-delete-var` | 9 | Node and sanitized native |
| `no-iterator` | 10 | Node and sanitized native |
| `no-compare-neg-zero` | 56 | Node and sanitized native |
| `no-async-promise-executor` | 24 | Node and sanitized native |
| `no-empty-pattern` | 44 | Node and sanitized native |
| `no-constructor-return` | 53 | Node and sanitized native |
| `no-multi-assign` | 70 | Node and sanitized native |
| `no-useless-catch` | 23 | Node and sanitized native |
| `no-unsafe-finally` | 61 | Node and sanitized native |
| `no-useless-concat` | 40 | Node and sanitized native |
| `no-empty-character-class` | 43 | Node and sanitized native |
| `no-ex-assign` | 19 | Node and sanitized native |
| `@typescript-eslint/no-unnecessary-type-constraint` | 43 | Node and sanitized native |
| `@typescript-eslint/prefer-namespace-keyword` | 22 | Node and sanitized native |

`TestBatch2Mutants` passed in 798.47s: twenty Node and twenty native mismatches.
The first run used the original controls. After adding regex and constructor
boundaries, those two families were rebuilt and rerun against the strengthened
controls; both remained caught (104.72s).

The combined regression additionally caught the existing decoration-range,
comment-fold and position-index mutants on both engines. The count-only mutant
produced `3` instead of Go's `2` on both engines while ordinary output stayed
identical. `TestTheOracleCatchesOneByte` independently passed in 7.61s.
These outcomes are observations from the saved logs, not general proofs that
unseen inputs agree.

## Throughput

77 compiler files, all fifty currently implemented rules enabled, 14,912
findings in every process. Count mode still parses, walks, constructs complete
findings and repair proposals, and sorts. It skips formatting output and
applying fixes. Builds and the full-output release parity check are outside
timed rounds. Five rounds interleave fresh Go, native and Node processes; no
other build, test or profiling jobs run during timing.

| Engine | Best of five seconds | Findings/s |
| --- | ---: | ---: |
| Go | 1.155717 | 12,902.82 |
| Native release | 6.012244 | 2,480.27 |
| Node | 2.849005 | 5,234.11 |

Native is about 5.20 times Go elapsed and 2.11 times Node elapsed. This is a
combined fifty-rule workload, on a different CPU from the earlier twenty-rule
performance report, so it is not a same-workload speed comparison.
Raw rounds and load observations are in [batch2_evidence/throughput.log](batch2_evidence/throughput.log).

Machine: Intel Xeon Platinum 8573C, Linux 6.18.44; `nproc` prints `5`,
`cpu.max` is `400000 100000`, 17.6 GB memory. Go 1.27.1, clang 20.1.8,
Node 24.19.0. Setup succeeded and printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (34s)
setup: done in 34s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Commands and outputs

Commands ran from `/workspace/adamic`; test output went directly to log files,
never through a pipe. Setup was `bash cloud/setup.sh` with output redirected;
the environment used afterwards was
`source /workspace/adamic-tools/env.sh`.
The corpus checkout was `git clone --depth 1 --branch v6.0.3
https://github.com/microsoft/TypeScript.git
/workspace/scratch/lint-batch2/typescript-6.0.3`, followed by the exact pin
check in the corpus test.

```bash
go test ./stage1/cohere/lint -run '^TestBatch2UpstreamAgree$' -count=1 -v -timeout=10m > /tmp/adamic-batch2-upstream.log 2>&1
go test ./stage1/cohere/lint -run '^TestBatch2Mutants$' -count=1 -v -timeout=30m > /tmp/adamic-batch2-mutants.log 2>&1
go test ./stage1/cohere/lint -run '^TestBatch2Mutants$/(no-self-compare|no-constructor-return)$' -count=1 -v -timeout=10m > /tmp/adamic-batch2-strengthened-mutants.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestBatch2Controls|TestRulesAgree|TestCompilerAndStage1Agree|TestNestedConstructorGap|TestOptionAndComparatorGaps|TestDecorationOptionMutant|TestCountGuardMutant|TestCommentFoldMutant|TestPositionIndexMutant)$' -count=1 -v -timeout=30m > /tmp/adamic-batch2-regression.log 2>&1
ADAMIC_LINT_BENCH=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/lint-batch2/typescript-6.0.3 go test ./stage1/cohere/lint -run '^TestThroughput$' -count=1 -v -timeout=20m > /tmp/adamic-batch2-throughput.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/adamic-batch2-filtered-oracle.log 2>&1
go vet ./stage1/cohere/lint > /tmp/adamic-batch2-vet.log 2>&1
git diff --check > /tmp/adamic-batch2-diff-check.log 2>&1
```

All final commands above exited zero. Upstream parity passed in 63.25s;
combined regression in 322.32s; throughput including release build and parity
in 106.43s. Vet and whitespace checks printed nothing.

The source gate binary was built with
`go build -o /workspace/scratch/lint-batch2/cohere ./command/cohere` inside
`cohere`. `--format-only` formatted the new files and the small integration.
Then `cohere --no-fix --no-cache` ran on all twenty rule files,
`rule_context.ts`, `batch2_registry.ts` and `lint.ts`:

```text
0 findings; 276 rules; 24 checked; 100% Adamic-ready (24 of 24)
```

Its unmodified footer is in [batch2_evidence/source-gate.log](batch2_evidence/source-gate.log).
Each final log above is copied verbatim under `batch2_evidence/`.
The local attributes keep raw log whitespace and line endings intact.

## Limits and initial failures

- `new Promise(@dec async () => {})` is malformed. Go recovers and produces
  no lint finding. Stage1 refuses `AtToken` at position 12 under both engines.
  Its exact refusal is asserted; it is not counted as successful parity.
- `x === -0_0;` and `x === -00;` have Go parse diagnostics. Their recovered
  findings match byte for byte; fixing invalid input is deliberately not
  attempted. These are exact enumerated recovery cases, not a broad skip.
- The existing method-signature batch has eight option/source combinations
  requiring unsupported parser recovery (four malformed sources, two styles).
  The combined gate preserves and asserts their explicit refusals. Parser
  recovery is outside this unit's territory.
- The first controls run refused assigning `empty.length = 0` in stage 0.
  Replacing the local empty-range array avoided the unsupported lowering;
  the final sanitized build and fixtures pass.
- The first oracle build downloaded cold regexp dependencies. The harness
  rejects any stderr, so that initial build failed despite a zero build exit.
  Dependencies were cached and the check reran successfully.
- Upstream fixtures caught the quoted-constructor parser shape difference.
  Rule-local normalization fixed it; all 53 constructor fixtures pass,
  including generator, static, accessor and nested function boundaries.
- No type-aware or binder-dependent rules are ported. No complete repository
  `go test ./...` gate was run. The exact touched-package selection and filtered
  external oracle are listed above; unchanged original-rule and volume-rule
  mutation suites were not repeated. No new instruction profile, allocation
  profile, RSS measurement or performance optimization is claimed.
