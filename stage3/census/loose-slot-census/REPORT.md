# Loose-slot census

## Control first

**Positive control passed.** Production head returns NotYet at `arrays.a:48:1`, ending “a cycle the cycle finder can't follow through unknown or object (#yeep0sx links these slots)”. The same complete program, copied alone into a control directory, yields **0 latent findings**, including the relation trailer, with `LATENT_FULL=1` and `LATENT_ASSERT_NO_OUTPUT=1`. Thus the latent zero demonstrably cannot count this stop.

The private type instrument sees **9 events at 9 distinct sites**, including that exact `:48:1` write. Disabling its stop judgment yields **0 events**, with the same H/candidate gathering. The mutant now replaces the head’s own `looseReaches` body with `return false` while keeping the collector active; it is not a compiler acceptance claim.

| Run | Observation | Wall seconds | Peak RSS KiB |
| --- | --- | ---: | ---: |
| production head | NotYet at arrays.a:48:1 | 0.194093 | 33048 |
| latent, LATENT_FULL=1 | 0 findings | 0.439692 | 37116 |
| type-level instrument | 9 events; includes arrays.a:48:1 | 0.314991 | 29708 |
| stop-disabled mutant | 0 events | 0.125977 | 30340 |

Each control event’s source line is quoted in [CONTROL-SITES.md](CONTROL-SITES.md). Raw control observations and per-process resource receipts are in [evidence/controls](evidence/controls). `latent/make_overlay.py` replaces the production orchestration through `lowerChecked` with a no-output return, and its units never invoke `findCycles`. This explains the observed blindness.



## Production next stops

**0 roots differ; 79/79 have the same exact first-stop tuple.** CommonJSBundle; production `LowerWithOptions` through the existing census inspector, one root per process. Both specified SHAs were freshly built and measured. The full table is [NEXT-STOPS.md](NEXT-STOPS.md); raw observations, times and peak RSS are under [evidence/main](evidence/main) and [evidence/head](evidence/head).

There are no changed roots to list. The new stop becomes the first stop for **0/79 roots**.

- 73 root(s): `sys.ts:1387:5`; Refused; a method read as a value (writeFile would lose its object, and this with it).
- 2 root(s): `—`; accepted; —.
- 1 root(s): `semver.ts:71:9`; NotYet; a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration.
- 1 root(s): `commandLineParser.ts:293:35`; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration.
- 1 root(s): `utilitiesPublic.ts:1284:52`; Refused; an unproven predicate argument for parameter predicate (argument "predicate").
- 1 root(s): `diagnosticInformationMap.generated.ts:13:45`; NotYet; reading an enum before its runtime initialization; move the call after the enum declaration.

The **73 roots** at main’s `sys.ts:1387:5` wall have **73 unchanged** exact tuples.

For the two main roots that lower whole:

- `corePublic.ts` at head: **accepted**; no first stop and no loose-slot NotYet.
- `hostErrors.ts` at head: **accepted**; no first stop and no loose-slot NotYet.

`findCycles` is late, not literally last: `lower.go:215` follows all module scans, declarations, bodies and preceding tail transformations; bind/view checks and finalization follow it. Earlier stopping gates cannot meet this stop first. Source-text order is not pass order: a type observation at sys.ts:364 does not precede the root’s existing production wall merely because its line number is smaller. `graph_types.go:132` invokes the loose-slot stop before the no-seeds return. The whole commit message and all of `loose_slots.go` were read.

## Sites past the walls: partial type-level source count

**22 observed events; 22 distinct exact source sites.** The exhaustive production site count remains **UNKNOWN**. This is a **partial type-level count, not a lowering**, and is conditional on the reconstructed source metadata described below. It is not a count of newly failing production roots. H was gathered over the complete 79-file input before judgment: 799863 non-type AST nodes; 33 candidate entries; 190 distinct explicit H types; H.every=True; 4457 shapes.

Whole-program type run: **29.502959 seconds**, **1125988 KiB** peak RSS. Each site, with its quoted source line, is in [SITES.md](SITES.md). The raw event list retains each event’s source, slot and holder types.

| Form | Events | Distinct sites |
| --- | ---: | ---: |
| captured | 7 | 7 |
| literal init | 9 | 9 |
| write | 6 | 6 |

| File | Events | Distinct sites | write events/sites | literal init events/sites | captured events/sites |
| --- | ---: | ---: | ---: | ---: | ---: |
| _namespaces/ts.moduleSpecifiers.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| _namespaces/ts.performance.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| _namespaces/ts.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| binder.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| builder.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| builderPublic.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| builderState.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| builderStatePublic.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| checker.ts | 2 | 2 | 1/1 | 0/0 | 1/1 |
| commandLineParser.ts | 10 | 10 | 2/2 | 5/5 | 3/3 |
| core.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| corePublic.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| debug.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| diagnosticInformationMap.generated.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| emitter.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| executeCommandLine.ts | 2 | 2 | 0/0 | 2/2 | 0/0 |
| expressionToTypeNode.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/baseNodeFactory.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/emitHelpers.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/emitNode.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/nodeChildren.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/nodeConverters.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/nodeFactory.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/nodeTests.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/parenthesizerRules.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/utilities.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| factory/utilitiesPublic.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| hostErrors.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| moduleNameResolver.ts | 1 | 1 | 0/0 | 0/0 | 1/1 |
| moduleSpecifiers.ts | 1 | 1 | 0/0 | 0/0 | 1/1 |
| parser.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| path.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| performance.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| performanceCore.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| program.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| programDiagnostics.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| resolutionCache.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| scanner.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| semver.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| sourcemap.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| symbolWalker.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| sys.ts | 3 | 3 | 1/1 | 1/1 | 1/1 |
| tracing.ts | 1 | 1 | 1/1 | 0/0 | 0/0 |
| transformer.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/classFields.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/classThis.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/declarations.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/declarations/diagnostics.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/destructuring.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2015.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2016.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2017.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2018.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2019.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2020.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/es2021.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/esDecorators.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/esnext.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/generators.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/jsx.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/legacyDecorators.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/module/esnextAnd2015.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/module/impliedNodeFormatDependent.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/module/module.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/module/system.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/namedEvaluation.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/taggedTemplate.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/ts.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/typeSerializer.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| transformers/utilities.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| tsbuild.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| tsbuildPublic.ts | 1 | 1 | 1/1 | 0/0 | 0/0 |
| types.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| utilities.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| utilitiesPublic.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| visitorPublic.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| watch.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |
| watchPublic.ts | 1 | 1 | 0/0 | 1/1 | 0/0 |
| watchUtilities.ts | 0 | 0 | 0/0 | 0/0 | 0/0 |

Distinct sites are `(adapted file, start line, start column)`, not source lines, edits or repeated root visits. Per-form site counts can overlap; events are never silently deduplicated. All 79 input files are listed. A zero means no event from this instrument, not a proof that lowering has no hidden stop.

### Method and limits

The declaration/overload boundary documented in `whole-program-scout/REPORT.md` at `0d078a8d` occurs before replaceable bodies. Placeholder bodies would also lose captures, concrete instantiations and stores. This work uses the explicitly permitted type-level fallback rather than claiming coherent continued IR. The private overlay adds a measurement-only API and no emission route.

The instrument calls head’s own `looseEntry` on the source walk, its shape/use/made methods, `noteLooseInit`, `holdsLoose`, `viewPairScalar` and `looseReaches`. It completes H before judging candidates. It reconstructs direct member/index assignments and recognized collection calls from source, gathers literal fields/elements/Map/Set entries with the head helper, and models captured cells by checker symbol identity and function nesting. The separate scratch mutant replaces the head’s own looseReaches body with return false; the collector remains active and still judges the candidates. k17 and k18 also match production at the exact captured-cell and literal-init sites, respectively; their receipts are in the controls directory.

**What this instrument cannot see or certify:** actual lowered IR write sites, synthetic library stores, concrete generic instantiations made only by lowering, explicit carried lowering slots, closed-frame-input proofs, transformed async/generator frames, accessor capture environments, static registration and bound-function metadata. Destructuring/loop/parameter-property stores are not fully reconstructed; computed collection method dispatch and record helper stores are not exhaustive. Source capture retention can overapproximate production. Checker query order also differs from production lowering; the commit itself warns about query-history sensitivity (#55xcnkc). It therefore establishes observed source/type judgments, not an exhaustive number of production stops or a proven upper/lower bound on one. No ordinary next-stop observation is inferred from this count. **Stopping reason:** an exhaustive entry-site census requires coherent declaration/signature recovery or validated reconstruction of the missing stores, instantiations and capture facts. This measurement has neither, so 22 is the observed partial count and the requested exhaustive total remains UNKNOWN.

## Setup, reproduction and resources

Main: `fe5a1a54f045e9bce4e5aa32ca74885a317fbebd`; head: `7e8cfaafb74a61f82f1610a4b738157b9918343d`, exactly one commit on main. Each used a fresh detached worktree under `/workspace/loose-slot-scratch`, initialized recursive shallow submodules, locked `npm ci` in `stage3/api`, and its own `stage3/apply.sh`. The result branch was created only in the head scratch worktree. The shared checkout was not used for edits or measurements.

TypeScript v6.0.3 source was freshly cloned at `050880ce59e30b356b686bd3144efe24f875ebc8`, supplied as ADAMIC_TYPESCRIPT_SOURCE. Completed adapted compiler manifests are identical at both SHAs: 79 TypeScript roots. Their hashes and submodule pins are in [evidence/provenance.json](evidence/provenance.json). GOTMPDIR, TMPDIR, GOCACHE and Adamic build caches were under scratch.

Scripts archive the exact driver/instrument as `.go.txt`. `build_drivers.py` overlays only the census command’s main; `build_type.py` overlays a private lower file plus its command. `groups.py` runs indices in groups of ten and pushes each completed group at both SHAs. `measure.py` records wall time and Linux wait4 peak RSS in KiB per child; these include load/check/lower or type scan, and exclude dependency installation, adaptation and Go build. Each reported observation retains its measured command and per-run resources in `.time.json`. Control receipts are the final runs after probe refinement; intermediate development reruns are not part of this final resource table.

- main: summed wall 1474.058278 s; largest per-root peak RSS 926504 KiB.
- head: summed wall 1519.070597 s; largest per-root peak RSS 954136 KiB.
- Interrupted scheduling-transition run `head-root-02.time.json`: 15.093474 s, 872092 KiB, exit -15; the root was rerun successfully. This receipt is excluded from the 79-root comparison.

The whole-program stop-disabled mutant also reads 0 events with identical gathering statistics: 15.027911 s and 617408 KiB. The ordinary table uses at most two simultaneous children; separate control/type runs can overlap that work. Per-child times include scheduling contention; they are resource receipts, not an isolated compiler-speed comparison.

Nothing was submitted or merged. Only the census directory is committed on the authorized new branch. No production compiler edit or native-output success is claimed.
