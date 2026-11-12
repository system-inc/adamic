# Paired latent census: overload proof 3a7ea290 against 17c6378d

Base `17c6378de039f08283edea73ee75cfd4ba32a55a`; head `3a7ea290fca9e3f3325ffa8206a861b22f441583`. The head is exactly one commit over its base. Report and evidence only; no compiler or adaptation edits.

## 1. Regressions

Previously finding-free units that stop on head: **0**. New panics: **0**. New refusal-scan sites: **0**. Raw regression entries: 0; distinct: 0.

There are no regressions to name.

## 2. Cleared, moved and new findings

| Paired delta | Raw | Distinct outermost accesses |
| --- | --- | --- |
| Cleared: no finding remains at identity | 3 | 3 |
| Moved: old reason gone, another remains | 0 | 0 |
| New findings | 3 | 3 |
| New sites | 3 | 3 |
| Changed unit first stops | 7 | 7 |
| Regressions | 0 | 0 |

Distinct identities use the unmodified base `stage3/census/pair` tool: source path, source SHA-256, outermost contiguous property/element access start, and exact reason. Raw identities retain kind, source position, reason and text. New reasons in already-blocked units are findings, not automatically regressions. Full ledgers: [delta.json](delta.json) and [pair-report.md](pair-report.md).

| Measure | Base | Head |
| --- | --- | --- |
| Resolved source files | 81 | 81 |
| Checker diagnostics | 0 | 0 |
| Units | 4565 | 4565 |
| Attempted units | 4061 | 4061 |
| Attempted units with no first stop | 3176 | 3178 |
| Overload-result first stops | 67 | 61 |


| Finding kind | Base raw | Head raw | Base distinct | Head distinct |
| --- | --- | --- | --- | --- |
| NotYet | 525 | 528 | 518 | 521 |
| Refused | 4427 | 4424 | 4232 | 4229 |


Cleared raw finding identities (the distinct view also clears three):

| Location | Kind | Exact reason |
| --- | --- | --- |
| src/compiler/core.ts:656:1 | Refused | overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined |
| src/compiler/emitter.ts:4949:5 | Refused | overload 1 of writeTokenText result void cannot be served by implementation result number |
| src/compiler/sourcemap.ts:547:5 | Refused | overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping |

New raw findings (the distinct view also has three):

| Location | Kind | Exact reason |
| --- | --- | --- |
| src/compiler/core.ts:675:12 | NotYet | an array of RequireOrImportCall |
| src/compiler/moduleNameResolver.ts:780:36 | NotYet | a dictionary member narrowed to an array without an element-checked view |
| src/compiler/sourcemap.ts:562:18 | NotYet | an uninitialized object field without a supported declared slot type |

Other changed unit first stops:

| Unit | Unit location | Base first stop | Head first stop |
| --- | --- | --- | --- |
| decodeMappings | src/compiler/sourcemap.ts:464:1 | Refused: overload 1 of captureMapping result Required<Mapping> cannot be served by implementation result Mapping at src/compiler/sourcemap.ts:547:5 | NotYet: an uninitialized object field without a supported declared slot type at src/compiler/sourcemap.ts:562:18 |


## 3. The five concatenate units

Overload indexes are **one-based among the eight overload declarations**, excluding the implementation. Both sides’ own pinned Adamic checker resolve the same call to the same declaration with zero diagnostics. Stock TypeScript independently agrees; its separate host-type diagnostics are retained in `concatenate.json` and are not the census diagnostics.

| Unit | Unit location | Call site and text | Resolved overload | Resolved declaration text | Base first stop | Head first stop |
| --- | --- | --- | --- | --- | --- | --- |
| getJSDocCommentRanges | src/compiler/utilities.ts:2761:1 | src/compiler/utilities.ts:2769:9 — concatenate(getTrailingCommentRanges(text, node.pos), getLeadingCommentRanges(text, node.pos)) | 7 / 8 | export function concatenate<T>(array1: T[] \| undefined, array2: T[] \| undefined): T[] \| undefined; | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 | Refused: overload 1 of filter result U[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:261:1 |
| getAllSuperTypeNodes | src/compiler/utilities.ts:5227:1 | src/compiler/utilities.ts:5229:29 — concatenate(singleElementArray(getEffectiveBaseTypeNode(node)), getEffectiveImplementsTypeNodes(node)) | 8 / 8 | export function concatenate<T>(array1: readonly T[] \| undefined, array2: readonly T[] \| undefined): readonly T[] \| undefined; | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 | No lowering stop |
| getRequiresAtTopOfFile | src/compiler/utilities.ts:10125:1 | src/compiler/utilities.ts:10133:24 — concatenate(requires, statement.declarationList.declarations.map(d => d.initializer)) | 4 / 8 | export function concatenate<T>(array1: T[] \| undefined, array2: T[]): T[] \| undefined; | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 | NotYet: an array of RequireOrImportCall at src/compiler/core.ts:675:12 |
| getConditions | src/compiler/moduleNameResolver.ts:755:1 | src/compiler/moduleNameResolver.ts:780:12 — concatenate(conditions, options.customConditions) | 3 / 8 | export function concatenate<T>(array1: T[], array2: T[] \| undefined): T[]; | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 | NotYet: a dictionary member narrowed to an array without an element-checked view at src/compiler/moduleNameResolver.ts:780:36 |
| getSemanticDiagnosticsOfFile | src/compiler/builder.ts:1058:1 | src/compiler/builder.ts:1065:12 — concatenate(<br>        getBinderAndCheckerDiagnosticsOfFile(state, sourceFile, cancellationToken, semanticDiagnosticsPerFile),<br>        state.program.getProgramDiagnostics(sourceFile),<br>    ) | 2 / 8 | export function concatenate<T>(array1: readonly T[], array2: readonly T[]): readonly T[]; | Refused: overload 1 of concatenate result T[] cannot be served by implementation result readonly T[] \| undefined at src/compiler/core.ts:656:1 | No lowering stop |


The declaration-level stop checks the overload set; it does not identify the overload selected by that call. Native evidence for both sides is in `concatenate-native-{base,head}.json`; call/declaration evidence is in [concatenate.json](concatenate.json).

## 4. getSupportedExtensions and writeTokenText

Each row gives the first stop on each side, preserving its exact reason words. Where the old guard clears, the head column names the next observed stop; an unchanged guard means no next stop was reached. Locations are one-based lines and UTF-16 columns.

| Unit | Unit location | Base first stop | Head first / next stop |
| --- | --- | --- | --- |
| getSupportedExtensions | src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| getSupportedExtensions | src/compiler/utilities.ts:9992:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| getSupportedExtensions | src/compiler/utilities.ts:9994:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| isSupportedSourceFileName | src/compiler/utilities.ts:10146:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| getFileNamesFromConfigSpecs | src/compiler/commandLineParser.ts:3950:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 | Refused: overload 1 of getSupportedExtensions result readonly Extension[][] cannot be served by implementation result readonly string[][] at src/compiler/utilities.ts:9990:1 |
| createPrinter | src/compiler/emitter.ts:1211:1 | Refused: overload 1 of writeTokenText result void cannot be served by implementation result number at src/compiler/emitter.ts:4949:5 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/factory/nodeFactory.ts:6111:21 |


All five getSupportedExtensions-blocked units retain the same overload-result first stop on this head. **No next stop is observed for these five, including no `adamic/invariant-mutable` stop.**

The adapted implementation contains `x.extension as Extension` in its later callback. The pinned head’s proof scans the entire body and rejects an `AsExpression` before its flow reasoning. See [the adapted implementation excerpt](get-supported-extensions-source.json), [overload-proof-head.go.txt](overload-proof-head.go.txt), and the retained source hashes and first-stop records.

## 5. The historical 43 overload-result stops

**43/43 clear at base; 43/43 clear at head**, using the earlier census’s criterion: the unit no longer first-stops on an overload-result refusal. This is not a claim that all these units lower. Exact historical definitions are in [historical-43.json](historical-43.json); mapping uses source file, declaration kind, function name and ordinal among same-named declarations.

| Unit | Current unit location | Base first stop | Head first stop |
| --- | --- | --- | --- |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1431:1 | No lowering stop | No lowering stop |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1432:1 | No lowering stop | No lowering stop |
| skipPartiallyEmittedExpressions | src/compiler/utilitiesPublic.ts:1433:1 | No lowering stop | No lowering stop |
| isLeftHandSideExpression | src/compiler/utilitiesPublic.ts:1995:1 | No lowering stop | No lowering stop |
| isUnaryExpression | src/compiler/utilitiesPublic.ts:2039:1 | No lowering stop | No lowering stop |
| isExpression | src/compiler/utilitiesPublic.ts:2086:1 | No lowering stop | No lowering stop |
| getThisContainer | src/compiler/utilities.ts:3180:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3182:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3184:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3186:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getThisContainer | src/compiler/utilities.ts:3187:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3245:24 |
| getNewTargetContainer | src/compiler/utilities.ts:3291:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:2258:13 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:2258:13 |
| getSuperContainer | src/compiler/utilities.ts:3333:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getSuperContainer | src/compiler/utilities.ts:3335:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| getSuperContainer | src/compiler/utilities.ts:3336:1 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 | Refused: a cast the runtime can't check at src/compiler/utilities.ts:3362:24 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3469:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3471:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3473:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3475:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeCanBeDecorated | src/compiler/utilities.ts:3477:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 | NotYet: checked predicate overload target NamedDeclaration & { name: DeclarationName; } at src/compiler/utilitiesPublic.ts:946:1 |
| nodeIsDecorated | src/compiler/utilities.ts:3523:1 | No lowering stop | No lowering stop |
| nodeIsDecorated | src/compiler/utilities.ts:3525:1 | No lowering stop | No lowering stop |
| nodeIsDecorated | src/compiler/utilities.ts:3527:1 | No lowering stop | No lowering stop |
| nodeIsDecorated | src/compiler/utilities.ts:3529:1 | No lowering stop | No lowering stop |
| getExpandoInitializer | src/compiler/utilities.ts:3992:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| isCompoundLikeAssignment | src/compiler/utilities.ts:4927:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| skipParentheses | src/compiler/utilities.ts:5039:1 | No lowering stop | No lowering stop |
| skipParentheses | src/compiler/utilities.ts:5041:1 | No lowering stop | No lowering stop |
| skipParentheses | src/compiler/utilities.ts:5043:1 | No lowering stop | No lowering stop |
| isAnonymousFunctionDefinition | src/compiler/utilities.ts:5483:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/factory/emitHelpers.ts:424:37 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/factory/emitHelpers.ts:424:37 |
| createParenthesizerRules | src/compiler/factory/parenthesizerRules.ts:54:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/transformers/destructuring.ts:192:9 |
| skipOuterExpressions | src/compiler/factory/utilities.ts:656:1 | No lowering stop | No lowering stop |
| skipOuterExpressions | src/compiler/factory/utilities.ts:658:1 | No lowering stop | No lowering stop |
| skipOuterExpressions | src/compiler/factory/utilities.ts:660:1 | No lowering stop | No lowering stop |
| visitFunctionBody | src/compiler/visitorPublic.ts:508:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:513:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:518:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:519:18 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:520:18 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:521:18 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| visitFunctionBody | src/compiler/visitorPublic.ts:522:1 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 | NotYet: a branded supertype inside a checked predicate's array target at src/compiler/visitorPublic.ts:526:9 |
| transformES2017 | src/compiler/transformers/es2017.ts:119:1 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at src/compiler/visitorPublic.ts:587:119 | Refused: overload 2 of visitEachChild parameter nodesVisitor cannot be served by implementation parameter nodesVisitor at src/compiler/visitorPublic.ts:587:119 |
| transformES2018 | src/compiler/transformers/es2018.ts:154:1 | NotYet: writing a possibly absent optional own field at src/compiler/factory/emitNode.ts:262:13 | NotYet: writing a possibly absent optional own field at src/compiler/factory/emitNode.ts:262:13 |


## Method, bounds and validation

Both scratch overlays are generated from their **own side’s tree**, including each side’s own `make_overlay.py` and `entry_overlay.py`. Both trees have identical meter instrumentation. No earlier measurement stream is reused. Both compilers use their exact recursive submodule pins. The adaptation-input diff is empty (see `adaptation-input-diff.txt`); the full head diff touches only four `internal/lower` files. Both sides read the single immutable tree produced by **17c6378d’s stage3/apply.sh**, pinned to TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. Source hashes and overlay provenance are retained.

As in 83e07d2f, this is the entry-root **first-error** latent census (`LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`), with per-unit eligibility, findings and timings. Ordinary Load and Lower remain disabled; no backend or usable IR is produced. A unit with no finding is not proof of complete compilation.

To meet the 30-minute process limit, each side runs eight disjoint attempted-unit shards over the complete project. Selection occurs after existing checker-body skips and generic deferral, before an independent attempt; it does not alter declarations, dependencies or eligibility. Shards 0 and 1 retain complete refusal scans; a scratch selector omits duplicate scans in shards 2–7. Both own-side overlay variants and executable hashes are retained. The two complete scan ledgers must agree exactly. Deferred/skipped unit records are repeated and deduplicated. Generic call witnesses are unioned using the original Go ordering and five-witness cap. Each attempted unit must occur exactly once; a stock-parser inventory checks every top-level unit and its source order. Shard relation-cache reports are process-scoped, retained separately, and never summed. The merger stops on any relation overflow or global finding.

Both binary variants on both sides passed the unsharded-versus-sharded fixture, including a real overloaded-value stop and cross-shard generic witnesses; both existing latent audits and all 20 pairing tests passed. The composition fixture validates this measurement mechanism on a bounded example, rather than claiming an unmeasured full unsharded run.

Two shards per side run concurrently with `GOMEMLIMIT=3GiB`. Every timing wrapper records its actual child status and exits with it (signals use the shell encoding). Each census child is terminated at 1,740 seconds, with at most 10 seconds to kill, so no child or wrapper reaches 30 minutes. No measured process timed out. `run.py`, `build.sh`, `measure.sh`, `measure-scan.sh`, `shard_overlay.py`, `scan_overlay.py`, `build-scan.sh`, `merge.py`, `analyze.py`, `report.py`, and `commands.md` retain the method. Go evidence is named `.go.txt`.

Census child measurements (overlapping runs; RSS is each child’s own peak):

| Side | Shard | Wall s | User s | System s | Peak RSS MiB | Child exit |
| --- | --- | --- | --- | --- | --- | --- |
| base | 0 | 1210.361 | 1179.469 | 19.365 | 2737.51 | 0 |
| base | 1 | 1164.034 | 1129.268 | 19.169 | 2714.46 | 0 |
| base | 2 | 134.443 | 130.685 | 2.487 | 1088.17 | 0 |
| base | 3 | 132.68 | 130.767 | 3.396 | 1203.26 | 0 |
| base | 4 | 127.661 | 128.316 | 3.366 | 1191.57 | 0 |
| base | 5 | 127.539 | 125.141 | 3.36 | 1007.93 | 0 |
| base | 6 | 141.173 | 136.266 | 3.277 | 1000.27 | 0 |
| base | 7 | 123.829 | 118.551 | 3.691 | 1014.76 | 0 |
| head | 0 | 1192.807 | 1156.185 | 20.502 | 2796.13 | 0 |
| head | 1 | 1161.979 | 1125.754 | 17.785 | 2783.62 | 0 |
| head | 2 | 135.256 | 131.974 | 3.381 | 1106.15 | 0 |
| head | 3 | 133.568 | 129.579 | 4.082 | 1298.0 | 0 |
| head | 4 | 131.384 | 129.103 | 4.014 | 1213.1 | 0 |
| head | 5 | 126.516 | 124.911 | 3.087 | 1006.7 | 0 |
| head | 6 | 135.915 | 132.0 | 3.376 | 1040.02 | 0 |
| head | 7 | 125.233 | 119.687 | 3.452 | 1031.94 | 0 |
