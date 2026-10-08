Held the fourteenth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows 924eccb0 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 31 probes in TestCheckedViewRanked14ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes) plus one pinned frontier in TestCheckedViewRanked14Frontiers.
Mutants: three execution catches (boolean element fields in both backends, and the frontier pin).
Limits: view_lazy.go's field-name fallback refuses every viewed `.text` read once tsc's EmitHelper.text (`string | ((name) => string)`) is declared; the fixtures model EmitHelper.text as string. InterfaceType.resolvedBaseTypes is not credited (untagged BaseType union).

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| ConstructorTypeNode.modifiers | 5 | `readonly modifiers?: NodeArray<Modifier>` | ranked14-constructor-type*.a | yes |
| ConstructorTypeNode.parameters | 5 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked14-constructor-type*.a | yes |
| ConstructorTypeNode.typeParameters | 5 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked14-constructor-type*.a | yes |
| EmitNode.helpers | 5 | `helpers?: EmitHelper[]` | ranked14-emit-helpers*.a | yes |
| InterfaceType.localTypeParameters | 5 | `localTypeParameters: TypeParameter[] \| undefined` | ranked14-interface-type*.a | yes |
| InterfaceType.resolvedBaseTypes | 5 | `resolvedBaseTypes: BaseType[]`, `BaseType = ObjectType \| IntersectionType \| TypeVariable` | ranked14-interface-type*.a | no: untagged union element, modeled as Type |
| InterfaceTypeWithDeclaredMembers.declaredProperties | 5 | `declaredProperties: Symbol[]` | ranked14-interface-type*.a | yes |
| JSDocSignature.parameters | 5 | `readonly parameters: readonly JSDocParameterTag[]` | ranked14-jsdoc-signature-parameters*.a | yes |
| JsxFragment.children | 5 | `readonly children: NodeArray<JsxChild>` | ranked14-jsx-fragment-children*.a | yes |
| MethodSignature.modifiers | 5 | `readonly modifiers?: NodeArray<Modifier>` | ranked14-method-signature-modifiers.a | yes |
| NamespaceExportDeclaration.modifiers | 5 | `readonly modifiers?: NodeArray<ModifierLike>` (internal) | ranked14-namespace-export-modifiers.a | yes |
| PropertySignature.modifiers | 5 | `readonly modifiers?: NodeArray<Modifier>` | ranked14-property-signature-modifiers-wrong-tag.a | yes |
| SignatureDeclaration \| JSDocSignature.parameters | 5 | cast to a union | none | no: lane 4 |
| TransientSymbol.declarations | 5 | `declarations?: Declaration[]` (Symbol) | ranked14-transient-declarations*.a | yes |
| UnionTypeNode.types | 5 | `readonly types: NodeArray<TypeNode>` | ranked14-union-type-types*.a | yes |
| CircularBuildOrder.circularDiagnostics | 4 | `circularDiagnostics: readonly Diagnostic[]` | ranked14-circular-diagnostics*.a | yes |
| DefaultClause.statements | 4 | `readonly statements: NodeArray<Statement>` | ranked14-default-clause*.a | yes |
| EmitNode.tokenSourceMapRanges | 4 | `tokenSourceMapRanges?: (SourceMapRange \| undefined)[]` | ranked14-token-source-map-ranges*.a | yes |
| EnumDeclaration.modifiers | 4 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked14-enum-modifiers*.a | yes |
| ExportAssignment.modifiers | 4 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked14-export-assignment-modifiers.a | yes |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts.
EmitNode is reached as an optional `emitNode` field of a node, as `destEmitNode.helpers`
is. EmitHelper keeps the Scoped/Unscoped union tagged by `scoped: true | false`.
tokenSourceMapRanges is an array whose element type admits undefined, read with a hole.

Frontier pinned before lowering (Node runs it): ranked14-text-name-fallback declares
an interface with tsc's `text: string | ((name: string) => string)`, never casts or
reads it, and reads `KeywordTypeNode.text` through a view; lowering refuses with
`checked view read of field text with unsupported callable contract`. The cause is
the conservative field-name fallback in `lazyViewReadRefusal`'s walk,
internal/lower/view_lazy.go (`family = unsupportedFields[field]`, commented "Wider
interfaces and instantiated helpers can give the same member a different checker type
id"). With the fallback removed, this program compiles and both backends print
`string`, as Node does; that shows the fallback is the only blocker here, not that it
can be removed in general. tsc declares EmitHelper.text this way, so with the full
declarations every viewed `.text` read (Identifier.text among them) would refuse.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| name-fallback-removed (view_lazy.go) | TestCheckedViewRanked14Frontiers/ranked14-text-name-fallback | the frontier program was admitted |
| native-boolean-field (object.c returns any slot for wanted 2) | ranked14-jsdoc-signature-parameters-wrong-bracketed | UBSan: invalid bool load |
| javascript-boolean-field (boolean check always true) | ranked14-jsdoc-signature-parameters-wrong-bracketed | a different, later diagnostic |

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 145 / 2,701 | 189 / 488 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The array remainder includes ten union-target pairs lane 4 owns (95 reads), four
index-signature pairs 0.1 refuses (32 reads) and InterfaceType.resolvedBaseTypes (5).
Tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked14' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked14-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

32 PASS lines (31 probes and the frontier), gofmt clean, vet passes.
