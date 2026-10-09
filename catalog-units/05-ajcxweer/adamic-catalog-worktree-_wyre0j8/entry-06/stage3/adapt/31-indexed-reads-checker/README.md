# Required indexed reads in checker.ts

This partition adapts only `src/compiler/checker.ts` in TypeScript 6.0.3,
pinned at `050880ce59e30b356b686bd3144efe24f875ebc8`. It follows adaptation 30:
one reviewed entry per required read, a class from `docs/tsc-strictness.md`,
and a one-line invariant obligation. `!` is a loud required-value check in
Adamic and erased syntax on Node. The invariant obligations cover populated
compiler-built or caller-provided arrays, not merely array bounds. Invalid
states may fail loudly natively even where erased Node syntax would continue.

Every read stays at its original evaluation point, including repeated reads
and callback boundaries. No optional chain, skip, hoist, default, or initializer
assertion is introduced. Only reviewed numeric bitwise U-zero operands may use
`?? 0`; none were selected in Wave A. Existing undefined!/null! initializers
remain untouched. Text insertions preserve CRLF; no source printer is used.

The adapter parses current text with stock npm TypeScript 6.0.3. Its addresses
are parsed element-access expressions plus occurrence and total counts, as in
adaptation 30. Ledger line/column numbers document the original source and are
never edit addresses. It validates every occurrence count, rejects defaulted
required reads, optional access chains and pure stores, plans every file before
writing, and checks that the source has not changed before applying edits.

## Tree, toolchain and waves

Base main: `e011f8f`. Required prerequisite merges: adaptation 10 at `a3ef0dc`
and adaptation 30 at `07f637c`; merge HEAD before this unit was `437fe1b`.
The tested tree contains setup, 10, 30, and this adapter. Adaptation 20 is absent.
No TypeScript source is committed. No files in the other partitions are edited.

Toolchain setup: Go 0s, clang 0s, Node 0s, submodules 0s, build cache 160s,
total 160s. `nproc=5`, CPU quota 4. Environment is
`source /workspace/adamic-tools/env.sh`; Node 24.19.0, Go 1.27.1, clang 20.1.8.

The fresh Adamic census has 796 findings across the five requested codes,
rather than the earlier stock census's 798. `wave-boundaries.json` freezes them
in numeric line, column, code order. Waves address ranks 1-200, 201-400,
401-600, and 601-796 of that frozen input, so declined findings do not consume
later waves repeatedly. `findings-wave-a.json` records every Wave A disposition.

## Wave A

Ranks 1-200, through `checker.ts:16133:46`: **102 required-read assertions**,
**6 declined indexed reads**, **164 findings removed**, **36 findings declined**.
The required-read ledger includes related reads needed to resolve a selected
finding, such as both populated lists of a parallel comparison. It never
asserts every indexed expression in a selected function or source range.

The real Adamic loader/checker checked 78 prepared roots with its unchanged
fixed strictness options and no implicit optional-declaration adaptation.
The census includes all five-code findings in checker.ts, whatever their cause.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| checker.ts before | 395 | 245 | 113 | 43 | 0 | 796 |
| checker.ts after Wave A | 309 | 191 | 94 | 38 | 0 | 632 |

`remaining-wave-a.json` lists **every remaining finding** with its complete
chain and reason. The 36 remaining Wave A findings concern explicit absence
handling or non-indexed iterator, Set, regex callback, and optional-return
contracts. Findings outside the first 200 are marked with their deferred wave,
not falsely claimed reviewed or repaired. This file does not yet pass the full
Adamic checker; other diagnostic codes and compiler feature refusals remain.

## Proof

Stock TypeScript emitted JavaScript is byte-identical to the before tree.
The verifier checks the site contract and idempotence independently of the
upstream suite. A real second CLI run reports **0 assertions, 0 U-zero edits**.
`proof-wave-a.json` records hashes, CRLF counts, and observations.

Mutant: the new `jsxFragmentPragma[0]!` at `checker.ts:2446:75` is replaced
with `(jsxFragmentPragma[0] ?? 0)` in an isolated checker snapshot. The stock
emitted-JavaScript comparison fails with exit 1, and the site contract fails
with `required read defaulted: checker.ts:2446:75`, exit 1. The oracle tree
never contained this mutant. The census and full oracle were not run on it.

Default stage 3 oracle: **pass**, 106,367 passing, 0 failing, 0 pending. Baseline differences: **0**. Install, build, and tests exits: 0, 0, 0. Wall time 529.947s. `baseline-wave-a.diff` records the exact diff.

The default oracle runs all suites, unfiltered, with four workers and
`--light=false`; lint is disabled by the harness. Native Adamic tsc compilation,
browser integration, ESLint-rule integrations, and the full repository Go gate
are not covered by this source-adaptation unit.

## Reproduction

Prepare a pinned scratch checkout with setup, adaptation 10 and adaptation 30,
then retain a before snapshot of checker.ts. Run with stock TypeScript 6.0.3
on NODE_PATH or CENSUS_TYPESCRIPT. All test output goes to logs.

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
unit=stage3/adapt/31-indexed-reads-checker
bash "$unit/census.sh" /tmp/checker-before /tmp/checker-before-census > /tmp/checker-before-census.log 2>&1
node "$unit/adapt.cjs" /tmp/checker-after > /tmp/checker-adapt.log 2>&1
bash "$unit/census.sh" /tmp/checker-after /tmp/checker-after-census > /tmp/checker-after-census.log 2>&1
node "$unit/verify.cjs" /tmp/checker-before /tmp/checker-after > /tmp/checker-verify.log 2>&1
node "$unit/adapt.cjs" /tmp/checker-after > /tmp/checker-idempotence.log 2>&1
stage3/oracle/run.sh /tmp/checker-after /tmp/checker-oracle > /tmp/checker-oracle.log 2>&1
```

Observed Wave A scratch tree: `/tmp/checker-wave-tree`; original checker
snapshot: `/tmp/checker-original`; census: `/tmp/checker-wave-{before,a-after}`;
oracle: `/tmp/checker-wave-a-oracle`. Command logs are
`/tmp/checker-wave-a-{adapt,verify,idempotence,census,oracle,mutant,mutant-contract}.log`.

## Site ledger

Each row represents exactly one parsed indexed read. Occurrence and count
addresses are retained in sites.json. One invariant applies at each original
read, even when several reads share a construction obligation.

| Site | Read | Class | Action | Invariant or decline reason |
| --- | --- | --- | --- | --- |
| checker.ts:2446:75 | `jsxFragmentPragma[0]` | U-endpoint | assert | A present parser pragma array contains the first parsed pragma; the scalar branch keeps its original value. |
| checker.ts:2492:55 | `jsxPragma[0]` | U-endpoint | assert | A present parser pragma array contains the first parsed pragma; the scalar branch keeps its original value. |
| checker.ts:5043:67 | `referenceParent.arguments[0]` | U-endpoint | assert | This resolved import-call reference has its required module-specifier argument in the parser-built argument list. |
| checker.ts:5311:54 | `mergedSymbols[symbol.mergeId]` | U-table | decline | Merged-symbol lookup explicitly tests absence and falls back to the original symbol. |
| checker.ts:5915:74 | `accessibleSymbolChain[0]` | U-endpoint | assert | A successful accessible-symbol lookup constructs a nonempty populated chain whose root is required. |
| checker.ts:6584:91 | `classDeclarations[0]` | U-endpoint | assert | The explicit positive or single-element length guard precedes this first read from a populated declaration, type-node, type, or signature list. |
| checker.ts:6602:67 | `decls[0]` | U-endpoint | assert | The explicit positive or single-element length guard precedes this first read from a populated declaration, type-node, type, or signature list. |
| checker.ts:6921:60 | `typeArgumentNodes[0]` | U-endpoint | assert | The explicit positive or single-element length guard precedes this first read from a populated declaration, type-node, type, or signature list. |
| checker.ts:6982:49 | `types[0]` | U-endpoint | assert | The explicit positive or single-element length guard precedes this first read from a populated declaration, type-node, type, or signature list. |
| checker.ts:7009:65 | `texts[0]` | U-parallel | assert | Template-literal construction supplies a head text and one following text for each interpolated type. |
| checker.ts:7014:112 | `texts[i + 1]` | U-parallel | assert | Template-literal construction supplies a head text and one following text for each interpolated type. |
| checker.ts:7376:43 | `resolved.callSignatures[0]` | U-endpoint | assert | The explicit positive or single-element length guard precedes this first read from a populated declaration, type-node, type, or signature list. |
| checker.ts:7382:43 | `resolved.constructSignatures[0]` | U-endpoint | assert | The explicit positive or single-element length guard precedes this first read from a populated declaration, type-node, type, or signature list. |
| checker.ts:7424:71 | `typeArguments[0]` | U-endpoint | assert | The recognized global Array or ReadonlyArray type has its one required element type argument. |
| checker.ts:7427:62 | `typeArguments[0]` | U-endpoint | assert | The recognized global Array or ReadonlyArray type has its one required element type argument. |
| checker.ts:7432:94 | `(type.target as TupleType).elementFlags[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7439:47 | `(type.target as TupleType).elementFlags[i]` | U-parallel | assert | Tuple construction aligns populated flags with constituent nodes at each visited tuple position. |
| checker.ts:7440:67 | `labeledElementDeclarations?.[i]` | U-parallel | decline | Tuple labels are optional and the following branch explicitly handles a missing label. |
| checker.ts:7447:97 | `tupleConstituentNodes[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7448:45 | `tupleConstituentNodes[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7452:179 | `tupleConstituentNodes[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7452:207 | `tupleConstituentNodes[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7453:104 | `tupleConstituentNodes[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7454:41 | `tupleConstituentNodes[i]` | U-parallel | assert | Tuple construction aligns populated element flags and constituent nodes by tuple position; each rewritten constituent remains defined. |
| checker.ts:7485:75 | `outerTypeParameters[i]` | U-loop | assert | The outer-type-parameter traversal bounds i by the populated parameter list, including after increment. |
| checker.ts:7489:81 | `outerTypeParameters[i]` | U-loop | assert | The outer-type-parameter traversal bounds i by the populated parameter list, including after increment. |
| checker.ts:7523:62 | `typeArguments[typeParameterCount - 1]` | U-parallel | assert | The positive remaining parameter count is bounded by both populated type-parameter and type-argument lists. |
| checker.ts:7524:63 | `type.target.typeParameters[typeParameterCount - 1]` | U-parallel | assert | The positive remaining parameter count is bounded by both populated type-parameter and type-argument lists. |
| checker.ts:7700:50 | `properties[properties.length - 1]` | U-endpoint | assert | The truncation branch runs inside populated property traversal and requires the final property. |
| checker.ts:7744:34 | `context.reverseMappedStack[context.reverseMappedStack.length - 1 - i]` | U-position | assert | The reverse-mapped stack has at least depth populated entries before the bounded backward traversal. |
| checker.ts:7909:50 | `types[0]` | U-endpoint | assert | The truncation length guards establish nonempty populated type input before first or last type serialization. |
| checker.ts:7913:50 | `types[types.length - 1]` | U-endpoint | assert | The truncation length guards establish nonempty populated type input before first or last type serialization. |
| checker.ts:7931:63 | `types[types.length - 1]` | U-endpoint | assert | The truncation length guards establish nonempty populated type input before first or last type serialization. |
| checker.ts:8013:36 | `getExpandedParameters(signature, /*skipUnionExpanding*/ true)[0]` | U-endpoint | assert | Skipping union expansion always returns a single populated outer parameter-list entry, even when that inner list is empty. |
| checker.ts:8027:65 | `expandedParams[expandedParams.length - 1]` | U-endpoint | decline | The callback compares a potentially absent endpoint directly; the outer list is asserted at its original read. |
| checker.ts:8208:43 | `expandedParams[pIndex]` | U-loop | assert | The parameter-scope loop traverses the populated expanded parameter list under its length bound. |
| checker.ts:8479:40 | `accessibleSymbolChain[0]` | U-endpoint | assert | A successful accessible-symbol lookup constructs a nonempty populated chain whose root is required. |
| checker.ts:8482:83 | `accessibleSymbolChain[0]` | U-endpoint | assert | A successful accessible-symbol lookup constructs a nonempty populated chain whose root is required. |
| checker.ts:8491:64 | `parents[i]` | U-position | assert | indices is a permutation of positions in the populated parents list; sorting changes only that permutation. |
| checker.ts:8559:28 | `chain[index]` | U-position | assert | lookupTypeParameterNodes asserts a valid nonnegative chain index; the successor read additionally checks index < length - 1. |
| checker.ts:8572:36 | `chain[index + 1]` | U-position | assert | lookupTypeParameterNodes asserts a valid nonnegative chain index; the successor read additionally checks index < length - 1. |
| checker.ts:8667:22 | `chain[0]` | U-endpoint | assert | lookupSymbolChain constructs a nonempty populated chain or a singleton containing the requested symbol. |
| checker.ts:8672:58 | `chain[0]` | U-endpoint | assert | lookupSymbolChain constructs a nonempty populated chain or a singleton containing the requested symbol. |
| checker.ts:8678:65 | `chain[0]` | U-endpoint | assert | lookupSymbolChain constructs a nonempty populated chain or a singleton containing the requested symbol. |
| checker.ts:8690:61 | `chain[0]` | U-endpoint | assert | lookupSymbolChain constructs a nonempty populated chain or a singleton containing the requested symbol. |
| checker.ts:8697:65 | `chain[0]` | U-endpoint | assert | lookupSymbolChain constructs a nonempty populated chain or a singleton containing the requested symbol. |
| checker.ts:8756:32 | `chain[index]` | U-position | assert | Symbol-chain serialization starts at length - 1 and recurses toward the root of a nonempty populated chain. |
| checker.ts:8888:32 | `chain[index]` | U-position | assert | Symbol-chain serialization starts at length - 1 and recurses toward the root of a nonempty populated chain. |
| checker.ts:8913:32 | `chain[index]` | U-position | assert | Symbol-chain serialization starts at length - 1 and recurses toward the root of a nonempty populated chain. |
| checker.ts:9379:42 | `chain[0]` | U-endpoint | assert | lookupSymbolChain constructs a nonempty populated chain or a singleton containing the requested symbol. |
| checker.ts:9498:41 | `group[0]` | U-endpoint | assert | The grouped reexport list has length greater than one before its first module specifier is read. |
| checker.ts:9518:88 | `statements[i]` | U-position | assert | indicesOf selects positions in the same populated statements list; filtering preserves those positions. |
| checker.ts:9519:114 | `statements[i]` | U-position | assert | indicesOf selects positions in the same populated statements list; filtering preserves those positions. |
| checker.ts:9588:41 | `symbols[symbols.length - 1]` | U-endpoint | assert | Truncation runs during symbol traversal and selects the last populated symbol-table snapshot entry. |
| checker.ts:9597:21 | `deferredPrivatesStack[deferredPrivatesStack.length - 1]` | U-position | assert | Private serialization pushes populated Map frames; the current or root frame is required until its matching pop. |
| checker.ts:9858:17 | `deferredPrivatesStack[isExternalImportAlias ? 0 : (deferredPrivatesStack.length - 1)]` | U-position | assert | Private serialization pushes populated Map frames; the current or root frame is required until its matching pop. |
| checker.ts:9965:61 | `props[props.length - 1]` | U-endpoint | assert | The truncation branch runs during populated property traversal and selects its final property. |
| checker.ts:9966:65 | `props[props.length - 1]` | U-endpoint | assert | The truncation branch runs during populated property traversal and selects its final property. |
| checker.ts:10098:38 | `memberProps[memberProps.length - 1]` | U-endpoint | assert | The truncation branch runs during populated property traversal and selects its final property. |
| checker.ts:10099:71 | `last.declarations[0]` | U-endpoint | decline | Enum initialization explicitly tests a missing first declaration and uses undefined when absent. |
| checker.ts:10099:108 | `last.declarations[0]` | U-endpoint | decline | Enum initialization explicitly tests a missing first declaration and uses undefined when absent. |
| checker.ts:10099:149 | `last.declarations[0]` | U-endpoint | decline | Enum initialization explicitly tests a missing first declaration and uses undefined when absent. |
| checker.ts:10212:40 | `props[0]` | U-endpoint | assert | Namespace serialization is invoked with nonempty populated properties belonging to the same parent symbol. |
| checker.ts:10997:65 | `signatures[i]` | U-parallel | assert | Equal signature-list lengths precede bounded comparison; both lists contain constructed signatures at i. |
| checker.ts:10997:80 | `baseSigs[i]` | U-parallel | assert | Equal signature-list lengths precede bounded comparison; both lists contain constructed signatures at i. |
| checker.ts:11021:29 | `signatures[0]` | U-endpoint | assert | A private or protected constructor flag found during signature traversal requires a nonempty populated signature list. |
| checker.ts:11192:23 | `types[i]` | U-loop | assert | formatUnionTypes traverses populated union constituents with i < types.length. |
| checker.ts:11199:86 | `types[i + count - 1]` | U-position | assert | Union constituents are nonempty; the i + count bound selects the populated final constituent of the matched enum span. |
| checker.ts:11199:140 | `(baseType as UnionType).types[count - 1]` | U-position | assert | Union constituents are nonempty; the i + count bound selects the populated final constituent of the matched enum span. |
| checker.ts:11226:51 | `type.symbol.declarations[0]` | U-endpoint | assert | A declared symbol carries a nonempty parser-built declarations list; the fallback name branch also checks its length. |
| checker.ts:11305:31 | `symbol.declarations[0]` | U-endpoint | assert | A declared symbol carries a nonempty parser-built declarations list; the fallback name branch also checks its length. |
| checker.ts:11505:45 | `resolutionTargets[i]` | U-parallel | assert | Resolution targets and property names are pushed and popped together; the backward loop stays in their populated active range. |
| checker.ts:11505:67 | `resolutionPropertyNames[i]` | U-parallel | assert | Resolution targets and property names are pushed and popped together; the backward loop stays in their populated active range. |
| checker.ts:12040:47 | `(symbol.escapedName as string).split("@")[1]` | U-endpoint | assert | Private names beginning __# are generated with an @ separator followed by the required original private identifier. |
| checker.ts:12061:47 | `(symbol.escapedName as string).split("@")[1]` | U-endpoint | assert | Private names beginning __# are generated with an @ separator followed by the required original private identifier. |
| checker.ts:12342:33 | `declarations[i]` | U-parallel | assert | The asserted equal lengths align populated declarations with types traversed by filter. |
| checker.ts:12583:73 | `declaration.statements[0]` | U-endpoint | assert | The JSON source-file branch returns for an empty statements list before reading its populated first expression statement. |
| checker.ts:13161:23 | `signatures[0]` | U-endpoint | assert | The mixin test checks one construct signature, then one rest parameter, before reading the respective populated first entries. |
| checker.ts:13163:54 | `s.parameters[0]` | U-endpoint | assert | The mixin test checks one construct signature, then one rest parameter, before reading the respective populated first entries. |
| checker.ts:13243:69 | `baseConstructorType.symbol.declarations[0]` | U-endpoint | assert | A declared symbol carries a nonempty parser-built declarations list; the fallback name branch also checks its length. |
| checker.ts:13313:69 | `type.elementFlags[i]` | U-parallel | assert | Tuple construction supplies one populated element flag for each tuple type parameter or element visited by the callback. |
| checker.ts:13347:49 | `constructors[0]` | U-endpoint | assert | The no-constructors error returns before selecting the first populated matching constructor signature. |
| checker.ts:13381:20 | `outerTypeParameters[last]` | U-parallel | assert | A present captured outer-parameter list is nonempty and has matching populated applied arguments in outer-to-inner order. |
| checker.ts:13381:57 | `typeArguments[last]` | U-parallel | assert | A present captured outer-parameter list is nonempty and has matching populated applied arguments in outer-to-inner order. |
| checker.ts:14190:32 | `sig.parameters[restIndex]` | U-endpoint | assert | signatureHasRestParameter requires a populated last parameter; restIndex selects that parameter. |
| checker.ts:14208:31 | `restType.target.elementFlags[i]` | U-parallel | assert | Tuple construction supplies one populated element flag for each tuple type parameter or element visited by the callback. |
| checker.ts:14219:134 | `type.target.elementFlags[i]` | U-parallel | assert | Tuple construction supplies one populated element flag for each tuple type parameter or element visited by the callback. |
| checker.ts:14231:48 | `names[i]` | U-position | assert | Duplicate positions were collected while traversing the same populated names list; rewrites preserve defined names. |
| checker.ts:14287:44 | `signatureLists[i]` | U-grid | assert | The bounded signature-list traversal selects a populated constituent list; an inner list may be empty and remains checked nearby. |
| checker.ts:14299:41 | `signatureLists[i]` | U-grid | assert | The bounded signature-list traversal selects a populated constituent list; an inner list may be empty and remains checked nearby. |
| checker.ts:14300:46 | `signatureLists[i]` | U-grid | assert | The bounded signature-list traversal selects a populated constituent list; an inner list may be empty and remains checked nearby. |
| checker.ts:14317:17 | `signatureLists[i]` | U-grid | assert | The bounded signature-list traversal selects a populated constituent list; an inner list may be empty and remains checked nearby. |
| checker.ts:14318:17 | `signatureLists[i]` | U-grid | assert | The bounded signature-list traversal selects a populated constituent list; an inner list may be empty and remains checked nearby. |
| checker.ts:14321:37 | `signatureLists[i]` | U-grid | assert | The bounded signature-list traversal selects a populated constituent list; an inner list may be empty and remains checked nearby. |
| checker.ts:14348:32 | `signatureLists[indexWithLengthOverOne == undefined ? indexWithLengthOverOne : 0]` | U-position | assert | The union has populated constituent signature lists; the chosen overload index or root index selects an existing list. |
| checker.ts:14375:28 | `sourceParams[i]` | U-parallel | assert | Equal populated parameter-list lengths precede the bounded comparison at their shared index. |
| checker.ts:14376:28 | `targetParams[i]` | U-parallel | assert | Equal populated parameter-list lengths precede the bounded comparison at their shared index. |
| checker.ts:14485:49 | `types[0]` | U-endpoint | assert | Union and intersection construction use populated nonempty constituent lists before selecting their first type. |
| checker.ts:14531:58 | `getSignaturesOfType(types[i], SignatureKind.Construct)[0]` | U-parallel | assert | The bounded mixin position selects a populated type; its mixin predicate requires one populated construct signature. |
| checker.ts:14531:78 | `types[i]` | U-parallel | assert | The bounded mixin position selects a populated type; its mixin predicate requires one populated construct signature. |
| checker.ts:14547:23 | `type.types[i]` | U-loop | assert | Intersection-member resolution traverses the populated constituent list at a bounded index. |
| checker.ts:14582:30 | `indexInfos[i]` | U-loop | assert | appendIndexInfo traverses populated index records under its length bound and preserves defined records on replacement. |
| checker.ts:14782:42 | `types[0]` | U-endpoint | assert | The length-two intersection guard precedes the first populated constituent flags read. |
| checker.ts:14987:66 | `(type as IntersectionType).types[0]` | U-endpoint | assert | Union and intersection construction use populated nonempty constituent lists before selecting their first type. |
| checker.ts:15172:87 | `type.target.elementFlags[i]` | U-parallel | assert | Tuple construction supplies one populated element flag for each tuple type parameter or element visited by the callback. |
| checker.ts:15439:77 | `t.target.elementFlags[i]` | U-parallel | assert | Tuple construction supplies one populated element flag for each tuple type parameter or element visited by the callback. |
| checker.ts:15958:67 | `(isReadonlyArraySymbol(t.symbol.parent) ? globalReadonlyArrayType : globalArrayType).typeParameters[0]` | U-endpoint | assert | The recognized global array interface has a required first type parameter used by its member mapper. |
| checker.ts:16133:46 | `typeParameters[i]` | U-loop | assert | The minimum-type-argument loop traverses populated parameters under i < typeParameters.length. |

## Remaining Wave A findings

| Site | Code | Reason |
| --- | --- | --- |
| checker.ts:1779:9 | TS2322 | Not an indexed read: arrayFrom(Set<Signature>) returns a possibly-undefined iterator payload under the current loader typing. |
| checker.ts:5311:45 | TS2322 | Declined U-table probe: mergedSymbols lookup tests absence and falls back to the original symbol. |
| checker.ts:8062:153 | TS18048 | Not an indexed read: regex split/map callback parameter is possibly undefined under the current string/regex typing. |
| checker.ts:8566:17 | TS2322 | Not an indexed read: copying an optional Set infers a possibly-undefined element type. |
| checker.ts:8861:21 | TS2322 | Not an indexed read: copying an optional Set infers a possibly-undefined element type. |
| checker.ts:9363:17 | TS2322 | Not an indexed read: copying an optional Set infers a possibly-undefined element type. |
| checker.ts:9591:37 | TS2345 | Not an indexed read: Array.from(symbolTable.values()) iterator payload is inferred as possibly undefined. |
| checker.ts:10022:63 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10022:86 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10026:97 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10033:67 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10033:79 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10053:53 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10061:90 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10062:69 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10063:69 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10064:47 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10064:93 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10065:122 | TS18048 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10066:103 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10070:50 | TS2345 | Not an indexed read: arrayFrom of namespace export Map/Set iterators introduces possibly-undefined Symbol payloads propagated through callbacks. |
| checker.ts:10319:83 | TS2345 | Not an indexed read: getNonInheritedProperties returns arrayFrom(seen.values()), whose iterator payload is possibly undefined. |
| checker.ts:10320:51 | TS2345 | Not an indexed read: getNonInheritedProperties returns arrayFrom(seen.values()), whose iterator payload is possibly undefined. |
| checker.ts:10336:86 | TS2345 | Not an indexed read: getNonInheritedProperties returns arrayFrom(seen.values()), whose iterator payload is possibly undefined. |
| checker.ts:12034:30 | TS18048 | Not an indexed read: getSourceFileOfNode has an optional return; checking its input declaration cannot change that return contract. |
| checker.ts:13986:47 | TS2345 | Not an indexed read: arrayFrom(assignments.values()) propagates a possibly-undefined iterator payload into the for-of binding. |
| checker.ts:13990:49 | TS2345 | Not an indexed read: arrayFrom(assignments.values()) propagates a possibly-undefined iterator payload into the for-of binding. |
| checker.ts:14643:66 | TS2345 | Not an indexed read: arrayFrom(Map.values()) infers a possibly-undefined Symbol iterator payload. |
| checker.ts:15145:9 | TS2322 | Not an indexed read: arrayFrom(Map.values()) infers a possibly-undefined Symbol iterator payload. |
| checker.ts:15695:41 | TS18048 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |
| checker.ts:15697:22 | TS18048 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |
| checker.ts:15697:47 | TS18048 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |
| checker.ts:15700:51 | TS18048 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |
| checker.ts:15701:42 | TS2345 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |
| checker.ts:15704:43 | TS2345 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |
| checker.ts:15706:52 | TS2345 | Not an indexed read: arrayFrom(propSet.values()) propagates a possibly-undefined Symbol iterator payload into the for-of binding. |

## Wave B

Frozen ranks 201-400: 132 new required-read assertions, 7 reviewed read declines; 196 frozen findings removed, 4 declined. Net census reduction: 196. Related alias repairs can also remove findings across a wave boundary.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| checker.ts before | 309 | 191 | 94 | 38 | 0 | 632 |
| checker.ts after | 211 | 143 | 53 | 29 | 0 | 436 |

Default oracle: **106,367 passing, 0 failing, 0 pending**, all phase exits 0, **empty baseline diff**, 352.725s. `oracle-wave-b.json` and `baseline-wave-b.diff` retain this observation.

Stock emitted JavaScript is byte-identical; site contract and CLI idempotence pass. CRLF preserved. The isolated mutant replaces `declaration.parameters[i]!` at checker.ts:16203 with `(declaration.parameters[i] ?? 0)`; the emitted-JavaScript verifier exits 1. No mutant is run in the real oracle tree. `proof-wave-b.json` records hashes and counts. Every remaining finding is recorded in `remaining-wave-b.json`; every frozen-wave disposition is recorded in `findings-wave-b.json`.

Commands use the reproduction above with /tmp/checker-wave-b-final-after and /tmp/checker-wave-b-final-oracle. Output logs: /tmp/checker-wave-b-final-{adapt,census,verify,idempotence,oracle}.log and /tmp/checker-wave-b-mutant.log.

| Site | Read | Class | Action | Invariant or reason |
| --- | --- | --- | --- | --- |
| checker.ts:16165:63 | `typeParameters[i]` | U-loop | assert | The bounded parameter traversal reads a populated type-parameter or parser parameter list at its current index. |
| checker.ts:16203:31 | `declaration.parameters[i]` | U-loop | assert | The bounded parameter traversal reads a populated type-parameter or parser parameter list at its current index. |
| checker.ts:16376:34 | `symbol.declarations[i - 1]` | U-position | assert | The overload-implementation check requires i > 0 and reads the populated previous declaration in the same bounded declaration traversal. |
| checker.ts:16554:49 | `signature.parameters[signature.parameters.length - 1]` | U-endpoint | assert | signatureHasRestParameter identifies a signature with a populated required last rest parameter. |
| checker.ts:16706:43 | `declaration.parameters[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:16780:89 | `typeParameters[index]` | U-position | assert | The infer declaration is a member of the enclosing parser-built type-argument list; its index selects the matching populated type parameter. |
| checker.ts:16878:33 | `types[i]` | U-loop | assert | Type-ID compression bounds each current or lookahead position by the populated type list length. |
| checker.ts:16880:46 | `types[i + count]` | U-loop | assert | Type-ID compression bounds each current or lookahead position by the populated type list length. |
| checker.ts:17032:80 | `typeArguments[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17032:129 | `typeArguments[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17262:114 | `(checkNode as TupleTypeNode).elements[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17262:156 | `(extendsNode as TupleTypeNode).elements[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17352:65 | `typeArgs[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17353:64 | `typeArgs[1]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17855:42 | `elementFlags[0]` | U-endpoint | assert | The single- or two-element shape guard precedes the fixed-position read in a populated type-argument, tuple-element, flag, or parameter list. |
| checker.ts:17886:31 | `elementFlags[i]` | U-parallel | assert | Tuple construction aligns populated element types and flags; the bounded traversal and callbacks use their shared position. |
| checker.ts:17943:69 | `target.elementFlags[i]` | U-parallel | assert | Tuple construction aligns populated element types and flags; the bounded traversal and callbacks use their shared position. |
| checker.ts:17945:75 | `target.elementFlags[i]` | U-parallel | assert | Tuple construction aligns populated element types and flags; the bounded traversal and callbacks use their shared position. |
| checker.ts:17962:26 | `elementTypes[i]` | U-parallel | assert | Tuple construction aligns populated element types and flags; the bounded traversal and callbacks use their shared position. |
| checker.ts:17963:27 | `target.elementFlags[i]` | U-parallel | assert | Tuple construction aligns populated element types and flags; the bounded traversal and callbacks use their shared position. |
| checker.ts:17984:63 | `type.target.elementFlags[n]` | U-parallel | assert | Tuple construction aligns populated element types and flags; the bounded traversal and callbacks use their shared position. |
| checker.ts:17984:92 | `type.target.labeledElementDeclarations?.[n]` | U-parallel | decline | Absence is handled nearby: tuple labels are optional metadata and the original optional indexed read is passed through unchanged. |
| checker.ts:17998:17 | `expandedFlags[i]` | U-parallel | assert | Expanded tuple flags are pushed with expanded types; required/rest bounds and the slice callback select populated flag positions. |
| checker.ts:18002:142 | `expandedFlags[firstRestIndex + i]` | U-parallel | assert | Expanded tuple flags are pushed with expanded types; required/rest bounds and the slice callback select populated flag positions. |
| checker.ts:18095:124 | `typeSet[len - 1]` | U-endpoint | assert | The positive type-set length guard precedes selection of its populated final ordered type. |
| checker.ts:18143:28 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18205:23 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18225:27 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18244:31 | `(type as IntersectionType).types[0]` | U-position | assert | The constrained-type-variable marker identifies a populated two-part intersection; index and 1 - index select its two constituents. |
| checker.ts:18256:35 | `(type as IntersectionType).types[0]` | U-position | assert | The constrained-type-variable marker identifies a populated two-part intersection; index and 1 - index select its two constituents. |
| checker.ts:18258:48 | `(type as IntersectionType).types[1 - index]` | U-position | assert | The constrained-type-variable marker identifies a populated two-part intersection; index and 1 - index select its two constituents. |
| checker.ts:18269:34 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18271:39 | `(type as IntersectionType).types[0]` | U-position | assert | The constrained-type-variable marker identifies a populated two-part intersection; index and 1 - index select its two constituents. |
| checker.ts:18272:29 | `(type as IntersectionType).types[index]` | U-position | assert | The constrained-type-variable marker identifies a populated two-part intersection; index and 1 - index select its two constituents. |
| checker.ts:18272:114 | `(type as IntersectionType).types[1 - index]` | U-position | assert | The constrained-type-variable marker identifies a populated two-part intersection; index and 1 - index select its two constituents. |
| checker.ts:18318:20 | `types[0]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18321:47 | `types[0]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18321:83 | `types[1]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18323:27 | `types[0]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18323:41 | `types[1]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18324:24 | `types[index]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18324:50 | `types[1 - index]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18382:24 | `namedUnions[0]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18437:20 | `types[0]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18452:39 | `types[0]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18452:84 | `types[1]` | U-endpoint | assert | The explicit singleton or length-two guard precedes each fixed or complementary position read in a populated type list. |
| checker.ts:18518:23 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18565:23 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18583:35 | `types[i]` | U-loop | assert | removeFromEach traverses populated types under its length bound and replaces each slot with a defined filtered type. |
| checker.ts:18600:23 | `types[i]` | U-loop | assert | Backward or bounded type traversal selects a populated constituent; removals preserve all positions needed by subsequent iterations. |
| checker.ts:18713:20 | `typeSet[0]` | U-position | assert | Positive singleton or length-two guards select populated intersection constituents and their complementary variable/primitive positions. |
| checker.ts:18716:34 | `typeSet[0]` | U-position | assert | Positive singleton or length-two guards select populated intersection constituents and their complementary variable/primitive positions. |
| checker.ts:18717:34 | `typeSet[typeVarIndex]` | U-position | assert | Positive singleton or length-two guards select populated intersection constituents and their complementary variable/primitive positions. |
| checker.ts:18718:35 | `typeSet[1 - typeVarIndex]` | U-position | assert | Positive singleton or length-two guards select populated intersection constituents and their complementary variable/primitive positions. |
| checker.ts:18758:78 | `(t as UnionType).types[0]` | U-endpoint | assert | Canonical union construction retains at least two populated constituents; nullable-order checks inspect its first two positions. |
| checker.ts:18763:79 | `(t as UnionType).types[0]` | U-endpoint | assert | Canonical union construction retains at least two populated constituents; nullable-order checks inspect its first two positions. |
| checker.ts:18763:131 | `(t as UnionType).types[1]` | U-endpoint | assert | Canonical union construction retains at least two populated constituents; nullable-order checks inspect its first two positions. |
| checker.ts:18819:21 | `types[j]` | U-position | assert | Cross-product traversal bounds j by populated inputs; nonempty union factors supply the modulo-selected constituent. |
| checker.ts:18822:39 | `sourceTypes[n % length]` | U-position | assert | Cross-product traversal bounds j by populated inputs; nonempty union factors supply the modulo-selected constituent. |
| checker.ts:18851:41 | `types[1 - emptyIndex]` | U-position | assert | The empty-index lookup runs only for two populated types; a successful lookup selects the complementary type. |
| checker.ts:19052:20 | `texts[0]` | U-parallel | assert | Template construction has a populated head text and one following text per populated interpolation type, including nested template spans. |
| checker.ts:19065:63 | `newTypes[0]` | U-endpoint | assert | The singleton normalized-type guard precedes the populated first pattern-type read and return. |
| checker.ts:19066:24 | `newTypes[0]` | U-endpoint | assert | The singleton normalized-type guard precedes the populated first pattern-type read and return. |
| checker.ts:19078:27 | `types[i]` | U-parallel | assert | Template construction has a populated head text and one following text per populated interpolation type, including nested template spans. |
| checker.ts:19081:29 | `texts[i + 1]` | U-parallel | assert | Template construction has a populated head text and one following text per populated interpolation type, including nested template spans. |
| checker.ts:19084:29 | `(t as TemplateLiteralType).texts[0]` | U-parallel | assert | Template construction has a populated head text and one following text per populated interpolation type, including nested template spans. |
| checker.ts:19086:29 | `texts[i + 1]` | U-parallel | assert | Template construction has a populated head text and one following text per populated interpolation type, including nested template spans. |
| checker.ts:19091:28 | `texts[i + 1]` | U-parallel | assert | Template construction has a populated head text and one following text per populated interpolation type, including nested template spans. |
| checker.ts:19149:25 | `texts[0]` | U-endpoint | decline | Absence is handled by the equality comparison at this exact read; the required charAt/slice/type reads in the selected branch are separate ledger entries. |
| checker.ts:19149:52 | `texts[0]` | U-parallel | assert | Template heads are populated; an empty head in this mapping path belongs to a template with its required first interpolation type. |
| checker.ts:19149:87 | `texts[0]` | U-parallel | assert | Template heads are populated; an empty head in this mapping path belongs to a template with its required first interpolation type. |
| checker.ts:19149:126 | `texts[0]` | U-endpoint | decline | Absence is handled by the equality comparison at this exact read; the required charAt/slice/type reads in the selected branch are separate ledger entries. |
| checker.ts:19149:174 | `types[0]` | U-parallel | assert | Template heads are populated; an empty head in this mapping path belongs to a template with its required first interpolation type. |
| checker.ts:19151:25 | `texts[0]` | U-endpoint | decline | Absence is handled by the equality comparison at this exact read; the required charAt/slice/type reads in the selected branch are separate ledger entries. |
| checker.ts:19151:52 | `texts[0]` | U-parallel | assert | Template heads are populated; an empty head in this mapping path belongs to a template with its required first interpolation type. |
| checker.ts:19151:87 | `texts[0]` | U-parallel | assert | Template heads are populated; an empty head in this mapping path belongs to a template with its required first interpolation type. |
| checker.ts:19151:126 | `texts[0]` | U-endpoint | decline | Absence is handled by the equality comparison at this exact read; the required charAt/slice/type reads in the selected branch are separate ledger entries. |
| checker.ts:19151:174 | `types[0]` | U-parallel | assert | Template heads are populated; an empty head in this mapping path belongs to a template with its required first interpolation type. |
| checker.ts:20194:34 | `types[types.length - 1]` | U-endpoint | assert | A canonical intersection has a nonempty populated constituent list before its last type is merged into a spread. |
| checker.ts:20414:28 | `(node as TupleTypeNode).elements[0]` | U-endpoint | assert | The one-element tuple guard precedes reading the populated first parser type node. |
| checker.ts:20543:30 | `items[i]` | U-loop | assert | instantiateList traverses populated caller input under its length bound, retaining each read at its original callback point. |
| checker.ts:20549:50 | `items[i]` | U-loop | assert | instantiateList traverses populated caller input under its length bound, retaining each read at its original callback point. |
| checker.ts:20573:59 | `sources[0]` | U-parallel | assert | Unary or array mappers align populated source and target positions; a missing target list retains its existing anyType branch. |
| checker.ts:20573:81 | `targets[0]` | U-parallel | assert | Unary or array mappers align populated source and target positions; a missing target list retains its existing anyType branch. |
| checker.ts:20585:42 | `targets[i]` | U-parallel | assert | Unary or array mappers align populated source and target positions; a missing target list retains its existing anyType branch. |
| checker.ts:20742:13 | `type.symbol.declarations[0]` | U-endpoint | assert | The anonymous non-reference type originates from a declared symbol with a populated first declaration. |
| checker.ts:20820:31 | `tp.symbol.declarations[0]` | U-endpoint | assert | The exactly-one-declaration guard precedes the first declaration and the nested reference traversal uses that same symbol contract. |
| checker.ts:20841:47 | `tp.symbol.declarations[0]` | U-endpoint | assert | The exactly-one-declaration guard precedes the first declaration and the nested reference traversal uses that same symbol contract. |
| checker.ts:20936:27 | `elementFlags[i]` | U-parallel | assert | Mapped tuple construction supplies a populated flag for each element type visited at the same position. |
| checker.ts:21037:29 | `activeTypeMappersCaches[index == -1 ? index : activeTypeMappersCount - 1]` | U-position | assert | An existing active mapper has its populated cache; pushActiveMapper creates the new active cache before the fallback last-index read. |
| checker.ts:21435:36 | `resultObj.errors[resultObj.errors.length - 1]` | U-endpoint | assert | The failed relation check appends a diagnostic before elaboration reads its populated last error entry. |
| checker.ts:21486:25 | `resultObj.errors[resultObj.errors.length - 1]` | U-endpoint | assert | The failed relation check appends a diagnostic before elaboration reads its populated last error entry. |
| checker.ts:21488:29 | `target.symbol.declarations[0]` | U-endpoint | assert | The positive declaration-length guard selects a populated first declaration for diagnostic provenance. |
| checker.ts:21501:25 | `resultObj.errors[resultObj.errors.length - 1]` | U-endpoint | assert | The failed relation check appends a diagnostic before elaboration reads its populated last error entry. |
| checker.ts:21582:46 | `resultObj.errors[resultObj.errors.length - 1]` | U-endpoint | assert | The failed relation check appends a diagnostic before elaboration reads its populated last error entry. |
| checker.ts:21596:96 | `targetProp.declarations[0]` | U-endpoint | assert | The positive declaration-length guard selects a populated first declaration for diagnostic provenance. |
| checker.ts:21596:126 | `target.symbol.declarations[0]` | U-endpoint | assert | The positive declaration-length guard selects a populated first declaration for diagnostic provenance. |
| checker.ts:21689:27 | `node.children[i]` | U-loop | assert | The bounded JSX-child or array-element traversal reads populated parser nodes; omitted syntax is a defined OmittedExpression node. |
| checker.ts:21777:35 | `validChildren[0]` | U-endpoint | assert | The single-valid-child elaboration branch requires the populated first filtered JSX child. |
| checker.ts:21828:26 | `node.elements[i]` | U-loop | assert | The bounded JSX-child or array-element traversal reads populated parser nodes; omitted syntax is a defined OmittedExpression node. |
| checker.ts:21914:50 | `s.parameters[0]` | U-endpoint | assert | The top-signature predicate checks exactly one parameter before selecting its populated rest parameter. |
| checker.ts:21915:55 | `getTypeArguments(paramType)[0]` | U-endpoint | assert | A recognized array type has its populated required first element type argument. |
| checker.ts:22140:115 | `types[0]` | U-endpoint | assert | The at-least-three union guard precedes nullable flags reads at its populated first two positions. |
| checker.ts:22141:25 | `types[1]` | U-endpoint | assert | The at-least-three union guard precedes nullable flags reads at its populated first two positions. |
| checker.ts:22149:51 | `(type as UnionType).types[0]` | U-endpoint | assert | A canonical union has populated nonempty constituents before its first type is selected for undefined classification. |
| checker.ts:22153:58 | `(type as UnionType).types[0]` | U-endpoint | assert | A canonical union has populated nonempty constituents before its first type is selected for undefined classification. |
| checker.ts:22550:69 | `args[0]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22550:78 | `args[1]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22566:148 | `args[0]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22566:157 | `args[1]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22570:159 | `args[0]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22570:168 | `args[1]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22570:177 | `args[2]` | U-position | assert | The selected diagnostic message code determines two or three populated positional substitution arguments reused in the rewritten message. |
| checker.ts:22806:57 | `types[0]` | U-endpoint | assert | Length-two or length-three union guards precede reads of populated nullable and complementary constituent positions. |
| checker.ts:22806:95 | `types[1]` | U-endpoint | decline | Absence is handled nearby by candidate && before normalization; this candidate-selection read is retained. |
| checker.ts:22807:43 | `types[0]` | U-endpoint | assert | Length-two or length-three union guards precede reads of populated nullable and complementary constituent positions. |
| checker.ts:22807:82 | `types[1]` | U-endpoint | assert | Length-two or length-three union guards precede reads of populated nullable and complementary constituent positions. |
| checker.ts:22807:120 | `types[2]` | U-endpoint | decline | Absence is handled nearby by candidate && before normalization; this candidate-selection read is retained. |
| checker.ts:22844:86 | `calls[0]` | U-endpoint | assert | Positive call or construct signature lengths guard their populated first signatures. |
| checker.ts:22845:91 | `constructs[0]` | U-endpoint | assert | Positive call or construct signature lengths guard their populated first signatures. |
| checker.ts:23204:45 | `sourceTypes[i]` | U-loop | assert | Related-type traversal bounds i by the populated source constituent list. |
| checker.ts:23215:19 | `(source as UnionType).types[0]` | U-endpoint | assert | Both recognized canonical unions have populated first constituents for the undefined-order comparison. |
| checker.ts:23215:82 | `(target as UnionType).types[0]` | U-endpoint | assert | Both recognized canonical unions have populated first constituents for the undefined-order comparison. |
| checker.ts:23229:36 | `sourceTypes[i]` | U-loop | assert | Related-type traversal bounds i by the populated source constituent list. |
| checker.ts:23236:61 | `(undefinedStrippedTarget as UnionType).types[i % (undefinedStrippedTarget as UnionType).types.length]` | U-position | assert | The target is a nonempty canonical union; the divisibility check maps a source index modulo its populated target length. |
| checker.ts:23261:62 | `variances[i]` | U-parallel | assert | The loop is bounded by both populated argument lists; a present variance position is separately bounded before its required flag read. |
| checker.ts:23265:31 | `sources[i]` | U-parallel | assert | The loop is bounded by both populated argument lists; a present variance position is separately bounded before its required flag read. |
| checker.ts:23266:31 | `targets[i]` | U-parallel | assert | The loop is bounded by both populated argument lists; a present variance position is separately bounded before its required flag read. |
| checker.ts:23459:41 | `maybeKeys[i]` | U-position | assert | The active maybe-relation stack populates every key from maybeStart through maybeCount before reset traverses those positions. |
| checker.ts:23461:38 | `maybeKeys[i]` | U-position | assert | The active maybe-relation stack populates every key from maybeStart through maybeCount before reset traverses those positions. |
| checker.ts:23595:57 | `sourceTypes[i]` | U-parallel | assert | Equal template text lists imply aligned populated source and target interpolation type lists, traversed at the shared index. |
| checker.ts:23595:73 | `targetTypes[i]` | U-parallel | assert | Equal template text lists imply aligned populated source and target interpolation type lists, traversed at the shared index. |
| checker.ts:23654:109 | `getTypeArguments(source)[0]` | U-endpoint | assert | The single-element generic tuple predicate supplies its populated sole type argument. |
| checker.ts:23655:186 | `getTypeArguments(target)[0]` | U-endpoint | assert | The single-element generic tuple predicate supplies its populated sole type argument. |
| checker.ts:24221:40 | `sourcePropertiesFiltered[i]` | U-parallel | assert | Filtered source properties are populated; cartesianProduct creates one populated combination entry per property, selected by the shared bounded index. |
| checker.ts:24238:48 | `sourcePropertiesFiltered[i]` | U-parallel | assert | Filtered source properties are populated; cartesianProduct creates one populated combination entry per property, selected by the shared bounded index. |
| checker.ts:24243:112 | `combination[i]` | U-parallel | assert | cartesianProductWorker appends one defined constituent per filtered source property; its callback indexes the populated combination at the shared bounded property position. |
| checker.ts:24287:45 | `properties[i]` | U-loop | assert | excludeProperties traverses populated property input under its length bound; both repeated reads remain at their original points. |
| checker.ts:24289:37 | `properties[i]` | U-loop | assert | excludeProperties traverses populated property input under its length bound; both repeated reads remain at their original points. |

| Remaining reviewed site | Code | Reason |
| --- | --- | --- |
| checker.ts:16689:70 | TS2345 | Not an indexed read: arrayFrom exports Map iterator values with possibly-undefined payloads under the current loader typing. |
| checker.ts:16693:62 | TS2322 | Not an indexed read: siblingSymbols inherits the Map iterator payload union from arrayFrom; its existing undefined-list handling is retained. |
| checker.ts:18662:15 | TS2322 | Not an indexed read: arrayFrom(typeMembershipMap.values()) widens iterator payloads to Type | undefined. |
| checker.ts:20618:50 | TS2345 | Not an indexed read: the TypeMapper Function object debugInfo optional-field assignment violates exactOptionalPropertyTypes. |

## Wave C

Frozen ranks 401-600: 134 new required-read assertions, 0 reviewed read declines; 197 frozen findings removed, 3 declined. Net census reduction: 200. Related alias repairs can also remove findings across a wave boundary.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| checker.ts before | 211 | 143 | 53 | 29 | 0 | 436 |
| checker.ts after | 122 | 73 | 19 | 22 | 0 | 236 |

Default oracle: **106,367 passing, 0 failing, 0 pending**, all phase exits 0, **empty baseline diff**, 356.196s. `oracle-wave-c.json` and `baseline-wave-c.diff` retain this observation.

Stock emitted JavaScript is byte-identical; site contract and CLI idempotence pass. CRLF preserved. The isolated mutant replaces `source.target.elementFlags[sourcePosition]!` at checker.ts:24478 with `(source.target.elementFlags[sourcePosition] ?? 0)`; the emitted-JavaScript verifier exits 1. No mutant is run in the real oracle tree. `proof-wave-c.json` records hashes and counts. Every remaining finding is recorded in `remaining-wave-c.json`; every frozen-wave disposition is recorded in `findings-wave-c.json`.

Commands use the reproduction above with /tmp/checker-wave-c-final-after and /tmp/checker-wave-c-final-oracle. Output logs: /tmp/checker-wave-c-final-{adapt,census,verify,idempotence,oracle}.log and /tmp/checker-wave-c-mutant.log.

| Site | Read | Class | Action | Invariant or reason |
| --- | --- | --- | --- | --- |
| checker.ts:24413:66 | `unmatchedProperty.declarations[0]` | U-endpoint | assert | The positive unmatched-property declaration length precedes the populated first diagnostic declaration. |
| checker.ts:24478:67 | `source.target.elementFlags[sourcePosition]` | U-parallel | assert | Tuple relation arity checks bound source and rest-adjusted target positions; argument and flag lists are populated in that shared tuple shape. |
| checker.ts:24485:45 | `target.target.elementFlags[targetPosition]` | U-parallel | assert | Tuple relation arity checks bound source and rest-adjusted target positions; argument and flag lists are populated in that shared tuple shape. |
| checker.ts:24515:62 | `sourceTypeArguments[sourcePosition]` | U-parallel | assert | Tuple relation arity checks bound source and rest-adjusted target positions; argument and flag lists are populated in that shared tuple shape. |
| checker.ts:24516:44 | `targetTypeArguments[targetPosition]` | U-parallel | assert | Tuple relation arity checks bound source and rest-adjusted target positions; argument and flag lists are populated in that shared tuple shape. |
| checker.ts:24629:45 | `sourceSignatures[0]` | U-endpoint | assert | Both positive signature-length guards precede each populated first constructor signature read. |
| checker.ts:24630:45 | `targetSignatures[0]` | U-endpoint | assert | Both positive signature-length guards precede each populated first constructor signature read. |
| checker.ts:24641:59 | `sourceSignatures[0]` | U-endpoint | assert | Both positive signature-length guards precede each populated first constructor signature read. |
| checker.ts:24641:80 | `targetSignatures[0]` | U-endpoint | assert | Both positive signature-length guards precede each populated first constructor signature read. |
| checker.ts:24660:56 | `sourceSignatures[i]` | U-parallel | assert | Signature counts are equal by assertion or explicit guard; the bounded comparison traverses populated signatures at the shared index. |
| checker.ts:24660:77 | `targetSignatures[i]` | U-parallel | assert | Signature counts are equal by assertion or explicit guard; the bounded comparison traverses populated signatures at the shared index. |
| checker.ts:24660:168 | `sourceSignatures[i]` | U-parallel | assert | Signature counts are equal by assertion or explicit guard; the bounded comparison traverses populated signatures at the shared index. |
| checker.ts:24660:189 | `targetSignatures[i]` | U-parallel | assert | Signature counts are equal by assertion or explicit guard; the bounded comparison traverses populated signatures at the shared index. |
| checker.ts:24762:60 | `sourceSignatures[i]` | U-parallel | assert | Signature counts are equal by assertion or explicit guard; the bounded comparison traverses populated signatures at the shared index. |
| checker.ts:24762:81 | `targetSignatures[i]` | U-parallel | assert | Signature counts are equal by assertion or explicit guard; the bounded comparison traverses populated signatures at the shared index. |
| checker.ts:24956:80 | `types[i]` | U-loop | assert | The union discriminator loop bounds i by the populated target constituent list. |
| checker.ts:25105:18 | `variances[i]` | U-parallel | assert | The caller supplies variance flags for the same generic parameters as the populated type-argument list; the bounded loop shares that index. |
| checker.ts:25105:92 | `typeArguments[i]` | U-parallel | assert | The caller supplies variance flags for the same generic parameters as the populated type-argument list; the bounded loop shares that index. |
| checker.ts:25251:27 | `stack[i]` | U-position | assert | The recursion stack is populated through the caller supplied depth; its bounded traversal selects a live type. |
| checker.ts:25414:27 | `source.typeParameters[i]` | U-parallel | assert | Signature identity rejects unequal type-parameter counts before traversing their populated aligned parameter lists. |
| checker.ts:25415:27 | `target.typeParameters[i]` | U-parallel | assert | Signature identity rejects unequal type-parameter counts before traversing their populated aligned parameter lists. |
| checker.ts:25486:20 | `types[0]` | U-endpoint | assert | The exactly-one type or base guard selects the populated sole constituent, including both original conditional branch reads. |
| checker.ts:25571:65 | `bases[0]` | U-endpoint | assert | The exactly-one type or base guard selects the populated sole constituent, including both original conditional branch reads. |
| checker.ts:25571:92 | `bases[0]` | U-endpoint | assert | The exactly-one type or base guard selects the populated sole constituent, including both original conditional branch reads. |
| checker.ts:25748:27 | `typeArguments[i]` | U-parallel | assert | Tuple arity and slice bounds select populated argument and flag positions; structure matching first establishes equal arity. |
| checker.ts:25749:35 | `type.target.elementFlags[i]` | U-parallel | assert | Tuple arity and slice bounds select populated argument and flag positions; structure matching first establishes equal arity. |
| checker.ts:25758:86 | `t2.target.elementFlags[i]` | U-parallel | assert | Tuple arity and slice bounds select populated argument and flag positions; structure matching first establishes equal arity. |
| checker.ts:26481:56 | `getTypeArguments(source)[0]` | U-endpoint | assert | A recognized array type supplies its populated required element type argument. |
| checker.ts:26601:29 | `source.texts[0]` | U-endpoint | assert | Template literal text lists always include a populated head and tail, even with no interpolations. |
| checker.ts:26602:29 | `target.texts[0]` | U-endpoint | assert | Template literal text lists always include a populated head and tail, even with no interpolations. |
| checker.ts:26603:27 | `source.texts[source.texts.length - 1]` | U-endpoint | assert | Template literal text lists always include a populated head and tail, even with no interpolations. |
| checker.ts:26604:27 | `target.texts[target.texts.length - 1]` | U-endpoint | assert | Template literal text lists always include a populated head and tail, even with no interpolations. |
| checker.ts:26672:99 | `(source as TemplateLiteralType).types[0]` | U-parallel | assert | A two-text template shape has exactly one populated interpolation type; only that required type argument is asserted. |
| checker.ts:26681:95 | `target.types[i]` | U-parallel | assert | Equal template text lists or successful template inference align populated target interpolation types at the callback position. |
| checker.ts:26689:104 | `target.types[i]` | U-parallel | assert | Equal template text lists or successful template inference align populated target interpolation types at the callback position. |
| checker.ts:26717:33 | `sourceTexts[0]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:26718:31 | `sourceTexts[lastSourceIndex]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:26721:33 | `targetTexts[0]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:26722:31 | `targetTexts[lastTargetIndex]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:26732:27 | `targetTexts[i]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:26759:46 | `sourceTexts[index]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:26765:22 | `sourceTexts[seg]` | U-position | assert | Template text lists are populated; delimiter bounds and match segment transitions select positions within their head-to-tail range. |
| checker.ts:27091:46 | `variances[i]` | U-parallel | assert | Generic reference inference traverses aligned populated variance and source/target argument lists for the same generic target. |
| checker.ts:27092:49 | `sourceTypes[i]` | U-parallel | assert | Generic reference inference traverses aligned populated variance and source/target argument lists for the same generic target. |
| checker.ts:27092:65 | `targetTypes[i]` | U-parallel | assert | Generic reference inference traverses aligned populated variance and source/target argument lists for the same generic target. |
| checker.ts:27095:36 | `sourceTypes[i]` | U-parallel | assert | Generic reference inference traverses aligned populated variance and source/target argument lists for the same generic target. |
| checker.ts:27095:52 | `targetTypes[i]` | U-parallel | assert | Generic reference inference traverses aligned populated variance and source/target argument lists for the same generic target. |
| checker.ts:27158:44 | `sources[i]` | U-loop | assert | The source-constituent loop bounds i by its populated union-or-singleton source list. |
| checker.ts:27290:46 | `matches[i]` | U-parallel | assert | Successful template matching produces one populated match per target interpolation; the loop is bounded by the populated target type list. |
| checker.ts:27291:36 | `types[i]` | U-parallel | assert | Successful template matching produces one populated match per target interpolation; the loop is bounded by the populated target type list. |
| checker.ts:27388:48 | `getTypeArguments(source)[i]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27388:77 | `elementTypes[i]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27396:44 | `getTypeArguments(source)[i]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27396:73 | `elementTypes[i]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27398:100 | `source.target.elementFlags[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27400:46 | `getTypeArguments(source)[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27402:48 | `elementFlags[i]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27402:128 | `elementTypes[i]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27408:37 | `elementFlags[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27408:65 | `elementFlags[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27410:80 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27413:144 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27414:130 | `elementTypes[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27417:42 | `elementFlags[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27417:95 | `elementFlags[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27420:75 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27424:137 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27425:138 | `elementTypes[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27428:42 | `elementFlags[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27428:91 | `elementFlags[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27431:75 | `elementTypes[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27439:138 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27440:71 | `elementTypes[startLength + 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27444:60 | `elementFlags[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27447:56 | `target.target.elementFlags[targetArity - 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27449:64 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27451:60 | `elementFlags[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27455:62 | `elementTypes[startLength]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27461:44 | `getTypeArguments(source)[sourceArity - i - 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27461:91 | `elementTypes[targetArity - i - 1]` | U-parallel | assert | Tuple inference splits populated argument/flag lists into bounded fixed ends and one- or two-element middles; each selected position belongs to that established segment. |
| checker.ts:27500:57 | `sourceSignatures[sourceIndex]` | U-parallel | assert | Positive source count and bounded target traversal select populated signatures; Math.max maps excess targets to the required first source signature. |
| checker.ts:27500:108 | `targetSignatures[i]` | U-parallel | assert | Positive source count and bounded target traversal select populated signatures; Math.max maps excess targets to the required first source signature. |
| checker.ts:27609:27 | `context.inferences[index]` | U-position | assert | Inference callers select an index of the populated context inference list, preserving the required inference record at its original read. |
| checker.ts:28545:32 | `originFiltered[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:28820:59 | `signatures[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:28820:90 | `signatures[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:28906:24 | `antecedents[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:28961:24 | `(flow as FlowLabel).antecedent[0]` | U-endpoint | assert | A loop flow junction has its populated first non-looping antecedent leading to the loop top, as documented by flow-loop processing. |
| checker.ts:29056:36 | `sharedFlowTypes[i]` | U-parallel | assert | The live shared-flow or loop range populates aligned node, key, and type cache entries; only the required type entry is selected. |
| checker.ts:29084:32 | `(flow as FlowLabel).antecedent[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:29225:200 | `flow.node.arguments[predicate.parameterIndex]` | U-position | assert | The assertion predicate parameter index is explicitly nonnegative and below the populated call argument list length. |
| checker.ts:29395:77 | `flowLoopTypes[i]` | U-parallel | assert | The live shared-flow or loop range populates aligned node, key, and type cache entries; only the required type entry is selected. |
| checker.ts:29396:71 | `flowLoopTypes[i]` | U-parallel | assert | The live shared-flow or loop range populates aligned node, key, and type cache entries; only the required type entry is selected. |
| checker.ts:29834:31 | `clauseTypes[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:29925:32 | `switchStatement.caseBlock.clauses[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:29936:36 | `switchStatement.caseBlock.clauses[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:30127:38 | `callExpression.arguments[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:30749:155 | `symbol.declarations[0]` | U-endpoint | assert | An existing declaration list on a resolved declared symbol is populated by binding; synthetic undefined/globalThis symbols exit through earlier disjuncts. |
| checker.ts:30999:97 | `contextualSignature.parameters[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:31953:55 | `args[indexOfParameter]` | U-position | assert | The parameter membership, rest-parameter predicate, or fixed tuple bound identifies a populated required argument/parameter/type position. |
| checker.ts:32182:50 | `signature.parameters[restIndex]` | U-position | assert | The parameter membership, rest-parameter predicate, or fixed tuple bound identifies a populated required argument/parameter/type position. |
| checker.ts:32517:33 | `elements[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:32531:46 | `getTypeArguments(t)[index]` | U-position | assert | The parameter membership, rest-parameter predicate, or fixed tuple bound identifies a populated required argument/parameter/type position. |
| checker.ts:32531:77 | `t.target.elementFlags[index]` | U-position | assert | The parameter membership, rest-parameter predicate, or fixed tuple bound identifies a populated required argument/parameter/type position. |
| checker.ts:32925:9 | `activeTypeMappersCaches[activeTypeMappersCount]` | U-position | assert | pushActiveMapper creates each live mapper cache before incrementing the count; pop and clear traverse only those populated live slots. |
| checker.ts:32939:13 | `activeTypeMappersCaches[i]` | U-position | assert | pushActiveMapper creates each live mapper cache before incrementing the count; pop and clear traverse only those populated live slots. |
| checker.ts:33201:27 | `target.parameters[targetParameterCount]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:33206:64 | `target.parameters[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:33246:54 | `signatureList[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:33258:49 | `signatureList[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:33258:89 | `signatureList[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:33342:23 | `elements[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:33401:62 | `elementFlags[i]` | U-parallel | assert | Array or tuple construction aligns populated flags with element types; generic argument filling and count checks align populated parameters and arguments. |
| checker.ts:33499:26 | `properties[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:33505:48 | `properties[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:33506:46 | `properties[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:33507:53 | `properties[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:34136:24 | `propertiesOfJsxElementAttribPropInterface[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:35142:57 | `baseTypes[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:35906:25 | `args[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:35953:25 | `args[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:35991:61 | `typeParameters[i]` | U-parallel | assert | Array or tuple construction aligns populated flags with element types; generic argument filling and count checks align populated parameters and arguments. |
| checker.ts:35998:38 | `typeArgumentTypes[i]` | U-parallel | assert | Array or tuple construction aligns populated flags with element types; generic argument filling and count checks align populated parameters and arguments. |
| checker.ts:36185:25 | `args[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:36206:60 | `args[argCount]` | U-position | assert | The positive rest-argument count or successful spread-index search selects populated parser argument positions. |
| checker.ts:36207:81 | `args[argCount]` | U-position | assert | The positive rest-argument count or successful spread-index search selects populated parser argument positions. |
| checker.ts:36207:101 | `args[args.length - 1]` | U-position | assert | The positive rest-argument count or successful spread-index search selects populated parser argument positions. |
| checker.ts:36224:36 | `errorOutputContainer.errors[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:36291:29 | `args[i]` | U-loop | assert | The bounded traversal reads populated parser nodes or compiler-built type/property entries; omitted syntax remains a defined node. |
| checker.ts:36296:39 | `spreadType.target.elementFlags[i]` | U-parallel | assert | Array or tuple construction aligns populated flags with element types; generic argument filling and count checks align populated parameters and arguments. |
| checker.ts:36413:44 | `args[spreadIndex]` | U-position | assert | The positive rest-argument count or successful spread-index search selects populated parser argument positions. |
| checker.ts:36501:25 | `signatures[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:36611:73 | `candidates[0]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |
| checker.ts:36681:34 | `candidatesForArgumentError[candidatesForArgumentError.length - 1]` | U-endpoint | assert | Positive, singleton, or bounded nonempty-list guards establish the populated first or last parser node, signature, type, property, or diagnostic. |

| Remaining reviewed site | Code | Reason |
| --- | --- | --- |
| checker.ts:25972:13 | TS2322 | Not an indexed read: arrayFrom(names.values()) widens Map iterator payloads to Symbol | undefined. |
| checker.ts:25974:9 | TS2322 | Not an indexed read: resolvedProperties is an optional cached property; the preceding Map iterator assignment does not establish a required array under current loader typing. |
| checker.ts:35291:79 | TS2345 | Not an indexed read: candidates comes from arrayFrom(symbols.values()), whose iterator payload union includes undefined. |

## Wave D

Frozen ranks 601-796: 80 new required-read assertions, 5 reviewed read declines; 169 frozen findings removed, 27 declined. Net census reduction: 166. Related alias repairs can also remove findings across a wave boundary.

| File | TS2345 | TS18048 | TS2532 | TS2322 | TS2538 | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| checker.ts before | 122 | 73 | 19 | 22 | 0 | 236 |
| checker.ts after | 29 | 28 | 1 | 12 | 0 | 70 |

Default oracle: **106,367 passing, 0 failing, 0 pending**, all phase exits 0, **empty baseline diff**, 350.388s. `oracle-wave-d.json` and `baseline-wave-d.diff` retain this observation.

Stock emitted JavaScript is byte-identical; site contract and CLI idempotence pass. CRLF preserved. The isolated mutant replaces `candidates[candidateIndex]!` at checker.ts:36809 with `(candidates[candidateIndex] ?? 0)`; the emitted-JavaScript verifier exits 1. No mutant is run in the real oracle tree. `proof-wave-d.json` records hashes and counts. Every remaining finding is recorded in `remaining-wave-d.json`; every frozen-wave disposition is recorded in `findings-wave-d.json`.

Final refinement: the Wave B assertion on `(type as IntersectionType).types[index]` at checker.ts:18272:29 was removed. Equality already handles absence at that probe; the complementary containsType input remains required. The final ledger has **447 assertions, 18 declined reads, zero defaults**. Historical Wave B/C reports describe their pushed revisions. A tree adapted with the older assertion needs a fresh pinned checkout for this revised decline contract; unexpected asserted declines deliberately fail rather than being silently erased. Exact fresh adapter replay is byte-identical to the final source, and rejects the same defaulting mutant. The corrected Wave D logs are /tmp/checker-wave-d-corrected-{census,verify,idempotence,oracle}.log; the original Wave D oracle was stopped before completion and is not claimed as proof.

Commands use the reproduction above with /tmp/checker-wave-d-final-after and /tmp/checker-wave-d-corrected-oracle. Output logs: /tmp/checker-wave-d-final-{adapt,census,verify,idempotence,oracle}.log and /tmp/checker-wave-d-mutant.log.

| Site | Read | Class | Action | Invariant or reason |
| --- | --- | --- | --- | --- |
| checker.ts:18272:29 | `(type as IntersectionType).types[index]` | U-position | decline | Absence is handled at this equality probe; only the complementary constituent used as a required containsType input is asserted. This refines the earlier Wave B ledger entry. |
| checker.ts:36727:45 | `allDiagnostics[minIndex]` | U-position | assert | Each failed overload appends a diagnostic list; minIndex records the position of the least-error populated entry before selection. |
| checker.ts:36740:55 | `diags[0]` | U-endpoint | assert | The preceding assertion establishes a nonempty diagnostic list produced by failed overload checking; every selected first diagnostic is populated. |
| checker.ts:36740:86 | `diags[0]` | U-endpoint | assert | The preceding assertion establishes a nonempty diagnostic list produced by failed overload checking; every selected first diagnostic is populated. |
| checker.ts:36740:116 | `diags[0]` | U-endpoint | assert | The preceding assertion establishes a nonempty diagnostic list produced by failed overload checking; every selected first diagnostic is populated. |
| checker.ts:36741:57 | `diags[0]` | U-endpoint | assert | The preceding assertion establishes a nonempty diagnostic list produced by failed overload checking; every selected first diagnostic is populated. |
| checker.ts:36747:57 | `candidatesForArgumentError[0]` | U-endpoint | assert | Overload failure retains a nonempty candidate list; singleton/non-generic or positive signature/base guards select its populated first entry. |
| checker.ts:36797:35 | `candidates[0]` | U-endpoint | assert | Overload failure retains a nonempty candidate list; singleton/non-generic or positive signature/base guards select its populated first entry. |
| checker.ts:36809:35 | `candidates[candidateIndex]` | U-loop | assert | Overload candidate traversal is bounded by the populated signature candidate list length. |
| checker.ts:36921:13 | `candidates[0]` | U-endpoint | assert | Overload failure retains a nonempty candidate list; singleton/non-generic or positive signature/base guards select its populated first entry. |
| checker.ts:36954:27 | `candidates[bestIndex]` | U-position | assert | getLongestCandidateIndex starts at zero and updates only within the nonempty populated overload candidate list. |
| checker.ts:36974:60 | `typeParameters[typeArguments.length]` | U-position | assert | The argument-filling loop bounds the current argument count below the populated parameter count, retaining both original parameter reads. |
| checker.ts:36974:130 | `typeParameters[typeArguments.length]` | U-position | assert | The argument-filling loop bounds the current argument count below the populated parameter count, retaining both original parameter reads. |
| checker.ts:36990:31 | `candidates[i]` | U-loop | assert | Overload candidate traversal is bounded by the populated signature candidate list length. |
| checker.ts:37164:48 | `constructSignatures[0]` | U-endpoint | assert | Overload failure retains a nonempty candidate list; singleton/non-generic or positive signature/base guards select its populated first entry. |
| checker.ts:37218:27 | `baseTypes[0]` | U-endpoint | assert | Overload failure retains a nonempty candidate list; singleton/non-generic or positive signature/base guards select its populated first entry. |
| checker.ts:37942:27 | `node.arguments[0]` | U-position | assert | Import checking returns on zero arguments; positive option-length guards and the extra-argument loop bound select populated parser arguments. |
| checker.ts:37944:79 | `node.arguments[1]` | U-position | assert | Import checking returns on zero arguments; positive option-length guards and the extra-argument loop bound select populated parser arguments. |
| checker.ts:37947:35 | `node.arguments[i]` | U-position | assert | Import checking returns on zero arguments; positive option-length guards and the extra-argument loop bound select populated parser arguments. |
| checker.ts:37959:91 | `node.arguments[1]` | U-position | assert | Import checking returns on zero arguments; positive option-length guards and the extra-argument loop bound select populated parser arguments. |
| checker.ts:38391:41 | `elements[index]` | U-position | assert | The binding-pattern index is a nonnegative tuple position below its populated fixed binding-element count. |
| checker.ts:38421:20 | `signature.parameters[pos]` | U-position | assert | Parameter positions are nonnegative caller positions and explicitly below the populated non-rest parameter count. |
| checker.ts:38429:34 | `tupleType.elementFlags[index]` | U-position | decline | Declined: tuple label lookup has no local bound on pos - paramCount, so populated flag existence is not established; only the numeric bitwise U-zero rule permits defaulting and this is a function argument. |
| checker.ts:38441:27 | `signature.parameters[pos]` | U-position | assert | Parameter positions are nonnegative caller positions and explicitly below the populated non-rest parameter count. |
| checker.ts:38487:26 | `signature.parameters[pos]` | U-position | assert | Parameter positions are nonnegative caller positions and explicitly below the populated non-rest parameter count. |
| checker.ts:38507:39 | `signature.parameters[pos]` | U-position | assert | Parameter positions are nonnegative caller positions and explicitly below the populated non-rest parameter count. |
| checker.ts:38513:46 | `signature.parameters[paramCount]` | U-endpoint | assert | signatureHasRestParameter identifies a nonempty signature whose populated last parameter is the rest symbol. |
| checker.ts:38562:46 | `signature.parameters[length - 1]` | U-endpoint | assert | signatureHasRestParameter identifies a nonempty signature whose populated last parameter is the rest symbol. |
| checker.ts:38576:50 | `signature.parameters[signature.parameters.length - 1]` | U-endpoint | assert | signatureHasRestParameter identifies a nonempty signature whose populated last parameter is the rest symbol. |
| checker.ts:38608:46 | `signature.parameters[signature.parameters.length - 1]` | U-endpoint | assert | signatureHasRestParameter identifies a nonempty signature whose populated last parameter is the rest symbol. |
| checker.ts:38616:46 | `signature.parameters[signature.parameters.length - 1]` | U-endpoint | assert | signatureHasRestParameter identifies a nonempty signature whose populated last parameter is the rest symbol. |
| checker.ts:38643:33 | `signature.parameters[i]` | U-loop | assert | The signature traversal bounds i by its populated non-rest parameter count. |
| checker.ts:38679:31 | `signature.parameters[i]` | U-loop | assert | The signature traversal bounds i by its populated non-rest parameter count. |
| checker.ts:40244:26 | `properties[propertyIndex]` | U-position | assert | The parser-node caller loop or local binding-pattern loop bounds its nonnegative position within the populated property/element list. |
| checker.ts:40298:17 | `node.elements[i]` | U-position | assert | The parser-node caller loop or local binding-pattern loop bounds its nonnegative position within the populated property/element list. |
| checker.ts:40308:25 | `elements[elementIndex]` | U-position | assert | The parser-node caller loop or local binding-pattern loop bounds its nonnegative position within the populated property/element list. |
| checker.ts:41428:23 | `patternElements[i]` | U-position | assert | The parser-node caller loop or local binding-pattern loop bounds its nonnegative position within the populated property/element list. |
| checker.ts:41606:40 | `a[i]` | U-parallel | assert | Inference contexts describe the same generic signature parameters, creating populated aligned inference arrays; the bounded merge uses that shared position. |
| checker.ts:41606:72 | `b[i]` | U-parallel | assert | Inference contexts describe the same generic signature parameters, creating populated aligned inference arrays; the bounded merge uses that shared position. |
| checker.ts:41615:41 | `target[i]` | U-parallel | assert | Inference contexts describe the same generic signature parameters, creating populated aligned inference arrays; the bounded merge uses that shared position. |
| checker.ts:41615:78 | `source[i]` | U-parallel | assert | Inference contexts describe the same generic signature parameters, creating populated aligned inference arrays; the bounded merge uses that shared position. |
| checker.ts:41616:29 | `source[i]` | U-parallel | assert | Inference contexts describe the same generic signature parameters, creating populated aligned inference arrays; the bounded merge uses that shared position. |
| checker.ts:42061:83 | `signature.parameters[typePredicate.parameterIndex]` | U-position | assert | A nonnegative predicate parameterIndex comes from resolving its name in this populated signature parameter list. |
| checker.ts:42398:64 | `declaration.parameters[0]` | U-endpoint | assert | The exactly-one index-signature parameter guard precedes each populated first parameter read. |
| checker.ts:42399:57 | `declaration.parameters[0]` | U-endpoint | assert | The exactly-one index-signature parameter guard precedes each populated first parameter read. |
| checker.ts:42670:40 | `node.typeArguments[index]` | U-position | assert | The enclosing infer-argument lookup selects a valid generic parameter index; the explicit syntax bound or missing-argument filling supplies its populated type. |
| checker.ts:42672:16 | `getEffectiveTypeArguments(node, typeParameters)[index]` | U-position | assert | The enclosing infer-argument lookup selects a valid generic parameter index; the explicit syntax bound or missing-argument filling supplies its populated type. |
| checker.ts:42684:61 | `typeParameters[i]` | U-parallel | assert | Constraint checking traverses populated type parameters and fills missing argument types before reading their aligned index. |
| checker.ts:42691:21 | `typeArguments[i]` | U-parallel | assert | Constraint checking traverses populated type parameters and fills missing argument types before reading their aligned index. |
| checker.ts:42848:47 | `typeParameters[typeArgumentPosition]` | U-position | decline | Declined: parser type-argument membership does not establish that the generic parameter list covers this position for malformed excess type arguments. |
| checker.ts:42981:39 | `node.members[0]` | U-endpoint | assert | The positive mapped-type member length guard precedes its populated first parser member. |
| checker.ts:43094:126 | `overloads[0]` | U-endpoint | assert | Function/constructor overload processing supplies its nonempty populated declaration list to canonical-overload selection. |
| checker.ts:43095:86 | `overloads[0]` | U-endpoint | assert | Function/constructor overload processing supplies its nonempty populated declaration list to canonical-overload selection. |
| checker.ts:43624:48 | `type.aliasTypeArguments[0]` | U-endpoint | assert | isAwaitedTypeInstantiation checks aliasTypeArguments length exactly one before unwrapAwaitedType selects the populated required argument. |
| checker.ts:44254:37 | `tags[i]` | U-loop | assert | The JSDoc tag loop starts at one and bounds i by the populated filtered tag list length. |
| checker.ts:46398:35 | `globalType.typeParameters[0]` | U-position | assert | The recognized global Generator/Iterator method belongs to the validated three-parameter global generic type; mapping selects its populated yield/return/next parameter. |
| checker.ts:46399:35 | `globalType.typeParameters[1]` | U-position | assert | The recognized global Generator/Iterator method belongs to the validated three-parameter global generic type; mapping selects its populated yield/return/next parameter. |
| checker.ts:46400:59 | `globalType.typeParameters[2]` | U-position | assert | The recognized global Generator/Iterator method belongs to the validated three-parameter global generic type; mapping selects its populated yield/return/next parameter. |
| checker.ts:46914:30 | `typeParameterDeclarations[i]` | U-loop | assert | The parser type-parameter traversal and captured earlier-index loop select populated declaration nodes within their established bounds. |
| checker.ts:46931:25 | `typeParameterDeclarations[j]` | U-loop | assert | The parser type-parameter traversal and captured earlier-index loop select populated declaration nodes within their established bounds. |
| checker.ts:46947:68 | `typeParameters[i]` | U-loop | assert | The parser type-parameter traversal and captured earlier-index loop select populated declaration nodes within their established bounds. |
| checker.ts:46995:32 | `sourceParameters[i]` | U-parallel | assert | Type-parameter count checks align the populated source declarations and target generic parameters before the bounded comparison. |
| checker.ts:46996:32 | `targetParameters[i]` | U-parallel | assert | Type-parameter count checks align the populated source declarations and target generic parameters before the bounded comparison. |
| checker.ts:47140:38 | `baseTypes[0]` | U-endpoint | assert | The class-base path uses the populated first base type after its nonempty base-type guard. |
| checker.ts:47446:33 | `signatures[0]` | U-endpoint | assert | The singleton constructor/signature or positive enum-member guard selects a populated first entry. |
| checker.ts:48091:41 | `enumDeclaration.members[0]` | U-endpoint | assert | The singleton constructor/signature or positive enum-member guard selects a populated first entry. |
| checker.ts:48833:155 | `symbol.declarations[0]` | U-endpoint | assert | A present resolved-symbol declaration list is populated by binding; synthetic undefined/globalThis symbols exit through earlier disjuncts. |
| checker.ts:49308:38 | `statements[i]` | U-position | assert | A successful statement membership search and bounded forward/backward scans retain populated start/end statement indices. |
| checker.ts:49318:38 | `statements[i]` | U-position | assert | A successful statement membership search and bounded forward/backward scans retain populated start/end statement indices. |
| checker.ts:49326:29 | `statements[first]` | U-position | assert | A successful statement membership search and bounded forward/backward scans retain populated start/end statement indices. |
| checker.ts:49327:27 | `statements[last]` | U-position | assert | A successful statement membership search and bounded forward/backward scans retain populated start/end statement indices. |
| checker.ts:50153:121 | `copy.declarations[0]` | U-endpoint | decline | Declined: mapDefined filters optional index-info declarations; a nonempty input does not establish a nonempty declaration result in this branch. |
| checker.ts:50849:53 | `signaturesOfSymbol[0]` | U-endpoint | assert | The singleton constructor/signature or positive enum-member guard selects a populated first entry. |
| checker.ts:52306:38 | `list[0]` | U-endpoint | assert | A parser list with hasTrailingComma contains its populated first element, including recovery nodes for malformed syntax. |
| checker.ts:52325:31 | `parameters[i]` | U-loop | assert | Grammar parameter traversal bounds i by the populated parser parameter list length. |
| checker.ts:52403:113 | `node.typeParameters[0]` | U-endpoint | assert | Present parser type-parameter lists are populated, including missing-node recovery entries; the arrow grammar branch selects the required first parameter. |
| checker.ts:52405:36 | `node.typeParameters[0]` | U-endpoint | assert | Present parser type-parameter lists are populated, including missing-node recovery entries; the arrow grammar branch selects the required first parameter. |
| checker.ts:52416:27 | `node.parameters[0]` | U-endpoint | decline | Absence is handled nearby: an empty index-signature parameter list deliberately produces the invalid-arity diagnostic before later alias reads; asserting this initializer would change that path. |
| checker.ts:52516:57 | `heritageClause.types[1]` | U-position | assert | Explicit positive or length-greater-than-one guards, or membership of the reported property, select populated parser list positions. |
| checker.ts:52838:53 | `variableList.declarations[1]` | U-position | assert | Explicit positive or length-greater-than-one guards, or membership of the reported property, select populated parser list positions. |
| checker.ts:52840:42 | `declarations[0]` | U-position | assert | Explicit positive or length-greater-than-one guards, or membership of the reported property, select populated parser list positions. |
| checker.ts:53398:39 | `node.parent.members[0]` | U-position | assert | Explicit positive or length-greater-than-one guards, or membership of the reported property, select populated parser list positions. |
| checker.ts:53652:50 | `nodeArguments[1]` | U-position | assert | Explicit positive or length-greater-than-one guards, or membership of the reported property, select populated parser list positions. |
| checker.ts:54122:23 | `t1.elementFlags[i]` | U-parallel | assert | Tuple comparison checks equal flag-list lengths before traversing their populated aligned flag positions. |
| checker.ts:54122:44 | `t2.elementFlags[i]` | U-parallel | assert | Tuple comparison checks equal flag-list lengths before traversing their populated aligned flag positions. |

| Remaining reviewed site | Code | Reason |
| --- | --- | --- |
| checker.ts:38430:64 | TS2345 | Declined indexed-read alias: tupleType.elementFlags[index] has no local upper bound for a rest-label position; populated flag existence is unproven, and a numeric default is not permitted at this argument. |
| checker.ts:42849:65 | TS2345 | Declined indexed-read alias: typeParameters[typeArgumentPosition] is not bounded by generic arity for malformed excess type arguments; populated parameter existence is unproven. |
| checker.ts:45935:19 | TS2322 | Not an indexed read: the ErrorOutputContainer object explicitly assigns undefined to its optional errors property under exactOptionalPropertyTypes. |
| checker.ts:45960:19 | TS2322 | Not an indexed read: the ErrorOutputContainer object explicitly assigns undefined to its optional errors property under exactOptionalPropertyTypes. |
| checker.ts:46114:128 | TS2345 | Not an indexed expression: array destructuring of generic type arguments yields possibly-undefined bindings; asserting a callback/argument alias would not follow the required-read adaptation method. |
| checker.ts:46114:208 | TS2345 | Not an indexed expression: array destructuring of generic type arguments yields possibly-undefined bindings; asserting a callback/argument alias would not follow the required-read adaptation method. |
| checker.ts:46128:128 | TS2345 | Not an indexed expression: array destructuring of generic type arguments yields possibly-undefined bindings; asserting a callback/argument alias would not follow the required-read adaptation method. |
| checker.ts:46734:64 | TS2345 | Not an indexed read: forEachKey callback caughtName is typed __String | undefined under the current table/callback typing. |
| checker.ts:46736:167 | TS2345 | Not an indexed read: forEachKey callback caughtName is typed __String | undefined under the current table/callback typing. |
| checker.ts:50153:121 | TS2532 | Declined indexed read: mapDefined filters optional index-info declarations; nonempty infos does not establish a nonempty copy.declarations in this branch. |
| checker.ts:51465:110 | TS2345 | Not an indexed read: arrayFrom(getMembersOfSymbol(sym).values()) widens Map iterator payloads to Symbol | undefined. |
| checker.ts:51530:21 | TS18048 | Not an indexed read: the for-of variable s comes from arrayFrom(exports.values()), with a possibly-undefined iterator payload. |
| checker.ts:51532:25 | TS18048 | Not an indexed read: getMergedSymbol(s) inherits the possibly-undefined Map iterator input. |
| checker.ts:51533:41 | TS18048 | Not an indexed read: merged inherits the Map iterator payload contract, rather than a required indexed read. |
| checker.ts:52426:13 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52427:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52429:35 | TS2345 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52430:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52432:13 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52433:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52435:13 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52436:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52438:14 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52439:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52441:42 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52443:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
| checker.ts:52446:39 | TS18048 | Absence handled nearby: parameter = node.parameters[0] intentionally permits absence for the invalid-arity diagnostic; later alias reads remain after the exactly-one guard, which the current loader does not correlate with that prior read. |
