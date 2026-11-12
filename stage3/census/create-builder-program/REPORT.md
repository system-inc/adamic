# createBuilderProgram: report-only adaptation measurements

Pinned main **`219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`**, stock TypeScript **6.0.3**, and the exact original adapted tree from `overload-results-main`. All **747** source/input hashes are reverified. Only this report directory is committed; both candidate trees are scratch copies.

## 1. Entire implementation, callers and return paths

The complete original implementation (including every nested body) is retained in `createBuilderProgram.original.ts.txt`; its checker-derived signature is:

```ts
(kind: BuilderProgramKind, { newProgram, host, oldProgram, configFileParsingDiagnostics }: BuilderCreationParameters): BuilderProgram | undefined
```

Its overloads are:

1. `src/compiler/builder.ts:1686:1`

```ts
export function createBuilderProgram(
    kind: BuilderProgramKind.SemanticDiagnosticsBuilderProgram,
    builderCreationParameters: BuilderCreationParameters,
): SemanticDiagnosticsBuilderProgram;
```

2. `src/compiler/builder.ts:1691:1`

```ts
export function createBuilderProgram(
    kind: BuilderProgramKind.EmitAndSemanticDiagnosticsBuilderProgram,
    builderCreationParameters: BuilderCreationParameters,
): EmitAndSemanticDiagnosticsBuilderProgram;
```

There are exactly **two returns from createBuilderProgram itself**. The remaining 25 returns belong to nested helpers/callbacks; every one is listed in the appendix and `base-inspection.json`.

| Return | Checker type | Reachability and promise |
| --- | --- | --- |
| src/compiler/builder.ts:1705:9 | BuilderProgram \| undefined | Fast reuse, before either kind branch. Can reach either overload with exactly the supplied oldProgram object. No kind/capability validation. |
| src/compiler/builder.ts:1741:5 | BuilderProgram | Fresh creation: kind 0 installs getSemanticDiagnosticsOfNextAffectedFile; kind 1 installs that method plus emitNextAffectedFile and overrides emitBuildInfo. Each valid kind therefore constructs its promised shape; an invalid kind throws via notImplemented before this return. |


The exact reuse condition is:

```ts
let oldState = oldProgram && oldProgram.state;
if (oldState && newProgram === oldState.program && configFileParsingDiagnostics === newProgram.getConfigFileParsingDiagnostics()) {
    newProgram = undefined!;
    oldState = undefined;
    return oldProgram;
}
```

The diagnostics test is **array object identity**, not equal contents. Clearing local newProgram/oldState does not replace or enrich oldProgram. A truthy oldState implies a present oldProgram at runtime; the inferred `| undefined` reflects the checker’s lack of correlation between those locals. There is no actual undefined-return path in this body. A plain createAbstractBuilder result has state undefined, so it does not by itself take this reuse path. The counterexample below retains real builder state while omitting the subtype-only method.

### All checkout callers

A full-checkout grep (including harness, services and tests) finds only the two compiler files. The AST inventory covers all **712 .ts files under src**, including harness/services; all six identifiers are three declarations, one import specifier and these two calls. No other references or aliases appear. `whole-checkout-grep.txt` and the checker reference inventory are retained.

| Call site | Enclosing caller | Resolved overload | Promise |
| --- | --- | --- | --- |
| src/compiler/builderPublic.ts:184:12 | createSemanticDiagnosticsBuilderProgram | 1 | SemanticDiagnosticsBuilderProgram |
| src/compiler/builderPublic.ts:223:12 | createEmitAndSemanticDiagnosticsBuilderProgram | 2 | EmitAndSemanticDiagnosticsBuilderProgram |


Every caller’s selected declaration text and signature are retained in `base-inspection.json`. The first passes kind 0 and getBuilderCreationParameters(...); the second passes kind 1 and the same parameter-building helper. The public wrappers each have two narrow overloads and one implementation, accounting for six census units; the internal helper’s two overloads and implementation account for the other three.

## 2. Is the promise true?

**The internal overload contract is false for admitted internal calls, with an observable missing-method result. The two existing checkout callers preserve a stronger invariant and do not exhibit that counterexample.**

Their invariant is: an old builder already provides the capabilities promised by the requested wrapper. The semantic wrapper’s old parameter is SemanticDiagnosticsBuilderProgram; the emit wrapper’s is EmitAndSemanticDiagnosticsBuilderProgram. An emit builder may also serve the semantic wrapper because the emit interface extends the semantic one, so “created by exactly the same kind” is stronger than necessary. getBuilderCreationParameters preserves the old object’s identity but erases that subtype in its non-generic BuilderCreationParameters result. The internal overloads do not themselves encode or check the invariant.

Two cast-free internal counterexamples type-check with zero diagnostics alongside all 79 roots (`counterexample.ts.txt`, `counterexample-check.json`):

- Pass a real semantic old builder to kind 1 with the same Program and its diagnostics array. The overload promises EmitAndSemanticDiagnosticsBuilderProgram; the fast path returns that same semantic object, whose emitNextAffectedFile is undefined.
- Destructure only getSemanticDiagnosticsOfNextAffectedFile away from that real semantic old builder. The rest object is a structurally valid plain BuilderProgram retaining its real state. Pass it to kind 0 with the same Program/diagnostics. The overload promises SemanticDiagnosticsBuilderProgram, but returns the same plain object with the promised method absent.

The runtime probe executes the **entire extracted adapted function**, not a reimplementation of its condition, using a real stock 6.0.3 builder/state. Both identity tests pass; both promised method calls throw TypeError. This establishes the helper’s contract hole without claiming an existing in-checkout call exercises it. The helper is internal, not a published typescript.d.ts function.

## 3. Two measured adaptation shapes

Two shapes were measured, within the requested maximum of three. Neither is ready to land: both remove the result-family stop only by moving every target unit to a different first stop. A split-function design was not measured, because the old-input invariant and honest result boundary can be tested with smaller, separately reviewable edits.

### Checked: honest wide internal results, guarded public narrowing

Both kind-specific internal overload results become `BuilderProgram | undefined`, exactly matching the inferred implementation result. Each public wrapper captures that result, tests presence of its required method(s) and `typeof ... === "function"`, then returns the original object with its narrow static type. The check is executable; the final structural assertion remains in the source. This experiment does not claim that a presence/function test proves an arbitrary structural method signature, nor that Adamic has approved the assertion: its first stops occur earlier. Source patches for both files are retained.

The 79 compiler roots have **zero stock diagnostics**. Exactly **one of 79 emitted JavaScript files differs**, builderPublic.js; builder.js is byte-identical. These are all exact emitted hunks:

```diff
--- base/builderPublic.js
+++ checked/builderPublic.js
@@ -1,9 +1,17 @@
 import { BuilderProgramKind, createBuilderProgram, createRedirectedBuilderProgram, getBuilderCreationParameters, } from "./_namespaces/ts.js";
 export function createSemanticDiagnosticsBuilderProgram(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences) {
-    return createBuilderProgram(BuilderProgramKind.SemanticDiagnosticsBuilderProgram, getBuilderCreationParameters(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences));
+    const builder = createBuilderProgram(BuilderProgramKind.SemanticDiagnosticsBuilderProgram, getBuilderCreationParameters(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences));
+    if (builder === undefined || !("getSemanticDiagnosticsOfNextAffectedFile" in builder) || typeof builder.getSemanticDiagnosticsOfNextAffectedFile !== "function") {
+        throw new Error("Builder program does not have the requested kind");
+    }
+    return builder;
 }
 export function createEmitAndSemanticDiagnosticsBuilderProgram(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences) {
-    return createBuilderProgram(BuilderProgramKind.EmitAndSemanticDiagnosticsBuilderProgram, getBuilderCreationParameters(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences));
+    const builder = createBuilderProgram(BuilderProgramKind.EmitAndSemanticDiagnosticsBuilderProgram, getBuilderCreationParameters(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences));
+    if (builder === undefined || !("getSemanticDiagnosticsOfNextAffectedFile" in builder) || typeof builder.getSemanticDiagnosticsOfNextAffectedFile !== "function" || !("emitNextAffectedFile" in builder) || typeof builder.emitNextAffectedFile !== "function") {
+        throw new Error("Builder program does not have the requested kind");
+    }
+    return builder;
 }
 export function createAbstractBuilder(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences) {
     const { newProgram, configFileParsingDiagnostics: newConfigFileParsingDiagnostics } = getBuilderCreationParameters(newProgramOrRootNames, hostOrOptions, oldProgramOrHost, configFileParsingDiagnosticsOrOldProgram, configFileParsingDiagnostics, projectReferences);
```

Runtime behavior is therefore not declared byte-preserving: the two wrappers gain property reads/checks and may throw on a malformed old object. The public overload signatures and builder interfaces remain unchanged.

### Typed-old: preserve the old builder subtype through parameters

`BuilderCreationParameters<T extends BuilderProgram = BuilderProgram>` stores `oldProgram: T | undefined`; getBuilderCreationParameters becomes generic over the old subtype. Its existing three old-program assertions become `as T`, preserving the existing positional dispatch assumption. The two internal kind overloads require BuilderCreationParameters<SemanticDiagnosticsBuilderProgram> and BuilderCreationParameters<EmitAndSemanticDiagnosticsBuilderProgram> respectively. The implementation remains broad. This makes the existing callers’ type-level capability correlation explicit, rather than using it as an unencoded global assumption.

The 79 roots have **zero stock diagnostics**. All **79 emitted JavaScript files are byte-identical** to baseline; `typed-old-emitted.patch` is empty. The result proof still does not get approved: the mutable narrow oldProgram field now meets the broad implementation parameter first, and the census refuses that parameter boundary. No public overload or interface changes.

### Summary and all nine units

“Moved” below means the same declaration unit still has a first stop, with a different reason. “Cleared” means it has no first stop. The overload-result family itself disappears from all nine first stops in both shapes, but **no unit clears**. Since the adapted source changes, units correspond by source path, function name and overload-set ordinal, rather than claiming same-input source-hash identity pairing.

| Shape | 79-root diagnostics | JS identical | Unit cleared | Unit moved | New stop from previously clean unit | New panic | Result first stops |
| --- | --- | --- | --- | --- | --- | --- | --- |
| checked | 0 | False | 0 | 9 | 0 | 0 | 0 |
| typed-old | 0 | True | 0 | 9 | 0 | 0 | 0 |


Baseline first stop for every row: `overload 1 of createBuilderProgram result SemanticDiagnosticsBuilderProgram cannot be served by implementation result BuilderProgram | undefined`, at `src/compiler/builder.ts:1686:1`. These fresh nine baseline findings match the earlier inventory **exactly**, including full diagnostic text. The stock checker resolves overload 2 in the emit wrapper, but censusOverloads validates the whole overload set and stops at incompatible overload 1; the finding number is not the call-resolution index.

### checked unit outcomes

| Unit / overload-set ordinal | Base unit | Candidate unit | Outcome | Candidate first stop | Exact reason |
| --- | --- | --- | --- | --- | --- |
| createBuilderProgram #1 | src/compiler/builder.ts:1686:1 | src/compiler/builder.ts:1686:1 | moved | src/compiler/builder.ts:267:1 | a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) |
| createBuilderProgram #2 | src/compiler/builder.ts:1691:1 | src/compiler/builder.ts:1691:1 | moved | src/compiler/builder.ts:267:1 | a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) |
| createBuilderProgram #3 | src/compiler/builder.ts:1696:1 | src/compiler/builder.ts:1696:1 | moved | src/compiler/builder.ts:267:1 | a checked predicate target whose check would read past the depth bound (BuilderProgramStateWithDefinedProgram) |
| createSemanticDiagnosticsBuilderProgram #1 | src/compiler/builderPublic.ts:162:1 | src/compiler/builderPublic.ts:162:1 | moved | src/compiler/builderPublic.ts:195:36 | in when a possibly absent spread has no property presence descriptors |
| createSemanticDiagnosticsBuilderProgram #2 | src/compiler/builderPublic.ts:168:1 | src/compiler/builderPublic.ts:168:1 | moved | src/compiler/builderPublic.ts:195:36 | in when a possibly absent spread has no property presence descriptors |
| createSemanticDiagnosticsBuilderProgram #3 | src/compiler/builderPublic.ts:176:1 | src/compiler/builderPublic.ts:176:1 | moved | src/compiler/builderPublic.ts:195:36 | in when a possibly absent spread has no property presence descriptors |
| createEmitAndSemanticDiagnosticsBuilderProgram #1 | src/compiler/builderPublic.ts:201:1 | src/compiler/builderPublic.ts:205:1 | moved | src/compiler/builderPublic.ts:238:36 | in when a possibly absent spread has no property presence descriptors |
| createEmitAndSemanticDiagnosticsBuilderProgram #2 | src/compiler/builderPublic.ts:207:1 | src/compiler/builderPublic.ts:211:1 | moved | src/compiler/builderPublic.ts:238:36 | in when a possibly absent spread has no property presence descriptors |
| createEmitAndSemanticDiagnosticsBuilderProgram #3 | src/compiler/builderPublic.ts:215:1 | src/compiler/builderPublic.ts:219:1 | moved | src/compiler/builderPublic.ts:238:36 | in when a possibly absent spread has no property presence descriptors |

### typed-old unit outcomes

| Unit / overload-set ordinal | Base unit | Candidate unit | Outcome | Candidate first stop | Exact reason |
| --- | --- | --- | --- | --- | --- |
| createBuilderProgram #1 | src/compiler/builder.ts:1686:1 | src/compiler/builder.ts:1686:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createBuilderProgram #2 | src/compiler/builder.ts:1691:1 | src/compiler/builder.ts:1691:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createBuilderProgram #3 | src/compiler/builder.ts:1696:1 | src/compiler/builder.ts:1696:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createSemanticDiagnosticsBuilderProgram #1 | src/compiler/builderPublic.ts:162:1 | src/compiler/builderPublic.ts:162:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createSemanticDiagnosticsBuilderProgram #2 | src/compiler/builderPublic.ts:168:1 | src/compiler/builderPublic.ts:168:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createSemanticDiagnosticsBuilderProgram #3 | src/compiler/builderPublic.ts:176:1 | src/compiler/builderPublic.ts:176:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createEmitAndSemanticDiagnosticsBuilderProgram #1 | src/compiler/builderPublic.ts:201:1 | src/compiler/builderPublic.ts:201:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createEmitAndSemanticDiagnosticsBuilderProgram #2 | src/compiler/builderPublic.ts:207:1 | src/compiler/builderPublic.ts:207:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |
| createEmitAndSemanticDiagnosticsBuilderProgram #3 | src/compiler/builderPublic.ts:215:1 | src/compiler/builderPublic.ts:215:1 | moved | src/compiler/builder.ts:1688:5 | overload 1 of createBuilderProgram parameter builderCreationParameters cannot be served by implementation parameter 2 |


The checked shape has three unique new first-stop sites (one predicate target and one property-presence site per wrapper). Typed-old has one unique new parameter-stop site. These are new reasons/sites in already-stopped units, not regressions from clean units. Coverage here is exactly these nine units; no whole-project regression claim is made.

### Public API status, for both shapes

| Name | 6.0.3 typescript.d.ts | stage3/lane/sanctioned-api.json |
| --- | --- | --- |
| createBuilderProgram | False | False |
| BuilderProgramKind | False | False |
| BuilderCreationParameters | False | False |
| getBuilderCreationParameters | False | False |
| BuilderProgram | True | True |
| SemanticDiagnosticsBuilderProgram | True | False |
| EmitAndSemanticDiagnosticsBuilderProgram | True | False |
| createSemanticDiagnosticsBuilderProgram | True | False |
| createEmitAndSemanticDiagnosticsBuilderProgram | True | False |


The helper, BuilderProgramKind enum, BuilderCreationParameters and getBuilderCreationParameters are internal and absent from both files. The two narrower builder interfaces and their public factories are published. They are absent from the sanction table; that absence is not permission to change them. BuilderProgram occurs in other sanction-table contexts. Automated AST comparison confirms that all public factory overload texts and all three public builder interface texts are identical to baseline in both shapes. No sanctioned-API edit is proposed.

## 4. Existing adaptation pins

The retained searches cover all of stage3/adapt: `adaptation-grep.txt` has 2,994 matching lines for createBuilderProgram or builder.ts; `affected-owner-grep.txt` has 32 lines for the helper parameter type, parameter-building helper and public wrappers. Long evidence lines are preview-truncated at 500 characters with their exact file/line retained; the original checkout remains the authoritative full text.

The directly relevant exact owner pins are in **99-correctness-axolotl-x/owners.json**: `createBuilderProgram.emitBuildInfo`, `createBuilderProgram.emitNextAffectedFileOrDtsErrors`, and `createBuilderProgram.getWriteFileCallback`. Its adapt.cjs checks both the owner hashes and exact text, plus the core maybeBind/import inventory. All three nested owner hashes remain equal to their pinned after hashes on baseline and both scratch shapes (nine checks in api-pins.json). Neither proposal edits those nested bodies. A split/rename of the outer helper would need an explicit owner-pin/ownership review.

Other builder.ts adaptations are present: 32-indexed-reads-program, 76-truthful-casts, 99-correctness-axolotl-j and 99-correctness-axolotl-k, with diagnostic/serialization owners elsewhere in the file; 70-readonly-views also records builder diagnostic/state ownership evidence. The k exact pins are convertOrRepopulateDiagnosticMessageChain and convertOrRepopulateDiagnosticMessageChainArray, and j touches toReusableCompilerOptionValue / createBuilderProgramUsingIncrementalBuildInfo, not the factory signatures or parameter helper edited here. The affected-owner search finds no exact pin naming BuilderCreationParameters, getBuilderCreationParameters or either public wrapper declaration. This is a targeted pin audit, not a claim that rerunning the entire adaptation pipeline on the candidates is approved or tested.

## Method, limitations and validation

The original own-main entry-census overlay is reused. First-error mode does not honor the full-mode LATENT_ONLY_FILE selector; two initial full-project pilots were explicitly terminated (child -15, shell 143) and retained separately, with no partial results used. A scratch-only selector was then layered over the validated scan-selector overlay, selecting exactly the nine top-level declarations by file and function name. It preserves complete source resolution, registration, eligibility, program facts, normal first-error lowering and both no-output guards. No refusal is suppressed inside a unit. Repeated whole-project refusal scans are omitted; no refusal-scan census or delta is claimed. The selected baseline’s nine full findings exactly reproduce the previous census. All final runs complete with zero diagnostics, nine attempted units/timings, no panic, and no relation overflow/global finding.

All checks use stock TypeScript 6.0.3. The 79 roots are every recursive .ts root under src/compiler, with main’s bundled Node declaration index supplied as ambient context. Checking uses the same strict settings as main’s loader. Full program JavaScript emission uses ES2024/ESNext, LF and no maps/declarations, disables noEmit/allowImportingTsExtensions for emission, and hashes all 79 outputs; exact changed bytes and hunks are retained. The complete checkout search includes harness/services/tests even though the requested diagnostic measurement covers the 79 compiler roots.

The first typed-old scratch edit had a mechanical THost token error (one diagnostic); the corrected final source and hashes were rechecked and produce zero diagnostics. Superseded inspection files and the terminated pilot streams are clearly named; no preliminary diagnostics or partial census contributes to final counts. Every potentially long child is bounded by run.py to 1,740 seconds with at most ten seconds to kill. Wrappers report and propagate their own child status. All final measurements/checks exit zero. Go evidence is named .go.txt. Commands, scripts, exact patches, complete streams, timings and hashes are retained.

## Appendix: every return in the complete original implementation

| Owner | Location | Return statement | Checker type |
| --- | --- | --- | --- |
| createBuilderProgram | src/compiler/builder.ts:1705:9 | return oldProgram; | BuilderProgram \| undefined |
| createBuilderProgram | src/compiler/builder.ts:1741:5 | return builderProgram; | BuilderProgram |
| emitBuildInfo | src/compiler/builder.ts:1754:13 | return result; | EmitResult |
| emitBuildInfo | src/compiler/builder.ts:1756:9 | return emitSkippedWithNoDiagnostics; | EmitResult |
| emitNextAffectedFileOrDtsErrors | src/compiler/builder.ts:1800:25 | return {<br>                            result: { emitSkipped: true, diagnostics: pendingForDiagnostics.diagnostics },<br>                            affected: pendingForDiagnostics.affectedFile,<br>                        }; | { result: { emitSkipped: true; diagnostics: readonly Diagnostic[]; }; affected: SourceFile; } |
| emitNextAffectedFileOrDtsErrors | src/compiler/builder.ts:1825:25 | return {<br>                            result: { emitSkipped: true, diagnostics },<br>                            affected: state.program,<br>                        }; | { result: { emitSkipped: true; diagnostics: Diagnostic[]; }; affected: Program; } |
| emitNextAffectedFileOrDtsErrors | src/compiler/builder.ts:1835:72 | return undefined; | undefined |
| emitNextAffectedFileOrDtsErrors | src/compiler/builder.ts:1842:17 | return { result, affected }; | { result: EmitResult; affected: Program; } |
| emitNextAffectedFileOrDtsErrors | src/compiler/builder.ts:1896:9 | return { result, affected }; | { result: EmitResult; affected: SourceFile \| Program; } |
| &lt;arrow/function expression&gt; | src/compiler/builder.ts:1904:26 | return; | undefined |
| emitNextAffectedFile | src/compiler/builder.ts:1923:9 | return emitNextAffectedFileOrDtsErrors(<br>            writeFile,<br>            cancellationToken,<br>            emitOnlyDtsFiles,<br>            customTransformers,<br>            /*isForDtsErrors*/ false,<br>        ); | AffectedFileResult&lt;EmitResult&gt; |
| getWriteFileCallback | src/compiler/builder.ts:1937:58 | return writeFile \|\| host.writeFile?.bind(host); | WriteFileCallback \| undefined |
| getWriteFileCallback | src/compiler/builder.ts:1938:9 | return (fileName, text, writeByteOrderMark, onError, sourceFiles, data) =&gt; {<br>            if (isDeclarationFileName(fileName)) {<br>                if (!state.compilerOptions.outFile) {<br>                    Debug.assert(sourceFiles?.length === 1);<br>                    let emitSignature;<br>                    if (!customTransformers) {<br>                        const file = sourceFiles[0]!;<br>                        const info = state.fileInfos.get(file.resolvedPath)!;<br>                        if (info.signature === file.version) {<br>                            const signature = computeSignatureWithDiagnostics(<br>                                state.program,<br>                                file,<br>                                text,<br>                                host,<br>                                data,<br>                            );<br>                            // With d.ts diagnostics they are also part of the signature so emitSignature will be different from it since its just hash of d.ts<br>                            if (!data?.diagnostics?.length) emitSignature = signature;<br>                            if (signature !== file.version) { // Update it<br>                                if (host.storeSignatureInfo) (state.signatureInfo ??= new Map()).set(file.resolvedPath, SignatureInfo.StoredSignatureAtEmit);<br>                                if (state.affectedFiles) {<br>                                    // Keep old signature so we know what to undo if cancellation happens<br>                                    const existing = state.oldSignatures?.get(file.resolvedPath);<br>                                    if (existing === undefined) (state.oldSignatures ??= new Map()).set(file.resolvedPath, info.signature \|\| false);<br>                                    info.signature = signature;<br>                                }<br>                                else {<br>                                    // These are directly committed<br>                                    info.signature = signature;<br>                                }<br>                            }<br>                        }<br>                    }<br><br>                    // Store d.ts emit hash so later can be compared to check if d.ts has changed.<br>                    // Currently we do this only for composite projects since these are the only projects that can be referenced by other projects<br>                    // and would need their d.ts change time in --build mode<br>                    if (state.compilerOptions.composite) {<br>                        const filePath = sourceFiles[0]!.resolvedPath;<br>                        emitSignature = handleNewSignature(state.emitSignatures?.get(filePath), emitSignature);<br>                        if (!emitSignature) return data!.skippedDtsWrite = true;<br>                        (state.emitSignatures ??= new Map()).set(filePath, emitSignature);<br>                    }<br>                }<br>                else if (state.compilerOptions.composite) {<br>                    const newSignature = handleNewSignature(state.outSignature, /*newSignature*/ undefined);<br>                    if (!newSignature) return data!.skippedDtsWrite = true;<br>                    state.outSignature = newSignature;<br>                }<br>            }<br>            if (writeFile) writeFile(fileName, text, writeByteOrderMark, onError, sourceFiles, data);<br>            else if (host.writeFile) host.writeFile(fileName, text, writeByteOrderMark, onError, sourceFiles, data);<br>            else state.program.writeFile(fileName, text, writeByteOrderMark, onError, sourceFiles, data);<br><br>            /**<br>             * Compare to existing computed signature and store it or handle the changes in d.ts map option from before<br>             * returning undefined means that, we dont need to emit this d.ts file since its contents didnt change<br>             */<br>            function handleNewSignature(oldSignatureFormat: EmitSignature \| undefined, newSignature: string \| undefined) {<br>                const oldSignature = !oldSignatureFormat \|\| isString(oldSignatureFormat) ? oldSignatureFormat : oldSignatureFormat[0];<br>                newSignature ??= computeSignature(text, host, data);<br>                // Dont write dts files if they didn't change<br>                if (newSignature === oldSignature) {<br>                    // If the signature was encoded as string the dts map options match so nothing to do<br>                    if (oldSignatureFormat === oldSignature) return undefined;<br>                    // Mark as differsOnlyInMap so that --build can reverse the timestamp so that<br>                    // the downstream projects dont detect this as change in d.ts file<br>                    else if (data) data.differsOnlyInMap = true;<br>                    else data = { differsOnlyInMap: true };<br>                }<br>                else {<br>                    state.hasChangedEmitSignature = true;<br>                    state.latestChangedDtsFile = fileName;<br>                }<br>                return newSignature;<br>            }<br>        }; | (fileName: string, text: string, writeByteOrderMark: boolean, onError: ((message: string) =&gt; void) \| undefined, sourceFiles: readonly SourceFile[] \| undefined, data: WriteFileCallbackData \| undefined) =&gt; true \| undefined |
| &lt;arrow/function expression&gt; | src/compiler/builder.ts:1978:45 | return data!.skippedDtsWrite = true; | true |
| &lt;arrow/function expression&gt; | src/compiler/builder.ts:1984:40 | return data!.skippedDtsWrite = true; | true |
| handleNewSignature | src/compiler/builder.ts:2002:62 | return undefined; | undefined |
| handleNewSignature | src/compiler/builder.ts:2012:17 | return newSignature; | string |
| emit | src/compiler/builder.ts:2040:21 | return result; | EmitResult |
| emit | src/compiler/builder.ts:2065:17 | return {<br>                    emitSkipped,<br>                    diagnostics: diagnostics \|\| emptyArray,<br>                    emittedFiles,<br>                    sourceMaps,<br>                }; | { emitSkipped: boolean; diagnostics: Diagnostic[]; emittedFiles: string[]; sourceMaps: SourceMapEmitResult[]; } |
| emit | src/compiler/builder.ts:2094:9 | return emitResult; | EmitResult |
| getDeclarationDiagnostics | src/compiler/builder.ts:2129:13 | return (<br>                !sourceFile ? diagnostics : state.emitDiagnosticsPerFile?.get(sourceFile.resolvedPath)<br>            ) as readonly DiagnosticWithLocation[] \| undefined \|\| emptyArray; | readonly DiagnosticWithLocation[] |
| getDeclarationDiagnostics | src/compiler/builder.ts:2141:13 | return result; | readonly DiagnosticWithLocation[] |
| getSemanticDiagnosticsOfNextAffectedFile | src/compiler/builder.ts:2162:17 | return undefined; | undefined |
| getSemanticDiagnosticsOfNextAffectedFile | src/compiler/builder.ts:2198:13 | return { result, affected }; | { result: readonly Diagnostic[]; affected: SourceFile \| Program; } |
| getSemanticDiagnostics | src/compiler/builder.ts:2217:13 | return getSemanticDiagnosticsOfFile(state, sourceFile, cancellationToken); | readonly Diagnostic[] |
| getSemanticDiagnostics | src/compiler/builder.ts:2226:60 | return affectedResult.result; | readonly Diagnostic[] |
| getSemanticDiagnostics | src/compiler/builder.ts:2237:9 | return diagnostics \|\| emptyArray; | Diagnostic[] |


Final children:

| Child | Wall seconds | Own exit |
| --- | --- | --- |
| selector-build | 16.63 | 0 |
| base-inspect | 35.026 | 0 |
| checked-inspect | 24.856 | 0 |
| typed-old-inspect | 24.197 | 0 |
| base-census | 9.16 | 0 |
| checked-census | 10.629 | 0 |
| typed-old-census | 8.561 | 0 |
| api-pins | 0.565 | 0 |
