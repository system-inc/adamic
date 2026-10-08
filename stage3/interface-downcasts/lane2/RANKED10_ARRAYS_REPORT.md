Held the tenth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows 257568b0 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 34 probes in TestCheckedViewRanked10ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes) plus one pinned frontier in TestCheckedViewRanked10Frontiers.
Mutants: two execution catches (literal unions on optional fields, in both backends).
Limits: assigning an array to an optional array field through a view is a NotYet (slotContract has no array certificate); CompilerOptions.<dynamic-key> is outside lane 2 because 0.1 refuses index signatures.

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| ArrowFunction.typeParameters | 11 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked10-arrow-type-parameters*.a | yes |
| NamedExports.elements | 11 | `readonly elements: NodeArray<ExportSpecifier>` | ranked10-named-exports*.a | yes |
| SourceFile.referencedFiles | 11 | `referencedFiles: readonly FileReference[]` | ranked10-references*.a | yes |
| SourceFile.typeReferenceDirectives | 11 | `typeReferenceDirectives: readonly FileReference[]` | ranked10-references*.a | yes |
| ArrayBindingPattern.elements | 10 | `readonly elements: NodeArray<ArrayBindingElement>` | ranked10-array-binding*.a | yes |
| CallExpression.typeArguments | 10 | `readonly typeArguments?: NodeArray<TypeNode>` | ranked10-call-type-arguments*.a | yes |
| ClassExpression.heritageClauses | 10 | `readonly heritageClauses?: NodeArray<HeritageClause>` | ranked10-class-expression*.a | yes |
| ClassExpression.members | 10 | `readonly members: NodeArray<ClassElement>` | ranked10-class-expression*.a | yes |
| CompilerOptions.<dynamic-key> | 10 | `[option: string]: CompilerOptionsValue \| TsConfigSourceFile \| undefined` | none | no: 0.1 refuses index signatures |
| GetAccessorDeclaration.modifiers | 10 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked10-get-modifiers.a | yes |
| InterfaceType.typeParameters | 10 | `typeParameters: TypeParameter[] \| undefined` | ranked10-interface-type*.a | yes |
| JSDocTemplateTag.typeParameters | 10 | `readonly typeParameters: NodeArray<TypeParameterDeclaration>` | ranked10-template-tag*.a | yes |
| ObjectBindingPattern.elements | 10 | `readonly elements: NodeArray<BindingElement>` | ranked10-object-binding*.a | yes |
| SetAccessorDeclaration.modifiers | 10 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked10-set-modifiers*.a | yes |
| Symbol \| undefined.declarations | 10 | `declarations?: Declaration[]` read as `symbol?.declarations` | ranked10-symbol-declarations*.a | yes |
| TypeReference.resolvedTypeArguments | 10 | `resolvedTypeArguments?: readonly Type[]` | ranked10-resolved-arguments*.a | reads only |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts.
BindingName is `Identifier | ObjectBindingPattern | ArrayBindingPattern` and
ArrayBindingElement `BindingElement | OmittedExpression`, as in tsc, so the recursive
`bindingText` walks arrays inside elements inside arrays. FileReference keeps
`resolutionMode?: ResolutionMode` as the literal union `1 | 99` (ModuleKind.CommonJS
and ESNext) and `preserve?: boolean`. ExportSpecifier.name is `Identifier |
StringLiteral` (ModuleExportName). Symbol.declarations is reached with optional
chaining on `Symbol | undefined`, as `moduleSymbol?.declarations` does.

CompilerOptions.<dynamic-key> reads `compilerOptions[flag]` through tsc's index
signature. 0.1 refuses an index signature anywhere ("use a Map"), before any view is
involved, so the pair is outside this family and gets no fixture. It stays in the
remainder below, marked.

Frontier pinned before lowering (Node runs it): ranked10-resolved-arguments-assign
assigns `[{flags: 32}, {flags: 64}]` to the optional `resolvedTypeArguments` through
the view. `setProperty` (internal/lower/class.go) requires a write certificate for an
optional field written through a view, from `slotContract`
(internal/lower/view_writes.go), and `slotContract` returns none for an array
representation, so lowering says `stage 0 can't lower a checked write without a
reifiable source-slot type certificate yet`. ranked7-related-assign is the same
refusal. Whole-array assignment to a required array field works (ranked5, ranked6,
ranked7, ranked9).

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-nullable-literal (view_nullish.go literal test disabled) | ranked10-references-wrong-mode | exit 0 printing `2` |
| javascript-nullable-literal (allowed literals ignored in adamicViewNullish) | ranked10-references-wrong-mode | exit 0 printing `2` |

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 80 / 2,274 | 254 / 915 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The array remainder includes the two union-target pairs lane 4 owns
(FunctionLikeDeclaration.parameters, `ClassDeclaration | ClassExpression`.members,
43 reads) and CompilerOptions.<dynamic-key> (10 reads). Tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked10' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked10-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

35 PASS lines (34 probes and the frontier), gofmt clean, vet passes.
