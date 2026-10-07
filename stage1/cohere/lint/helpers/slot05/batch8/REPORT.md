Built: ParseCandidate, ParseVariant and LoadDesignSystem, each in its own .a file; eighteen prerequisites removed across six rules, zero final blockers.
Commits: claim 85c71d9 pushed before code; implementation 6df1743a9c4c0a612db579f57c2fe3359521e56f; main base e8ba3d5.
Checks: four-way three-helper Go gate PASS 209.625s; expanded variant gate PASS 35.439s; final coverage 34,057 cases; vet/types/format and filtered uncached oracle pass.
Mutants: all nine compiling semantic mutants caught against actual Go, three per helper; missing oracle-data failures are not credited.
Not covered: full repository gate, whole-rule findings/fixes/suggestions, shared parser integration, standalone CSS/registry/normalizer/filesystem implementations or exhaustive fuzzing.

## Landing and claim

The only pushed branch owned by this unit is codex/lint-helpers-05. Before any new claim, fresh fetch confirmed its pushed head 66a88f6 contained current origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965. The prior landing gate covers every older helper on that base, and batch seven's eleven-mutant four-way gate was already green and pushed. Main remained unchanged at the final fetch. No rebase was needed this turn; no main or area branch was pushed.

Explicit wildcard fetch inspected every claims file on all eighteen origin codex/lint-helpers* branches, plus shared HELPERS.md's comments bundle. Every higher concrete-symbol count is reserved. ParseCandidate, ParseVariant and LoadDesignSystem tie the highest unclaimed count, six consumers each. Claim 85c71d9 was pushed before writing code. Refreshed ownership scans found no competing reservations. evidence/ownership.json records the original claim SHA/time and tested main base. No fourth helper is reserved.

## Delivered behavior

ParseCandidate preserves ordered readings, prefix rejection, innermost-first variants, trailing-versus-leading importance precedence, static-first yields, empty-third-modifier truthiness, arbitrary-property guards, CSS-variable shorthand rewriting, typehint scanning, named fractions and each return/continue distinction. Segment, decoder, modifier, validity and root-search implementations remain explicit separately owned dependencies. It executes this batch's real ParseVariant rather than supplying a predicted whole-parser result.

ParseVariant preserves arbitrary-selector validation/wrapping, relative selectors, static/functional guards, ordered root splits, variable shorthand, compound compatibility, recursion and not/has/in modifier forwarding. Nil values/modifiers use presence bits; compound edges are numeric arena indexes. Children precede parents, avoiding recursive ownership. A failed parse can leave an unreachable appended child; callers read only reachable indexes and retain the arena for the result's lifetime.

LoadDesignSystem preserves configuration/error precedence, resolver selection, absolute-path resolution, one fresh collector and one graph load, graph-error short circuit, nil versus empty framework overrides, framework/theme/repository registration sequencing, optional evaluator construction, original collector fields and one successful counter increment. Theme, registry and evaluator objects are opaque engine handles retained by their provider. This helper orchestrates existing components rather than implementing their filesystem, CSS parser or registry internals. README.md states the adapter and dependency contracts.

## Independent oracle and coverage

Pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db decides behavior. Private overlays instrument the actual Go helper bodies, wrapping original dependency calls and preserving their return values. The cohere worktree is unchanged; upstream file hashes are in evidence/upstream.sha256. Parser dependency answers and loader operation traces come from actual Go execution, not handwritten expected verdicts. Loader cases really create/read temporary CSS files and run the actual Go collector and registry. Baselines compare Go output with source Node, emitted JavaScript and native under ASan/UBSan, byte for byte.

All nonempty Go string literals from every inventory-listed test file of all six consumers contribute. Counts are 259, 370, 110, 131, 247 and 239; deduplication yields 639 distinct literals, including prose/options as well as source. Missing coverage and Go-pin drift fail. The corpus also includes whitespace fields from those strings, all 5,056 pinned candidate fixture inputs and targeted branch controls. Fixture expected engine outputs and repository-specific registrations are not copied; actual Go parsing in a real loaded framework-plus-custom design system provides the expected readings.

Final coverage is 13,288 candidate cases, 15,014 variant cases and 5,755 loader cases, 34,057 total. The candidate prefix matrix tests empty and tw. Variant coverage additionally includes actual bracket-aware segments from class inputs and argument/modifier/compound controls for every generated framework registration. ParseVariant does not read Prefix, so both prefix configurations agree there as Go does.

Loader cases place each consumer literal in a CSS comment, escaping comment terminators. They vary nil/empty/custom framework overrides, absent/present utility definitions and malformed CSS. Four extra controls cover empty entry, missing package/resolver, default resolver selection and missing file. Collector contents, exact dependency call order, error text and returned fields/counters are observed. Full import graphs, host cwd failure and arbitrary IO are outside this bounded load orchestration comparison.

Go's loaded design-system implementation itself approximates child compounding compatibility by the parent registration kind. The port preserves that Go behavior; it does not claim to fix the separate engine parity gap described upstream in design_system.go. The adapter assumes valid, immutable dependency state and callbacks. Raw invalid UTF-8, mutated/cyclic arenas, unbounded recursion and concurrent mutation are not covered.

## Mutants and superseded failures

Every final mutant compiled and exited zero without stderr before comparison with Go. A compile failure, panic or sanitizer failure is never a credited kill. Exact anchors and witnesses are in evidence/mutants.json.

| Helper | Mutation | Independent first witness |
|---|---|---|
| ParseCandidate | erase trailing importance | verdict 18083: false instead of true |
| ParseCandidate | reverse returned variant order | verdict 17524: &:hover instead of &:focus |
| ParseCandidate | omit named fraction | verdict 39: empty instead of red-500/50 |
| ParseVariant | omit arbitrary selector wrapping | expanded verdict 2522: --my-var:red instead of &:is(--my-var:red) |
| ParseVariant | erase relative flag | expanded verdict 2505: false instead of true |
| ParseVariant | invert compound compatibility | expanded verdict 8721: absent instead of present |
| LoadDesignSystem | treat explicit empty framework as absent | verdict 128: framework call instead of register a static |
| LoadDesignSystem | omit evaluator construction | verdict 205: count call instead of evaluator |
| LoadDesignSystem | reverse repository variant ordering | verdict 8: register 😀 static instead of register a static |

The first private test build failed after 0.244s because the oracle used a bool-map-only sort helper with other map types, and its generated main had an unfinished loop. These were fixed only in owned test files.

The next gate failed after 198.835s: a candidate mutant changed parsing order and requested a dependency input the original case never queried, panicking at exit 70. It is not credited. Reversing the final variant result preserves dependency inputs and gives the intended semantic order witness; the full gate then passed 209.625s.

Expanded variant coverage initially failed after 18.130s because omitting selector wrapping changed a child passed to compounding, with no oracle answer for that changed selector. No kill is credited. The Go oracle now explicitly queries the real compatibility method for that mutated selector as well. The repaired expanded variant gate passes all three semantic mutants at 35.439s. No production helper or shared harness/compiler was changed to resolve either data gap.

## Commands and outputs

All test output went directly to log files. Setup succeeded in 31s; nproc printed 5:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (31s)
setup: done in 31s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Source /workspace/adamic-tools/env.sh before commands:

```sh
bash cloud/setup.sh > /tmp/lint05-batch8-setup.log 2>&1
ADAMIC_SLOT05_BATCH8_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch8/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch8 -count=1 -v -timeout=20m > /tmp/lint05-batch8-final.log 2>&1
ADAMIC_SLOT05_BATCH8_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch8/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch8 -run '^TestSlot05Variant$' -count=1 -v -timeout=20m > /tmp/lint05-batch8-variant-final.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch8 > /tmp/lint05-batch8-vet-final.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch8/main.a > /tmp/lint05-batch8-types.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch8 > /tmp/lint05-batch8-format-final.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > /tmp/lint05-batch8-oracle.log 2>&1
```

Three-helper package PASS 209.625s: candidate 56.14s, original 13,288-case variant 30.01s, loader 123.47s. Expanded 15,014-case variant PASS 35.439s supersedes that helper's smaller corpus. All nine distinct semantic mutations are caught. Vet/format logs are empty; types prints ordinary inferred-type records. Filtered uncached oracle PASS 1.183s, six input fixtures, zero cache hits and six probe misses.

Evidence saves consumer counts, corpus hashes and raw-observation log copies. Parser inputs are deterministic; loader hashes include temporary absolute paths and are run-specific. Committed logs trim trailing whitespace; raw /tmp logs remain unchanged. The complete repository gate was not run; the bounded owned-package gate and filtered external oracle were used.

## Rules unblocked

Each helper removes one listed prerequisite from all six:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Eighteen dependency occurrences are removed, but no rule loses its final blocker. Cumulative slot 05 retains twenty-three helpers, removes 187 prerequisites across sixty-four unique rules and removes four final blockers, raising conditional helper readiness from 46 to 50. CONSUMERS.md maps every helper; readiness.json retains every residual blocker. No rule inventory status is rewritten. Production parser/linter integration and full findings, spans, fixes and suggestions remain uncovered. The three claimed helper comparisons have no remaining shared-harness blocker.
