# Wave 1 slot 05 continuation

Built two .a rule candidates and validated the earlier no-cond-assign candidate; Google font display is blocked.
Claim pushed first: f6e65dc0; draft 85dd2939; Tailwind a061e72d; descriptions d7d89680; scope 93bdda10.
Four-way supported corpus PASS: 451 upstream cases, 136,399 bytes; compiler/stage1: 214 files, 12,710,560 bytes; three owned mutants PASS.
Assignment-token omission, removed RTL exemption, and ignored description each compiled and ran cleanly, then failed output comparison on Node, emitted JavaScript and sanitized native.
No complete three-rule certification: Google needs JSX; two JSX cases and one malformed input excluded; default .a discovery and the profile helper mismatch remain blocked.

## Selection and integration

Fetched all origin heads and searched direct claim Markdown on 297 origin refs, plus main's ports, before claiming. Helper-ready list had one eligible entry: structure/tailwind-no-physical-direction. Next syntax-only inventory entries were @eslint-community/eslint-comments/require-description and @next/next/google-font-display. Claim f6e65dc0 was pushed before code. The inventory queue includes syntax-waiting-on-helpers and syntax-ready entries, excluding checker/type-aware/binding rules.

The registration foundation requires rule.ts, contrary to the task's .a requirement. The merged helper profile refers to the old portFiles variable and cannot compile against the registration function. Shared infrastructure was not edited. compatibility.patch is the scratch migration from origin/codex/lint-wave1-12; validate.py additionally preserves upstream fixture filename extensions and chooses TSX in the Go driver. This patch is evidence only, not applied to the worktree. Default integration is still unavailable.

Google font display is entirely JSX. The executable parser probe in jsx.log exits 70 with `parser slice expected GreaterThanToken, got Identifier at 32`. Its source is google-font.tsx.txt. Go's own TestGoogleFontDisplay passes, including that missing-display finding. No Google port or mutant was fabricated; parser support is needed first.

## Reproduction

Source `/workspace/adamic-tools/env.sh`. Obtain TypeScript v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8, at the ADAMIC_TYPESCRIPT_SOURCE path. From the repository root:

```sh
python3 stage1/cohere/lint/claims/wave1-05-evidence/validate.py --scratch /tmp/lint-wave1-05-repro
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-05-typescript
go test -overlay=/tmp/lint-wave1-05-repro/overlay.json ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestWave05SupportedCorpus|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=20m > /tmp/wave05-corpus.log 2>&1
go test -overlay=/tmp/lint-wave1-05-repro/overlay.json ./stage1/cohere/lint -run '^TestMutants$/(compound_assignment_omitted|rtl_exemption_removed|reason_ignored)$' -count=1 -v -timeout=10m > /tmp/wave05-mutants.log 2>&1
go test -overlay=/tmp/lint-wave1-05-repro/overlay.json ./stage1/cohere/lint -run '^TestWave05Throughput$' -count=1 -v -timeout=10m > /tmp/wave05-throughput.log 2>&1
go test -overlay=/tmp/lint-wave1-05-repro/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/wave05-registry.log 2>&1
go vet -overlay=/tmp/lint-wave1-05-repro/overlay.json ./stage1/cohere/lint/... > /tmp/wave05-vet.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v > /tmp/wave05-oracle-control.log 2>&1
```

Each comparison checks serialized findings, ranges, messages and complete fixed output against unchanged upstream Go rules. Sanitized native is used for correctness; throughput uses unsanitized native. Owned witnesses matched 17,445 bytes. The final bounded upstream capture retained 454 unique source/rule/options/filename cases and explicitly excluded exactly two JSX cases and `x;\n/* eslint-disable`, leaving 451. The initial unrestricted TestRulesAgree failed because the inherited driver replayed JSX as .ts; parity.log records that failure and the passing compiler/stage1 comparison, not an overall passing command.

## Throughput observations

Best of three interleaved rounds, end-to-end driver time including startup and parsing. Corpus: 77 compiler files plus one 1,000-statement stress file per rule. Compiler alone had 52 no-cond-assign findings, zero Tailwind findings and 125 description findings. Stress prevents reporting an uninformative zero-findings Tailwind rate. The count output matched Go in every measured round.

| Rule | Findings | native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| no-cond-assign | 1,052 | 942.20 | 1,345.32 | 6,208.24 |
| structure/tailwind-no-physical-direction | 2,000 | 1,611.03 | 2,357.34 | 10,227.00 |
| @eslint-community/eslint-comments/require-description | 1,125 | 582.62 | 1,038.27 | 3,761.95 |

These are observations for this corpus, not parser-only benchmarks. No Google rate is available.

## Setup and limits

nproc: 5. Initial cloud/setup.sh completed: go 0s, clang 1s, Node 1s, submodules 1s, cache warm 75s, done 75s. On continuation, Go/clang/Node/submodules each reported 0s, then cache warming failed in profile_test.go:32: cannot range over portFiles (function). There were no completion timing lines on that failed run. The scratch overlay worked around this mismatch for focused validation.

No full repository gate was run. Focused lint, registry, vet and the one-byte oracle control were used. No shared parser, compiler, registry or harness source was changed. Full JSX fidelity, malformed-input behavior and automatic discovery of .a rule directories remain unverified/blocked. These implementations remain candidates rather than complete ports under the requested bar.
