Built: the same latent measurement on cumulative-2 and cumulative-2 plus newest nested functions, with adaptations 10/20/30/45/46.
Tonight: both configurations have 26/78 checker-clean files (33.33%), up from 19/78; measured on a checker-rejected program.
Commands/results: both builds and runs exit 0; checker total 1424 each, delta -561 against previous cumulative, measured on a checker-rejected program.
Mutants: extra NotYet on the real corpus, output guards, body scope/misattribution, checker ratios, variance ownership and report artifacts tested and caught.
Limits: zero own-file diagnostics is the score, not project acceptance; first lowering error per unit remains; no native or full-gate correctness claim.

# Tonight's number

**26/78 = 33.33%** for both configurations,
**measured on a checker-rejected program**. The previous cumulative run was
19/78 = 24.36%. The increase is seven files.
A checker-clean file has zero diagnostics attributed to its own source in the same
whole-project checker run as before. The 78-root project still has checker errors;
this score does not imply that any file's complete imported program compiles.
No message has been sent to @system_adamic or another worker.

# Method and provenance

All checker totals, per-file/reason counts, ratios, unit counts, variance counts and
deltas here and in JSON are **measured on a checker-rejected program**. The method
is unchanged: one whole-project checker run, diagnose bodies by exact byte overlap,
skip/count top-level functions with their own body diagnostics, measure every other
function and every top-level statement independently, scan refusal syntax past its
first failure, and deduplicate `(kind, location, reason, exact text)` across attempts.
Checker-diagnosed dependencies remain SkippedDependency boundaries, separate from
NotYet/Refused. Ordinary errors/panics remain separate. Source declarations were
added as metadata only; the lowering and eligibility algorithms are unchanged.

The original driver, production Load/LoadOverlay, and lower.Lower output path remain
disabled in this scratch overlay binary. Every corpus run enables the nil-IR and
loader guards; no emitter/native backend is called. Delivery changes stay under
stage3/census/latent/. Scratch integration commits and branches are never pushed.

The previous cumulative head is `15eb079bc200cb1f3f8137fa762fc53bdbc45f22`, delivered by
`70456b766243ee155104c4c9c6928cc53b7b5640`. Its unchanged raw data is `data/cumulative.jsonl.gz`;
the full previous report is archived as `data/previous-REPORT.md` and
`data/previous-REPORT.json.gz`. Its 10/20-only source hashes remain there.
This rerun's changed adapted-source hashes are recorded in REPORT.json. Deltas against
the previous run combine the new adaptations, checker rulings and compiler commits.
They cannot isolate each adaptation's contribution. Exact reason strings may shift
with changed types; exact locations may shift with inserted source lines.
The nested-vs-cumulative-2 delta uses identical source bytes and differs by the newly
merged nested-functions commits and their scratch conflict resolution.

The remote's default fetch refspec tracks only main. Requested branches were fetched
explicitly. Main at the fetch snapshot has no stage3/adapt directory, so the named
adaptations were taken from their branches. Flags at 246ecc07 includes 4097384;
nested-functions' newest fetched commit is 59fe216b. Cumulative-2 retains the previous
cumulative's older nested-functions implementation; the second run adds its new commits.

| Configuration | Resolved scratch head | Never-pushed branch |
|---|---|---|
| cumulative-2 | 5d6cac54cc4a5165a178c1f5ec4bc3147559edb0 | scratch/latent-cumulative2 |
| cumulative-2-nested | 8231a3258298ba59b21c60dbe265162ff8621869 | scratch/latent-cumulative2-nested |

Pinned newly merged commits:

| Branch | Commit |
|---|---|
| origin/main | e011f8f60899586d6373a5ccb07335ad82cfbf3c |
| origin/codex/stage3-optional-declarations | b7e379e054fc245b5a7d002d5abc62ec5ea42b20 |
| origin/codex/stage3-indexed-reads | ebae9b2f9e58f2ff42264ba8e075b9f117729d4c |
| origin/codex/stage3-regex-captures | 347b594807eba1e37b00a4566921846d1a2b7bb1 |
| origin/codex/fallthrough-and-implicit-returns | 5f77d3315f3a6472ee2a6cc28e23e2f7cb1eb8ff |
| origin/codex/flag-enums | 246ecc073993de6a5f1bcb64346cde7938deec43 |
| origin/codex/nested-functions | 59fe216b148efb2d1bc01b9e080541cad1d53198 |

The starting cumulative already contains taste at aa896b5d, namespaces at ce8a2acf,
and nested functions at b15216da, plus the base pipeline and original census.
All requested commit ancestries were checked before measurement.
Dependency pins remain cohere 715ba94f and typescript-go 8d550c83. Canonical dependency
paths in scratch Go workspaces share the warm cache; this does not change source or
checker options. Exact generated overlay and binary hashes are in JSON.

Scratch fallthrough conflicts preserve taste's label/continue support and enum
ErasableSyntaxOnly=false, while removing NoImplicitReturns/NoFallthroughCasesInSwitch
as ruled by the fallthrough branch. IR retains label and depth fields for the two
representations. The newest enum merge retains namespace initialization (which
already calls enum initialization) and both namespace/enum refusal hooks. The
nested merge combines undefined-only/narrowed-field storage with void storage and
retains the TypeScript notice. Whole compiler/native correctness was not tested;
these integrations are used only by the disabled-output measurement driver.
`data/rerun2/` preserves conflict-resolution source, merge logs, final integration
diffs, and each resolved branch's head. Unknown/new conflicts would require review.

# Adapted tree

`apply.sh` from cumulative-2 discovered setup plus exactly 10, 20, 30, 45 and 46,
in numeric order, on TypeScript 6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. Both compilers use identical adapted bytes.
There are still exactly the same 78 compiler roots, including the generated diagnostic
map. Adaptation 46 is a deliberate fork fix for the empty single-quoted pragma argument,
not an erased soundness assertion. This rerun measures it without revalidating its oracle.
`data/rerun2/apply.log.gz` and `patch-set.md` preserve the exact adapter outputs.

Incremental source edits

Compared with the pinned pristine tree after upstream diagnostic generation.
Rows measure incremental edits; total measures the final tree against pristine.

| Adaptation | Files | Lines added | Lines removed |
|---|---:|---:|---:|
| 00-setup | 0 | 0 | 0 |
| 10-type-imports | 72 | 3719 | 3719 |
| 20-optional-declarations | 24 | 371 | 371 |
| 30-indexed-reads | 9 | 170 | 170 |
| 45-regex-captures | 7 | 16 | 14 |
| 46-fix-pragma-empty-argument | 1 | 1 | 1 |
| **Total** | 74 | 4276 | 4274 |


# Checker, lowering and variance totals

Every number is **measured on a checker-rejected program**. N/R are unique NotYet/
Refused sites, not attempted-unit counts. Dependency skips are excluded from N/R.

| Configuration | Checker | Clean files / 78 | NotYet | Refused | Variance Refused subset | Functions attempted | Bodies skipped | Dependency skips | Errors/panics |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| cumulative | 1985 | 19/78 | 1032 | 1983 | 487 | 2500 | 255 | 25 | 1/0 |
| cumulative-2 | 1424 | 26/78 | 1148 | 2230 | 494 | 2634 | 121 | 5 | 1/0 |
| cumulative-2-nested | 1424 | 26/78 | 1146 | 2230 | 494 | 2634 | 121 | 5 | 1/0 |

Each new configuration records one ordinary error whose text says the checker gave
`src/compiler/parser.ts:1472:9` a declaration no symbol. It exposes no structured
location and is excluded from N/R. No panic was recorded.

# Deltas

Every delta is **measured on a checker-rejected program**. More functions become
eligible when their body diagnostics disappear, so added lowerer findings can be
newly exposed failures rather than regressions. The report makes no causal attribution
of the combined source/compiler delta. The optional-declaration revision is narrower
than the previous revision, and the indexed/regex adaptations discharge additional
obligations, so this is not a pure feature comparison against identical sources.

| Configuration relative to previous cumulative | Checker delta | Clean-file delta | NotYet delta | Refused delta | Variance delta |
|---|---:|---:|---:|---:|---:|
| cumulative-2 | -561 | +7 | +116 | +247 | +7 |
| cumulative-2-nested | -561 | +7 | +114 | +247 | +7 |

Newest nested commits relative to cumulative-2: checker +0, clean files +0, NotYet -2, Refused +0, all **measured on a checker-rejected program**.

Newly checker-clean files (both configurations, **measured on a checker-rejected program**):

- src/compiler/factory/emitHelpers.ts
- src/compiler/factory/parenthesizerRules.ts
- src/compiler/factory/utilities.ts
- src/compiler/transformers/declarations/diagnostics.ts
- src/compiler/transformers/legacyDecorators.ts
- src/compiler/types.ts
- src/compiler/utilitiesPublic.ts

No previously checker-clean file loses that status.

All 26 checker-clean files in both configurations:

- src/compiler/_namespaces/ts.moduleSpecifiers.ts
- src/compiler/_namespaces/ts.performance.ts
- src/compiler/_namespaces/ts.ts
- src/compiler/builderPublic.ts
- src/compiler/builderStatePublic.ts
- src/compiler/corePublic.ts
- src/compiler/factory/baseNodeFactory.ts
- src/compiler/factory/emitHelpers.ts
- src/compiler/factory/nodeChildren.ts
- src/compiler/factory/nodeConverters.ts
- src/compiler/factory/nodeTests.ts
- src/compiler/factory/parenthesizerRules.ts
- src/compiler/factory/utilities.ts
- src/compiler/factory/utilitiesPublic.ts
- src/compiler/performance.ts
- src/compiler/symbolWalker.ts
- src/compiler/transformers/declarations/diagnostics.ts
- src/compiler/transformers/es2016.ts
- src/compiler/transformers/es2019.ts
- src/compiler/transformers/es2021.ts
- src/compiler/transformers/legacyDecorators.ts
- src/compiler/transformers/module/impliedNodeFormatDependent.ts
- src/compiler/transformers/taggedTemplate.ts
- src/compiler/tsbuild.ts
- src/compiler/types.ts
- src/compiler/utilitiesPublic.ts

# Per file

Every cell is **measured on a checker-rejected program**. Checker columns count all
primary diagnostic chains in that file. N/R columns count unique actual diagnostic
sites. JSON includes per-file checker/lowering deltas, reasons, eligibility and exact
finding texts. Zero checker diagnostics are the ratio criterion above.

| File under src/compiler | Previous checker | Cum-2 checker | New nested checker | Previous N/R | Cum-2 N/R | New nested N/R |
|---|---:|---:|---:|---:|---:|---:|
| _namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0 | 0/1 | 0/1 | 0/1 |
| _namespaces/ts.performance.ts | 0 | 0 | 0 | 0/1 | 0/1 | 0/1 |
| _namespaces/ts.ts | 0 | 0 | 0 | 0/73 | 0/73 | 0/73 |
| binder.ts | 16 | 4 | 4 | 3/4 | 3/6 | 3/6 |
| builder.ts | 43 | 40 | 40 | 24/24 | 26/26 | 26/26 |
| builderPublic.ts | 0 | 0 | 0 | 6/0 | 6/0 | 6/0 |
| builderState.ts | 2 | 2 | 2 | 1/15 | 1/15 | 1/15 |
| builderStatePublic.ts | 0 | 0 | 0 | 0/0 | 0/0 | 0/0 |
| checker.ts | 969 | 824 | 824 | 5/17 | 5/13 | 5/13 |
| commandLineParser.ts | 31 | 25 | 25 | 58/73 | 60/73 | 60/73 |
| core.ts | 88 | 7 | 7 | 167/66 | 226/146 | 225/146 |
| corePublic.ts | 0 | 0 | 0 | 0/1 | 0/1 | 0/1 |
| debug.ts | 37 | 36 | 36 | 15/61 | 15/62 | 15/62 |
| diagnosticInformationMap.generated.ts | 1 | 1 | 1 | 1/0 | 1/0 | 1/0 |
| emitter.ts | 45 | 31 | 31 | 11/9 | 11/9 | 11/9 |
| executeCommandLine.ts | 7 | 6 | 6 | 18/20 | 18/20 | 18/20 |
| expressionToTypeNode.ts | 6 | 1 | 1 | 2/1 | 2/1 | 2/1 |
| factory/baseNodeFactory.ts | 0 | 0 | 0 | 1/0 | 1/0 | 1/0 |
| factory/emitHelpers.ts | 5 | 0 | 0 | 3/0 | 5/10 | 5/10 |
| factory/emitNode.ts | 5 | 2 | 2 | 25/7 | 25/8 | 25/8 |
| factory/nodeChildren.ts | 0 | 0 | 0 | 4/2 | 4/2 | 4/2 |
| factory/nodeConverters.ts | 0 | 0 | 0 | 1/11 | 1/11 | 1/11 |
| factory/nodeFactory.ts | 8 | 4 | 4 | 8/7 | 8/6 | 8/6 |
| factory/nodeTests.ts | 0 | 0 | 0 | 0/227 | 0/227 | 0/227 |
| factory/parenthesizerRules.ts | 1 | 0 | 0 | 1/2 | 2/22 | 2/22 |
| factory/utilities.ts | 22 | 0 | 0 | 41/101 | 44/120 | 44/120 |
| factory/utilitiesPublic.ts | 0 | 0 | 0 | 1/2 | 1/2 | 1/2 |
| moduleNameResolver.ts | 37 | 10 | 10 | 50/45 | 53/47 | 53/47 |
| moduleSpecifiers.ts | 9 | 4 | 4 | 9/19 | 11/28 | 11/28 |
| parser.ts | 48 | 3 | 3 | 36/179 | 39/170 | 39/170 |
| path.ts | 8 | 7 | 7 | 29/0 | 29/0 | 29/0 |
| performance.ts | 0 | 0 | 0 | 6/0 | 6/0 | 6/0 |
| performanceCore.ts | 2 | 2 | 2 | 1/1 | 1/1 | 1/1 |
| program.ts | 69 | 67 | 67 | 19/45 | 18/32 | 18/32 |
| programDiagnostics.ts | 6 | 6 | 6 | 0/0 | 0/0 | 0/0 |
| resolutionCache.ts | 22 | 19 | 19 | 10/3 | 10/3 | 10/3 |
| scanner.ts | 24 | 2 | 2 | 28/10 | 32/35 | 31/35 |
| semver.ts | 15 | 4 | 4 | 1/2 | 4/7 | 4/7 |
| sourcemap.ts | 8 | 6 | 6 | 7/20 | 8/21 | 8/21 |
| symbolWalker.ts | 0 | 0 | 0 | 1/7 | 1/7 | 1/7 |
| sys.ts | 58 | 57 | 57 | 16/25 | 16/25 | 16/25 |
| tracing.ts | 30 | 30 | 30 | 3/8 | 3/8 | 3/8 |
| transformer.ts | 9 | 9 | 9 | 2/3 | 2/3 | 2/3 |
| transformers/classFields.ts | 14 | 9 | 9 | 4/3 | 4/3 | 4/3 |
| transformers/classThis.ts | 1 | 1 | 1 | 3/2 | 3/2 | 3/2 |
| transformers/declarations/diagnostics.ts | 1 | 0 | 0 | 1/12 | 2/15 | 2/15 |
| transformers/declarations.ts | 4 | 2 | 2 | 2/8 | 2/8 | 2/8 |
| transformers/destructuring.ts | 13 | 13 | 13 | 6/17 | 6/17 | 6/17 |
| transformers/es2015.ts | 33 | 32 | 32 | 0/0 | 0/0 | 0/0 |
| transformers/es2016.ts | 0 | 0 | 0 | 1/11 | 1/11 | 1/11 |
| transformers/es2017.ts | 8 | 8 | 8 | 1/1 | 1/1 | 1/1 |
| transformers/es2018.ts | 4 | 4 | 4 | 0/0 | 0/0 | 0/0 |
| transformers/es2019.ts | 0 | 0 | 0 | 1/1 | 1/1 | 1/1 |
| transformers/es2020.ts | 11 | 11 | 11 | 0/0 | 0/0 | 0/0 |
| transformers/es2021.ts | 0 | 0 | 0 | 1/0 | 1/0 | 1/0 |
| transformers/esDecorators.ts | 6 | 5 | 5 | 0/0 | 0/0 | 0/0 |
| transformers/esnext.ts | 6 | 6 | 6 | 0/1 | 0/1 | 0/1 |
| transformers/generators.ts | 27 | 27 | 27 | 1/1 | 1/1 | 1/1 |
| transformers/jsx.ts | 5 | 5 | 5 | 0/0 | 0/0 | 0/0 |
| transformers/legacyDecorators.ts | 1 | 0 | 0 | 0/0 | 1/19 | 1/19 |
| transformers/module/esnextAnd2015.ts | 6 | 5 | 5 | 0/0 | 0/0 | 0/0 |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0 | 1/0 | 1/0 | 1/0 |
| transformers/module/module.ts | 5 | 4 | 4 | 0/0 | 0/0 | 0/0 |
| transformers/module/system.ts | 6 | 4 | 4 | 0/0 | 0/0 | 0/0 |
| transformers/namedEvaluation.ts | 1 | 1 | 1 | 14/2 | 14/2 | 14/2 |
| transformers/taggedTemplate.ts | 0 | 0 | 0 | 1/3 | 1/3 | 1/3 |
| transformers/ts.ts | 6 | 3 | 3 | 0/0 | 0/0 | 0/0 |
| transformers/typeSerializer.ts | 5 | 5 | 5 | 0/0 | 0/0 | 0/0 |
| transformers/utilities.ts | 9 | 8 | 8 | 17/47 | 17/47 | 17/47 |
| tsbuild.ts | 0 | 0 | 0 | 1/0 | 1/0 | 1/0 |
| tsbuildPublic.ts | 26 | 33 | 33 | 48/81 | 48/47 | 48/47 |
| types.ts | 10 | 0 | 0 | 0/10 | 0/10 | 0/10 |
| utilities.ts | 104 | 5 | 5 | 227/429 | 254/553 | 254/553 |
| utilitiesPublic.ts | 23 | 0 | 0 | 26/179 | 30/190 | 30/190 |
| visitorPublic.ts | 1 | 1 | 1 | 33/18 | 33/18 | 33/18 |
| watch.ts | 8 | 7 | 7 | 17/55 | 16/33 | 16/33 |
| watchPublic.ts | 6 | 12 | 12 | 6/6 | 6/2 | 6/2 |
| watchUtilities.ts | 4 | 4 | 4 | 6/4 | 6/4 | 6/4 |

# Variance refusals for readonly views

Every number is **measured on a checker-rejected program**. The subset contains
mutable invariance, method-parameter bivariance, contravariant overrides, widened
return overrides and readonly-to-mutable overrides. Classification uses the exact
rule tags/fixes and named override reasons. Enum domain, nominal ancestry, casts
and type predicates are excluded. There are 440 mutable-invariance and 54 method-
parameter-bivariance sites in each run; the other variance families are zero.
Previous cumulative had 422 mutable-invariance and 65 bivariance sites (487 total).

The owning declaration is the nearest named AST declaration containing the refusal's
actual diagnostic byte position: function/method, variable/parameter, property,
interface/type alias, class, enum or namespace. Names are qualified by enclosing named
declarations. The attempting top-level unit is retained separately, so a dependency
failure is charged once to its source declaration, not once to every caller. Owning
declarations can contain several distinct sites/reasons. Arrow/function expressions
without a source name belong to the nearest enclosing named declaration. File scope
is an explicit fallback when no named declaration contains a site.

Stock typescript@6.0.3 independently parses all 78 adapted sources. Every Go overlay
declaration span/name/kind/location matches that AST ledger. The report audit calculates
owners from the stock spans and UTF-16 diagnostic locations, independently of the
reported ownership. Each run has 494 sites in 304 owning declarations. Per-owner
findings, family counts, locations, exact text and attempting units are in JSON.
The previous run did not record declaration spans; its variance total is recounted,
but no previous per-owner breakdown is invented.

| Owning declaration location | Qualified declaration | Cum-2 variance sites | New nested variance sites |
|---|---|---:|---:|
| src/compiler/builder.ts:2452:1 | createRedirectedBuilderProgram | 2 | 2 |
| src/compiler/builder.ts:535:1 | convertOrRepopulateDiagnosticMessageChain | 2 | 2 |
| src/compiler/builder.ts:554:11 | convertOrRepopulateDiagnosticMessageChain / next | 1 | 1 |
| src/compiler/builder.ts:790:1 | removeDiagnosticsOfLibraryFiles | 1 | 1 |
| src/compiler/builder.ts:805:1 | handleDtsMayChangeOfAffectedFile | 1 | 1 |
| src/compiler/builder.ts:841:1 | handleDtsMayChangeOf | 3 | 3 |
| src/compiler/builderState.ts:312:15 | BuilderState / create / useOldState | 1 | 1 |
| src/compiler/builderState.ts:499:5 | BuilderState / getAllFileNames | 1 | 1 |
| src/compiler/builderState.ts:549:5 | BuilderState / getAllFilesExcludingDefaultLibraryFile | 1 | 1 |
| src/compiler/builderState.ts:565:9 | BuilderState / getAllFilesExcludingDefaultLibraryFile / addSourceFile | 1 | 1 |
| src/compiler/commandLineParser.ts:1884:1 | parseListTypeOption | 3 | 3 |
| src/compiler/commandLineParser.ts:2277:1 | parseConfigFileTextToJson | 2 | 2 |
| src/compiler/commandLineParser.ts:2437:1 | convertConfigFileToObject | 1 | 1 |
| src/compiler/commandLineParser.ts:2491:5 | convertToJson / convertObjectLiteralExpressionToJson | 3 | 3 |
| src/compiler/commandLineParser.ts:2538:5 | convertToJson / convertPropertyValueToJson | 5 | 5 |
| src/compiler/commandLineParser.ts:2780:1 | serializeOptionBaseObject | 3 | 3 |
| src/compiler/commandLineParser.ts:2795:19 | serializeOptionBaseObject / value | 1 | 1 |
| src/compiler/commandLineParser.ts:288:14 | optionsForWatch | 3 | 3 |
| src/compiler/commandLineParser.ts:2971:1 | convertToOptionsWithAbsolutePaths | 2 | 2 |
| src/compiler/commandLineParser.ts:2990:1 | convertToOptionValueWithAbsolutePaths | 1 | 1 |
| src/compiler/commandLineParser.ts:3074:11 | parseJsonConfigFileContentWorker / parsedConfig | 1 | 1 |
| src/compiler/commandLineParser.ts:3198:5 | parseJsonConfigFileContentWorker / getProjectReferences | 1 | 1 |
| src/compiler/commandLineParser.ts:3260:1 | handleOptionConfigDirTemplateSubstitution | 1 | 1 |
| src/compiler/commandLineParser.ts:3279:27 | handleOptionConfigDirTemplateSubstitution / listResult | 1 | 1 |
| src/compiler/commandLineParser.ts:3294:5 | handleOptionConfigDirTemplateSubstitution / setOptionValue | 1 | 1 |
| src/compiler/commandLineParser.ts:3517:1 | getExtendsConfigPathOrArray | 1 | 1 |
| src/compiler/commandLineParser.ts:3542:19 | getExtendsConfigPathOrArray / fileName | 1 | 1 |
| src/compiler/commandLineParser.ts:3667:11 | getExtendsConfigPath / resolved | 1 | 1 |
| src/compiler/commandLineParser.ts:3685:1 | getExtendedConfig | 1 | 1 |
| src/compiler/commandLineParser.ts:3777:1 | convertOptionsFromJson | 1 | 1 |
| src/compiler/commandLineParser.ts:3794:1 | createDiagnosticForNodeInSourceFileOrCompilerDiagnostic | 2 | 2 |
| src/compiler/commandLineParser.ts:3801:1 | convertJsonOption | 1 | 1 |
| src/compiler/commandLineParser.ts:3854:11 | validateJsonOptionValue / d | 1 | 1 |
| src/compiler/commandLineParser.ts:637:7 | commandOptionsWithoutBuild | 6 | 6 |
| src/compiler/core.ts:1038:1 | toSorted | 1 | 1 |
| src/compiler/core.ts:1570:11 | createQueue / elements | 1 | 1 |
| src/compiler/core.ts:1783:1 | cast | 1 | 1 |
| src/compiler/core.ts:2377:1 | createGetCanonicalFileName | 1 | 1 |
| src/compiler/core.ts:2531:1 | cartesianProduct | 1 | 1 |
| src/compiler/core.ts:350:1 | sameMap | 1 | 1 |
| src/compiler/core.ts:375:1 | flatten | 2 | 2 |
| src/compiler/core.ts:399:1 | flatMap | 1 | 1 |
| src/compiler/core.ts:725:1 | deduplicate | 4 | 4 |
| src/compiler/core.ts:985:1 | addRange | 1 | 1 |
| src/compiler/debug.ts:1101:9 | Debug / formatControlFlowGraph / renderFlowNode | 1 | 1 |
| src/compiler/debug.ts:187:5 | Debug / shouldAssertFunction | 1 | 1 |
| src/compiler/debug.ts:196:5 | Debug / fail | 1 | 1 |
| src/compiler/debug.ts:206:5 | Debug / failBadSyntaxKind | 1 | 1 |
| src/compiler/debug.ts:230:5 | Debug / assertLessThan | 1 | 1 |
| src/compiler/debug.ts:236:5 | Debug / assertLessThanOrEqual | 1 | 1 |
| src/compiler/debug.ts:242:5 | Debug / assertGreaterThanOrEqual | 1 | 1 |
| src/compiler/debug.ts:273:5 | Debug / assertNever | 1 | 1 |
| src/compiler/debug.ts:349:5 | Debug / assertMissingNode | 1 | 1 |
| src/compiler/debug.ts:789:25 | Debug / enableDebugInfo / value | 1 | 1 |
| src/compiler/debug.ts:962:9 | Debug / formatControlFlowGraph / hasAntecedents | 1 | 1 |
| src/compiler/debug.ts:994:9 | Debug / formatControlFlowGraph / buildGraphNode | 1 | 1 |
| src/compiler/emitter.ts:600:5 | createAddOutput / addOutput | 1 | 1 |
| src/compiler/emitter.ts:605:5 | createAddOutput / getOutputs | 1 | 1 |
| src/compiler/emitter.ts:6374:1 | getEmitListItem | 3 | 3 |
| src/compiler/executeCommandLine.ts:109:15 | countLines / lineCount | 1 | 1 |
| src/compiler/executeCommandLine.ts:908:1 | performCompilation | 1 | 1 |
| src/compiler/factory/emitHelpers.ts:148:1 | createEmitHelperFactory | 4 | 4 |
| src/compiler/factory/emitNode.ts:132:1 | getSourceMapRange | 2 | 2 |
| src/compiler/factory/emitNode.ts:156:11 | setTokenSourceMapRange / tokenSourceMapRanges | 1 | 1 |
| src/compiler/factory/emitNode.ts:183:1 | getCommentRange | 2 | 2 |
| src/compiler/factory/nodeChildren.ts:14:1 | getNodeChildren | 1 | 1 |
| src/compiler/factory/nodeConverters.ts:118:5 | createNodeConverters / convertToObjectAssignmentElement | 3 | 3 |
| src/compiler/factory/nodeConverters.ts:147:5 | createNodeConverters / convertToObjectAssignmentPattern | 1 | 1 |
| src/compiler/factory/nodeConverters.ts:160:5 | createNodeConverters / convertToArrayAssignmentPattern | 1 | 1 |
| src/compiler/factory/nodeConverters.ts:54:5 | createNodeConverters / convertToFunctionBlock | 2 | 2 |
| src/compiler/factory/nodeConverters.ts:63:5 | createNodeConverters / convertToFunctionExpression | 1 | 1 |
| src/compiler/factory/nodeConverters.ts:82:5 | createNodeConverters / convertToClassExpression | 1 | 1 |
| src/compiler/factory/nodeConverters.ts:98:5 | createNodeConverters / convertToArrayAssignmentElement | 2 | 2 |
| src/compiler/factory/nodeFactory.ts:7394:1 | makeSynthetic | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:391:5 | createParenthesizerRules / parenthesizeLeftSideOfAccess | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:412:5 | createParenthesizerRules / parenthesizeOperandOfPostfixUnary | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:417:5 | createParenthesizerRules / parenthesizeOperandOfPrefixUnary | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:422:5 | createParenthesizerRules / parenthesizeExpressionsOfCommaDelimitedList | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:427:5 | createParenthesizerRules / parenthesizeExpressionForDisallowedComma | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:435:5 | createParenthesizerRules / parenthesizeExpressionOfExpressionStatement | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:442:23 | createParenthesizerRules / parenthesizeExpressionOfExpressionStatement / updated | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:463:5 | createParenthesizerRules / parenthesizeConciseBodyOfArrowFunction | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:54:1 | createParenthesizerRules | 4 | 4 |
| src/compiler/factory/parenthesizerRules.ts:676:49 | nullParenthesizerRules / _ | 1 | 1 |
| src/compiler/factory/parenthesizerRules.ts:677:50 | nullParenthesizerRules / _ | 1 | 1 |
| src/compiler/factory/utilities.ts:1280:5 | BinaryExpressionState / enter | 1 | 1 |
| src/compiler/factory/utilities.ts:1294:5 | BinaryExpressionState / left | 1 | 1 |
| src/compiler/factory/utilities.ts:1312:5 | BinaryExpressionState / operator | 1 | 1 |
| src/compiler/factory/utilities.ts:1326:5 | BinaryExpressionState / right | 1 | 1 |
| src/compiler/factory/utilities.ts:1344:5 | BinaryExpressionState / exit | 1 | 1 |
| src/compiler/factory/utilities.ts:1365:5 | BinaryExpressionState / done | 1 | 1 |
| src/compiler/factory/utilities.ts:1507:1 | elideNodes | 1 | 1 |
| src/compiler/factory/utilities.ts:1518:1 | getNodeForGeneratedName | 1 | 1 |
| src/compiler/factory/utilities.ts:1522:13 | getNodeForGeneratedName / node | 1 | 1 |
| src/compiler/factory/utilities.ts:1695:1 | isSyntheticParenthesizedExpression | 1 | 1 |
| src/compiler/factory/utilities.ts:193:15 | createMemberAccessForPropertyName / expression | 1 | 1 |
| src/compiler/factory/utilities.ts:218:15 | createJsxFactoryExpressionFromEntityName / right | 1 | 1 |
| src/compiler/factory/utilities.ts:309:1 | createForOfBindingStatement | 2 | 2 |
| src/compiler/factory/utilities.ts:328:15 | createForOfBindingStatement / updatedExpression | 1 | 1 |
| src/compiler/factory/utilities.ts:334:1 | createExpressionFromEntityName | 2 | 2 |
| src/compiler/factory/utilities.ts:338:15 | createExpressionFromEntityName / right | 1 | 1 |
| src/compiler/factory/utilities.ts:348:1 | createExpressionForPropertyName | 2 | 2 |
| src/compiler/factory/utilities.ts:362:1 | createExpressionForAccessorDeclaration | 3 | 3 |
| src/compiler/factory/utilities.ts:411:1 | createExpressionForPropertyAssignment | 2 | 2 |
| src/compiler/factory/utilities.ts:424:1 | createExpressionForShorthandPropertyAssignment | 2 | 2 |
| src/compiler/factory/utilities.ts:437:1 | createExpressionForMethodDeclaration | 3 | 3 |
| src/compiler/factory/utilities.ts:516:1 | expandPreOrPostfixIncrementOrDecrementExpression | 5 | 5 |
| src/compiler/factory/utilities.ts:685:11 | getExternalHelpersModuleName / emitNode | 1 | 1 |
| src/compiler/factory/utilities.ts:692:11 | hasRecordedExternalHelpers / emitNode | 1 | 1 |
| src/compiler/factory/utilities.ts:786:1 | getLocalNameForExternalImport | 2 | 2 |
| src/compiler/moduleNameResolver.ts:1887:19 | nodeModuleNameResolverWorker / diagnosticState | 1 | 1 |
| src/compiler/moduleNameResolver.ts:1894:19 | nodeModuleNameResolverWorker / diagnosticResult | 1 | 1 |
| src/compiler/moduleNameResolver.ts:1913:5 | nodeModuleNameResolverWorker / tryResolve | 5 | 5 |
| src/compiler/moduleNameResolver.ts:2263:19 | getEntrypointsFromPackageJsonInfo / exportResolutions | 1 | 1 |
| src/compiler/moduleNameResolver.ts:2332:23 | loadEntrypointsFromExportMap / loadEntrypointsFromTargetExports / result | 1 | 1 |
| src/compiler/moduleNameResolver.ts:2619:1 | loadModuleFromExports | 2 | 2 |
| src/compiler/moduleNameResolver.ts:2656:1 | loadModuleFromImports | 4 | 4 |
| src/compiler/moduleNameResolver.ts:3242:1 | tryFindNonRelativeModuleNameInCache | 2 | 2 |
| src/compiler/moduleNameResolver.ts:821:15 | getAutomaticTypeDirectiveNames / typeRoots | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:1197:11 | tryGetModuleNameAsNodeModule / preferences | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:1257:19 | tryGetModuleNameAsNodeModule / tryDirectoryWithPackageJson / importMode | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:276:1 | getModuleSpecifier | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:301:1 | getModuleSpecifierWorker | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:318:1 | tryGetModuleSpecifiersFromCache | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:386:1 | getModuleSpecifiersWithCacheInfo | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:453:1 | computeModuleSpecifiers | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:463:11 | computeModuleSpecifiers / preferences | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:725:11 | forEachFileNameOfModule / importedFileNames | 1 | 1 |
| src/compiler/moduleSpecifiers.ts:920:11 | tryGetModuleNameFromAmbientModule / ambientModuleDeclare | 1 | 1 |
| src/compiler/parser.ts:10087:5 | IncrementalParser / moveElementEntirelyPastChangeRange | 1 | 1 |
| src/compiler/parser.ts:10096:9 | IncrementalParser / moveElementEntirelyPastChangeRange / visitNode | 1 | 1 |
| src/compiler/parser.ts:10246:9 | IncrementalParser / updateTokenPositionsAndMarkElements / visitNode | 1 | 1 |
| src/compiler/parser.ts:10591:1 | processCommentPragmas | 2 | 2 |
| src/compiler/parser.ts:10718:15 | extractPragmas / pragma | 1 | 1 |
| src/compiler/parser.ts:10770:11 | addPragmaForMatch / pragma | 1 | 1 |
| src/compiler/parser.ts:1253:11 | forEachChild / fn | 1 | 1 |
| src/compiler/parser.ts:1399:1 | updateSourceFile | 1 | 1 |
| src/compiler/parser.ts:1603:5 | Parser / parseSourceFile | 2 | 2 |
| src/compiler/parser.ts:1847:5 | Parser / withJSDoc | 1 | 1 |
| src/compiler/parser.ts:1862:5 | Parser / reparseTopLevelAwait | 1 | 1 |
| src/compiler/parser.ts:2498:5 | Parser / parseExpectedMatchingBrackets | 1 | 1 |
| src/compiler/parser.ts:2553:5 | Parser / parseTokenNode | 1 | 1 |
| src/compiler/parser.ts:2560:5 | Parser / parseTokenNodeJSDoc | 1 | 1 |
| src/compiler/parser.ts:2600:5 | Parser / finishNode | 2 | 2 |
| src/compiler/parser.ts:2619:5 | Parser / createMissingNode | 1 | 1 |
| src/compiler/parser.ts:3116:5 | Parser / parseListElement | 1 | 1 |
| src/compiler/parser.ts:4036:5 | Parser / parseParameterWorker | 1 | 1 |
| src/compiler/parser.ts:4268:5 | Parser / parsePropertyOrMethodSignature | 1 | 1 |
| src/compiler/parser.ts:4438:5 | Parser / parseTupleElementType | 2 | 2 |
| src/compiler/parser.ts:4547:5 | Parser / parseImportType | 1 | 1 |
| src/compiler/parser.ts:4794:5 | Parser / parseFunctionOrConstructorTypeToError | 1 | 1 |
| src/compiler/parser.ts:6048:5 | Parser / parseJsxElementOrSelfClosingElementOrFragment | 2 | 2 |
| src/compiler/parser.ts:6133:5 | Parser / parseJsxChild | 1 | 1 |
| src/compiler/parser.ts:6393:5 | Parser / tryReparseOptionalChain | 1 | 1 |
| src/compiler/parser.ts:6415:5 | Parser / parsePropertyAccessExpressionRest | 1 | 1 |
| src/compiler/parser.ts:6505:5 | Parser / parseTaggedTemplateRest | 1 | 1 |
| src/compiler/parser.ts:6699:5 | Parser / parseObjectLiteralElement | 2 | 2 |
| src/compiler/parser.ts:7467:5 | Parser / parseDeclaration | 1 | 1 |
| src/compiler/parser.ts:7502:5 | Parser / parseDeclarationWorker | 1 | 1 |
| src/compiler/parser.ts:7765:5 | Parser / tryParseConstructorDeclaration | 2 | 2 |
| src/compiler/parser.ts:7782:5 | Parser / parseMethodDeclaration | 1 | 1 |
| src/compiler/parser.ts:7851:5 | Parser / parseAccessorDeclaration | 2 | 2 |
| src/compiler/parser.ts:7935:5 | Parser / parseClassStaticBlockDeclaration | 1 | 1 |
| src/compiler/parser.ts:8068:5 | Parser / parseClassElement | 1 | 1 |
| src/compiler/parser.ts:8132:5 | Parser / parseDecoratedExpression | 1 | 1 |
| src/compiler/parser.ts:8373:5 | Parser / parseNamespaceExportDeclaration | 1 | 1 |
| src/compiler/parser.ts:8456:5 | Parser / parseImportAttributes | 1 | 1 |
| src/compiler/parser.ts:9446:13 | Parser / JSDocParser / parseJSDocCommentWorker / parseNestedTypeLiteral | 1 | 1 |
| src/compiler/parser.ts:9575:23 | Parser / JSDocParser / parseJSDocCommentWorker / parseExpressionWithTypeArgumentsForAugments / node | 1 | 1 |
| src/compiler/parser.ts:9610:13 | Parser / JSDocParser / parseJSDocCommentWorker / parseTypedefTag | 1 | 1 |
| src/compiler/parser.ts:9668:13 | Parser / JSDocParser / parseJSDocCommentWorker / parseJSDocTypeNameWithNamespace | 1 | 1 |
| src/compiler/parser.ts:9691:13 | Parser / JSDocParser / parseJSDocCommentWorker / parseCallbackTagParameters | 1 | 1 |
| src/compiler/parser.ts:9947:5 | IncrementalParser / updateSourceFile | 2 | 2 |
| src/compiler/program.ts:1323:1 | getConfigFileParsingDiagnostics | 1 | 1 |
| src/compiler/program.ts:5059:1 | handleNoEmitOptions | 1 | 1 |
| src/compiler/program.ts:520:1 | changeCompilerHostLikeToUseCache | 1 | 1 |
| src/compiler/program.ts:636:1 | getPreEmitDiagnostics | 2 | 2 |
| src/compiler/program.ts:670:15 | formatDiagnostic / { line, character } | 1 | 1 |
| src/compiler/program.ts:712:11 | formatCodeSpan / { line: firstLine, character: firstLineChar } | 1 | 1 |
| src/compiler/program.ts:713:11 | formatCodeSpan / { line: lastLine, character: lastLineChar } | 1 | 1 |
| src/compiler/program.ts:714:11 | formatCodeSpan / lastLineInFile | 1 | 1 |
| src/compiler/program.ts:732:15 | formatCodeSpan / lineStart | 1 | 1 |
| src/compiler/program.ts:733:15 | formatCodeSpan / lineEnd | 1 | 1 |
| src/compiler/program.ts:767:11 | formatLocation / { line: firstLine, character: firstLineChar } | 1 | 1 |
| src/compiler/scanner.ts:1060:9 | createScanner / scanner | 4 | 4 |
| src/compiler/scanner.ts:409:7 | tokenStrings | 1 | 1 |
| src/compiler/scanner.ts:423:7 | regExpFlagCharCodes | 1 | 1 |
| src/compiler/semver.ts:207:1 | VersionRange | 1 | 1 |
| src/compiler/semver.ts:46:1 | Version | 1 | 1 |
| src/compiler/sourcemap.ts:483:9 | decodeMappings / next | 14 | 14 |
| src/compiler/symbolWalker.ts:30:1 | createGetSymbolWalker | 1 | 1 |
| src/compiler/symbolWalker.ts:44:5 | createGetSymbolWalker / getSymbolWalker | 4 | 4 |
| src/compiler/symbolWalker.ts:49:23 | createGetSymbolWalker / getSymbolWalker / type | 1 | 1 |
| src/compiler/symbolWalker.ts:59:25 | createGetSymbolWalker / getSymbolWalker / symbol | 1 | 1 |
| src/compiler/sys.ts:137:11 | setCustomPollingValues / pollingIntervalChanged | 1 | 1 |
| src/compiler/sys.ts:1523:15 | sys / getNodeSystem / nodeSystem | 3 | 3 |
| src/compiler/sys.ts:1618:27 | sys / getNodeSystem / nodeSystem / modulePath | 1 | 1 |
| src/compiler/sys.ts:185:1 | pollWatchedFileQueue | 3 | 3 |
| src/compiler/sys.ts:301:5 | createDynamicPriorityPollingWatchFile / pollLowPollingIntervalQueue | 1 | 1 |
| src/compiler/sys.ts:376:5 | createDynamicPriorityPollingWatchFile / scheduleNextPoll | 1 | 1 |
| src/compiler/sys.ts:773:5 | createDirectoryWatcherSupportingRecursive / onTimerToUpdateChildWatches | 1 | 1 |
| src/compiler/sys.ts:826:5 | createDirectoryWatcherSupportingRecursive / updateChildWatches | 2 | 2 |
| src/compiler/sys.ts:873:9 | createDirectoryWatcherSupportingRecursive / updateChildWatches / addChildDirectoryWatcher | 1 | 1 |
| src/compiler/tracing.ts:199:5 | tracingEnabled / getLocation | 2 | 2 |
| src/compiler/transformer.ts:127:1 | getScriptTransformers | 1 | 1 |
| src/compiler/transformers/classThis.ts:104:1 | injectClassThisAssignmentIfMissing | 2 | 2 |
| src/compiler/transformers/declarations/diagnostics.ts:164:1 | createGetSymbolAccessibilityDiagnosticForNodeName | 1 | 1 |
| src/compiler/transformers/declarations/diagnostics.ts:719:5 | createGetIsolatedDeclarationErrors / createAccessorTypeError | 2 | 2 |
| src/compiler/transformers/declarations/diagnostics.ts:733:5 | createGetIsolatedDeclarationErrors / addParentDeclarationRelatedInfo | 1 | 1 |
| src/compiler/transformers/declarations/diagnostics.ts:751:5 | createGetIsolatedDeclarationErrors / createReturnTypeError | 1 | 1 |
| src/compiler/transformers/declarations/diagnostics.ts:760:5 | createGetIsolatedDeclarationErrors / createVariableOrPropertyError | 1 | 1 |
| src/compiler/transformers/declarations/diagnostics.ts:766:5 | createGetIsolatedDeclarationErrors / createParameterError | 1 | 1 |
| src/compiler/transformers/declarations/diagnostics.ts:790:5 | createGetIsolatedDeclarationErrors / createExpressionError | 3 | 3 |
| src/compiler/transformers/destructuring.ts:107:9 | flattenDestructuringAssignment / location | 1 | 1 |
| src/compiler/transformers/destructuring.ts:244:1 | flattenDestructuringBinding | 2 | 2 |
| src/compiler/transformers/destructuring.ts:559:15 | createDestructuringPropertyAccess / argumentExpression | 1 | 1 |
| src/compiler/transformers/destructuring.ts:99:1 | flattenDestructuringAssignment | 3 | 3 |
| src/compiler/transformers/es2016.ts:116:5 | transformES2016 / visitExponentiationExpression | 1 | 1 |
| src/compiler/transformers/es2016.ts:60:5 | transformES2016 / visitExponentiationAssignmentExpression | 9 | 9 |
| src/compiler/transformers/legacyDecorators.ts:183:5 | transformLegacyDecorators / transformDecoratorsOfClassElements | 1 | 1 |
| src/compiler/transformers/legacyDecorators.ts:236:5 | transformLegacyDecorators / transformClassDeclarationWithClassDecorators | 2 | 2 |
| src/compiler/transformers/legacyDecorators.ts:430:5 | transformLegacyDecorators / finishClassElement | 1 | 1 |
| src/compiler/transformers/legacyDecorators.ts:502:5 | transformLegacyDecorators / visitParameterDeclaration | 1 | 1 |
| src/compiler/transformers/legacyDecorators.ts:717:5 | transformLegacyDecorators / transformDecoratorsOfParameter | 1 | 1 |
| src/compiler/transformers/legacyDecorators.ts:823:5 | transformLegacyDecorators / trySubstituteClassAlias | 2 | 2 |
| src/compiler/transformers/namedEvaluation.ts:179:1 | injectClassNamedEvaluationHelperBlockIfMissing | 2 | 2 |
| src/compiler/transformers/taggedTemplate.ts:110:1 | getRawLiteral | 1 | 1 |
| src/compiler/transformers/utilities.ts:179:1 | collectExternalModuleInfo | 1 | 1 |
| src/compiler/transformers/utilities.ts:303:5 | collectExternalModuleInfo / addExportedNamesForExportDeclaration | 1 | 1 |
| src/compiler/transformers/utilities.ts:332:5 | collectExternalModuleInfo / addExportedFunctionDeclaration | 2 | 2 |
| src/compiler/transformers/utilities.ts:426:23 | IdentifierNameMap / toKey / baseName | 3 | 3 |
| src/compiler/transformers/utilities.ts:829:1 | getPrivateIdentifier | 1 | 1 |
| src/compiler/transformers/utilities.ts:839:1 | setPrivateIdentifier | 1 | 1 |
| src/compiler/transformers/utilities.ts:880:1 | rewriteModuleSpecifier | 1 | 1 |
| src/compiler/tsbuildPublic.ts:1351:1 | getOldProgram | 2 | 2 |
| src/compiler/tsbuildPublic.ts:2201:1 | watchPackageJsonFiles | 1 | 1 |
| src/compiler/tsbuildPublic.ts:331:1 | getCompilerOptionsOfBuildOptions | 1 | 1 |
| src/compiler/tsbuildPublic.ts:608:1 | createBuildOrder | 2 | 2 |
| src/compiler/tsbuildPublic.ts:622:5 | createBuildOrder / visit | 2 | 2 |
| src/compiler/tsbuildPublic.ts:659:1 | createStateBuildOrder | 9 | 9 |
| src/compiler/tsbuildPublic.ts:751:11 | enableCache / {<br>        originalReadFile,<br>        originalFileExists,<br>        originalDirectoryExists,<br>        originalCreateDirectory,<br>        originalWriteFile,<br>        getSourceFileWithCache,<br>        readFileWithCache,<br>    } | 1 | 1 |
| src/compiler/utilities.ts:10002:11 | getSupportedExtensions / extensions | 1 | 1 |
| src/compiler/utilities.ts:10015:1 | getSupportedExtensionsWithJsonIfResolveJsonModule | 2 | 2 |
| src/compiler/utilities.ts:10053:59 | usesExtensionsOnImports / hasExtension | 3 | 3 |
| src/compiler/utilities.ts:10095:15 | getModuleSpecifierEndingPreference / inferPreference / specifiers | 1 | 1 |
| src/compiler/utilities.ts:10124:1 | getRequiresAtTopOfFile | 1 | 1 |
| src/compiler/utilities.ts:10230:1 | tryParsePatterns | 1 | 1 |
| src/compiler/utilities.ts:1024:11 | nodePosToString / loc | 1 | 1 |
| src/compiler/utilities.ts:10644:1 | setTextRangePos | 1 | 1 |
| src/compiler/utilities.ts:10654:1 | setTextRangeEnd | 1 | 1 |
| src/compiler/utilities.ts:10687:1 | setNodeFlags | 1 | 1 |
| src/compiler/utilities.ts:10703:1 | setParent | 1 | 1 |
| src/compiler/utilities.ts:10722:1 | setParentRecursive | 1 | 1 |
| src/compiler/utilities.ts:11322:5 | createEvaluator / evaluate | 20 | 20 |
| src/compiler/utilities.ts:11424:5 | createEvaluator / evaluateTemplateExpression | 2 | 2 |
| src/compiler/utilities.ts:12172:1 | forEachDynamicImportOrRequireCall | 2 | 2 |
| src/compiler/utilities.ts:12328:1 | getSynthesizedDeepCloneWithReplacements | 1 | 1 |
| src/compiler/utilities.ts:12345:1 | getSynthesizedDeepCloneWorker | 2 | 2 |
| src/compiler/utilities.ts:12346:11 | getSynthesizedDeepCloneWorker / nodeClone | 1 | 1 |
| src/compiler/utilities.ts:12356:15 | getSynthesizedDeepCloneWorker / clone | 2 | 2 |
| src/compiler/utilities.ts:12374:1 | getSynthesizedDeepClones | 1 | 1 |
| src/compiler/utilities.ts:1238:1 | getTokenPosOfNode | 2 | 2 |
| src/compiler/utilities.ts:1276:1 | getNonDecoratorTokenPosOfNode | 1 | 1 |
| src/compiler/utilities.ts:1384:14 | getScriptTargetFeatures | 17 | 17 |
| src/compiler/utilities.ts:2036:1 | canUseOriginalText | 1 | 1 |
| src/compiler/utilities.ts:2525:1 | getErrorSpanForArrowFunction | 1 | 1 |
| src/compiler/utilities.ts:2528:15 | getErrorSpanForArrowFunction / { line: startLine } | 1 | 1 |
| src/compiler/utilities.ts:2529:15 | getErrorSpanForArrowFunction / { line: endLine } | 1 | 1 |
| src/compiler/utilities.ts:4381:1 | tryGetImportFromModuleSpecifier | 2 | 2 |
| src/compiler/utilities.ts:4658:1 | getJSDocCommentsAndTags | 1 | 1 |
| src/compiler/utilities.ts:4750:1 | getParameterSymbolFromJSDoc | 1 | 1 |
| src/compiler/utilities.ts:6060:1 | createDiagnosticCollection | 1 | 1 |
| src/compiler/utilities.ts:6124:5 | createDiagnosticCollection / getDiagnostics | 1 | 1 |
| src/compiler/utilities.ts:6129:15 | createDiagnosticCollection / getDiagnostics / fileDiags | 1 | 1 |
| src/compiler/utilities.ts:6207:1 | escapeString | 1 | 1 |
| src/compiler/utilities.ts:625:1 | getDeclarationOfKind | 2 | 2 |
| src/compiler/utilities.ts:639:1 | getDeclarationsOfKind | 2 | 2 |
| src/compiler/utilities.ts:6750:11 | getLineOfLocalPosition / lineStarts | 1 | 1 |
| src/compiler/utilities.ts:7879:1 | moveRangePastDecorators | 3 | 3 |
| src/compiler/utilities.ts:7891:1 | moveRangePastModifiers | 2 | 2 |
| src/compiler/utilities.ts:7944:1 | getLinesBetweenRangeEndAndRangeStart | 1 | 1 |
| src/compiler/utilities.ts:7950:1 | getLinesBetweenRangeEndPositions | 1 | 1 |
| src/compiler/utilities.ts:7960:1 | positionsAreOnSameLine | 1 | 1 |
| src/compiler/utilities.ts:7970:1 | getLinesBetweenPositionAndPrecedingNonWhitespaceCharacter | 1 | 1 |
| src/compiler/utilities.ts:7977:1 | getLinesBetweenPositionAndNextNonWhitespaceCharacter | 1 | 1 |
| src/compiler/utilities.ts:8583:1 | formatStringFromArgs | 1 | 1 |
| src/compiler/utilities.ts:8641:1 | attachFileToDiagnostic | 1 | 1 |
| src/compiler/utilities.ts:9003:19 | getSetExternalModuleIndicator / combined | 1 | 1 |
| src/compiler/utilities.ts:9399:1 | getCompilerOptionValue | 2 | 2 |
| src/compiler/utilities.ts:959:1 | aggregateChildData | 2 | 2 |
| src/compiler/utilities.ts:9832:11 | matchFiles / includeFileRegexes | 1 | 1 |
| src/compiler/utilities.ts:9838:11 | matchFiles / results | 1 | 1 |
| src/compiler/utilities.ts:9993:1 | getSupportedExtensions | 1 | 1 |
| src/compiler/utilitiesPublic.ts:766:1 | getOriginalNode | 3 | 3 |
| src/compiler/visitorPublic.ts:474:1 | addDefaultValueAssignmentForInitializer | 2 | 2 |
| src/compiler/visitorPublic.ts:598:1 | visitEachChild | 2 | 2 |
| src/compiler/visitorPublic.ts:603:11 | visitEachChild / fn | 1 | 1 |
| src/compiler/watch.ts:334:1 | isBuilderProgram | 1 | 1 |
| src/compiler/watch.ts:564:1 | emitFilesAndReportErrors | 1 | 1 |
| src/compiler/watch.ts:673:1 | createWatchHost | 4 | 4 |
| src/compiler/watch.ts:827:1 | setGetSourceFileAsHashVersioned | 1 | 1 |
| src/compiler/watchPublic.ts:151:11 | createIncrementalProgram / oldProgram | 1 | 1 |
| src/compiler/watchUtilities.ts:414:11 | updateSharedExtendedConfigFileWatcher / extendedConfigs | 1 | 1 |
| src/compiler/watchUtilities.ts:629:11 | isIgnoredFileFromWildCardWatching / builderProgram | 1 | 1 |

# Per checker reason

Every number is **measured on a checker-rejected program**.

| TS code | Previous | Cum-2 | New nested |
|---|---:|---:|---:|
| TS1484 | 1 | 1 | 1 |
| TS18046 | 8 | 8 | 8 |
| TS18048 | 380 | 353 | 353 |
| TS2304 | 6 | 6 | 6 |
| TS2307 | 1 | 1 | 1 |
| TS2320 | 6 | 0 | 0 |
| TS2322 | 119 | 97 | 97 |
| TS2339 | 28 | 28 | 28 |
| TS2345 | 738 | 605 | 605 |
| TS2375 | 15 | 19 | 19 |
| TS2379 | 15 | 14 | 14 |
| TS2412 | 6 | 27 | 27 |
| TS2420 | 1 | 1 | 1 |
| TS2430 | 10 | 0 | 0 |
| TS2488 | 11 | 11 | 11 |
| TS2532 | 225 | 177 | 177 |
| TS2538 | 7 | 5 | 5 |
| TS2556 | 2 | 2 | 2 |
| TS2591 | 54 | 54 | 54 |
| TS2684 | 2 | 2 | 2 |
| TS2722 | 2 | 1 | 1 |
| TS2740 | 1 | 1 | 1 |
| TS2769 | 10 | 9 | 9 |
| TS7006 | 1 | 1 | 1 |
| TS7029 | 83 | 0 | 0 |
| TS7030 | 252 | 0 | 0 |
| TS7031 | 1 | 1 | 1 |

# Per NotYet/Refused reason

Every number is **measured on a checker-rejected program**. These are the exact
reason families as before, including concrete types. SkippedDependency and ordinary
errors are preserved separately. Their exact text and every site are in JSON.

| Reason | Previous cumulative | Cum-2 | New nested |
|---|---:|---:|---:|
| NotYet: .length on a value | 1 | 1 | 1 |
| NotYet: ?.[] on a value | 1 | 1 | 1 |
| NotYet: Object.assign on a shape not proven by a plain literal or its const binding | 1 | 1 | 1 |
| NotYet: Object.entries on a shape not proven by a plain literal or its const binding | 3 | 3 | 3 |
| NotYet: RegExp with a nonconstant pattern | 1 | 1 | 1 |
| NotYet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | 1 | 1 | 1 |
| NotYet: a BinaryExpression with a number and a number | 1 | 1 | 1 |
| NotYet: a BinaryExpression with a string and a number | 2 | 2 | 2 |
| NotYet: a BinaryExpression with a union of differently held members and a union of differently held members | 1 | 1 | 1 |
| NotYet: a BinaryExpression with a value and a value | 8 | 8 | 8 |
| NotYet: a ClassExpression | 1 | 1 | 1 |
| NotYet: a Map of T | 1 | 1 | 1 |
| NotYet: a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions | 7 | 7 | 7 |
| NotYet: a NonNullExpression | 25 | 70 | 70 |
| NotYet: a PostfixUnaryExpression | 1 | 1 | 1 |
| NotYet: a PrefixUnaryExpression on a number | 1 | 1 | 1 |
| NotYet: a SpreadElement | 1 | 1 | 1 |
| NotYet: a YieldExpression as a statement | 0 | 1 | 1 |
| NotYet: a boolean \| undefined variable a function value captures | 1 | 1 | 1 |
| NotYet: a call returning T | 1 | 1 | 1 |
| NotYet: a call through ?. (an optional call) | 13 | 13 | 13 |
| NotYet: a case whose type differs from the switch's | 3 | 3 | 3 |
| NotYet: a computed field name | 4 | 4 | 4 |
| NotYet: a declaration directly in a case (wrap the case in a block) | 2 | 3 | 4 |
| NotYet: a destructured parameter beside a parameter with a default | 1 | 1 | 1 |
| NotYet: a field from a boolean \| undefined variable | 1 | 1 | 1 |
| NotYet: a field holding union of differently held members | 3 | 3 | 3 |
| NotYet: a field of type "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number> | 2 | 2 | 2 |
| NotYet: a field of type AnyBuildOrder \| undefined | 1 | 1 | 1 |
| NotYet: a field of type NodeArray<ParameterDeclaration> \| readonly JSDocParameterTag[] | 0 | 1 | 1 |
| NotYet: a field of type boolean \| (() => boolean) \| undefined | 2 | 2 | 2 |
| NotYet: a field of type boolean \| undefined | 37 | 37 | 37 |
| NotYet: a field of type false \| VersionPaths \| undefined | 1 | 1 | 1 |
| NotYet: a field of type false \| string[] \| undefined | 1 | 1 | 1 |
| NotYet: a field of type string \| DiagnosticMessageChain | 6 | 6 | 6 |
| NotYet: a field of type string \| false \| undefined | 1 | 1 | 1 |
| NotYet: a field of type string \| number \| undefined | 1 | 1 | 1 |
| NotYet: a field of type true \| Node \| undefined | 2 | 2 | 2 |
| NotYet: a field of type true \| undefined | 6 | 6 | 6 |
| NotYet: a function returning (AssignmentExpression<EqualsToken> & { readonly left: GeneratedIdentifier; }) \| undefined | 1 | 1 | 1 |
| NotYet: a function returning (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 |
| NotYet: a function returning AnyValidImportOrReExport | 1 | 1 | 1 |
| NotYet: a function returning AnyValidImportOrReExport \| undefined | 1 | 1 | 1 |
| NotYet: a function returning CanonicalKey | 1 | 1 | 1 |
| NotYet: a function returning ClassNamedEvaluationHelperBlock | 1 | 1 | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 | 1 | 1 |
| NotYet: a function returning CompilerOptionsValue | 3 | 3 | 3 |
| NotYet: a function returning ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 |
| NotYet: a function returning Extract<AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; }, Pick<...>> \| ... 7 more ... \| Extract<...> | 1 | 1 | 1 |
| NotYet: a function returning Extract<ClassDeclaration, Pick<...>> \| Extract<...> | 2 | 2 | 2 |
| NotYet: a function returning HasJSDoc \| undefined | 1 | 1 | 1 |
| NotYet: a function returning MemberName \| (Expression & (NumericLiteral \| StringLiteralLike)) | 1 | 1 | 1 |
| NotYet: a function returning ModeAwareCacheKey | 2 | 2 | 2 |
| NotYet: a function returning Node \| (TIn & undefined) \| (TVisited & undefined) | 1 | 1 | 1 |
| NotYet: a function returning NodeArray<Node> \| (TInArray & undefined) | 1 | 1 | 1 |
| NotYet: a function returning NodeArray<TOut> \| (TInArray & undefined) | 1 | 1 | 1 |
| NotYet: a function returning PackageJson[K] \| undefined | 4 | 4 | 4 |
| NotYet: a function returning Path | 5 | 5 | 5 |
| NotYet: a function returning Path \| undefined | 2 | 2 | 2 |
| NotYet: a function returning PathPathComponents | 1 | 1 | 1 |
| NotYet: a function returning ResolvedConfigFileName | 3 | 3 | 3 |
| NotYet: a function returning ResolvedConfigFilePath | 1 | 1 | 1 |
| NotYet: a function returning T | 48 | 50 | 50 |
| NotYet: a function returning T \| T[] | 1 | 1 | 1 |
| NotYet: a function returning T \| T[] \| undefined | 3 | 3 | 3 |
| NotYet: a function returning T \| readonly T[] | 1 | 1 | 1 |
| NotYet: a function returning T \| readonly T[] \| undefined | 3 | 3 | 3 |
| NotYet: a function returning T \| undefined | 62 | 72 | 72 |
| NotYet: a function returning T1 & T2 | 1 | 1 | 1 |
| NotYet: a function returning TEntry \| undefined | 1 | 1 | 1 |
| NotYet: a function returning TOut | 1 | 1 | 1 |
| NotYet: a function returning TOut \| (TIn & undefined) \| (TVisited & undefined) | 1 | 1 | 1 |
| NotYet: a function returning TOut \| undefined | 1 | 1 | 1 |
| NotYet: a function returning TPrivateEntry \| undefined | 1 | 1 | 1 |
| NotYet: a function returning TResult | 0 | 1 | 1 |
| NotYet: a function returning U | 2 | 2 | 2 |
| NotYet: a function returning U \| undefined | 13 | 17 | 17 |
| NotYet: a function returning V | 2 | 2 | 2 |
| NotYet: a function returning __String | 10 | 10 | 10 |
| NotYet: a function returning __String \| undefined | 4 | 4 | 4 |
| NotYet: a function returning any | 6 | 6 | 6 |
| NotYet: a function returning object | 2 | 2 | 2 |
| NotYet: a function returning object \| undefined | 1 | 1 | 1 |
| NotYet: a function returning readonly Node[] \| (TInArray & undefined) | 1 | 1 | 1 |
| NotYet: a function returning readonly TOut[] \| (TInArray & undefined) | 1 | 1 | 1 |
| NotYet: a function returning string \| object | 1 | 1 | 1 |
| NotYet: a function returning undefined | 1 | 1 | 0 |
| NotYet: a function returning unknown | 1 | 1 | 1 |
| NotYet: a function value returning boolean \| undefined | 2 | 2 | 2 |
| NotYet: a function value returning union of differently held members | 6 | 7 | 7 |
| NotYet: a function value taking string \| DiagnosticMessageChain \| undefined | 1 | 1 | 1 |
| NotYet: a function value taking string \| string[] | 1 | 1 | 1 |
| NotYet: a function value with an optional parameter | 6 | 9 | 9 |
| NotYet: a function with an optional or rest parameter, as a value | 5 | 5 | 5 |
| NotYet: a function without a body | 194 | 195 | 197 |
| NotYet: a generic function as a value | 8 | 15 | 15 |
| NotYet: a generic or unnamed nested function declaration | 0 | 1 | 0 |
| NotYet: a method call through a structural signature in a program with statics; use typeof the declaring class | 110 | 113 | 113 |
| NotYet: a namespace merged with a function; callable object properties, identity and receivers are not represented | 1 | 1 | 1 |
| NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 5 | 5 | 5 |
| NotYet: a parameter that isn't a plain name | 22 | 23 | 23 |
| NotYet: a tagged template other than the intrinsic String.raw | 2 | 2 | 2 |
| NotYet: a template interpolating an object, an array, a map, a function or undefined | 2 | 2 | 2 |
| NotYet: a union of differently held members variable a function value captures | 1 | 1 | 1 |
| NotYet: a value of type "" \| ResolvedConfigFileName \| undefined | 1 | 1 | 1 |
| NotYet: a value of type (AssignmentExpression<EqualsToken> & { readonly left: Identifier; readonly right: WrappedExpression<AnonymousFunctionDefinition>; } & BinaryExpression) \| (... & ... 1 more ... & BinaryExpression) | 1 | 1 | 1 |
| NotYet: a value of type (CompilerHost \| ProgramHost<T>) & { onUnRecoverableConfigFileDiagnostic?: DiagnosticReporter; } | 1 | 0 | 0 |
| NotYet: a value of type (ConstructorDeclaration & { body: Block; }) \| undefined | 1 | 1 | 1 |
| NotYet: a value of type (EmitNode & { autoGenerate: AutoGenerateInfo; }) \| (EmitNode & { autoGenerate: AutoGenerateInfo; }) | 2 | 2 | 2 |
| NotYet: a value of type (ModuleDeclaration & { name: StringLiteral; }) \| undefined | 0 | 1 | 1 |
| NotYet: a value of type AccessExpression \| RequireOrImportCall | 1 | 1 | 1 |
| NotYet: a value of type AccessorDeclaration & { readonly name: BigIntLiteral \| ComputedPropertyName \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral; } | 1 | 1 | 1 |
| NotYet: a value of type BindableStaticNameExpression | 2 | 2 | 2 |
| NotYet: a value of type BindingElement & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type CompilerOptionsValue | 3 | 4 | 4 |
| NotYet: a value of type EmitNode & { autoGenerate: AutoGenerateInfo; } | 1 | 1 | 1 |
| NotYet: a value of type EntityNameExpression \| (LeftHandSideExpression & BindableStaticNameExpression) | 1 | 1 | 1 |
| NotYet: a value of type ExportAssignment & { readonly expression: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type ExpressionWithTypeArguments & { readonly expression: Identifier \| PropertyAccessEntityNameExpression; } | 1 | 1 | 1 |
| NotYet: a value of type HasJSDoc | 2 | 2 | 2 |
| NotYet: a value of type HasJSDoc \| undefined | 1 | 1 | 1 |
| NotYet: a value of type IncludeTypeSpaceImports | 0 | 1 | 1 |
| NotYet: a value of type IncrementalBuildInfoFilePendingEmit | 1 | 1 | 1 |
| NotYet: a value of type IncrementalMultiFileEmitBuildInfoFileInfo | 1 | 1 | 1 |
| NotYet: a value of type JSDocImportTag \| CanHaveModuleSpecifier | 1 | 1 | 1 |
| NotYet: a value of type K | 3 | 3 | 3 |
| NotYet: a value of type NamedEvaluation | 1 | 1 | 1 |
| NotYet: a value of type NodeArray<Expression> & readonly [BindableStaticNameExpression, NumericLiteral \| StringLiteralLike, ObjectLiteralExpression] & Readonly<...> | 1 | 1 | 1 |
| NotYet: a value of type NonNullable<T> | 0 | 5 | 5 |
| NotYet: a value of type ParameterDeclaration & { readonly name: Identifier; readonly dotDotDotToken: undefined; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type Path | 20 | 20 | 20 |
| NotYet: a value of type PropertyAssignment & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type PropertyDeclaration & { readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type RequireOrImportCall | 1 | 1 | 1 |
| NotYet: a value of type ResolvedConfigFileName | 7 | 7 | 7 |
| NotYet: a value of type ResolvedConfigFileName \| undefined | 1 | 1 | 1 |
| NotYet: a value of type ResolvedConfigFilePath | 18 | 18 | 18 |
| NotYet: a value of type ResolvedModuleWithFailedLookupLocations & ResolvedTypeReferenceDirectiveWithFailedLookupLocations | 1 | 1 | 1 |
| NotYet: a value of type ShorthandPropertyAssignment & { readonly objectAssignmentInitializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type SourceFile | 1 | 1 | 1 |
| NotYet: a value of type T | 27 | 30 | 28 |
| NotYet: a value of type T \| Program | 4 | 4 | 4 |
| NotYet: a value of type T \| T[] | 2 | 2 | 2 |
| NotYet: a value of type T \| null \| undefined | 1 | 1 | 1 |
| NotYet: a value of type T \| readonly T[] | 1 | 1 | 1 |
| NotYet: a value of type T \| undefined | 10 | 11 | 11 |
| NotYet: a value of type T1 | 1 | 1 | 1 |
| NotYet: a value of type TData | 1 | 1 | 1 |
| NotYet: a value of type TEntry | 1 | 1 | 1 |
| NotYet: a value of type TInArray | 2 | 2 | 2 |
| NotYet: a value of type T["kind"] | 1 | 1 | 1 |
| NotYet: a value of type TypeNode & LiteralTypeNode & { readonly literal: StringLiteral; } | 1 | 1 | 1 |
| NotYet: a value of type TypeParameterDeclaration & { parent: JSDocTemplateTag; } | 1 | 1 | 1 |
| NotYet: a value of type U | 0 | 1 | 1 |
| NotYet: a value of type U \| readonly U[] \| undefined | 0 | 1 | 1 |
| NotYet: a value of type V | 1 | 1 | 1 |
| NotYet: a value of type VariableDeclaration & { readonly name: Identifier; readonly initializer: WrappedExpression<AnonymousFunctionDefinition>; } | 1 | 1 | 1 |
| NotYet: a value of type WatchFactoryHost & { trace?(s: string): void; } | 1 | 1 | 1 |
| NotYet: a value of type WrappedExpression<AnonymousFunctionDefinition> | 2 | 2 | 2 |
| NotYet: a value of type WrappedExpression<T> | 1 | 1 | 1 |
| NotYet: a value of type __String | 26 | 28 | 28 |
| NotYet: a value of type __String & string | 1 | 1 | 1 |
| NotYet: a value of type any | 31 | 33 | 33 |
| NotYet: a value of type false \| RegExpExecArray \| null | 0 | 1 | 1 |
| NotYet: a value of type never | 1 | 1 | 1 |
| NotYet: a value of type object | 4 | 5 | 5 |
| NotYet: a value of type object \| undefined | 1 | 1 | 1 |
| NotYet: a value of type string \| (void & { __escapedIdentifier: void; }) \| (string & { __escapedIdentifier: void; }) | 3 | 3 | 3 |
| NotYet: a value of type string \| null \| undefined | 2 | 2 | 2 |
| NotYet: a value of type undefined | 1 | 1 | 0 |
| NotYet: a value of type unknown | 11 | 11 | 11 |
| NotYet: a void call used as a value | 8 | 8 | 8 |
| NotYet: an ElementAccessExpression | 6 | 6 | 6 |
| NotYet: an array of T | 6 | 12 | 12 |
| NotYet: an array of T \| U | 1 | 1 | 1 |
| NotYet: an array of U | 0 | 3 | 3 |
| NotYet: an array of V | 0 | 1 | 1 |
| NotYet: an array of never | 6 | 8 | 8 |
| NotYet: an array of unknown | 1 | 1 | 1 |
| NotYet: an enum inside a function or block; declare it at module scope | 3 | 3 | 3 |
| NotYet: an optional chain longer than one step | 4 | 4 | 4 |
| NotYet: assigning a field of a value | 2 | 2 | 2 |
| NotYet: assigning an element of a value | 1 | 1 | 1 |
| NotYet: destructuring inside a namespace; use plain singleton bindings | 1 | 1 | 1 |
| NotYet: for...in over an array (holes and own enumerable properties are not represented; use for...of for elements) | 1 | 1 | 1 |
| NotYet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | 4 | 4 | 4 |
| NotYet: for...of over an object | 11 | 11 | 11 |
| NotYet: lastIndexOf with these arguments | 2 | 2 | 2 |
| NotYet: new Map from something that isn't [key, value] pairs | 1 | 1 | 1 |
| NotYet: new a ParenthesizedExpression | 3 | 3 | 3 |
| NotYet: new an Identifier | 7 | 8 | 8 |
| NotYet: optional chaining to .size on a value | 2 | 4 | 4 |
| NotYet: reading Error | 1 | 1 | 1 |
| NotYet: reading getOptionsNameMap | 2 | 2 | 2 |
| NotYet: regex replacement other than a string | 6 | 6 | 6 |
| NotYet: spreading an array of other elements | 1 | 1 | 1 |
| NotYet: storing any in a field | 1 | 1 | 1 |
| NotYet: storing boolean in a field | 1 | 1 | 1 |
| NotYet: storing string \| number in a field | 1 | 1 | 1 |
| NotYet: storing true \| Node \| undefined in a field | 2 | 2 | 2 |
| NotYet: this outside a method | 7 | 7 | 7 |
| Refused: Object.defineProperty | 1 | 1 | 1 |
| Refused: a cast the runtime can't check | 69 | 70 | 70 |
| Refused: a definite assignment assertion ! | 5 | 11 | 11 |
| Refused: a flag initializer outside the non-negative int32 bound | 3 | 3 | 3 |
| Refused: a function taking () => T seen as one taking () => T (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 3 | 3 |
| Refused: a function taking (node: Node) => T \| undefined seen as one taking (node: Node) => T \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking (symbol: Symbol) => boolean seen as one taking ((symbol: Symbol) => boolean) \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking (value: never, key: never, map: ReadonlyMap<never, never>) => void seen as one taking <TKey extends keyof PragmaPseudoMap>(value: PragmaPseudoMap[TKey][] \| PragmaPseudoMap[TKey], key: TKey, map: ReadonlyPragmaMap) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking BinaryExpressionStateMachine<TOuterState, TState, TResult> seen as one taking BinaryExpressionStateMachine<TOuterState, TState, TResult> (tsc relates a method's parameters both ways), so it can be handed what it can't take | 6 | 6 | 6 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 2 | 2 |
| Refused: a function taking CallExpression \| (IncludeTypeSpaceImports extends false ? never : ImportTypeNode \| JSDocImportTag) seen as one taking CallExpression \| ImportTypeNode \| JSDocImportTag (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 2 | 2 |
| Refused: a function taking Expression seen as one taking Expression \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 1 | 1 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 2 | 2 |
| Refused: a function taking LogLevel seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking Node seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking Node \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking NodeArray<Expression> seen as one taking readonly Expression[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 1 | 1 |
| Refused: a function taking NodeArray<TypeNode> \| undefined seen as one taking readonly TypeNode[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 1 | 1 |
| Refused: a function taking OrdinalParentheizerRuleSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking ParenthesizerRule<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking ParenthesizerRuleOrSelector<Node> \| undefined seen as one taking ParenthesizerRuleOrSelector<T> \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking PollingInterval seen as one taking number \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking SourceFile seen as one taking [file: SourceFile, options: CompilerOptions] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking T seen as one taking Expression (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 |
| Refused: a function taking T seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking TIn seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking [callback: (...args: any[]) => void, ms: number, ...args: any[]] seen as one taking (...args: any[]) => void (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking [data: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 4 | 0 | 0 |
| Refused: a function taking [fileName: string, languageVersionOrOptions: CreateSourceFileOptions \| ScriptTarget, onError?: ((message: string) => void) \| undefined, shouldCreateNewSourceFile?: boolean \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 |
| Refused: a function taking [fileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking [libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking [moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... \| undefined] seen as one taking readonly StringLiteralLike[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking [moduleNames: string[], containingFile: string, reusedNames: string[] \| undefined, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile?: SourceFile \| undefined] seen as one taking string[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking [name: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | 0 |
| Refused: a function taking [path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking [path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking [path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking [path: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 7 | 0 | 0 |
| Refused: a function taking [s: string] seen as one taking string (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 0 | 0 |
| Refused: a function taking [timeoutId: any] seen as one taking unknown (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking [typeDirectiveReferences: readonly (string \| FileReference)[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile \| undefined, reusedNames: ... \| undefined] seen as one taking readonly T[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking [typeReferenceDirectiveNames: string[] \| readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingFileMode?: ResolutionMode] seen as one taking string[] \| readonly FileReference[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking [writeByteOrderMark: boolean, onError?: ((message: string) => void) \| undefined, sourceFiles?: readonly SourceFile[] \| undefined, data?: WriteFileCallbackData \| undefined] seen as one taking boolean (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking boolean seen as one taking boolean \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 2 | 2 |
| Refused: a function taking never seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking number seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a function taking number seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 3 | 3 | 3 |
| Refused: a function taking readonly ParameterDeclaration[] seen as one taking readonly ParameterDeclaration[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 0 | 1 | 1 |
| Refused: a function taking readonly string[] seen as one taking readonly string[] \| undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking string seen as one taking [data: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking string seen as one taking [fileName: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 |
| Refused: a function taking string seen as one taking [name: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking string seen as one taking [path: string] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 0 | 0 |
| Refused: a function taking string seen as one taking any[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 2 | 2 | 2 |
| Refused: a function taking string \| undefined seen as one taking never[] (tsc relates a method's parameters both ways), so it can be handed what it can't take | 1 | 1 | 1 |
| Refused: a generator function | 4 | 5 | 5 |
| Refused: a method in object destructuring | 3 | 3 | 3 |
| Refused: a method read as a value (add would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (afterProgramCreate would lose its object, and this with it) | 3 | 3 | 3 |
| Refused: a method read as a value (afterProgramEmitAndDiagnostics would lose its object, and this with it) | 3 | 3 | 3 |
| Refused: a method read as a value (base64decode would lose its object, and this with it) | 0 | 1 | 1 |
| Refused: a method read as a value (base64encode would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (clearScreen would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (clearTimeout would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (compare would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (convertToArrayAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (convertToObjectAssignmentElement would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createDirectory would lose its object, and this with it) | 1 | 2 | 2 |
| Refused: a method read as a value (createHash would lose its object, and this with it) | 3 | 0 | 0 |
| Refused: a method read as a value (createIntersectionTypeNode would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocClassTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocDeprecatedTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocLink would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocLinkCode would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocLinkPlain would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocOverrideTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocPrivateTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocProtectedTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocPublicTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createJSDocReadonlyTag would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (createUnionTypeNode would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (deleteFile would lose its object, and this with it) | 4 | 1 | 1 |
| Refused: a method read as a value (directoryExists would lose its object, and this with it) | 3 | 3 | 3 |
| Refused: a method read as a value (enableCPUProfiler would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (fileExists would lose its object, and this with it) | 4 | 4 | 4 |
| Refused: a method read as a value (fill would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (getBuildInfo would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (getCanonicalFileName would lose its object, and this with it) | 6 | 2 | 2 |
| Refused: a method read as a value (getCommonSourceDirectory would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (getCurrentDirectory would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (getDefaultLibLocation would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (getDirectories would lose its object, and this with it) | 2 | 1 | 1 |
| Refused: a method read as a value (getEnvironmentVariable would lose its object, and this with it) | 4 | 2 | 2 |
| Refused: a method read as a value (getGlobalTypingsCacheLocation would lose its object, and this with it) | 1 | 2 | 2 |
| Refused: a method read as a value (getModifiedTime would lose its object, and this with it) | 4 | 1 | 1 |
| Refused: a method read as a value (getModuleResolutionCache would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (getNearestAncestorDirectoryWithPackageJson would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (getOrCreateCacheForModuleName would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (getParsedCommandLine would lose its object, and this with it) | 3 | 2 | 2 |
| Refused: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (getSourceFile would lose its object, and this with it) | 4 | 4 | 4 |
| Refused: a method read as a value (globalCacheResolutionModuleName would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (hasOwnProperty would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (liftToBlock would lose its object, and this with it) | 5 | 5 | 5 |
| Refused: a method read as a value (now would lose its object, and this with it) | 4 | 3 | 3 |
| Refused: a method read as a value (onWatchStatusChange would lose its object, and this with it) | 3 | 3 | 3 |
| Refused: a method read as a value (readDirectory would lose its object, and this with it) | 2 | 0 | 0 |
| Refused: a method read as a value (readFile would lose its object, and this with it) | 2 | 5 | 5 |
| Refused: a method read as a value (realpath would lose its object, and this with it) | 4 | 1 | 1 |
| Refused: a method read as a value (remove would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (repeat would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (replace would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportCyclicStructureError would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportInaccessibleThisError would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportInaccessibleUniqueSymbolError would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportInferenceFallback would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportLikelyUnsafeImportRequiredError would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportNonSerializableProperty would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportNonlocalAugmentation would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportPrivateInBaseOfClassExpression would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (reportTruncationError would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (resolveLibrary would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (resolveModuleNameLiterals would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (resolveModuleNames would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (resolveTypeReferenceDirectiveReferences would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (resolveTypeReferenceDirectives would lose its object, and this with it) | 1 | 0 | 0 |
| Refused: a method read as a value (setModifiedTime would lose its object, and this with it) | 4 | 1 | 1 |
| Refused: a method read as a value (setPrototypeOf would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (setTimeout would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (toKey would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (toString would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a method read as a value (trace would lose its object, and this with it) | 4 | 5 | 5 |
| Refused: a method read as a value (trackSymbol would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (useCaseSensitiveFileNames would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (watchDirectory would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (watchFile would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (writeFile would lose its object, and this with it) | 2 | 2 | 2 |
| Refused: a method read as a value (writeOutputIsTTY would lose its object, and this with it) | 1 | 1 | 1 |
| Refused: a non-exhaustive enum switch; missing ModuleKind.CommonJS | 1 | 1 | 1 |
| Refused: a non-exhaustive enum switch; missing ModuleResolutionKind.Classic | 1 | 1 | 1 |
| Refused: a non-exhaustive enum switch; missing SyntaxKind.Unknown | 33 | 38 | 38 |
| Refused: a number outside the proven flag domain assigned to CheckFlags.None; its domain is a closed union of non-negative int32 bit subsets | 2 | 2 | 2 |
| Refused: a number outside the proven flag domain assigned to Connection.None; its domain is a closed union of non-negative int32 bit subsets | 3 | 3 | 3 |
| Refused: a number outside the proven flag domain assigned to EmitFlags.None; its domain is a closed union of non-negative int32 bit subsets | 1 | 1 | 1 |
| Refused: a number outside the proven flag domain assigned to EscapeSequenceScanningFlags.String; its domain is a closed union of non-negative int32 bit subsets | 0 | 2 | 2 |
| Refused: a number outside the proven flag domain assigned to Extensions.TypeScript; its domain is a closed union of non-negative int32 bit subsets | 6 | 7 | 7 |
| Refused: a number outside the proven flag domain assigned to Extensions; its domain is a closed union of non-negative int32 bit subsets | 1 | 1 | 1 |
| Refused: a number outside the proven flag domain assigned to InternalEmitFlags.None; its domain is a closed union of non-negative int32 bit subsets | 2 | 2 | 2 |
| Refused: a number outside the proven flag domain assigned to LexicalEnvironmentFlags.None; its domain is a closed union of non-negative int32 bit subsets | 2 | 2 | 2 |
| Refused: a number outside the proven flag domain assigned to NodeFlags.NestedNamespace; its domain is a closed union of non-negative int32 bit subsets | 0 | 1 | 1 |
| Refused: a number outside the proven flag domain assigned to NodeFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 8 | 8 |
| Refused: a number outside the proven flag domain assigned to NodeResolutionFeatures.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 2 | 2 |
| Refused: a number outside the proven flag domain assigned to ObjectFlags.None; its domain is a closed union of non-negative int32 bit subsets | 3 | 3 | 3 |
| Refused: a number outside the proven flag domain assigned to TokenFlags.None; its domain is a closed union of non-negative int32 bit subsets | 0 | 2 | 2 |
| Refused: a type predicate | 581 | 587 | 587 |
| Refused: a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } \| { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string \| undefined where string is read | 1 | 1 | 1 |
| Refused: a value of type (sourceFile: SourceFile \| undefined, cancellationToken: CancellationToken \| undefined) => readonly DiagnosticWithLocation[] seen as (sourceFile?: SourceFile \| undefined, cancellationToken?: CancellationToken \| undefined) => readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 |
| Refused: a value of type (symbol: Symbol) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Symbol) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 |
| Refused: a value of type (symbolAccessibilityResult: SymbolAccessibilityResult) => { diagnosticMessage: DiagnosticMessage; errorNode: DeclarationDiagnosticProducing; typeName: DeclarationName \| undefined; } \| undefined seen as (symbolAccessibilityResult: SymbolAccessibilityResult) => SymbolAccessibilityDiagnostic \| undefined, which can write Node where DeclarationDiagnosticProducing is read | 1 | 1 | 1 |
| Refused: a value of type (type: Type) => { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) => { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 | 1 | 1 |
| Refused: a value of type AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type AccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type AmbientModuleDeclaration \| undefined seen as (AmbientModuleDeclaration & { name: StringLiteral; }) \| undefined, whose readonly field name becomes writable: a readonly field may hold something narrower than ModuleName, which a write of StringLiteral would replace | 0 | 1 | 1 |
| Refused: a value of type ArrayBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type BigIntLiteral \| Identifier \| NoSubstitutionTemplateLiteral \| NumericLiteral \| PrivateIdentifier \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type BigIntLiteral \| NoSubstitutionTemplateLiteral \| NumericLiteral \| StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type BinaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 |
| Refused: a value of type BindingElement seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 5 | 5 | 5 |
| Refused: a value of type BuilderProgram seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what BuilderProgram can't hold | 1 | 1 | 1 |
| Refused: a value of type BuilderProgram \| Program seen as BuilderProgram, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 |
| Refused: a value of type BuilderProgramStateWithDefinedProgram seen as BuilderProgramState, which can write Program \| undefined where Program is read | 5 | 5 | 5 |
| Refused: a value of type CallExpression seen as RequireOrImportCall, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & Identifier would replace | 1 | 1 | 1 |
| Refused: a value of type ClassDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ClassDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 0 | 1 | 1 |
| Refused: a value of type ClassElement seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 0 | 1 | 1 |
| Refused: a value of type ClassStaticBlockDeclaration seen as Mutable<ClassStaticBlockDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 | 1 | 1 |
| Refused: a value of type CommandLineOption seen as CommandLineOptionOfListType, which can write "list" \| "listOrElement" where "boolean" is read | 1 | 1 | 1 |
| Refused: a value of type CommandLineOption \| undefined seen as TsConfigOnlyOption, which can write "object" where "boolean" is read | 1 | 1 | 1 |
| Refused: a value of type CommandLineOptionOfBooleanType \| CommandLineOptionOfCustomType \| CommandLineOptionOfNumberType \| CommandLineOptionOfStringType \| TsConfigOnlyOption seen as CommandLineOptionOfCustomType, which can write Map<string, string \| number> where "boolean" is read | 1 | 1 | 1 |
| Refused: a value of type CompilerHost seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where WriteFileCallback is read | 2 | 1 | 1 |
| Refused: a value of type CompilerOptions & { types: string[]; } seen as CompilerOptions, which can write string[] \| undefined where string[] is read | 1 | 1 | 1 |
| Refused: a value of type CompilerOptionsValue seen as TsConfigSourceFile \| CompilerOptionsValue, which can write string \| number where string is read | 4 | 4 | 4 |
| Refused: a value of type CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 | 1 | 1 |
| Refused: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ConstructorDeclaration seen as Mutable<ConstructorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type Declaration seen as T, a type parameter whose constraint Declaration can be written, so it can write what Declaration can't hold | 2 | 2 | 2 |
| Refused: a value of type DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Diagnostic seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 |
| Refused: a value of type DiagnosticMessageChain seen as T, a type parameter whose constraint DiagnosticMessageChain \| ReusableDiagnosticMessageChain can be written, so it can write what DiagnosticMessageChain can't hold | 3 | 3 | 3 |
| Refused: a value of type DiagnosticWithDetachedLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where undefined is read | 4 | 4 | 4 |
| Refused: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | 10 | 10 | 10 |
| Refused: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile \| undefined where SourceFile is read | 10 | 10 | 10 |
| Refused: a value of type DiagnosticWithLocation \| undefined seen as Diagnostic \| undefined, which can write SourceFile \| undefined where SourceFile is read | 1 | 1 | 1 |
| Refused: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read | 2 | 3 | 3 |
| Refused: a value of type ElementAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type EvaluatorResult<number> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where number is read | 16 | 16 | 16 |
| Refused: a value of type EvaluatorResult<string \| undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string \| undefined is read | 1 | 1 | 1 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where string is read | 2 | 2 | 2 |
| Refused: a value of type EvaluatorResult<string> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where string is read | 1 | 1 | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| number \| undefined>, which can write string \| number \| undefined where undefined is read | 1 | 1 | 1 |
| Refused: a value of type EvaluatorResult<undefined> seen as EvaluatorResult<string \| undefined>, which can write string \| undefined where undefined is read | 1 | 1 | 1 |
| Refused: a value of type Expression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 6 | 12 | 12 |
| Refused: a value of type Expression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type Expression \| Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ExpressionWithTypeArguments seen as ExpressionWithTypeArguments & { expression: Identifier \| PropertyAccessEntityNameExpression; }, whose readonly field expression becomes writable: a readonly field may hold something narrower than LeftHandSideExpression, which a write of LeftHandSideExpression & (Identifier \| PropertyAccessEntityNameExpression) would replace | 1 | 1 | 1 |
| Refused: a value of type Extension[] seen as string[], which can write string where Extension is read | 1 | 1 | 1 |
| Refused: a value of type FlowArrayMutation \| FlowAssignment \| FlowCall \| FlowCondition \| FlowLabel \| FlowReduceLabel \| FlowStart \| FlowUnreachable seen as FlowNode, which can write BindingElement \| Expression \| VariableDeclaration where BinaryExpression \| CallExpression is read | 1 | 1 | 1 |
| Refused: a value of type FlowNode seen as FlowLabel, which can write undefined where BinaryExpression \| CallExpression is read | 1 | 1 | 1 |
| Refused: a value of type FunctionDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat" \| "packageJsonScope">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 2 | 2 | 2 |
| Refused: a value of type FutureSourceFile \| SourceFile seen as Pick<SourceFile, "fileName" \| "impliedNodeFormat">, whose readonly field fileName becomes writable: a readonly field may hold something narrower than string, which a write of string would replace | 3 | 3 | 3 |
| Refused: a value of type GeneratedIdentifier seen as GeneratedIdentifier \| Identifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Identifier \| PrivateIdentifier seen as Identifier \| PrivateIdentifier, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 |
| Refused: a value of type GeneratedIdentifier \| GeneratedPrivateIdentifier \| Node seen as Node, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 4 | 4 | 4 |
| Refused: a value of type GeneratedIdentifier \| Identifier seen as Identifier \| undefined, whose readonly field emitNode becomes writable: a readonly field may hold something narrower than EmitNode & { autoGenerate: AutoGenerateInfo; }, which a write of EmitNode \| undefined would replace | 1 | 1 | 1 |
| Refused: a value of type GetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type GetAccessorDeclaration \| SetAccessorDeclaration seen as Mutable<AccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Identifier seen as Mutable<Identifier>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type Identifier seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 3 | 3 |
| Refused: a value of type Identifier seen as T, a type parameter whose constraint Node can be written, so it can write what Identifier can't hold | 1 | 1 | 1 |
| Refused: a value of type Identifier seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type Identifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 3 | 3 |
| Refused: a value of type Identifier[][] seen as ModuleExportName[][], which can write ModuleExportName[] where Identifier[] is read | 4 | 4 | 4 |
| Refused: a value of type ImportTypeNode \| undefined seen as ValidImportTypeNode \| undefined, whose readonly field argument becomes writable: a readonly field may hold something narrower than TypeNode, which a write of LiteralTypeNode & { literal: StringLiteral; } would replace | 1 | 1 | 1 |
| Refused: a value of type JSDocNullableType seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type JsxOpeningFragment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type JsxTagNameExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type LeftHandSideExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 4 | 4 |
| Refused: a value of type LeftHandSideExpression \| UnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type LiteralLikeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Map<CharacterCodes, RegularExpressionFlags> seen as Map<CharacterCodes, number>, which can write number where RegularExpressionFlags is read | 1 | 1 | 1 |
| Refused: a value of type Map<Path, FileWatcher> seen as Map<string, FileWatcher>, which can write string where Path is read | 1 | 1 | 1 |
| Refused: a value of type Map<Path, string[]> seen as InvokeMap, which can write true \| string[] where string[] is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, BuildInfoCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where BuildInfoCacheEntry is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ConfigFileCacheEntry> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ConfigFileCacheEntry is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Map<Path, Date>> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Map<Path, Date> is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, ProgramUpdateLevel> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where ProgramUpdateLevel is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, Set<string> \| undefined> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where Set<string> \| undefined is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, T> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where T is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, UpToDateStatus> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where UpToDateStatus is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, readonly Diagnostic[]> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where readonly Diagnostic[] is read | 1 | 1 | 1 |
| Refused: a value of type Map<ResolvedConfigFilePath, true> seen as Map<ResolvedConfigFilePath, unknown>, which can write unknown where true is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, ImportsNotUsedAsValues> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ImportsNotUsedAsValues is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, JsxEmit> seen as Map<string, string \| number>, which can write string \| number where JsxEmit is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>> seen as ScriptTargetFeatures, which can write string where never is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, ModuleDetectionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleDetectionKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, ModuleResolutionKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where ModuleResolutionKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, NewLineKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where NewLineKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, PollingWatchKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where PollingWatchKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, SyntaxKind> seen as Map<string, number>, which can write number where SyntaxKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, WatchDirectoryKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchDirectoryKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, WatchFileKind> seen as "boolean" \| "list" \| "listOrElement" \| "number" \| "object" \| "string" \| Map<string, string \| number>, which can write string \| number where WatchFileKind is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string[] where never[] is read | 10 | 10 | 10 |
| Refused: a value of type Map<string, string> seen as Map<string, string \| number>, which can write string \| number where string is read | 1 | 1 | 1 |
| Refused: a value of type Map<string, string[] \| never[]> seen as Map<string, string[]> \| Map<string, never[]> \| Map<string, string[] \| never[]>, which can write string where never is read | 3 | 3 | 3 |
| Refused: a value of type MethodDeclaration seen as Mutable<MethodDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type MethodDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type MethodDeclaration \| PropertyDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type MissingDeclaration seen as Mutable<MissingDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type ModifierLike seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState, which can write ModuleResolutionHost where ModuleResolutionHost & GetPackageJsonEntrypointsHost is read | 0 | 1 | 1 |
| Refused: a value of type NamespaceExportDeclaration seen as Mutable<NamespaceExportDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Node seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 |
| Refused: a value of type Node seen as Node \| SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Node seen as Node \| TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type Node seen as SyntaxList, whose readonly field kind becomes writable: a readonly field may hold something narrower than SyntaxKind, which a write of SyntaxKind.SyntaxList would replace | 1 | 1 | 1 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node can be written, so it can write what Node can't hold | 4 | 4 | 4 |
| Refused: a value of type Node seen as T, a type parameter whose constraint Node \| undefined can be written, so it can write what Node can't hold | 1 | 1 | 1 |
| Refused: a value of type Node seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 |
| Refused: a value of type Node \| NodeArray<Node> seen as NodeArray<Node>, whose readonly field transformFlags becomes writable: a readonly field may hold something narrower than TransformFlags, which a write of TransformFlags would replace | 1 | 1 | 1 |
| Refused: a value of type Node \| SourceMapRange seen as SourceMapRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Node \| TextRange seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type NodeArray<ClassElement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 4 | 4 |
| Refused: a value of type NodeArray<Expression> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 0 | 1 | 1 |
| Refused: a value of type NodeArray<Statement> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type NodeArray<T> seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type NonNullExpression seen as Mutable<NonNullExpression>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type NonNullable<T> seen as T, a type parameter whose constraint any can be written, so it can write what NonNullable<T> can't hold | 2 | 6 | 6 |
| Refused: a value of type ObjectBindingPattern seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type OptionalTypeNode seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ParameterDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type ParameterDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 0 | 1 | 1 |
| Refused: a value of type ParameterDeclaration \| VariableDeclaration seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ParameterDeclaration \| undefined seen as Symbol \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of SymbolFlags would replace | 1 | 1 | 1 |
| Refused: a value of type ParenthesizedExpression seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ParseConfigHost seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 |
| Refused: a value of type Path[] seen as string[], which can write string where Path is read | 1 | 1 | 1 |
| Refused: a value of type PostfixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type PostfixUnaryExpression \| PrefixUnaryExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 3 | 3 | 3 |
| Refused: a value of type PrivateIdentifier seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type PropertyAccessExpression seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type PropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type PropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type PropertyName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type PropertySignature seen as Mutable<PropertySignature>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type QualifiedName seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Readonly<BuilderState> \| undefined seen as BuilderState \| undefined, whose readonly field fileInfos becomes writable: a readonly field may hold something narrower than Map<Path, FileInfo>, which a write of Map<Path, FileInfo> would replace | 1 | 1 | 1 |
| Refused: a value of type ResolvedModuleFull \| undefined seen as { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined, whose readonly field originalPath becomes writable: a readonly field may hold something narrower than string \| undefined, which a write of string \| true would replace | 1 | 1 | 1 |
| Refused: a value of type SearchResult<Resolved> seen as { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined, which can write { resolved: Resolved; isExternalLibraryImport: true; } \| undefined where Resolved \| undefined is read | 1 | 1 | 1 |
| Refused: a value of type SearchResult<undefined> seen as SearchResult<Resolved>, which can write Resolved \| undefined where undefined is read | 6 | 6 | 6 |
| Refused: a value of type SetAccessorDeclaration seen as Mutable<SetAccessorDeclaration>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type SetAccessorDeclaration seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as Mutable<PropertyAssignment \| ShorthandPropertyAssignment>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type ShorthandPropertyAssignment seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type SolutionBuilderHost<T> seen as CompilerHostLikeForCache, which can write WriteFileCallback \| undefined where ((path: string, data: string, writeByteOrderMark?: boolean \| undefined) => void) \| undefined is read | 1 | 0 | 0 |
| Refused: a value of type SourceFile seen as EmitNode \| undefined, whose readonly field flags becomes writable: a readonly field may hold something narrower than NodeFlags, which a write of EmitFlags would replace | 2 | 2 | 2 |
| Refused: a value of type SourceFile seen as Mutable<SourceFile>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type SourceFile seen as SourceFile \| SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 1 | 3 | 3 |
| Refused: a value of type SourceFile seen as SourceFileLike, which can write readonly number[] \| undefined where readonly number[] is read | 24 | 24 | 24 |
| Refused: a value of type StringLiteral seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Symbol[] seen as unknown[], which can write unknown where Symbol is read | 2 | 2 | 2 |
| Refused: a value of type System seen as ModuleResolutionHost, which can write boolean \| (() => boolean) \| undefined where boolean is read | 1 | 1 | 1 |
| Refused: a value of type T seen as Mutable<T>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of T["pos"] would replace | 6 | 6 | 6 |
| Refused: a value of type T seen as T, a type parameter whose constraint Node can be written, so it can write what T can't hold | 1 | 1 | 1 |
| Refused: a value of type T seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type T seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type T[] seen as (T \| undefined)[], which can write T \| undefined where T is read | 1 | 1 | 1 |
| Refused: a value of type T[][] seen as (readonly T[])[], which can write readonly T[] where T[] is read | 1 | 1 | 1 |
| Refused: a value of type TaggedTemplateExpression seen as Mutable<Node>, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type TemplateLiteralLikeNode seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Token<SyntaxKind> seen as T, a type parameter whose constraint Node can be written, so it can write what Token<SyntaxKind> can't hold | 2 | 2 | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as CompilerOptionsValue, which can write string \| number where string is read | 2 | 2 | 2 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 | 1 | 1 |
| Refused: a value of type TsConfigSourceFile \| CompilerOptionsValue seen as string[], which can write string where PluginImport is read | 1 | 1 | 1 |
| Refused: a value of type TypeCheckerHost seen as Program, whose readonly field redirectTargetsMap becomes writable: a readonly field may hold something narrower than RedirectTargetsMap, which a write of MultiMap<Path, string> would replace | 2 | 0 | 0 |
| Refused: a value of type TypeNode seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type Type[] seen as unknown[], which can write unknown where Type is read | 2 | 2 | 2 |
| Refused: a value of type VariableDeclaration \| DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 2 | 2 | 2 |
| Refused: a value of type VariableDeclarationList seen as TextRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type VisitEachChildTable seen as Record<SyntaxKind, VisitEachChildFunction<any> \| undefined>, which can write VisitEachChildFunction<any> \| undefined where VisitEachChildFunction<QualifiedName> is read | 1 | 1 | 1 |
| Refused: a value of type WatchedFileWithUnchangedPolls[] seen as (WatchedFileWithUnchangedPolls \| undefined)[], which can write WatchedFileWithUnchangedPolls \| undefined where WatchedFileWithUnchangedPolls is read | 1 | 1 | 1 |
| Refused: a value of type any seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what any can't hold | 4 | 2 | 2 |
| Refused: a value of type any seen as T, a type parameter whose constraint Node can be written, so it can write what any can't hold | 1 | 1 | 1 |
| Refused: a value of type never seen as T, a type parameter whose constraint any can be written, so it can write what never can't hold | 1 | 1 | 1 |
| Refused: a value of type never seen as U, a type parameter whose constraint {} can be written, so it can write what never can't hold | 0 | 1 | 1 |
| Refused: a value of type never[] seen as (JSDoc \| JSDocTag)[], which can write JSDoc \| JSDocTag where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as (SourceMapRange \| undefined)[], which can write SourceMapRange \| undefined where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as ChildDirectoryWatcher[], which can write ChildDirectoryWatcher where never is read | 2 | 2 | 2 |
| Refused: a value of type never[] seen as CommentRange[], which can write CommentRange where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as Comparator[][], which can write Comparator[] where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as Declaration[], which can write Declaration where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as Diagnostic[], which can write Diagnostic where never is read | 2 | 2 | 2 |
| Refused: a value of type never[] seen as ProjectReference[], which can write ProjectReference where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as RequireOrImportCall[], which can write RequireOrImportCall where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as ResolvedConfigFileName[], which can write ResolvedConfigFileName where never is read | 3 | 3 | 3 |
| Refused: a value of type never[] seen as SourceFile[], which can write SourceFile where never is read | 2 | 2 | 2 |
| Refused: a value of type never[] seen as StringLiteralLike[], which can write StringLiteralLike where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as TransformerFactory<Bundle \| SourceFile>[], which can write TransformerFactory<Bundle \| SourceFile> where never is read | 1 | 1 | 1 |
| Refused: a value of type never[] seen as string[] \| never[], which can write string where never is read | 3 | 3 | 3 |
| Refused: a value of type never[] seen as string[], which can write string where never is read | 7 | 8 | 8 |
| Refused: a value of type never[] \| SortedArray<DiagnosticWithLocation> seen as Diagnostic[], which can write Diagnostic where never is read | 0 | 1 | 1 |
| Refused: a value of type never[][] seen as string[][], which can write string where never is read | 0 | 1 | 1 |
| Refused: a value of type number[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where number is read | 1 | 1 | 1 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] \| undefined, which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 |
| Refused: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile \| undefined where SourceFile is read | 2 | 2 | 2 |
| Refused: a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read | 3 | 3 | 3 |
| Refused: a value of type readonly T[] \| undefined seen as unknown[], which can write unknown where T is read | 0 | 1 | 1 |
| Refused: a value of type readonly string[] \| undefined seen as RegExp[] \| undefined, which can write RegExp where string is read | 0 | 1 | 1 |
| Refused: a value of type string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined seen as string \| number \| boolean \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> \| TsConfigSourceFile \| null \| undefined, which can write string \| number where string is read | 1 | 1 | 1 |
| Refused: a value of type string \| number \| boolean \| string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] \| MapLike<string[]> seen as CompilerOptionsValue, which can write string \| number where string is read | 1 | 1 | 1 |
| Refused: a value of type string[] seen as CompilerOptionsValue, which can write string \| number where string is read | 4 | 4 | 4 |
| Refused: a value of type string[] seen as string \| (string \| number)[] \| undefined, which can write string \| number where string is read | 1 | 1 | 1 |
| Refused: a value of type string[] \| PluginImport[] \| ProjectReference[] \| (string \| number)[] seen as unknown[], which can write unknown where string is read | 2 | 2 | 2 |
| Refused: a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | 1 | 1 | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint BuilderProgram can be written, so it can write what undefined can't hold | 1 | 1 | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Declaration can be written, so it can write what undefined can't hold | 1 | 1 | 1 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint Node can be written, so it can write what undefined can't hold | 4 | 4 | 4 |
| Refused: a value of type undefined seen as T, a type parameter whose constraint WatchedFileWithIsClosed can be written, so it can write what undefined can't hold | 3 | 3 | 3 |
| Refused: a value of type { (fileName: string): DiagnosticWithLocation[]; (): Diagnostic[]; } seen as { (): Diagnostic[]; (fileName: string): DiagnosticWithLocation[]; }, which can write Diagnostic where DiagnosticWithLocation is read | 0 | 1 | 1 |
| Refused: a value of type { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } seen as CompilerOptions, which can write ModuleResolutionKind \| undefined where ModuleResolutionKind is read | 1 | 1 | 1 |
| Refused: a value of type { arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ... seen as ({ arguments: { name?: string; } & { path: string; }; range: CommentRange; } \| { arguments: { name: string; }; range: CommentRange; } \| { arguments: { factory: string; }; range: CommentRange; } \| ... 5 more ... \| ...)[] \| ... 8 more ... \| ..., which can write { name?: string; } & { path: string; } where never is read | 1 | 1 | 1 |
| Refused: a value of type { compilerOptions: CompilerOptions; traceEnabled: boolean; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 9 more ...; host: GetPackageJsonEntrypointsHost; } seen as ModuleResolutionState & { host: GetPackageJsonEntrypointsHost; }, which can write string[] \| undefined where never[] is read | 1 | 1 | 1 |
| Refused: a value of type { host: ModuleResolutionHost; traceEnabled: boolean; failedLookupLocations: string[] \| undefined; affectingLocations: string[] \| undefined; resultFromCache?: ResolvedModuleWithFailedLookupLocations; ... 8 more ...; reportDiagnostic: ...; } seen as ModuleResolutionState, which can write CompilerOptions where { all?: boolean; allowImportingTsExtensions?: boolean; allowJs?: boolean; allowNonTsExtensions?: boolean; allowArbitraryExtensions?: boolean; allowSyntheticDefaultImports?: boolean; allowUmdGlobalAccess?: boolean; ... 129 more ...; moduleResolution: ModuleResolutionKind; } is read | 1 | 1 | 1 |
| Refused: a value of type { id: number; flowNode: FlowNode; edges: never[]; text: string; lane: number; endLane: number; level: number; circular: false; } seen as FlowGraphNode, which can write FlowGraphEdge[] where never[] is read | 1 | 1 | 1 |
| Refused: a value of type { kind: "ambient" \| "node_modules" \| "paths" \| "redirect" \| "relative" \| undefined; moduleSpecifiers: readonly string[]; computedWithoutCache: false; } \| undefined seen as ModuleSpecifierResult \| undefined, which can write boolean where false is read | 1 | 1 | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string \| number; minor: number; patch: number; prerelease: string \| readonly string[]; build: string \| readonly string[]; }, which can write string \| number where number is read | 1 | 1 | 1 |
| Refused: a value of type { matchableStringSet: Set<string> \| undefined; patterns: Pattern[] \| undefined; } seen as ParsedPatterns, which can write ReadonlySet<string> \| undefined where Set<string> \| undefined is read | 1 | 1 | 1 |
| Refused: a value of type { path: string; originalPath: string \| true; extension: string; packageId: PackageId \| undefined; resolvedUsingTsExtension: boolean \| undefined; } \| undefined seen as Resolved \| undefined, which can write string \| true \| undefined where string \| true is read | 1 | 1 | 1 |
| Refused: a value of type { readonly args: readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...]; readonly kind: PragmaKindFlags.... seen as PragmaDefinition<string, string, string, string>, whose readonly field args becomes writable: a readonly field may hold something narrower than readonly [{ readonly name: "types"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "lib"; readonly optional: true; readonly captureSpan: true; }, { readonly name: "path"; readonly optional: true; readonly captureSpan: true; }, ..., ..., ...], which a write of readonly [PragmaArgumentSpecification<string>] \| readonly [PragmaArgumentSpecification<string>, PragmaArgumentSpecification<string>] \| ... \| ... \| undefined would replace | 0 | 2 | 2 |
| Refused: a value of type { resolved: Resolved; isExternalLibraryImport: true; } \| undefined seen as { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined, which can write boolean where true is read | 1 | 1 | 1 |
| Refused: a value of type { value: never; done: true; } seen as IteratorResult<Mapping, any>, which can write any where never is read | 14 | 14 | 14 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: false; }; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: false; } is read | 2 | 2 | 2 |
| Refused: a value of type { value: { resolved: Resolved; isExternalLibraryImport: true; } \| undefined; } \| undefined seen as SearchResult<{ resolved: Resolved; isExternalLibraryImport: boolean; }>, which can write { resolved: Resolved; isExternalLibraryImport: boolean; } \| undefined where { resolved: Resolved; isExternalLibraryImport: true; } \| undefined is read | 1 | 1 | 1 |
| Refused: a value without nominal ancestry seen as BinaryExpressionStateMachine<TOuterState, TState, TResult> | 11 | 12 | 12 |
| Refused: an arbitrary number or a value from another enum assigned to AccessKind.Read; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.EOF; its members are a closed union | 0 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.plus; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to CharacterCodes.slash; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Mjs; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to Extension.Ts; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to ForegroundColorEscapeSequences.Grey; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to GeneratedIdentifierFlags.None; its members are a closed union | 1 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to InternalSymbolName.Call; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to InvalidPosition; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ModifierFlags.None; its members are a closed union | 15 | 15 | 15 |
| Refused: an arbitrary number or a value from another enum assigned to ModuleKind.ESNext; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to NodeFlags.NestedNamespace; its members are a closed union | 1 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to NodeFlags.None; its members are a closed union | 17 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to NodeResolutionFeatures.None; its members are a closed union | 2 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to OperatorPrecedence.Invalid; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to OuterExpressionKinds.Parentheses; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to ParsingContext.SourceElements; its members are a closed union | 4 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to Phase.Parse; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to ProgramUpdateLevel.Update; its members are a closed union | 1 | 0 | 0 |
| Refused: an arbitrary number or a value from another enum assigned to SymbolFlags.None; its members are a closed union | 2 | 8 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayLiteralExpression; its members are a closed union | 6 | 6 | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrayType; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ArrowFunction; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BinaryExpression; its members are a closed union | 20 | 27 | 27 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.BindingElement; its members are a closed union | 4 | 8 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CallExpression; its members are a closed union | 7 | 8 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CaseBlock; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CatchClause; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassDeclaration; its members are a closed union | 4 | 5 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ClassExpression; its members are a closed union | 4 | 5 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.CommaToken; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ComputedPropertyName; its members are a closed union | 1 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ConditionalExpression; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Constructor; its members are a closed union | 2 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Decorator; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ElementAccessExpression; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.EnumDeclaration; its members are a closed union | 2 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportAssignment; its members are a closed union | 3 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportDeclaration; its members are a closed union | 13 | 13 | 13 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExportSpecifier; its members are a closed union | 4 | 6 | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExpressionStatement; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExpressionWithTypeArguments; its members are a closed union | 10 | 11 | 11 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ExternalModuleReference; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ForStatement; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionDeclaration; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.FunctionExpression; its members are a closed union | 3 | 5 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.GetAccessor; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.HeritageClause; its members are a closed union | 2 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Identifier; its members are a closed union | 15 | 17 | 17 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportClause; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportDeclaration; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportEqualsDeclaration; its members are a closed union | 5 | 5 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportSpecifier; its members are a closed union | 4 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ImportType; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.InferType; its members are a closed union | 0 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDoc; its members are a closed union | 1 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocEnumTag; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocReturnTag; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocSatisfiesTag; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTemplateTag; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypeExpression; its members are a closed union | 7 | 7 | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JSDocTypedefTag; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxElement; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.JsxNamespacedName; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.LabeledStatement; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MetaProperty; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MethodDeclaration; its members are a closed union | 2 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.MinusToken; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ModuleDeclaration; its members are a closed union | 3 | 5 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceExport; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NamespaceImport; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NewExpression; its members are a closed union | 2 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NoSubstitutionTemplateLiteral; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.NumericLiteral; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ObjectLiteralExpression; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.Parameter; its members are a closed union | 8 | 13 | 13 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ParenthesizedExpression; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ParenthesizedType; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PostfixUnaryExpression; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PrefixUnaryExpression; its members are a closed union | 7 | 7 | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyAccessExpression; its members are a closed union | 6 | 7 | 7 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyAssignment; its members are a closed union | 4 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.PropertyDeclaration; its members are a closed union | 5 | 8 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.QualifiedName; its members are a closed union | 3 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ReturnStatement; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SatisfiesExpression; its members are a closed union | 1 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SetAccessor; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.ShorthandPropertyAssignment; its members are a closed union | 4 | 4 | 4 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.SourceFile; its members are a closed union | 4 | 5 | 5 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.StringLiteral; its members are a closed union | 8 | 8 | 8 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TaggedTemplateExpression; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateExpression; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateHead; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateLiteralTypeSpan; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TemplateSpan; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeAliasDeclaration; its members are a closed union | 0 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeAssertionExpression; its members are a closed union | 0 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeParameter; its members are a closed union | 2 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.TypeReference; its members are a closed union | 3 | 3 | 3 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableDeclaration; its members are a closed union | 6 | 6 | 6 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.VariableStatement; its members are a closed union | 1 | 1 | 1 |
| Refused: an arbitrary number or a value from another enum assigned to SyntaxKind.YieldExpression; its members are a closed union | 0 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to TransformFlags.None; its members are a closed union | 2 | 2 | 2 |
| Refused: an arbitrary number or a value from another enum assigned to TypeReferenceSerializationKind.Unknown; its members are a closed union | 1 | 1 | 1 |
| Refused: an import cycle | 1 | 1 | 1 |
| Refused: an index signature | 7 | 8 | 8 |
| Refused: arguments | 0 | 1 | 1 |
| Refused: arithmetic assigned back into an enum; the result need not be one of its members | 42 | 15 | 15 |
| Refused: debugger | 1 | 1 | 1 |
| Refused: delete | 1 | 1 | 1 |
| Refused: export * | 75 | 75 | 75 |
| Refused: in | 2 | 2 | 2 |
| Refused: inherited library member compare read as an own field | 1 | 1 | 1 |
| Refused: inherited library member hasOwnProperty read as an own field | 1 | 1 | 1 |
| Refused: inherited library member prototype read as an own field | 1 | 1 | 1 |
| Refused: inherited library member replace read as an own field | 1 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 10 more ... \| undefined written where AbstractKeyword \| AccessorKeyword \| AsyncKeyword \| ConstKeyword \| DeclareKeyword \| Decorator \| ... 9 more ... \| StaticKeyword is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type ClassElement \| undefined written where ClassElement is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type CommandLineOption \| undefined written where CommandLineOption is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type CommentRange \| undefined written where CommentRange is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type Diagnostic \| undefined written where Diagnostic is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type EmitHelper \| undefined written where EmitHelper is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type JsxChild \| undefined written where JsxChild is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type T \| undefined written where T is read | 2 | 1 | 1 |
| Refused: instantiating a generic function makes a value of type VariableDeclaration \| undefined written where VariableDeclaration is read | 1 | 0 | 0 |
| Refused: instantiating a generic function makes a value of type string \| undefined written where string is read | 1 | 0 | 0 |
| Refused: the non-null assertion ! | 180 | 383 | 383 |
| Refused: this in a namespace function; a qualified call and a detached call have different receivers | 1 | 1 | 1 |
| Refused: var | 1 | 1 | 1 |
| Refused: yield (generators) | 4 | 5 | 5 |
| SkippedDependency: a dependency function whose body has checker diagnostics (measurement skipped) | 25 | 5 | 5 |
| error: lower: src/compiler/parser.ts:1472:9: the checker gave a declaration no symbol | 1 | 1 | 1 |

# Commands, checks and mutants

Required toolchain setup ran again: Go/clang/Node ready in 0s, submodules in 1s,
build cache warm and total 24s. `nproc` is 5, cgroup CPU quota four CPUs.
Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup/fetch logs are preserved.
All test output was redirected to logs, never piped.

```sh
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/tsc-latent2-adapted > /tmp/latent2-apply.log 2>&1 # in the resolved cumulative-2 tree
python3 stage3/census/latent/run_comparisons.py /workspace/adamic /tmp/tsc-latent2-adapted /tmp/latent2-runs /tmp/latent2-worktrees.json > /tmp/latent2-comparisons.log 2>&1
python3 stage3/census/latent/rerun2.py /tmp/latent2-runs /tmp/tsc-latent2-adapted > /tmp/latent2-summary.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node stage3/census/latent/declarations.cjs /tmp/tsc-latent2-adapted /tmp/latent2-manifest.json /tmp/latent2-declarations.json > /tmp/latent2-declarations.log 2>&1
python3 stage3/census/latent/audit_rerun2.py /tmp/tsc-latent2-adapted /tmp/latent2-declarations.json > /tmp/latent2-report-audit.log 2>&1
python3 stage3/census/latent/audit_corpus.py /tmp/latent2-runs /tmp/tsc-latent2-adapted cumulative-2 > /tmp/latent2-corpus-audit.log 2>&1
```

The stock AST oracle initially failed because NODE_PATH pointed at /root's cache;
using the actual /home/agent cache fixed it. The scratch merge helper initially lacked
gofmt on PATH; sourcing the toolchain and completing the same resolutions fixed it.
Neither failed attempt produced census results. Logs preserve the successful final runs.

Report mutants alter source hashes, checker totals/per-file attribution, zero-file
membership, the ratio numerator/denominator, NotYet/reason totals, skipped-body counts,
variance totals, owner metadata/counts, checker deltas and measurement labels. A raw
AST declaration-end mutant is caught only by the independent stock-span check.
All named mutants fail the intended check. `data/rerun2/report-audit.log` records them.

The real-corpus overlay-only NotYet at binder.ts:330:1 adds one site in binder.ts
and leaves every other file's entire record and all unit eligibility unchanged.
The baseline contains no mutant. Mutant audit JSON and compressed raw evidence are
preserved. The prior synthetic body-scope/misattribution and IR/loader guard mutants
were rerun with this instrumentation. Overlay vet and Python/Node syntax checks pass.
The older six-configuration report/audits remain archived as historical evidence.

# Limits

Measurement cannot emit native output. Lowering stops at its first returned error
within a unit; the refusal pre-scan sees additional syntax failures. Generic declarations
are attempted without invented specializations; registration is best effort. Diagnosed
dependency bodies remain skipped boundaries. Final module order, ownership, initializer
readiness, specialization completeness and backend passes are not validated. Native
runtime equivalence, adaptation oracle suites and the full Adamic gate were not rerun.
The ratio is zero own-file checker diagnostics out of 78, not native tsc readiness.
