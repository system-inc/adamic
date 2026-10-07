Built: owned .a deprecation/duplicate listeners and class-value traversal; unknown-class decision/reporting core with required live resolver and ignore-pattern adapters.
Commits: claim 0ad2c2db preceded all code; this report is included in the implementation commits on codex/lint-wave1-01.
Checks: 1,001 Go core observations, 51/54 fixtures and all 678 rule/source pairs over 339 compiler/stage1 files match on three backends; five .ts fixtures match independent parsing.
Mutants: restrictive major-version gate, inverted duplicate first-occurrence guard, and reversed nil-system guard all compile/run and fail only output comparison on all three backends.
Not covered: full integration certificate, three multiple-edit fixtures, live project CSS resolution, configured variable regexes, independent JSX parsing, or converged post-fix parsing.

The decision-core oracle calls actual private Go deprecationFor, reportDuplicates, reportTemplateDuplicates and classExistsIn through a temporary Go overlay. Rule bodies are not modified. Its 1,001 observations cover release gating, variants, importance, removed utilities, exact descriptions, duplicate diagnostics and disjoint edit lists, template holes and glued fragments, nil-system safety and marker/arbitrary-property exemptions. All three backend outputs match 184,433 bytes. Unknown-class observations cover nil systems and exemptions; they do not exercise the unported candidate/value resolver. Expected bytes are retained separately from runtime input.

The fixture capture executes actual Go upstream tests and preserves filenames/options. There are 24 deprecation and 30 duplicate fixtures. The owned Go adapter selects out exactly the three cases whose actual diagnostics have more than one automatic edit, because shared Finding cannot carry them. The other 51 match 17,695 bytes of findings, spans, edit ranges/texts, overlap decisions and first edit-plan source on each backend. That first-plan comparison uses Go edit.Resolve and is not a converged autofix or post-fix parse certificate. Source parsing independently matches the five .ts fixtures; .tsx fixtures use full projected Go ASTs.

All three candidates have owned descriptors, upstream adapters, .a modules, raw witnesses and mutant metadata. The earlier report-only directories broke registry discovery; adding valid owned descriptors repairs that mistake. The default unknown-class factory has no live program resolver, so it refuses explicitly. Its bind method accepts a resolver and ignore-pattern predicate for the reporting core; no current shared adapter supplies them. No guessed root table or missing-system clean verdict is used as a full-rule certificate. Duplicate reporting preserves every deletion in its core, and refuses when an integration diagnostic needs more than one edit. Configured variable patterns beyond the defaults refuse pending a Go-equivalent regex provider.

Exact shared gaps: Finding has one editStart/editEnd/replacement and shared Go testdata/oracle.go rejects len(d.Fixes)!=1. RuleContext has no program/project root/recorded stylesheet filesystem. LoadedDesignSystem, ParseCandidate and ClassValueResolvesIn still have no delivered native adapter/engine. Shared TestRulesAgree fails earlier on missing ./volume.ts, ./messages.ts and ./comments.ts imports (shared.log), so it cannot certify these candidates. New traversal remains in the owned rule directory; no shared generator, harness or compiler implementation is edited.

Commands, with output redirected to logs:

    bash cloud/setup.sh
    source /workspace/adamic-tools/env.sh
    go test ./stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes -run 'TestDecisionCores|TestListeners|TestRefusals' -count=1 -timeout 15m -v
    go test ./stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes -run TestListeners -count=1 -timeout 15m -v
    go test ./stage1/cohere/lint/rules/better-tailwindcss-no-deprecated-classes -run TestListenerCorpus -count=1 -timeout 30m -v
    go run ./cmd/lint-registry
    go test ./stage1/cohere/lint/registry -count=1 -timeout 10m -v
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 10m -v
    go test ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -timeout 15m -v

The focused core/listener/refusal gate passes in 33.778s; the per-rule rate gate passes in 15.769s. Registry tests pass in 0.036s. The filtered uncached oracle passes in 17.199s (six misses). Initial setup failed on the report-only directories; retry passes in 37s, nproc=5. Timing lines: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s; build cache warm 37s. Earlier failed development attempts exposed parameter properties, inferred never[], array length assignment, a temporary mutant nominal-import mismatch and a missing transport field; none is counted as a semantic mutant catch. Successful results follow their correction.

Fixture findings/s (startup included; Go parses source, Node/native consume projected ASTs):

| Rule | Native | Node | Go |
|---|---:|---:|---:|
| better-tailwindcss/no-deprecated-classes | 274.837 | 165.527 | 3384.396 |
| better-tailwindcss/no-duplicate-classes | 244.650 | 129.317 | 2601.457 |

No meaningful full-rule throughput is available for no-unknown-classes while its live resolver is blocked. These rates cover supported fixture subsets and are not comparable parser benchmarks. The full repository gate was not run. No next batch is claimed while this batch lacks its full integration certificate.

The corpus gate passes in 499.606s: all 339 source files (262 stage1 plus 77 TypeScript src/compiler files), all 678 source/rule pairs supported, 26,364,805 bytes per backend. TypeScript is pinned to v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8. This is projected-tree rule behavior, including initial findings/fixes and the first edit-plan source, not independent whole-corpus parser or converging-engine parity. No corpus pair was excluded for multiple edits.

The stricter refusal rerun passes in 12.196s. It requires exit code 70, the exact first stderr line, and empty stdout for each missing provider on source Node, emitted JavaScript and sanitized native. The failure does not print a clean result.
