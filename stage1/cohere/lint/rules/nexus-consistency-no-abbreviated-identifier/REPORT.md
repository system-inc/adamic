Built: nexus/consistency-no-abbreviated-identifier from batch 3 on the unified harness, using .a modules and handed Identifier nodes.
Base: origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; no shared harness files changed.
Checks: lint-registry, gofmt, vet, TestRulesAgree (13,133,047 identical bytes), TestOwnedWitnesses (97,599 identical bytes), and all 41 TestMutants passed.
Mutant: suppressing the whole-word props finding compiled and ran, then Go comparison caught it on source Node, emitted JavaScript and ASan/UBSan native.
Not covered: the full compiler/stage1 corpus, throughput and the complete repository gate were not rerun; inherited malformed method-signature recovery remains explicitly unsupported by the harness.

Source: origin/codex/stage1-lint-batch3 fa9781c2c0911c52312002ba97ffd4b560d178ae, rules/no_abbreviated_identifier.ts, as assigned by DEDUP_LEDGER.md. The descriptor prefix TestConsistencyNoAbbreviatedIdentifier covers all eleven real upstream Test names; their source/rule/options rows enter the 2,200-combination shared comparison. The Go adapter decodes field 5 into nexus.ConsistencyNoAbbreviatedIdentifierOptions, including all four option arrays. The default witness fires without an options sidecar, so this branch does not depend on the pending witness-options merge.

The four policy descriptions are copied verbatim. The vocabulary and binding cache are private to this registered rule. Existing RuleContext ancestry, child tests, scanning and reporting are reused. The foreign-name helper includes current Go's JSX exemptions, and late-prefix matching preserves Go's phase ordering. Go regex builders are expanded once against the fixed policy vocabulary into JS RegExp literals; no hand-written matchers or dynamic constructors remain. The millisecond pattern follows the codex/lint-regex table row, with a non-global replacement preserving Go's first-match behavior. Finding offsets continue through the shared UTF-16-to-byte wire serializer.

Exact commands and complete logs are in testdata/evidence/. The initial gate checked the owned mutant plus both parity tests; every discovered mutant was then rerun in three independent groups (14, 14 and 13), with every name appearing exactly once and no failures. No checks were skipped, changed or relaxed in this unit. No known blocker was reached by this rule.


## Identifier speed change, 2026-10-14

Baseline: `efc78ad6`. Regex literals now initialize at module scope, and their
unchanged non-global patterns are gated by literal prefixes/substrings. Segment
replacement runs only after a match. A module-initialized policy builds the
whole-word Map, allowed-name Set, phase lists, and capitalized words/replacements
in vocabulary order. The walk only reads these structures; no walk-time cache
is introduced. Go's whole/early/suffix/late/segment priority and its first
right-bounded replacement occurrence are preserved. No oracle, test gate,
deadline, sanitizer option, or language implementation changed.

Measured on AMD EPYC 9V74, 4-core cgroup quota (`400000 100000`), affinity CPUs
0–4 (5 available logical CPUs), Clang 20.1.8. Callgrind 3.24 runs the complete
findings protocol with `taskset -c 0`; wall runs use default affinity and no
concurrent measurement jobs. Both use only this rule on TypeScript v6.0.3
`checker.ts`, pin `050880ce59e30b356b686bd3144efe24f875ebc8`.

| Measurement | Before | After |
| --- | ---: | ---: |
| Release instructions | 32,441,192,391 | 9,200,792,329 |
| abbreviation() inclusive share | 81.53% | 34.85% |
| Release wall, three runs | 3.091–3.234 s | 1.063–1.090 s |
| Go wall, three runs | 0.324–0.391 s | 0.295–0.323 s |
| Median release / median Go | 9.50x | 3.49x |
| Exact sanitized side 4, isolated | 23.846 s | 3.659 s |
| Sanitized side 4 during concurrent agreement | — | 6.270 s |
| Count-only allocations | 33,025,468 | 1,520,005 |

Release uses `-O2`; the instruction binary adds `-g`. Sanitized uses
`-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all`. All use
`-std=c11 -Wall -Wextra -Werror -Wcast-function-type-strict -pedantic
-Wno-unused-variable -Wno-unused-but-set-variable -Wno-unused-function
-Wno-unused-parameter -Wno-self-assign -ffp-contract=off
-fno-optimize-sibling-calls`, and link with `-lm`. Count-only uses the release
flags plus `-DADAMIC_COUNT` and `--count`, so its allocation counts exclude wire
formatting. ASAN_OPTIONS, UBSAN_OPTIONS and LSAN_OPTIONS remain unset, exactly as
in the existing sanitized test. No LTO or architecture-specific flag is added.

All 16 existing RulesAgree shards passed (3,922 captured/generated cases), all
16 owned-witness shards passed, and all 16 JSX tree shards passed. These compare
source Node, emitted JavaScript and sanitized native with Go. The rule's existing
whole-word props mutant was caught on Node and emitted JavaScript; an additional
scratch build with `adamic build ... --sanitize` equaled mutated Node and differed
from Go on the owned witness. MutantsUnion, RulesAgreeUnion, OwnedWitnessesUnion,
registry generation and `go vet ./stage1/cohere/lint` also passed.

All 1,606 selected existing CompilerAndStage1Agree units passed. Selection covers
all five sides of the rule's 64 file buckets (rule hash group 1), its checker.ts
bucket (owner 1281, sides 0–3), its indivisible sanitized checker test, and all
five sides of the 256 non-checker all-rule convergence buckets. Setup validates
the complete census. The corpus contains 77 compiler files and 547 retained
Stage 1 files. The combined checker.ts/all case remains the existing deferred
heavy unit; it is not claimed here. No agreement test was edited or reshaped.

[Measurements, complete Callgrind profiles, selected test names and test logs](evidence/speed-20261014/)
are saved with this change. Before instructions sum both dumped Callgrind parts;
after has one complete part. Findings bytes match Go, including fixes and
suggestions. The isolated timings and concurrent run establish local headroom;
Loom performance still needs the user's submission.

Reproduction uses the existing `TestProfileArtifacts` with
`ADAMIC_TYPESCRIPT_SOURCE=<pinned compiler checkout>` and
`ADAMIC_LINT_PROFILE_DIR=<scratch output>`. Write a one-row manifest containing
`<compiler>/src/compiler/checker.ts<TAB>nexus/consistency-no-abbreviated-identifier`.
Run `taskset -c 0 valgrind --tool=callgrind --callgrind-out-file=<profile>
<scratch>/profiled --manifest <manifest>`, then three default-affinity runs each
of `<scratch>/scanner` and `<scratch>/oracle`. Run the exact
`TestCompilerAndStage1Agree_CheckerSanitized_nexus_consistency_no_abbreviated_identifier`
for side 4, and `<scratch>/counted --manifest <manifest> --count` for allocations.

Fixture/mutant command:
`go test ./stage1/cohere/lint -run
'^(TestRulesAgree_[0-9]+|TestRulesAgreeUnion|TestOwnedWitnesses_[0-9]+|TestOwnedWitnessesUnion|TestMutants_nexus_consistency_no_abbreviated_identifier|TestMutantsUnion|TestJsxLintTrees_[0-9]+)$' -v`.
Compiler agreement command uses `go test ./stage1/cohere/lint -run <selector> -v`,
where the selector is the anchored alternation of all names in the saved
`compiler-tests.txt`. Neither invocation changes parallelism or test budgets.
