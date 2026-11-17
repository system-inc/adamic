# Compiler overload sets — #ev14ch9

Base: cf468df0747075e50d326a97ff03e8530d7a1a88. Upstream: TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

The required pre-adaptation report was pushed first at 6e7c19568, before either candidate was constructed. The four requested lowering files have been read in full. createBuilderProgram is excluded and remains untouched.

| Set (upstream line) | Resolved calls | Classification | Road to investigate |
| --- | ---: | --- | --- |
| utilities.ts:6122 getDiagnostics | 11 | 6 bindings; 2 arguments; 1 returned; 1 property initializer; 1 member read | (iii) stop: empty string admits global diagnostics without locations |
| utilities.ts:10011 getSupportedExtensionsWithJsonIfResolveJsonModule | 4 | 2 bindings; 1 flatten argument; 1 conditional branch | (i) try dropping the narrower Extension overload; reject if a mutable widening moves to a caller |
| builder.ts:558 convertOrRepopulateDiagnosticMessageChainArray | 3 | 1 binding; 2 next property initializers; all pass T[] plus the matching (T) callback | (i) investigate a generic implementation matching the overload, amending 99-correctness-axolotl-k |
| commandLineParser.ts:3775 convertOptionsFromJson | 3 | 2 discarded calls with default objects; 1 returned watch result with undefined default | (iii) stop pending a truthful watch result boundary; a new accumulator shape or a cast is outside the authorized roads |

## Current checkpoint

Builder is now stopped: full single-file latent census has 453 baseline events and 449 candidate events, with 11 gone, **7 new**, and 0 changed after unchanged-line mapping. The requested overload refusal disappears, but new findings violate the acceptance gate. Full event tuples are in [latent-builder-diff.json](latent-builder-diff.json); raw streams, commands, exits and stderr are alongside it. The new findings include an unproven predicate argument, a generic function as a value, a mutable recursive-chain view, and three runtime-uncheckable casts newly reached by lowering. No landing branch is created.

The builder candidate passes stock TypeScript 6.0.3 with **0 diagnostics over all 79 roots**, just like the baseline. Both full-file JavaScript comparisons pass, and all 13 existing owner-contract refusal controls plus zero-write replay pass. These checks do not override the failed census gate. Baseline stage3/apply.sh passed. The matching scan-all runs, extension caller census, and restoring mutant are still running at this checkpoint. Product, lane and dts builds are not run for a rejected candidate.

## Caller audit

callers.cjs uses TypeScript 6.0.3 getResolvedSignature on call/new/tagged-template expressions. It recognizes declarations by their checker-resolved declaration and source owner, including DiagnosticCollection's returned interface signatures. It does not count text matches. adapted-callers.json records 21,234 TypeScript roots across the checkout, including src/harness, src/testRunner and tests; src-callers.json is a separate source-only cross-check. Each resolved call includes its source, complete call, immediate use, argument types and resolved overload declaration. These preliminary adapted-tree reads occurred while baseline application was still running; they will be repeated on the completed tree before being used as final proof. The completed final audit is [final-callers-decoded.json](final-callers-decoded.json): 21,236 TypeScript roots, no changes during its checker walk, and the same 11/4/3/3 counts. The JavaScript pass covers 18,289 roots including compiler declaration roots and finds no additional calls to these source owners. [CALLERS.md](CALLERS.md) gives every resolved caller and its result/argument classification.

The whole-checkout checker encounters one exception in the deliberately invalid tests/cases/conformance/importDefer/dynamicImportDefer.ts:10, import.defer("./a.js"). This exception is retained in resolutionErrors and is not counted as a resolved call. The JavaScript pass has 14 checker exceptions, including generated import.defer baselines and two deep LKG calls. [exceptions-reviewed.json](exceptions-reviewed.json) excludes all 15 exception sites: named callees bind to lib declarations outside the four source owners, and import meta-properties cannot denote those named functions. A first upstream audit aborted on this exception and its empty output is not evidence.

## Failure rules and candidate roads

getDiagnostics fails censusOverload's covariant-result relation: Diagnostic[] cannot serve mutable DiagnosticWithLocation[]. Presence, phantom brands, predicate proof and void do not apply. narrowOverload refuses arrays. The existing body's map.get call, local return and unshift prevent the body proof. More fundamentally, fileName === "" takes the all-diagnostics branch. Returning a file-only array there would change values. Widening the nested overload also requires changing DiagnosticCollection's contract and checking its consumers; claiming the string overload is sound would be false. Candidate roads requiring a ruling: a nonempty-path contract, a changed internal collection shape, or explicit caller narrowing.

getSupportedExtensionsWithJsonIfResolveJsonModule fails the contravariant parameter relation because readonly applies only to the outer array; the inner Extension[] is mutable and cannot be widened to string[]. Three calls pass string arrays; program.ts passes readonly Extension[][]. The proposed first candidate removes the unused narrower result promise, preserving emitted JavaScript. Acceptance remains conditional on scan-all and full latent censuses showing no moved caller refusal. Other candidate: a generic implementation preserving the incoming element type plus Extension.Json, which must also account for the two builtin identity returns without casts or copies.

convertOrRepopulateDiagnosticMessageChainArray's adapted overload admits arbitrary T[] while its nongeneric implementation receives a mutable union-element array. This loses T and breaks both mutable array invariance and the matching callback contract. 99-correctness-axolotl-k pins the entire owner; no new competing pin is permitted. The candidate is type-only generic implementation/predicate annotations that retain T while preserving its existing runtime guard and identity behavior. If that cannot type-check without changing the existing algorithm or adding a new shape, stop and report.

convertOptionsFromJson's watch overload promises WatchOptions | undefined while the generic implementation can return a CompilerOptions/TypeAcquisition object too. WatchOptions is an untagged interface; narrowOverload cannot check it, and the generic implementation independently excludes that road. The body assigns defaultOptions, loops, and returns a local/parameter after assignment, so censusReturnProof cannot prove the watch promise. The other two results are discarded. Dropping all overloads would expose the broad result at the watch worker's returned result; asserting it is WatchOptions would hide the same unproved claim. Candidate roads requiring a ruling: a separately typed accumulator/helper or changed internal option shape. No public declaration edit is proposed.

## Proof status at checkpoint

For every set: scan-all, full single-file latent census (owner and caller files), restoring mutant, stock tsc over 79 roots, adamic types, candidate apply, upstream npm ci + hereby dts equality, TestProduct_Stage3AdaptedTree, whole unfiltered lane, and landing ledger: **not run**. No landing branch or landing SHA exists. Baseline setup and measurement-tool preparation are in progress, not adaptation acceptance.

## Extension follow-up road, before construction

The first candidate fails scan-all: 0 gone, 1 new, 0 changed. The full new event is `refusal_scan/stringWideningRefusal` at `src/compiler/program.ts:1588:119`: `a value of type readonly Extension[][] seen as readonly string[][], which can write string where Extension is read`.

A second road (i) will be checked before stopping this set: one implementation signature taking and returning `readonly (Extension[] | string[])[]`. This keeps both existing mutable inner-array views as separate union members under the existing readonly outer array, instead of widening an Extension[] into string[]. It introduces no record or runtime shape and changes no public declaration. The four caller/result classifications remain the same. Full acceptance still requires the original site gone and zero new/changed events; no landing branch exists.
