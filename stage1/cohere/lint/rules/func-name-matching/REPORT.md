Built: five earlier claimed rules, each in its own directory; Google Fonts and one Tailwind fixture retain an explicit JSX parser blocker.
Commits: implementation SHAs are recorded below after the per-rule commits.
Commands: owned validation compares Go cohere, source Node, emitted JavaScript, and ASan/UBSan native; logs are in evidence/.
Mutants: one compiling semantic mutant per rule, caught by the complete Go output comparison on all three execution paths.
Not covered: a complete repository gate, independent JSX parsing, and shared harness integration pending the separately authored harness branch.

# Earlier claims finished before selecting more

This finishes consistent-this, func-name-matching, structure/tailwind-no-physical-direction, @eslint-community/eslint-comments/require-description, and @next/next/google-font-display. No compiler, parser, shared registration generator, or shared test harness was edited. Sources use the explicitly authorized temporary .ts extension because this checkout still requires it. The available origin/codex/lint-harness-dot-a commit 2650ad59 contains the separate .a/profile/suggestion handoff and is not merged here, preserving the rule-only territory.

All five Go rules are non-fixable and provide no suggestions. Findings include exact ids, descriptions and UTF-8 ranges. The complete comparison also asserts zero fixes, zero suggestions and identical unchanged fixed source. Options are the actual decoded Go options captured by the upstream tests, as required by the registration contract, including direction, descriptor/CommonJS flags, alias arrays, ignored directives and additional directive names.

consistent-this follows Go's source-file/function passes, initialized aliases, plain-assignment rescue, inner block/catch/switch/with scope gates, arrow assignment-only passes, compounds and destructuring exemptions. func-name-matching preserves the direct named-function restriction, computed string keys, Unicode identifier tables, all assignment operators and three descriptor forms. Empty descriptor targets retain their own empty name. Tailwind keeps Go's 20 ordered mappings, Unicode Fields whitespace, negative-family limits, direction variants, filename gate, regex eligibility and each template piece. Description comments use the shared parser-anchored helper and convert its byte ranges back to UTF-16 for reporting; description splitting uses ECMAScript whitespace, while fallback directives retain Go TrimSpace, block continuation markers and whole-word boundaries. Empty custom directive names are represented separately from 'no directive'. Google Fonts handles both opening JSX kinds, intrinsic link names, the first case-insensitive literal href, the pinned XHTML entity table, numeric entities and Go's deliberately loose first display parameter.

## Evidence boundaries

Original Go tests yield 135 unique func-name-matching configurations, 62 consistent-this, 54 Tailwind, 109 require-description and 16 Google Fonts: 376 configurations total. Deduplication includes the filename and decoded options. The independent parser comparison includes 359 original cases and four owned witnesses: 363 cases. It explicitly excludes the 16 Google Fonts JSX sources, the one Tailwind JSX source and the owned Google Fonts JSX witness. The require-description JSX comment fixture matches complete findings; this does not certify the parser's JSX tree, which an earlier owned probe observed to be wrong.

The other 17 original configurations compare complete rule findings using only Go AST kinds, geometry, text and child indexes projected into the arena. Go does not supply findings to the Adamic implementation. This proves rule semantics on that adapter; it is not an independent parser result. The same test sends positive JSX fixtures to the unmodified independent parser and requires exit 70 with 'parser slice expected GreaterThanToken' on Node, emitted JavaScript and sanitized native. Google Fonts remains integration-blocked until that parser supports JSX; Tailwind otherwise works on the plain-source cases and compiler/stage1 corpus.

The Google Fonts mutant likewise uses the projected JSX geometry. The other four mutants use independently parsed raw source. Every credited mutant compiles, exits zero, has empty stderr and differs only through Go's output comparison. No compiler crash, sanitizer failure or parse refusal is counted as a semantic catch.

## Reproduce

Source /workspace/adamic-tools/env.sh and run from the repository root, with stdout/stderr redirected directly to a log:

```
ADAMIC_WAVE104_EVIDENCE=/workspace/adamic/stage1/cohere/lint/rules/func-name-matching/evidence ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript ADAMIC_LINT_BENCH=1 go test ./stage1/cohere/lint/rules/func-name-matching -count=1 -v -timeout=15m > stage1/cohere/lint/rules/func-name-matching/evidence/final.log 2>&1
go test ./stage1/cohere/lint/rules/func-name-matching -run '^TestRulesAndSuggestions$' -count=1 -v -timeout=5m > .../evidence/final-fixtures.log 2>&1
go test ./stage1/cohere/lint/rules/func-name-matching -run '^TestDecodedOptionCorners$' -count=1 -v -timeout=5m > .../evidence/option-corners.log 2>&1
```

The compiler checkout is pinned to 050880ce59e30b356b686bd3144efe24f875ebc8. All 77 compiler sources and every .ts/.a stage1 source are included, without exclusions. The throughput test uses the identical corpus, best of three unsanitized native/Node/Go runs, includes parsing and traversal, and first verifies equal finding counts. Zero findings means zero findings/s, not proof that a rule never fires: the positive fixtures and semantic mutants are separate checks.

## Toolchain and shared gate

nproc reports 5; cpu.max is 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0. The earlier authorized bash cloud/setup.sh run reached go/clang/node/submodules in 0s each, then cache warming failed with profile_test.go:32:23: cannot range over portFiles (value of type func(t *testing.T) []string). Its exact log is in ../typescript-no-non-null-asserted-optional-chain/evidence/setup.log. /opt/adamic-tools/env.sh does not exist here; sourcing /workspace/adamic-tools/env.sh works. Owned packages bypass the unrelated shared package compilation without patching it. A full gate is not claimed.

## Mutants and superseded failures

- func-name-matching: reverse the always comparison; a named bar assigned to foo becomes silent.
- consistent-this: reverse the plain '=' condition; a valid that=this assignment reports.
- Tailwind: remove direction-aware exemption; rtl:ml-4 produces an extra finding.
- require-description: reverse the non-empty reason guard; a bare directive becomes silent.
- Google Fonts: omit block from the discouraged values; the projected block href becomes silent.

The first option-casing run failed: shared Settings lowercases keys, so camel-case flag lookups missed the CommonJS option. Corrected locally. The initial consistent-this mutant witness had only incorrect aliases and survived; it now includes a valid plain assignment. The first long corpus attempt overlapped live source corrections and saw differing fixed-source text; it is superseded by the final frozen-source run, not counted as parity. Relevant failed logs are retained, labeled as such. The final empty-name corrections have their own independent Go comparison and fixture rerun.

## Final results and commits

The complete owned gate passed in 341.324s. Independent fixtures: 82,729 identical bytes over 363 cases. Compiler/stage1: 62,099,842 identical bytes over 233 files and 1,165 selected rule/file pairs, in 133.86s. Projected JSX semantics: 4,925 identical bytes over 17 original cases; parser refusals pass. All five semantic mutants pass in 81.63s. Final fixture rerun passes in 24.254s; empty-name option corners pass with 1,012 identical bytes in 17.409s. The filtered uncached input oracle passes in 1.073s with six probe misses; bounded vet exits zero with an empty log.

| Rule | Findings | Native seconds / findings per second | Node seconds / findings per second | Go seconds / findings per second |
|---|---:|---:|---:|---:|
| func-name-matching | 0 | 1.336459 / 0.00 | 0.878433 / 0.00 | 0.224532 / 0.00 |
| consistent-this | 0 | 1.459725 / 0.00 | 0.902947 / 0.00 | 0.221726 / 0.00 |
| Tailwind physical direction | 7 | 1.311276 / 5.34 | 0.955975 / 7.32 | 0.210966 / 33.18 |
| require-description | 125 | 9.342320 / 13.38 | 2.956863 / 42.27 | 0.351626 / 355.49 |
| Google Fonts display | 0 | 1.279285 / 0.00 | 1.008265 / 0.00 | 0.219704 / 0.00 |

- consistent-this: 653204e6
- func-name-matching: d4808055
- structure/tailwind-no-physical-direction: 54bb0385
- @eslint-community/eslint-comments/require-description: c41967f2
- @next/next/google-font-display: aad17d79
