Held the eleventh ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows 60d9bf9f on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 29 probes in TestCheckedViewRanked11ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes).
Mutants: two execution catches (the intersection tag check on array elements, in both backends).
Limits: pushing a ReverseMappedSymbol (it holds an object field) stays a named runtime refusal; three pairs on this stretch are not lane 2's (two union-target casts, one index-signature declaration).

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| CallExpression \| NewExpression.arguments | 9 | cast `parent as CallExpression \| NewExpression` | none | no: union-target cast, lane 4 |
| ClassDeclaration.typeParameters | 9 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked11-class-type-parameters*.a | yes |
| CompilerOptions.lib | 9 | `lib?: string[]` on an interface with `[option: string]: ...` | none | no: 0.1 refuses index signatures |
| ExportDeclaration.modifiers | 9 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked11-export-modifiers.a | yes |
| FunctionExpression.parameters | 9 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked11-function-expression-parameters*.a | yes |
| ImportAttributes.elements | 9 | `readonly elements: NodeArray<ImportAttribute>` | ranked11-import-attributes*.a | yes |
| IndexInfo.components | 9 | `components?: ElementWithComputedPropertyName[]` | ranked11-index-components*.a | yes |
| InterfaceDeclaration.heritageClauses | 9 | `readonly heritageClauses?: NodeArray<HeritageClause>` | ranked11-interface-heritage*.a | yes |
| NodeBuilderContext.reverseMappedStack | 9 | `reverseMappedStack: ReverseMappedSymbol[] \| undefined` (checker.ts) | ranked11-reverse-mapped*.a | reads only |
| SourceFile.libReferenceDirectives | 9 | `libReferenceDirectives: readonly FileReference[]` | ranked11-lib-references*.a | yes |
| VariableStatement.modifiers | 9 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked11-variable-modifiers*.a | yes |
| ArrowFunction.modifiers | 8 | `readonly modifiers?: NodeArray<Modifier>` | ranked11-arrow-modifiers*.a | yes |
| ClassExpression.modifiers | 8 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked11-class-expression-modifiers.a | yes |
| ConstructorDeclaration.modifiers | 8 | `readonly modifiers?: NodeArray<ModifierLike> \| undefined` | ranked11-constructor-modifiers-wrong-tag.a | yes |
| HasDecorators.modifiers | 8 | cast to the HasDecorators union alias | none | no: union-target cast, lane 4 |
| ImportEqualsDeclaration.modifiers | 8 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked11-import-equals-modifiers-absent.a | yes |
| InterfaceDeclaration.members | 8 | `readonly members: NodeArray<TypeElement>` | ranked11-interface-members*.a | yes |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts, and
NodeBuilderContext from src/compiler/checker.ts. ArrowFunction and FunctionExpression
modifiers are `NodeArray<Modifier>`, without Decorator, so a decorator there refuses at
its kind read (ranked11-arrow-modifiers-decorator). ElementWithComputedPropertyName is
tsc's intersection `(ClassElement | ObjectLiteralElement) & { name: ComputedPropertyName }`,
modeled with one ClassElement member; it is reached as an element of
`ResolvedType.indexInfos`, as the census does. ImportAttributes.token is the literal
union `118 | 132` (WithKeyword, AssertKeyword). ReverseMappedSymbol keeps an object
`links` field.

Observed: the intersection's tag (`name.kind`) is checked when an element of
`components` is read, before the program reads `.pos`, so
ranked11-index-components-wrong-name refuses before printing the line Node prints
first. That is stricter than a field-by-field reading, not unsound: the element
violates its declared type either way.

ranked11-function-expression-parameters-unread-kind gives a parameter a `name` whose
`kind` is wrong but never read; lazy checking admits it and prints what Node prints.

Named runtime refusal pinned (Node succeeds): ranked11-reverse-mapped-push, `element
write failed: <array write> expected object, found uncertified source element
contract`, because ReverseMappedSymbol holds an object field (`links`). tsc's
createAnonymousTypeNode pushes onto reverseMappedStack.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-intersection-tag (view_intersections.go literal check disabled) | ranked11-index-components-wrong-name | a different, later diagnostic |
| javascript-intersection-tag (no allowed literals) | ranked11-index-components-wrong-name | a different, later diagnostic |

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 94 / 2,395 | 240 / 794 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The array remainder includes four union-target pairs lane 4 owns (60 reads) and two
index-signature pairs 0.1 refuses (19 reads). Tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked11' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked11-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

29 PASS lines, gofmt clean, vet passes.
