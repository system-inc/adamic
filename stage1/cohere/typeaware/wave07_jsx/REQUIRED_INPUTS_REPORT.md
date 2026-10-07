Built: rebased twelve existing ports onto d65a8f93 and supplied pinned external correctness inputs; no new claims.
Commits: previous pushed tip fd393c8b6; tested source commit is pinned in the refreshed evidence metadata.
Checks: 18 owned landing steps and 16 external root tests pass, with 100 passing test events, zero skips and zero failures.
Mutants: all twelve rule, three named-dispatch, nine listener, graph/fact/data/ownership guards and external comparison witnesses are caught.
Not covered: three parked React analyses, four production bridge routes, shared checker context, own emitted-JavaScript comparison and the full repository gate.

Latest rebase and rerun: [LATEST_LANDING_REPORT.md](LATEST_LANDING_REPORT.md).

## Rebase and source scope

The branch was rebased cleanly onto origin/area/stage1-lint at
d65a8f931c98655936ae04c6899f38f14862b73e, containing current main
39638d9e278d38bb5aeae887f46d55a70e47aaad. Incoming runtime allocation,
string builder and reverse-search changes, their tests and their evidence were
retained. No shared or protected source was edited. Existing rule bodies remain
unchanged. Only codex/typeaware-wave-07 is pushed.

The owned gate adds TestRuntimeLastIndexOfMatchesNode alongside the existing
31 Node fixtures, one-byte oracle mutant and iterator refusals. Native release,
ASan/UBSan/LSan and emitted JavaScript agree with Node on the new reverse-search
fixture. Targeted TestRuntimeReleasePaths and TestRuntimeStringEquality also
pass ordinary and sanitizer builds. All 18 owned steps exit 0; exact commands
and output are archived in evidence/landing.

## Required external inputs, with no selected skips

required_inputs.py installs nothing and changes no shared harness. It verifies
exact library versions and the corpus tag, sets every selected external-library
input, runs the existing upstream checks through go test -p 2 -json -count=1
-timeout 30m with an explicit test filter, and rejects any skip or failed event.
The npm packages were installed separately into /workspace/wave-07-required-oracles
with --ignore-scripts --no-audit --no-fund. Repository dependencies were untouched.
The supplied corpus is v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.

| Input | Pinned version |
| --- | --- |
| prettier | 3.9.6 |
| postcss | 8.5.16 |
| postcss-scss | 4.0.9 |
| graphql | 17.0.2 |
| postcss-selector-parser | 2.2.3 |
| postcss-values-parser | 2.0.1 |
| postcss-media-query-parser | 0.2.3 |

The receipt records 16 passing root tests across eleven packages, 100 passing
test events including subtests, zero skips and zero failures. This is a filtered
correctness gate, not a claim that the complete repository gate ran. Benchmarks
and artifact-only checks outside this filter are not represented as tested.
The complete argv, input environment, versions, package lockfile and every Go
JSON event are archived in evidence/required-inputs.

| Package | Selected root check |
| --- | --- |
| stage1/typescript/parser | TestCompilerExpressionsAgree |
| stage1/typescript/scanner | TestScannerAgreesWithTypescriptGo |
| stage1/cohere/css | TestThePortParsesAsGoCohereDoes |
| stage1/cohere/lint | TestCompilerAndStage1Agree |
| stage1/cohere/cssstrings | TestCSSStrings |
| stage1/cohere/cssnumbers | TestCSSNumbers |
| stage1/cohere/graphql | TestThePortParsesAsGoCohereDoes |
| stage1/cohere/selector | TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars |
| stage1/cohere/selector | TestThePortParsesAsGoCohereDoes |
| stage1/cohere/values | TestThePortParsesAsGoCohereDoes |
| stage1/cohere/mediaquery | TestThePortParsesAsGoCohereDoes |
| stage1/cohere/json | TestUpstreamNumericSeparatorGap |
| stage1/cohere/json | TestExternalComparisonCatchesThreePrinterMutants |
| stage1/cohere/css | TestCSSPrinterAgreesWithGo |
| stage1/cohere/css | TestCSSPrinterBoundaryProofs |
| stage1/cohere/json | TestUpstreamRepositoryCorpusParity |

Observed TypeScript comparisons: compiler expressions cover 77 files and
28,836,875 identical bytes; scanner covers 30,671,195 identical answer bytes.
CSS, GraphQL, selector, values, media-query and Prettier comparisons use their
pinned actual libraries. CSS string/number checks no longer omit their external
library leg. Existing known upstream JSON gaps are checked by their unchanged
oracles, not relaxed. Every selected test ran; none was deleted or disabled.

The external log explicitly enumerates 60 passing mutant/catch test names in
mutant-witnesses.json, with detector messages in events.jsonl. Shared emitted
JavaScript and second-suggestion-edit mutants were also caught by the focused
shared-model gate. Existing rule and bridge mutant details are in
AREA_LANDING_REPORT.md; their fresh command streams are archived alongside it.
Every rule mutant compiles, runs successfully with empty stderr and disagrees
with production Go bytes. The three JSX dispatch mutants change a real listener
to Unknown and miss findings. Nine declaration mutations disagree with Go's
listener keys. Graph and four fact mutants disagree with findings/suggestions;
three retained-program registry mutants miss required panic 70. The valid library
copyright data mutation disagrees with the independent external-byte oracle.

## Refreshed type-aware evidence

Original trio: 283 controls, 173 findings and 47,635 identical bytes, ordinary
and sanitized. Its repository corpus matches 16 findings / 22,753 bytes and
compiler matches 24 / 9,282. The next six ports pass their existing positive
controls, option variants, corpora, sanitizers and mutants, including nonempty
regex suggestions. Four bridge questions succeed live and refuse after release;
unregistered production routes are still explicitly refused.

JSX: 131 controls, default 65 findings / 28,512 bytes; element 48 / 24,534;
globals 63 / 27,829; both 46 / 23,851. Every option combination matches under
sanitizers. Repository corpus is 18,485 bytes; compiler is 5,318, both zero
findings. Three message and three named-dispatch mutants are caught again.
Production Go listener keys match declarations/manifests on 184 bytes; earlier
nine listener keys match on 358 bytes. All 114 library sources match 3,793,522
external bytes under sanitizers, and the valid library data mutant is caught.

## Quiet native versus Go timings

Three separate-process pairs per JSX rule after both gates and all heavy work
finished. Every timed output also agrees byte for byte. Times include startup,
checker creation, parsing and handling; no handler-only speedup is inferred.

| Corpus | Rule alias | Native seconds | Go seconds | Native/Go |
| --- | --- | ---: | ---: | ---: |
| repository | fragments | 0.344452 | 0.174185 | 1.98 |
| repository | undef | 0.339348 | 0.149408 | 2.27 |
| repository | adjacent | 0.340341 | 0.155188 | 2.19 |
| compiler | fragments | 2.084074 | 0.313460 | 6.65 |
| compiler | undef | 1.989083 | 0.289765 | 6.86 |
| compiler | adjacent | 2.027918 | 0.296696 | 6.84 |

## Setup and remaining scope

Setup timing lines: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready
0s; build cache warm 94s; done 94s. nproc is 5, with cgroup cpu.max 400000 100000
and 17.6 GB memory. Source /workspace/adamic-tools/env.sh before commands.
All test output went to logs, with no test pipeline. No correctness failure was
observed, so there is no failing input to suppress or repair in this unit.

The three React hook claims remain parked on native high-level IR, SSA/value
flow and capture analysis pending #dnv6f2c. Four production bridge dispatcher
routes remain private overlays under shared-file ownership restrictions. The
shared RuleContext still has no checker-program handle, so standalone type-aware
handlers are not represented as registered there. Own type-aware emitted-JavaScript
comparison and the full repository gate were not run. Shared harness .a support
and suggestion serialization are present and their focused tests pass.
No new rules were claimed before the final post-push ranking audit.
