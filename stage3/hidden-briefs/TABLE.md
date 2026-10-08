# Hidden boundary briefs, ranks 11–60

Census source: `codex/stage3-hidden-source` at `388096e6`; read RESULT.json and README.md plus its full.jsonl.gz and stock.json.gz evidence. Compiler measured by that census: `ed6e29751ee47d86fad450cd1674139883bc0f70`. Current-main source search: `89ac4a8c1de0b02d95965be72f7f8bf1c92433d2`. These are different compiler baselines. The census pin and its recorded measured compiler have no diff in internal/lower, internal/load or stage3/adapt, so the isolated replay uses the measured lowering and adaptation sources.

RESULT.json contains causes only for the top ten. Reconstruct all residual regions from `files[*].hidden_ranges`, sort by descending bytes, then file and byte start, and take ranks 11–60. Join their heads to Boundary records in the pinned raw census. Prefer the smallest containing boundary; duplicate attempts do not count twice. If there is no Boundary at the head, record the smallest containing checker-skipped body instead. A connected region can have multiple downstream blockers: assignment to one head is a conservative prioritization convention, not proof that one fix reveals its entirety.

Totals: **431,255 bytes / 50 regions**. Checked views and conservative representation exclusions: **154,365 bytes**. Checker-only/dependency blocks: **106,264 bytes**. Eligible groups: **170,626 bytes**; the fifteen briefs cover **158,725 bytes**. Two smaller eligible groups (computed field name, for...in origin) remain in the complete ledger below.

View exclusion assumption: exclude explicit cast/view reasons and structural/nominal view representations (`Declaration & HasModifiers`, `__String`, `AliasDeclarationNode`, `PrivateIdentifierInExpression`, `CapturedThis`, `PropertyDeclaration | ParameterPropertyDeclaration`, `SourceFile`, and the Map head whose key is `__String`). The latter exclusions are conservative inferences from the represented type names, not observed proof that every one needs a checked cast. Generic parameters, optional callbacks and optional arrays are retained because their immediate boundary does not establish a view requirement. Revisit an inferred exclusion if a sound ordinary representation is established.

| Brief | Boundary | Raising function | Hidden bytes | Region ranks | Existing fix |
|---|---|---|---:|---|---|
| [01](01-overload-parameter.md) | `overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem` | `censusOverload (historical; absent on origin/main)` | 20,524 | 16, 43 | Absent on main; no verified fix |
| [02](02-any-return.md) | `a function returning any` | `signature` | 15,966 | 13 | None verified |
| [03](03-structural-statics.md) | `a method call through a structural signature in a program with statics; use typeof the declaring class` | `callOrMethod` | 13,111 | 39, 55 | codex/notyet-statics, b6aa4f000 (source diff) |
| [04](04-generic-optional-array.md) | `a value of type readonly T[] \| undefined` | `typeOf` | 13,061 | 40, 56 | None verified |
| [05](05-optional-node-callback.md) | `a value of type ((node: Node) => boolean) \| undefined` | `typeOf` | 11,826 | 49, 59 | None verified |
| [06](06-generic-tnode.md) | `a value of type TNode` | `typeOf` | 11,417 | 19 | None verified |
| [07](07-any-value.md) | `a value of type any` | `typeOf` | 9,904 | 23 | None verified |
| [08](08-never-array.md) | `an array of never` | `elementType` | 9,790 | 24 | None verified |
| [09](09-var.md) | `var` | `variables` | 8,452 | 26 | None verified |
| [10](10-object-iteration.md) | `for...of over an object` | `forOf` | 8,175 | 29 | None verified |
| [11](11-derived-binding.md) | `reading derived` | `enumNeverValue` | 8,143 | 31 | None verified |
| [12](12-name-binding.md) | `reading name` | `enumNeverValue` | 7,705 | 36 | None verified |
| [13](13-block-overload-result.md) | `overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody` | `censusOverload (historical; absent on origin/main)` | 7,289 | 38 | Absent on main; no verified fix |
| [14](14-evaluator-overload-result.md) | `overload 1 of evaluate result EvaluatorResult<string \| undefined> cannot be served by implementation result EvaluatorResult<string \| number \| undefined>` | `censusOverload (historical; absent on origin/main)` | 7,102 | 42 | Absent on main; no verified fix |
| [15](15-old-source-files.md) | `reading oldSourceFiles` | `enumNeverValue` | 6,260 | 48 | None verified |

## Complete region ledger

Half-open offsets are UTF-8 bytes in the adapted source. The recorded diagnostic can be in a dependency or much later than the residual head. Boundary reasons shown are the actual diagnostic What, not the generic census rollback reason.

| Rank | File | Range | Bytes | Category | Reason and raising function |
|---:|---|---|---:|---|---|
| 11 | `checker.ts` | [326156, 343821) | 17,665 | checked views / representations | `a function returning Declaration & HasModifiers`; `signature` |
| 12 | `checker.ts` | [546205, 563029) | 16,824 | checked views / representations | `a value of type __String`; `typeOf` |
| 13 | `utilities.ts` | [48518, 64484) | 15,966 | eligible | `a function returning any`; `signature` |
| 14 | `tsbuildPublic.ts` | [64324, 78537) | 14,213 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 15 | `checker.ts` | [1120078, 1133897) | 13,819 | checked views / representations | `a value of type __String \| undefined`; `typeOf` |
| 16 | `transformers/declarations.ts` | [68180, 81805) | 13,625 | eligible | `overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem`; `censusOverload (historical; absent on origin/main)` |
| 17 | `checker.ts` | [2825793, 2837232) | 11,439 | checked views / representations | `a value of type AliasDeclarationNode`; `typeOf` |
| 18 | `checker.ts` | [1084088, 1095517) | 11,429 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 19 | `transformers/esDecorators.ts` | [61324, 72741) | 11,417 | eligible | `a value of type TNode`; `typeOf` |
| 20 | `checker.ts` | [907070, 917940) | 10,870 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 21 | `watchPublic.ts` | [19425, 30255) | 10,830 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 22 | `checker.ts` | [631149, 641885) | 10,736 | checked views / representations | `a function returning Declaration & HasModifiers`; `signature` |
| 23 | `checker.ts` | [1155628, 1165532) | 9,904 | eligible | `a value of type any`; `typeOf` |
| 24 | `utilities.ts` | [356194, 365984) | 9,790 | eligible | `an array of never`; `elementType` |
| 25 | `builder.ts` | [11292, 20800) | 9,508 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 26 | `parser.ts` | [81993, 90445) | 8,452 | eligible | `var`; `variables` |
| 27 | `checker.ts` | [2011504, 2019788) | 8,284 | checked views / representations | `a value of type __String \| undefined`; `typeOf` |
| 28 | `transformers/classFields.ts` | [103917, 112133) | 8,216 | checked views / representations | `a value of type PrivateIdentifierInExpression`; `typeOf` |
| 29 | `checker.ts` | [1997085, 2005260) | 8,175 | eligible | `for...of over an object`; `forOf` |
| 30 | `checker.ts` | [425305, 433466) | 8,161 | checked views / representations | `a function returning Declaration & HasModifiers`; `signature` |
| 31 | `checker.ts` | [2778143, 2786286) | 8,143 | eligible | `reading derived`; `enumNeverValue` |
| 32 | `builder.ts` | [54527, 62640) | 8,113 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 33 | `checker.ts` | [2985768, 2993846) | 8,078 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 34 | `transformers/es2015.ts` | [200153, 208201) | 8,048 | checked views / representations | `a function returning CapturedThis`; `signature` |
| 35 | `moduleNameResolver.ts` | [140732, 148485) | 7,753 | checker-only | `latent census: skipped checker-diagnosed dependency at moduleNameResolver.ts:2753:5`; `checker eligibility / census dependency skip` |
| 36 | `binder.ts` | [29754, 37459) | 7,705 | eligible | `reading name`; `enumNeverValue` |
| 37 | `sys.ts` | [64317, 71913) | 7,596 | checker-only | `latent census: skipped checker-diagnosed dependency at sys.ts:1469:5`; `checker eligibility / census dependency skip` |
| 38 | `transformers/es2017.ts` | [30237, 37526) | 7,289 | eligible | `overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody`; `censusOverload (historical; absent on origin/main)` |
| 39 | `transformers/module/module.ts` | [16356, 23548) | 7,192 | eligible | `a method call through a structural signature in a program with statics; use typeof the declaring class`; `callOrMethod` |
| 40 | `checker.ts` | [620304, 627479) | 7,175 | eligible | `a value of type readonly T[] \| undefined`; `typeOf` |
| 41 | `transformers/classFields.ts` | [64862, 72007) | 7,145 | checked views / representations | `checked view field left of type LeftHandSideExpression`; `view (historical; absent on origin/main)` |
| 42 | `utilities.ts` | [455532, 462634) | 7,102 | eligible | `overload 1 of evaluate result EvaluatorResult<string \| undefined> cannot be served by implementation result EvaluatorResult<string \| number \| undefined>`; `censusOverload (historical; absent on origin/main)` |
| 43 | `transformers/declarations.ts` | [82928, 89827) | 6,899 | eligible | `overload 1 of visitBindingElement parameter elem cannot be served by implementation parameter elem`; `censusOverload (historical; absent on origin/main)` |
| 44 | `checker.ts` | [3060634, 3067472) | 6,838 | checked views / representations | `a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions`; `mapTypes` |
| 45 | `checker.ts` | [129007, 135700) | 6,693 | checked views / representations | `a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions` | 6,838 | 44 |
| `a value of type PropertyDeclaration \| ParameterPropertyDeclaration`; `typeOf` |
| 46 | `checker.ts` | [1323705, 1330360) | 6,655 | checked views / representations | `a value of type __String`; `typeOf` |
| 47 | `checker.ts` | [2570813, 2577207) | 6,394 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 48 | `program.ts` | [117206, 123466) | 6,260 | eligible | `reading oldSourceFiles`; `enumNeverValue` |
| 49 | `transformers/es2018.ts` | [35690, 41768) | 6,078 | eligible | `a value of type ((node: Node) => boolean) \| undefined`; `typeOf` |
| 50 | `checker.ts` | [1195137, 1201199) | 6,062 | checked views / representations | `a cast the runtime can't check`; `castProof` |
| 51 | `checker.ts` | [593883, 599899) | 6,016 | checked views / representations | `a function returning Declaration & HasModifiers`; `signature` |
| 52 | `checker.ts` | [1212364, 1218326) | 5,962 | checked views / representations | `a cast the runtime can't check`; `castProof` |
| 53 | `scanner.ts` | [5112, 11073) | 5,961 | eligible | `a computed field name`; `objectLiteral` |
| 54 | `moduleSpecifiers.ts` | [44553, 50493) | 5,940 | eligible | `for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly)`; `forIn` |
| 55 | `moduleSpecifiers.ts` | [67576, 73495) | 5,919 | eligible | `a method call through a structural signature in a program with statics; use typeof the declaring class`; `callOrMethod` |
| 56 | `checker.ts` | [1623136, 1629022) | 5,886 | eligible | `a value of type readonly T[] \| undefined`; `typeOf` |
| 57 | `resolutionCache.ts` | [38947, 44749) | 5,802 | checked views / representations | `a value of type SourceFile`; `typeOf` |
| 58 | `tsbuildPublic.ts` | [17492, 23252) | 5,760 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |
| 59 | `transformers/es2015.ts` | [144178, 149926) | 5,748 | eligible | `a value of type ((node: Node) => boolean) \| undefined`; `typeOf` |
| 60 | `moduleNameResolver.ts` | [21621, 27341) | 5,720 | checker-only | `checker diagnostics in own body; body skipped`; `checker eligibility / census dependency skip` |

## Checked-view exclusions, grouped

Source inspection also identifies the intersections: `PrivateIdentifierInExpression = BinaryExpression & { left: PrivateIdentifier; token: InKeyword }`; `CapturedThis = GeneratedIdentifier & { escapedText: __String & "__this" }`; `ParameterPropertyDeclaration = ParameterDeclaration & { parent: ConstructorDeclaration; name: Identifier }`. `AliasDeclarationNode` includes the intersection `BindingElementOfBareOrAccessedRequire`, with a narrowed nested parent contract. `__String` contains a string brand intersection. These observations support treating those as view representations. `SourceFile` remains a conservative structural-view inference.

| Boundary | Bytes | Ranks |
|---|---:|---|
| `a function returning Declaration & HasModifiers` | 42,578 | 11, 22, 30, 51 |
| `a value of type __String` | 23,479 | 12, 46 |
| `a value of type __String \| undefined` | 22,103 | 15, 27 |
| `a cast the runtime can't check` | 12,024 | 50, 52 |
| `a value of type AliasDeclarationNode` | 11,439 | 17 |
| `a value of type PrivateIdentifierInExpression` | 8,216 | 28 |
| `a function returning CapturedThis` | 8,048 | 34 |
| `checked view field left of type LeftHandSideExpression` | 7,145 | 41 |
| `a value of type PropertyDeclaration \| ParameterPropertyDeclaration` | 6,693 | 45 |
| `a value of type SourceFile` | 5,802 | 57 |

Map exclusion observation: the head at `checker.ts:52598:22` creates `new Map<__String, DeclarationMeaning>()`. Its key is the branded view representation already excluded above, so the generic Map reason is excluded too; a Map with an ordinary string or undefined key is not a faithful reduction of this head.

## Other head blockers

The smallest-head convention does not remove broader enclosing failures. These additional head causes were present in the raw evidence:

- Region 12: `/tmp/hidden-adapted/src/compiler/checker.ts:6600:18: stage 0 can't lower a function returning Declaration & HasModifiers yet`; blocked [326156, 653463).
- Region 19: `/tmp/hidden-adapted/src/compiler/transformers/esDecorators.ts:670:14: stage 0 can't lower a function returning ImmediatelyInvokedArrowFunction yet`; blocked [9813, 127378).
- Region 36: `/tmp/hidden-adapted/src/compiler/binder.ts:630:47: stage 0 can't lower a value of type __String yet`; blocked [17474, 188866).
- Region 38: `/tmp/hidden-adapted/src/compiler/transformers/es2017.ts:736:5: Adamic 0.1 refuses overload 1 of transformAsyncFunctionBody result Block cannot be served by implementation result ConciseBody; make the implementation result covariant with every overload result`; blocked [3036, 45738).
- Region 39: `/tmp/hidden-adapted/src/compiler/transformers/module/module.ts:1926:43: stage 0 can't lower a value of type InitializedVariableDeclaration yet`; blocked [4518, 111519).
- Region 40: `/tmp/hidden-adapted/src/compiler/checker.ts:6600:18: stage 0 can't lower a function returning Declaration & HasModifiers yet`; blocked [326156, 653463).
- Region 41: `/tmp/hidden-adapted/src/compiler/transformers/classFields.ts:671:55: stage 0 can't lower a value of type PrivateIdentifierInExpression yet`; blocked [10532, 154733).
- Region 42: `/tmp/hidden-adapted/src/compiler/utilities.ts:11316:35: Adamic 0.1 refuses a method in object destructuring; call it on its receiver or wrap that call in an arrow; destructuring would lose this`; blocked [455254, 463614).
- Region 49: `/tmp/hidden-adapted/src/compiler/transformers/es2018.ts:1220:5: Adamic 0.1 refuses overload 1 of transformFunctionBody result Block cannot be served by implementation result ConciseBody; make the implementation result covariant with every overload result`; blocked [3955, 69978).
- Region 55: `/tmp/hidden-adapted/src/compiler/moduleSpecifiers.ts:1254:35: stage 0 can't lower a method call through a structural signature in a program with statics; use typeof the declaring class yet`; blocked [67576, 67682).
- Region 56: `/tmp/hidden-adapted/src/compiler/core.ts:145:26: stage 0 can't lower a value of type readonly T[] | undefined yet`; blocked [1623136, 1629022).
- Region 57: `/tmp/hidden-adapted/src/compiler/resolutionCache.ts:631:11: stage 0 can't lower a value of type Path yet`; blocked [24063, 81888).
- Region 59: `/tmp/hidden-adapted/src/compiler/transformers/es2015.ts:825:14: stage 0 can't lower a function returning CapturedThis yet`; blocked [16600, 231787).

Checker-only regions have no honest `internal/lower` raising function. Their exact diagnostics and selected attempted units are:

### Region 14

Unit: `/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:1469:1`.

```text
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:1776:54: error TS2345: Argument of type 'string | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

### Region 18

Unit: `/tmp/hidden-adapted/src/compiler/checker.ts:1486:1`.

```text
/tmp/hidden-adapted/src/compiler/checker.ts:1683:9: error TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & ...) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBuilderF...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'ArrowFunction' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'ArrowFunction' is not assignable to type 'ArrowFunction & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'ArrowFunction' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
/tmp/hidden-adapted/src/compiler/checker.ts:1779:9: error TS2322: Type '(call: CallLikeExpression, editingArgument: Node) => (Signature | undefined)[]' is not assignable to type '(call: CallLikeExpression, editingArgument: Node) => Signature[]'.
  Type '(Signature | undefined)[]' is not assignable to type 'Signature[]'.
    Type 'Signature | undefined' is not assignable to type 'Signature'.
      Type 'undefined' is not assignable to type 'Signature'.
```

### Region 20

Unit: `/tmp/hidden-adapted/src/compiler/checker.ts:1486:1`.

```text
/tmp/hidden-adapted/src/compiler/checker.ts:1683:9: error TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & ...) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBuilderF...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'ArrowFunction' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'ArrowFunction' is not assignable to type 'ArrowFunction & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'ArrowFunction' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
/tmp/hidden-adapted/src/compiler/checker.ts:1779:9: error TS2322: Type '(call: CallLikeExpression, editingArgument: Node) => (Signature | undefined)[]' is not assignable to type '(call: CallLikeExpression, editingArgument: Node) => Signature[]'.
  Type '(Signature | undefined)[]' is not assignable to type 'Signature[]'.
    Type 'Signature | undefined' is not assignable to type 'Signature'.
      Type 'undefined' is not assignable to type 'Signature'.
```

### Region 21

Unit: `/tmp/hidden-adapted/src/compiler/watchPublic.ts:420:1`.

```text
/tmp/hidden-adapted/src/compiler/watchPublic.ts:506:5: error TS2412: Type '(() => boolean) | undefined' is not assignable to type '() => boolean' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '() => boolean'.
/tmp/hidden-adapted/src/compiler/watchPublic.ts:529:5: error TS2412: Type '((moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... | undefined) => ...) | undefined' is not assignable to type '(moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... | undefined) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... | undefined) => ...'.
/tmp/hidden-adapted/src/compiler/watchPublic.ts:530:5: error TS2412: Type '((moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile | undefined) => ...) | undefined' is not assignable to type '(moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile | undefined) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile | undefined) => ...'.
/tmp/hidden-adapted/src/compiler/watchPublic.ts:534:5: error TS2412: Type '((typeDirectiveReferences: readonly (string | FileReference)[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile | undefined, reusedNames: ... | undefined) => ...) | undefined' is not assignable to type '<T extends FileReference | string>(typeDirectiveReferences: readonly T[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile | undefined, reusedNames: ... | undefined) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '<T extends FileReference | string>(typeDirectiveReferences: readonly T[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile | undefined, reusedNames: ... | undefined) => ...'.
/tmp/hidden-adapted/src/compiler/watchPublic.ts:535:5: error TS2412: Type '((typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => ...) | undefined' is not assignable to type '(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => ...'.
/tmp/hidden-adapted/src/compiler/watchPublic.ts:542:5: error TS2412: Type '(() => ModuleResolutionCache | undefined) | undefined' is not assignable to type '() => ModuleResolutionCache | undefined' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '() => ModuleResolutionCache | undefined'.
```

### Region 25

Unit: `/tmp/hidden-adapted/src/compiler/builder.ts:321:1`.

```text
/tmp/hidden-adapted/src/compiler/builder.ts:395:85: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'Path'.
  Type 'undefined' is not assignable to type 'Path'.
    Type 'undefined' is not assignable to type 'string'.
/tmp/hidden-adapted/src/compiler/builder.ts:395:118: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'Path'.
  Type 'undefined' is not assignable to type 'Path'.
    Type 'undefined' is not assignable to type 'string'.
```

### Region 32

Unit: `/tmp/hidden-adapted/src/compiler/builder.ts:1235:1`.

```text
/tmp/hidden-adapted/src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
/tmp/hidden-adapted/src/compiler/builder.ts:1258:65: error TS2488: Type '[Path, FileInfo] | undefined' must have a '[Symbol.iterator]()' method that returns an iterator.
/tmp/hidden-adapted/src/compiler/builder.ts:1292:61: error TS2488: Type '[Path, FileInfo] | undefined' must have a '[Symbol.iterator]()' method that returns an iterator.
/tmp/hidden-adapted/src/compiler/builder.ts:1345:64: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'Path'.
  Type 'undefined' is not assignable to type 'Path'.
    Type 'undefined' is not assignable to type 'string'.
/tmp/hidden-adapted/src/compiler/builder.ts:1347:41: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'Path'.
  Type 'undefined' is not assignable to type 'Path'.
    Type 'undefined' is not assignable to type 'string'.
/tmp/hidden-adapted/src/compiler/builder.ts:1347:97: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'Path'.
  Type 'undefined' is not assignable to type 'Path'.
    Type 'undefined' is not assignable to type 'string'.
```

### Region 33

Unit: `/tmp/hidden-adapted/src/compiler/checker.ts:1486:1`.

```text
/tmp/hidden-adapted/src/compiler/checker.ts:1683:9: error TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & ...) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBuilderF...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'ArrowFunction' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'ArrowFunction' is not assignable to type 'ArrowFunction & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'ArrowFunction' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
/tmp/hidden-adapted/src/compiler/checker.ts:1779:9: error TS2322: Type '(call: CallLikeExpression, editingArgument: Node) => (Signature | undefined)[]' is not assignable to type '(call: CallLikeExpression, editingArgument: Node) => Signature[]'.
  Type '(Signature | undefined)[]' is not assignable to type 'Signature[]'.
    Type 'Signature | undefined' is not assignable to type 'Signature'.
      Type 'undefined' is not assignable to type 'Signature'.
```

### Region 47

Unit: `/tmp/hidden-adapted/src/compiler/checker.ts:1486:1`.

```text
/tmp/hidden-adapted/src/compiler/checker.ts:1683:9: error TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & ...) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBuilderF...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'ArrowFunction' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'ArrowFunction' is not assignable to type 'ArrowFunction & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'ArrowFunction' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
/tmp/hidden-adapted/src/compiler/checker.ts:1779:9: error TS2322: Type '(call: CallLikeExpression, editingArgument: Node) => (Signature | undefined)[]' is not assignable to type '(call: CallLikeExpression, editingArgument: Node) => Signature[]'.
  Type '(Signature | undefined)[]' is not assignable to type 'Signature[]'.
    Type 'Signature | undefined' is not assignable to type 'Signature'.
      Type 'undefined' is not assignable to type 'Signature'.
```

### Region 58

Unit: `/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:429:1`.

```text
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:438:5: error TS2412: Type '((moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... | undefined) => ...) | undefined' is not assignable to type '(moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... | undefined) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: ... | undefined) => ...'.
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:439:5: error TS2412: Type '((typeDirectiveReferences: readonly (string | FileReference)[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile | undefined, reusedNames: ... | undefined) => ...) | undefined' is not assignable to type '<T extends FileReference | string>(typeDirectiveReferences: readonly T[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile | undefined, reusedNames: ... | undefined) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '<T extends FileReference | string>(typeDirectiveReferences: readonly T[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile: SourceFile | undefined, reusedNames: ... | undefined) => ...'.
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:440:5: error TS2412: Type '((libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string) => ResolvedModuleWithFailedLookupLocations) | undefined' is not assignable to type '(libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string) => ResolvedModuleWithFailedLookupLocations' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string) => ResolvedModuleWithFailedLookupLocations'.
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:441:5: error TS2412: Type '((moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile | undefined) => ...) | undefined' is not assignable to type '(moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile | undefined) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(moduleNames: string[], containingFile: string, reusedNames: string[] | undefined, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingSourceFile?: SourceFile | undefined) => ...'.
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:442:5: error TS2412: Type '((typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => ...) | undefined' is not assignable to type '(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => ...' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '(typeReferenceDirectiveNames: string[] | readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference | undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => ...'.
/tmp/hidden-adapted/src/compiler/tsbuildPublic.ts:443:5: error TS2412: Type '(() => ModuleResolutionCache | undefined) | undefined' is not assignable to type '() => ModuleResolutionCache | undefined' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the type of the target.
  Type 'undefined' is not assignable to type '() => ModuleResolutionCache | undefined'.
```

### Region 60

Unit: `/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:538:1`.

```text
/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:627:9: error TS2375: Type '{ primary: boolean; resolvedFileName: string; originalPath: string | undefined; packageId: PackageId | undefined; isExternalLibraryImport: boolean; }' is not assignable to type 'ResolvedTypeReferenceDirective' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'originalPath' are incompatible.
    Type 'string | undefined' is not assignable to type 'string'.
      Type 'undefined' is not assignable to type 'string'.
/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:635:5: error TS2375: Type '{ resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective | undefined; failedLookupLocations: string[] | undefined; affectingLocations: ... | undefined; resolutionDiagnostics: ... | undefined; }' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'resolutionDiagnostics' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:642:143: error TS2345: Argument of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations | undefined' is not assignable to parameter of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
  Type 'undefined' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:644:144: error TS2345: Argument of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations | undefined' is not assignable to parameter of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
  Type 'undefined' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:647:35: error TS2345: Argument of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations | undefined' is not assignable to parameter of type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
  Type 'undefined' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
/tmp/hidden-adapted/src/compiler/moduleNameResolver.ts:648:5: error TS2322: Type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations | undefined' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
  Type 'undefined' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations'.
```


## Branch search

Searched the fetched 72 origin `codex/notyet-*` and `compiler/*` tips, checking diagnostic templates in their raising files. Exact instantiated type strings are assembled at runtime, so grep the constant prefix. A retained generic prefix cannot prove a specific instantiation is still blocked; absence cannot prove a fix when the branch predates the guard. The one positively identified deletion is `codex/notyet-statics` at `b6aa4f000eb7279da65353b2e39ee359c7171082`; `git log -S` and its diff show the guard replaced by actual receiver lookup. No candidate was merged or its backend tests claimed.

The overload refusal template is absent on main and these 29 tips (not verified fixes):

`origin/codex/notyet-case-declaration-topic`, `origin/codex/notyet-class-construction-topic`, `origin/codex/notyet-class-set-property-topic`, `origin/codex/notyet-destructuring-topic`, `origin/codex/notyet-element-type-topic`, `origin/codex/notyet-enum-values-topic`, `origin/codex/notyet-function-values-topic`, `origin/codex/notyet-generic-returns-t-topic`, `origin/codex/notyet-library-small-object-ruling`, `origin/codex/notyet-new-expression-topic`, `origin/codex/notyet-object-property-topic`, `origin/codex/notyet-optional-call-topic`, `origin/codex/notyet-void-value-topic-0730`, `origin/compiler/area-fix-1`, `origin/compiler/area-fix-2`, `origin/compiler/area-main-4e0bfda5`, `origin/compiler/area-merge-main-recovery`, `origin/compiler/class-wrong-output-flow`, `origin/compiler/input-spread`, `origin/compiler/landing-batch-1`, `origin/compiler/landing-batch-2`, `origin/compiler/landing-batch-3`, `origin/compiler/nullable-references-reconciled`, `origin/compiler/private-generic-2`, `origin/compiler/reland-fixes-3`, `origin/compiler/stage3-front`, `origin/compiler/stage3-front-2`, `origin/compiler/super-closure`, `origin/compiler/taste-land`.

Every searched tip and its fetched SHA:

| Tip | SHA |
|---|---|
| `origin/codex/notyet-binary` | `0c2120231ad4f154b72ae9f541ee2e3e721d5995` |
| `origin/codex/notyet-case-declaration` | `94c137835daf9c15361091089e77b322018428eb` |
| `origin/codex/notyet-case-declaration-topic` | `1acd8e30d23de010ba65f5686b7f1c7e9af180ed` |
| `origin/codex/notyet-class-construction` | `6353ee98cc9a6f9ffb39b8fdea1bf4d67e0455f0` |
| `origin/codex/notyet-class-construction-topic` | `f05e54dece394a40422402bc21352fa0c6c5a2c5` |
| `origin/codex/notyet-class-set-property` | `7f85562feb536a3723fcd270943b40eff9eb7869` |
| `origin/codex/notyet-class-set-property-topic` | `446fbec88f2bd658f28a839d318b21f7aec82dbc` |
| `origin/codex/notyet-destructuring` | `903ba7b229bad116798714b05a0ef93164da7316` |
| `origin/codex/notyet-destructuring-topic` | `d516d4e1dcc960e8a9373e161abd3adb94bb7c6b` |
| `origin/codex/notyet-element-access` | `638c9f530c6b9ad20a8c021b332a33deb05fe176` |
| `origin/codex/notyet-element-type` | `8c8611b0ee51d1bdd342ac34750ba05f50b3eede` |
| `origin/codex/notyet-element-type-topic` | `66047ac917eaf362e80a3765850046fe966230da` |
| `origin/codex/notyet-enum-values-topic` | `f8bfba3d10488a72f9427d6c598f5091ba13d5fa` |
| `origin/codex/notyet-for-in` | `115302f2e5fdbc2d966d9cd00e0eecbf1b4f3254` |
| `origin/codex/notyet-for-of-object` | `fbc6e6702ea70bf406ecbb9a735905a46026b8df` |
| `origin/codex/notyet-function-values` | `7a3bc40cded34297da49334ac110de9721c3c759` |
| `origin/codex/notyet-function-values-topic` | `ae8ff5831550ca46a7da7a5e57bf696eb842e0c3` |
| `origin/codex/notyet-generic-returns-t` | `2ec964a8e5f241f4048df8dd19c84c4614f68bc4` |
| `origin/codex/notyet-generic-returns-t-topic` | `6f2c92fc10882b1064222d31642468766b2976ef` |
| `origin/codex/notyet-library-small` | `e62b9ef2b6b72463c049d1ed6015a36624d5875c` |
| `origin/codex/notyet-library-small-object-ruling` | `53adbbce09fcf4507e06e08c67477b5afacb9d7d` |
| `origin/codex/notyet-namespace-reads` | `e1ebac4520adf1db9ebba4981a143d671225f2b5` |
| `origin/codex/notyet-new-expression` | `a625251c9557d42c62b8bbdcd2c30f4d70739af4` |
| `origin/codex/notyet-new-expression-topic` | `5439d953229f09c4c9c84de4d00d03f411b96538` |
| `origin/codex/notyet-object-property` | `72ebd024ae1291c7ad2d23f0d2aed7eb0e8c7aff` |
| `origin/codex/notyet-object-property-topic` | `e3102b7a1cc87408f5ead42e9ce9dd87e6d5e428` |
| `origin/codex/notyet-object-small` | `f8463395eaa174dc5b8a6f1e8b139cd95e087ebc` |
| `origin/codex/notyet-optional-call` | `8b69eda98c6d3474c9c3131a8eab34bc40bfd2fb` |
| `origin/codex/notyet-optional-call-topic` | `3db9290c11d71eb00e1727463c30cbf73edbfd3b` |
| `origin/codex/notyet-overloads` | `46ac7bc77ea33c5ff48899402c846e2ea3c2990b` |
| `origin/codex/notyet-predicates` | `c191f97b65a8b00db3305117e037cde2f6a6a65a` |
| `origin/codex/notyet-predicates-topic` | `3fa87ffb99a736de1410cf9ad90e4a269bedf198` |
| `origin/codex/notyet-predicates-views` | `faeefe95573efc3abb649057fbf2989466603b11` |
| `origin/codex/notyet-representations` | `37ff860f082f81aad6ad44b49a014c5c2d1c6b83` |
| `origin/codex/notyet-rest-args` | `d206f30cf84eb4cf6695437a2c9f2dff14bb525a` |
| `origin/codex/notyet-statements-small` | `742bada352d4fb898bb0bf65f52e01f0f80a97e3` |
| `origin/codex/notyet-statics` | `1db2b32429c7edd8e4b6cfcd36ac98cea8df4a9a` |
| `origin/codex/notyet-syntax-kinds` | `03e2eaf5d709c9c5a587d2aa5eecb1180bc1cf2e` |
| `origin/codex/notyet-this-outside` | `5cdfbf7168cbc627f5386f3c5d5f30cc459a98cd` |
| `origin/codex/notyet-void-or-undefined` | `fb3985f1c2e4d136f162f0ef1896a3c28f7fa031` |
| `origin/codex/notyet-void-value` | `89dfb2eb133a0198b28caedcaf4c002667f50dc4` |
| `origin/codex/notyet-void-value-topic-0730` | `5963afb742f823912817fc334dce2901edc8defe` |
| `origin/compiler/area-closure` | `d9b1062dfc57617bc370d6b8524cf0ec59c82499` |
| `origin/compiler/area-fix-1` | `3d193f8a88d0aa23dbaa33d3492659496ab1c6dd` |
| `origin/compiler/area-fix-2` | `25981c38601d02d1222ff2337d322c76dc1a3928` |
| `origin/compiler/area-fix-39` | `1e5e25d281066458bec47fcf74c36320bb15ccd2` |
| `origin/compiler/area-fix-4` | `865f729eb6f0d98b47cef8619e06599c91c34ec6` |
| `origin/compiler/area-fix-5` | `0fbcea80759b0295e3ce627f64eaa28bbbd7a82f` |
| `origin/compiler/area-gaps` | `7f98e622903c1b1dd78708f43d942267b222cf2a` |
| `origin/compiler/area-main-4e0bfda5` | `5b9321f9a3ef541ecd8049fbb0ef3c08e84eda21` |
| `origin/compiler/area-main-c6761c24` | `09e711f559e076833ae423e5c4d83433cfa395b8` |
| `origin/compiler/area-main-f4efdd23` | `de111681883c99f201170f6fc38c665ffc2cc62f` |
| `origin/compiler/area-merge-main-recovery` | `95d6be8934e7607bbf3351a9798eb25047d3eea2` |
| `origin/compiler/area-next` | `337aa466bf7b519ffdc50c6f16e976a330a4a4f5` |
| `origin/compiler/area-next-drop` | `784b577a7488ddd0ce4cb2b82a96fa6535395896` |
| `origin/compiler/area-stack` | `e0e234ea2339edb433f1e78e1a14211b83192a99` |
| `origin/compiler/area-views-wip` | `3d74d3b756f60086c646937d8bce8636310e6c20` |
| `origin/compiler/class-wrong-output-flow` | `9b5f5fb531223d875b5277c4cbd9b39068eb9cd1` |
| `origin/compiler/decode-runtime-modules` | `addb9fbdeac002c0dccc8a643a4f1655d211342b` |
| `origin/compiler/input-spread` | `85328b26aa5c0c614ea21c6e0455ece8185e6afe` |
| `origin/compiler/landing-batch-1` | `e5d34c09923d5b2f7aab1ab0fda43f3379878580` |
| `origin/compiler/landing-batch-2` | `1821239302b9393a79387c699ff4f325fdc41623` |
| `origin/compiler/landing-batch-3` | `a37ebdb0913eca9d3d6e8adbe8f01bb733da52eb` |
| `origin/compiler/nonnull-refuse-a` | `de9795bc5a9f15e56bc5b21fe0c41ac0db385812` |
| `origin/compiler/nullable-references-reconciled` | `bbe9f350dad4d95e21990bf083aa361a1e81e734` |
| `origin/compiler/private-generic-2` | `9d8c53829844c49aa236e96f04dd37b0d7b46374` |
| `origin/compiler/reland-fixes-3` | `904cae75fb01630c1d1f85a961cbf92bbf655da0` |
| `origin/compiler/stage3-front` | `829b8a62d517577c186f14e7d6343157794de1b8` |
| `origin/compiler/stage3-front-2` | `391b3e9c378c0f5aea5ec34aec89a22eb3b59e73` |
| `origin/compiler/stage3-front-3` | `a36d1c0472649ef2ee549cc2505b93a84e893b59` |
| `origin/compiler/super-closure` | `20610643cd2fc90d763201640a380d76ee314ba8` |
| `origin/compiler/taste-land` | `bb61f951d7d5845c1cf4342995a891ca8a690c77` |

## Replay preparation and validation limits

Prepare an isolated detached census-pin worktree; do not merge an evidence branch into the delivery branch. The census pin already contains the replay tool from codex/stage3-census-replay (`9a1f14c5` is an ancestor):

```sh
git worktree add --detach /tmp/hidden-census-source 388096e6
# Both pins use cohere 7945d102a6c18dd36adf9114a758ce646e8b2359.
# Reference the initialized submodule from this repository; do not copy it.
test "$(git -C cohere rev-parse HEAD)" = 7945d102a6c18dd36adf9114a758ce646e8b2359
rmdir /tmp/hidden-census-source/cohere
ln -s "$PWD/cohere" /tmp/hidden-census-source/cohere
cd /tmp/hidden-census-source
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /tmp/hidden-adapted > /tmp/hidden-briefs-apply.log 2>&1
```

Before each replay, verify the adapted file hash against the pinned RESULT.json `files[file].sha256`; offsets are not portable to a different adaptation. The brief contains the exact census diagnostic selector. Replay is not present on main; run the inherited tool in /tmp/hidden-census-source so its compiler pin matches the hidden census. All fifteen selectors were executed with the prebuilt replay worker. Eleven exactly reproduced position, kind and reason; three mismatched (02, 04, 05), and 03 timed out at 180s. The reductions for 02 and 03 reproduce their reasons on current main. Exact caller-context reproduction remains open for 04 and 05, and the original-position retries for 02 and 03 remain open. Node witnesses and matching reductions are distinguished from those historical source-position checks.

Remeasure using the hidden README workflow and the same adapted-source hashes. For original interval [start,end), intersect the new `hidden_ranges` with that interval, union intersections, and subtract their length from the original byte length. Record the resulting bytes revealed and any next stop. Do not add overlap lengths or infer revealed bytes from disappearance of a reason string.

No compiler changes, no new oracle fixtures, no counts refresh required for a fixture addition, and no compiler mutants or backend equivalence run claimed. The per-brief mutants are acceptance criteria for later implementation units. Node observation checks included deliberately changed expected outputs; these are comparison controls, not compiler mutants. Ranking validation and its mutants are reported below.

## Documentation validation

`python3 /tmp/hidden-briefs/probes.py` ran fifteen Node witnesses: all exited 0 with the displayed stdout and empty stderr. `python3 /tmp/hidden-briefs/validate.py` independently compared the complete ledger with the pinned RESULT.json, checked all fifteen brief totals and descending order, and passed. Three in-memory documentation mutants (one byte added, wrong rank, shifted start endpoint) each failed the independent ledger assertion. Logs: `/tmp/hidden-briefs/node.log` and `/tmp/hidden-briefs/validation.log`. No production code was mutated.

Exact pinned file hashes for the replay heads:

| Brief | File | SHA-256 |
|---|---|---|
| 01 | `transformers/declarations.ts` | `3b69b0e68abb0f4a0745d5dd45651f43c9c336fe5b72883fe8d4dde1f6c4d6ed` |
| 02 | `utilities.ts` | `ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef` |
| 03 | `transformers/module/module.ts` | `d1e29a4b33d1b7c31d8ed53e8c6321beffe682ffbdc4f81560dd190c191f3c6c` |
| 04 | `checker.ts` | `938b2634f27cb57a202675254cf9af0935a891fecddb8f9d94a12d69cf0f93fb` |
| 05 | `transformers/es2018.ts` | `85b478ff113c6dfa39fd489f0c81c654ceb24a0ee8ef82592c16264301ef7b88` |
| 06 | `transformers/esDecorators.ts` | `3d77006c23ed173fb28a1399b0404a6cf51fe9976993747a22fb14d8a6059633` |
| 07 | `checker.ts` | `938b2634f27cb57a202675254cf9af0935a891fecddb8f9d94a12d69cf0f93fb` |
| 08 | `utilities.ts` | `ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef` |
| 09 | `parser.ts` | `9c8e41e143294677580750ada80e5ba18128645e280e7950b001e054b4ae0df4` |
| 10 | `checker.ts` | `938b2634f27cb57a202675254cf9af0935a891fecddb8f9d94a12d69cf0f93fb` |
| 11 | `checker.ts` | `938b2634f27cb57a202675254cf9af0935a891fecddb8f9d94a12d69cf0f93fb` |
| 12 | `binder.ts` | `8ff0292668cdbe20584f06f5ca54c0a9ae419b16bc099c8d6261b1bd595a2d39` |
| 13 | `transformers/es2017.ts` | `a5ad5e9213c7af3f4f266f50553cdfc4e31296cbe54fe8464d6a792e4d2f2755` |
| 14 | `utilities.ts` | `ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef` |
| 15 | `program.ts` | `7aaf5804093401db924e8278b142db40fcb2b14cef0566081cfaf870477b5284` |

## Toolchain setup observed

Ran `export GOPROXY='https://proxy.golang.org|direct'` then `bash cloud/setup.sh > /tmp/hidden-briefs-setup.log 2>&1`, exit 0, and sourced `/workspace/adamic-tools/env.sh`. Two cache-download HTTP 500 responses recovered through the setup script; the cohere cache restored the required pin. `nproc`: 5; cgroup `cpu.max`: `400000 100000`. Node 24.19.0, Go 1.27.1, clang 20.1.8. These setup timing lines are cumulative unless marked step-duration:

```text
setup: node ready (0.476s)
setup: go ready (0.719s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (6.318s)
setup: markdown dependencies installed (npm ci, integrity verified); step-duration=11.750s
setup: markdown dependencies ready (13.845s)
setup: submodules ready (133.177s)
setup: go build ready (2929.234s)
setup: test binaries deferred (use --warm-tests) (2929.730s)
setup: build cache warm (2929.734s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (2930.129s)
```

The only whole-repository build was the user-required setup. No whole-package tests or full gate were run. A preliminary scratch compiler build was canceled to avoid duplicate cold dependency compilation. Preliminary replay-build attempts exposed a missing module update requirement and then a duplicate workspace replacement; an isolated scratch workspace with absolute cohere references resolved both. The delivery repository's module files and compiler sources were not edited.

## Observed head replay and reduction checks

`bash /tmp/hidden-census-source/stage3/apply.sh /tmp/hidden-adapted > /tmp/hidden-briefs/apply.log 2>&1` exited 0. SHA-256 comparison of all 82 compiler files against pinned RESULT.json found zero mismatches. The replay worker was built successfully with the inherited `make_overlay.py` overlay and an isolated workspace pointing to the existing cohere submodule:

```sh
cd /tmp/hidden-census-source
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/hidden-briefs/replay-overlay > /tmp/hidden-briefs/replay-overlay.log 2>&1
GOWORK=/tmp/hidden-briefs/replay.work go build -buildvcs=false \
  -overlay=/tmp/hidden-briefs/replay-overlay/overlay.json \
  -o /tmp/hidden-briefs/replay-worker ./stage3/census/latent/replay/worker > /tmp/hidden-briefs/replay-build.log 2>&1
```

The worker executes the load/lower output guards and then matches position, kind and reason exactly. Each brief records its complete flags. The first three probes were sequential, bounded at 180s; remaining independent probes ran with three workers, `GOMAXPROCS=1` per worker and 120s limits. A task-owned sequential runner was stopped after its third completed result to avoid duplicate probes; completed results were retained. The replacement oldSourceFiles probe passed. No compiler sources were changed to run these checks.

| Brief | Exact census head replay | Seconds | Current-main witness boundary check |
|---|---|---:|---|
| 01 | PASS, exit 0 | 28.751 | Different earlier stop, exit 1 |
| 02 | Mismatch, exit 1 | 35.753 | Same reason, exit 1 |
| 03 | Timeout, 180s | 180.180 | Same reason, exit 1 |
| 04 | Mismatch, exit 1 | 21.605 | Different earlier stop, exit 1 |
| 05 | Mismatch, exit 1 | 23.161 | Different earlier stop, exit 1 |
| 06 | PASS, exit 0 | 22.956 | C emitted, exit 0; not a stop reproduction |
| 07 | PASS, exit 0 | 39.481 | Same reason, exit 1 |
| 08 | PASS, exit 0 | 39.521 | Same reason, exit 1 |
| 09 | PASS, exit 0 | 63.153 | Same reason, exit 1 |
| 10 | PASS, exit 0 | 19.438 | Different earlier stop, exit 1 |
| 11 | PASS, exit 0 | 16.628 | C emitted, exit 0; not a stop reproduction |
| 12 | PASS, exit 0 | 25.428 | C emitted, exit 0; not a stop reproduction |
| 13 | PASS, exit 0 | 28.711 | Different earlier stop, exit 1 |
| 14 | PASS, exit 0 | 24.788 | Different earlier stop, exit 1 |
| 15 | PASS, exit 0 | 29.401 | Not rerun; exact census replay passed |

The main compiler build `go build -o /tmp/hidden-briefs/adamic ./cmd/adamic > /tmp/hidden-briefs/adamic-build.log 2>&1` exited 0. Fifteen initial witness checks ran as `/tmp/hidden-briefs/adamic c /tmp/hidden-briefs/NN.a`, with output in `reduction-NN.log`; five matched the intended reasons (02, 03, 07, 08, 09). Four emitted C (06, 11, 12 and the subsequently excluded Map witness); emitting C is not backend execution or proof that a historical head is fixed. Six hit other stops. The replaced Map check is retained only as `reduction-map-excluded.log`. The new 15 has its exact head replay and Node observation instead.

A wrong-reason replay mutant appended ` MUTANT` to brief 01's exact reason. It exited 1 with `replay signature did not reproduce` and never claimed reproduction; the unmutated selector exited 0. Catcher: exact position/kind/reason matcher. Logs: `replay-mutant.log` and `replay-mutant-run.log`. Together with the byte-count, rank, endpoint and Map-category documentation mutants, five deliberate mutants were caught. The fifteen changed Node expected-output controls also rejected their altered goldens; none is a compiler mutation.

For 04 and 05, small standalone census reductions were tried with explicit un-narrowed local copies. They still did not reproduce the retained reasons: 04 stopped later at `an array of T`, and 05 emitted no lowering finding. Their failed logs are `replay-04-reduction.log` and `replay-05-reduction.log`. A separate caller-attempt selector or a faithful caller-context reduction is needed before those two briefs meet the exact head-reproduction requirement. This documentation unit does not change the replay tool.

Final documentation scope checks passed: exactly sixteen Markdown files, no em dashes, and `git diff --cached --check` clean. Independent category sums matched 154,365 view bytes, 106,264 checker-only bytes and 170,626 eligible bytes. A fourth documentation mutant moved the branded Map region back into eligible; the independent category-sum assertion caught it. Log: `/tmp/hidden-briefs/category-mutant.log`. The completed set therefore has four documentation mutants plus the wrong-reason replay mutant, all caught. No compiler mutant or backend equivalence is claimed.

## Overload regions measured on c68bf26c

Source pin: `388096e6`. Compiler: `c68bf26cb4225e1f24a84f1a412f10e7eaa3da9a`.
All 82 adapted source hashes match the pin. The measurement attempts every independently
 overlapping declaration in these seven intervals, with the entire project registered.
It is a measurement of a checker-rejected program, not proof that the compiler corpus compiles.
The original ledger above remains the historical census.

| Region | File and byte interval | Hidden before | Hidden on c68bf26c | Revealed |
|---|---|---:|---:|---:|
| 01 large | declarations.ts [68180, 81805) | 13,625 | 13,625 | 0 |
| 01 small | declarations.ts [82928, 89827) | 6,899 | 6,899 | 0 |
| 05 large | es2018.ts [35690, 41768) | 6,078 | 6,078 | 0 |
| 05 small | es2015.ts [144178, 149926) | 5,748 | 5,748 | 0 |
| 06 | esDecorators.ts [61324, 72741) | 11,417 | 11,417 | 0 |
| 13 | es2017.ts [30237, 37526) | 7,289 | 7,289 | 0 |
| 14 | utilities.ts [455532, 462634) | 7,102 | 7,102 | 0 |

The exact replay matcher confirms visitNode's parameter refusal, visitNodes' visitor
parameter refusal, transformAsyncFunctionBody's Block result refusal, and evaluate's
EvaluatorResult refusal. Hidden-01 instead stops at declarations.ts:612:90 on
`an indirect value of an overload requiring a checked implementation boundary`.
The historical optional-callback selector instead reaches visitNode's parameter refusal.
Both selectors exit 1 because the expected signature did not reproduce; neither means a fix.

Remaining groups, by assigned bytes: hidden-01 callback contracts (20,524), hidden-05
node and callback contracts (11,826), hidden-06 (11,417), hidden-13 structural result
(7,289), hidden-14 structural and returned callback results (7,102).
Hidden-06's worker separately revealed 8,048 bytes with constraint-backed TNode storage.
That unlanded dependency is absent from c68bf26c and was not merged for this baseline.
No new compiler fix or additional revealed byte is claimed in this measurement group.

Evidence: [region counts](../../docs/overload-results/groups/c68/regions.json),
[replay commands](../../docs/overload-results/groups/c68/replays.json),
[raw census](../../docs/overload-results/groups/c68/census.jsonl.gz).

## Callback follow-up on 0c84ce9c

Compiler `0c84ce9c93451b4e5aa7bbeb0d2fb2e38f8e7a9a` adds immediate callback
arguments when the implementation directly serves the consumer's concrete
parameters and result, with matching call and readonly scalar field storage.
The original closure preserves identity and captures. Its new fixture agrees
with Node in both backends, including sanitizers; four compiler mutants are caught.

All seven intervals were measured again against the c68bf26c records with the
same scope and all 82 source hashes verified. The counts above are unchanged:
58,158 hidden bytes, zero revealed. This callback case does not establish the
full visitor or structural result contracts. Hidden-06's unlanded TNode storage
change remains separate. The structural result group is not delivered.

Evidence: [counts](../../docs/overload-results/groups/callback/regions.json),
[tests and limitations](../../docs/overload-results/groups/callback/REPORT.md),
[compiler mutants](../../docs/overload-results/groups/callback/mutants.json).
