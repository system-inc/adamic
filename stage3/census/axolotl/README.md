# Diagnostic classification

Measured main: `a23ba096388ceb3774b94d2db70009316d90b0cf`.

Reproduced **144 whole-program diagnostics** and **145 own-file diagnostics across 21 files**, matching the requested baseline.

Checker acceptance remains 2/79 single-root source loads; own-file acceptance is 58/79. External and unlocated diagnostic unions are both zero.

[All diagnostics as Markdown](diagnostics.md), [JSON table](diagnostics.json), [machine-readable summary](summary.json), [method and exact options](METHOD.md), [commands](commands.sh). Raw census and independent scratch-extractor observations are retained as compressed JSONL.

## By cause

| Cause | Own-file count | Whole-program count | Representative |
| --- | ---: | ---: | --- |
| adaptation | 144 | 143 | D001 at src/compiler/checker.ts:1683:9 |
| prelude or options | 0 | 0 | None observed |
| language | 0 | 0 | None observed |
| unsure | 1 | 1 | D055 at src/compiler/sys.ts:1598:69 |

Each diagnostic has one cause. Zero language findings means no undecided language rule was identified in these checker errors; it is not an assertion that lowering will accept the program.

## By code and cause

| Code | Cause | Count | Representative | Source line |
| --- | --- | ---: | --- | --- |
| TS2307 | unsure | 1 | D055 src/compiler/sys.ts:1598:69 |                     (require("source-map-support") as typeof import("source-map-support")).install(); |
| TS2322 | adaptation | 13 | D001 src/compiler/checker.ts:1683:9 |         signatureToSignatureDeclaration: nodeBuilder.signatureToSignatureDeclaration, |
| TS2339 | adaptation | 3 | D098 src/compiler/transformers/generators.ts:1881:50 |                         if (containsYield(clause.expression) &amp;&amp; pendingClauses.length &gt; 0) { |
| TS2345 | adaptation | 38 | D010 src/compiler/checker.ts:38430:64 |             return getTupleElementLabel(associatedName, index, elementFlags, restParameter); |
| TS2375 | adaptation | 9 | D030 src/compiler/checker.ts:54261:5 |     return { |
| TS2412 | adaptation | 25 | D004 src/compiler/checker.ts:17890:21 |                     property.links.tupleLabelDeclaration = namedMemberDeclarations?.[i]; |
| TS2488 | adaptation | 1 | D009 src/compiler/checker.ts:22498:32 |                 reportError(...stack[0]); |
| TS2532 | adaptation | 15 | D016 src/compiler/checker.ts:50153:121 |                     copy.parent = type.aliasSymbol ? type.aliasSymbol : type.symbol ? type.symbol : getSymbolAtLocation(copy.declarations[0].parent); |
| TS2538 | adaptation | 2 | D065 src/compiler/transformers/classFields.ts:2251:45 |         const superStatement = statementsIn[superStatementIndex]; |
| TS2556 | adaptation | 1 | D008 src/compiler/checker.ts:22498:29 |                 reportError(...stack[0]); |
| TS2722 | adaptation | 1 | D007 src/compiler/checker.ts:20595:32 |                         return targets[i](); |
| TS2740 | adaptation | 1 | D037 src/compiler/core.ts:1637:11 |     const set: Set&lt;TElement&gt; = { |
| TS2769 | adaptation | 2 | D005 src/compiler/checker.ts:17946:29 |                     mapType(elementTypes[unionIndex], t =&gt; createNormalizedTupleType(target, replaceElement(elementTypes, unionIndex, t))) : |
| TS18046 | adaptation | 4 | D032 src/compiler/commandLineParser.ts:2301:91 |         return createCompilerDiagnostic(Diagnostics.Cannot_read_file_0_Colon_1, fileName, e.message); |
| TS18048 | adaptation | 29 | D003 src/compiler/checker.ts:12034:30 |         reference.flowNode = file.endFlowNode; |

## By code

| Code | Own-file count | Whole-program count |
| --- | ---: | ---: |
| TS2307 | 1 | 1 |
| TS2322 | 13 | 12 |
| TS2339 | 3 | 3 |
| TS2345 | 38 | 38 |
| TS2375 | 9 | 9 |
| TS2412 | 25 | 25 |
| TS2488 | 1 | 1 |
| TS2532 | 15 | 15 |
| TS2538 | 2 | 2 |
| TS2556 | 1 | 1 |
| TS2722 | 1 | 1 |
| TS2740 | 1 | 1 |
| TS2769 | 2 | 2 |
| TS18046 | 4 | 4 |
| TS18048 | 29 | 29 |

## Cause examples and routing

### adaptation

D001, TS2322, `src/compiler/checker.ts:1683:9`.

```ts
        signatureToSignatureDeclaration: nodeBuilder.signatureToSignatureDeclaration,
```

nodeBuilder can return an ArrowFunction whose typeArguments property explicitly admits undefined, while the TypeChecker overload intersection excludes present-undefined. Align that signature intersection with the actual returned node contract.

### unsure

D055, TS2307, `src/compiler/sys.ts:1598:69`.

```ts
                    (require("source-map-support") as typeof import("source-map-support")).install();
```

The optional source-map-support import has no declaration in this load graph. Need the supported-host policy and a pinned, audited package API to choose a trusted loader binding versus a source adaptation of this optional hook.

## Per file

| File | Own-file diagnostics |
| --- | ---: |
| src/compiler/checker.ts | 30 |
| src/compiler/transformers/generators.ts | 27 |
| src/compiler/program.ts | 11 |
| src/compiler/tsbuildPublic.ts | 10 |
| src/compiler/watchPublic.ts | 10 |
| src/compiler/commandLineParser.ts | 6 |
| src/compiler/transformer.ts | 6 |
| src/compiler/transformers/classFields.ts | 6 |
| src/compiler/transformers/es2017.ts | 6 |
| src/compiler/transformers/es2015.ts | 5 |
| src/compiler/transformers/esnext.ts | 5 |
| src/compiler/transformers/module/esnextAnd2015.ts | 5 |
| src/compiler/sys.ts | 4 |
| src/compiler/moduleNameResolver.ts | 3 |
| src/compiler/transformers/esDecorators.ts | 3 |
| src/compiler/transformers/jsx.ts | 2 |
| src/compiler/watch.ts | 2 |
| src/compiler/core.ts | 1 |
| src/compiler/resolutionCache.ts | 1 |
| src/compiler/sourcemap.ts | 1 |
| src/compiler/transformers/declarations.ts | 1 |

## Why the own-file sum is one higher

There are 145 distinct formatted strings but only 144 distinct (code, file, line, column) locations. TS2322 at src/compiler/commandLineParser.ts:2682:9 has two object-property-order renderings under different root contexts. The aggregate includes one; the own-file union retains both because the meter deduplicates complete strings. Both describe the same originalPath: undefined incompatibility, not a second failing source expression.

D034, TS2322, `src/compiler/commandLineParser.ts:2682:9`.

```ts
        references: map(configParseResult.projectReferences, r => ({ ...r, path: r.originalPath ? r.originalPath : "", originalPath: undefined })),
```

Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }[] | undefined' is not assignable to type 'readonly ProjectReference[] | undefined'.
  Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.
    Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
      Types of property 'originalPath' are incompatible.
        Type 'undefined' is not assignable to type 'string'.

The object may explicitly store undefined in originalPath, but the destination declaration permits only absence or a defined value. Give the declaration or construction a truthful present-undefined contract, including nested signatures.

Observed singleton roots: ['src/compiler/diagnosticInformationMap.generated.ts'], ['src/compiler/types.ts'].

Same-location aggregate rendering:

Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }[] | undefined' is not assignable to type 'readonly ProjectReference[] | undefined'.
  Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.
    Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
      Types of property 'originalPath' are incompatible.
        Type 'undefined' is not assignable to type 'string'.

All root memberships, including the aggregate marker, are retained in diagnostics.json.

## Validation

The unmodified meter report function was run on the fresh ordinary census. Its counts agree with the independent scratch Go extractor, record by record for all 79 source singletons and the aggregate. The report table uses the meter primary-location ownership rule and deduplicates the complete formatted string. JSON preserves every chain, source line, and observing root. All rows are checked against the fresh adapted source.

These are classifications for work allocation, not implemented fixes. The custom createSet wrapper needs a truthful protocol rather than removal of the Set operations from the prelude. The source-map-support finding is left unsure pending a supported-host decision and audited dependency contract; it is not a recommendation to add an unverified ambient declaration.

The measured tip differs from the historical baseline only outside the loader, adaptations, and checker pin relevant to these diagnostics. No source fixes, language changes, or lowering claims were made.
