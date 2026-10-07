Built `no-fallthrough` in .a with directory registration and exact Go messages.
Commits: pushed claim 9b082b6d before code; prior work was pushed through 39d19398; implementation SHA is in the final delivery.
Commands/output: final bounded go test PASS 346.558s, 159 upstream cases, 238 corpus files / 714 cases / 38,193,081 identical bytes across all runtimes.
Mutants: fallthrough_findings_disabled compiled and finished; only Go output comparison caught it on Node, emitted JavaScript and ASan/UBSan native.
Not covered: full repository gate, unmodified shared profile/registry-copy tests; general commentPattern expressions; one shared-parser JSX fixture.

The structural exit analysis reproduces the Go rule, including conditional branches, try/catch/finally, nested switches, labeled/with statements, and the narrower do-while analysis that credits return/throw but not a loop-owned break. Empty clauses are skipped unless a blank line separates them from the next clause, with allowEmptyCase overriding that exception.

The excusing comment is the last comment in the sole-block interior window first, then the ordinary between-clauses window. Only the Go rule's four exact default spellings count. eslint-/oxlint- directives never excuse a clause. Configured patterns are case-insensitive and use the bounded implementation in ../default-case/pattern.a; general runtime patterns remain blocked by stage 0's constant-pattern restriction. reportUnusedFallthroughComment reports the comment's exact byte-derived range when the clause always exits. Normal findings point at the clause fallen into, with distinct case/default IDs and the exact Go description; no fixes are offered.

All 94 non-JSX upstream cases pass with full findings/fixed-source output. One originally clean fixture in Fallthrough.tsx contains td elements. The existing shared parser refuses it before the listener: expected GreaterThanToken, got Identifier at 84. The original source, file and expected outcome are committed in ../default-case/testdata/jsx-gap.json, with the actual exit-70 probe in the evidence. This fixture is excluded explicitly, not reported as clean. General configured regexp support is likewise an explicit incomplete boundary.

Findings per second, native / Node / Go: **492.41 / 215.91 / 4237.88**, best of five wall-clock runs including process startup. Each timing corpus contains 77 TypeScript src/compiler files plus a 1,000-line positive file. Native timing is unsanitized; all correctness comparisons use ASan/UBSan and require empty stderr on success. Default-case counted 1,286 findings, no-extra-label 1,001, no-fallthrough 1,009; counts agree across all engines in every round. No speedup is inferred from these measurements.

The final source tree was held by this exact command, with test output written directly to a log:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/default-case/validate.py > /tmp/wave10-fourth-overlay.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-10-next/typescript GOFLAGS=-overlay=/tmp/lint-wave1-10-next/overlay.json go test -count=1 -timeout 15m -v ./stage1/cohere/lint -run '^TestFourth(Witnesses|Upstream|Corpus|Mutants|Throughput|Shapes|PatternRefusal)$' > /tmp/wave10-fourth-final.log 2>&1
```

Observed final results:

- Witnesses: PASS 19.94s, 1,459 identical bytes.
- Upstream: PASS 25.71s, 34,918 identical bytes; 27 default-case, 38 no-extra-label, 94 no-fallthrough. The original Go tests passed and were captured through their assertion harness. One clean JSX source is explicitly excluded from replay.
- Corpus: PASS 95.17s; 238 source files / 714 selected-rule cases / 38,193,081 identical bytes, including exact messages, UTF-8 diagnostic ranges, repair metadata, independent edit ranges and Go fixed source.
- Mutants: PASS 109.26s; three baseline-positive compiling semantic mutants were caught by output comparison on all three runtimes.
- Throughput: PASS 75.19s; 78 files, five interleaved rounds, identical counts.
- Shapes: PASS 20.56s, 6,411 identical bytes; 19 cases cover Unicode byte offsets/case folding, comment ordering, literal custom patterns, known-invalid fallback and nested control flow.
- Pattern refusal: PASS 0.72s; all three runtimes agree on the explicit unsupported-pattern boundary, and its compiling guard-removal mutant is caught.
- Touched-package vet through the same overlay: exit 0, empty log.
- Registry generation: exit 0, 17 discovered rule names. No shared registration source was changed.

Setup was rerun: ordinary setup exited 1 at shared profile_test.go:32 because it still ranges over the portFiles function. All four ready steps printed 0s. The existing isolated compatibility overlay lets setup complete: Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 19s, done 19s; nproc 5, cpu.max 400000 100000, 17.6 GB. Shared registry tests separately fail because their fixture copier still assumes rule.ts, for example rules/arrow-body-style/rule.ts. These failures belong to the published .a harness integration, not a shared-file edit made by this unit. Latest origin/codex/lint-harness-dot-a is f4d98cab; it still rejects multiple fixes and the existing parser still rejects JSX.

All three new rules use full output/fix comparison in the scratch overlay. The overlay's diagnostic-only special case is confined to earlier arrow-body-style/no-extra-bind rules; none of these three names takes that path. It adapts the old shared driver to .a module discovery/emitted-JS execution/independent edit spans in temporary copies and does not replace rule answers. The external Go callbacks, messages and actual fixer remain unmodified.

Preliminary failures are retained, not counted as passes: an early return in the owned Pattern constructor reached clang as a non-void return mismatch, and was replaced by a nested conditional within the owned file. First mutant forms used false && void calls, which stage 0 refused as BinaryExpression statements; these were not credited, and the final variants mutate ordinary if conditions. One corpus run caught an input changing while it ran, when pattern.a was improved during validation. The final complete gate ran with frozen sources and passed. Earlier rates and failed combined runs are superseded by the final log.

Claim selection checked all 348 fetched origin refs. Selection JSON files in another worker's evidence list rules with claims: []; these inventory mentions are not reservations. Actual Markdown claim documents and main's executable dispatches leave these three first in the 62-rule measured helper handoff. default-case had been skipped incorrectly by the previous name-only scan and is now correctly claimed. The public selection evidence is in ../default-case/testdata/selection.json. No syntax-only fallback was needed, and no further rules were claimed after these three.

Public raw logs reside under ../default-case/evidence. No protected compiler file, shared generator, shared dispatcher or shared test harness was edited. Authored Adamic files use .a; auxiliary Go drivers use .go.txt and raw witnesses use .ts.txt. No .ts Adamic source was authored, and no pull request was opened.
