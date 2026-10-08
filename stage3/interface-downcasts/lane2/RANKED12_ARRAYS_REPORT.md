Held the twelfth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows dc925ee2 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 33 probes in TestCheckedViewRanked12ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes).
Mutants: two execution catches on the string-or-array union member check; one single native mutant survives because native checks that union twice.
Limits: pushing a Diagnostic whose original elements hold undefined-typed fields stays a named runtime refusal; two pairs here are not lane 2's (one index signature, one union-target cast).

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| JSDocFunctionType.parameters | 8 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked12-jsdoc-function-parameters*.a | yes |
| MethodDeclaration.typeParameters | 8 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked12-method-type-parameters.a | yes |
| ModuleDeclaration.modifiers | 8 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked12-module-modifiers*.a | yes |
| ParsedCommandLine.errors | 8 | `errors: Diagnostic[]` | ranked12-command-errors*.a | reads only |
| SourceFile.moduleAugmentations | 8 | `moduleAugmentations: readonly (StringLiteral \| Identifier)[]` | ranked12-module-augmentations*.a | yes |
| SourceFile.packageJsonLocations | 8 | `packageJsonLocations?: readonly string[]` | ranked12-package-json-locations*.a | yes |
| TemplateLiteralTypeNode.templateSpans | 8 | `readonly templateSpans: NodeArray<TemplateLiteralTypeSpan>` | ranked12-template-literal-type*.a | yes |
| TypeAliasDeclaration.typeParameters | 8 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked12-type-alias-type-parameters.a | yes |
| TypeLiteralNode.members | 8 | `readonly members: NodeArray<TypeElement>` | ranked12-type-literal-members*.a | yes |
| CaseClause.statements | 7 | `readonly statements: NodeArray<Statement>` | ranked12-case-statements*.a | yes |
| CompilerOptions.typeRoots | 7 | `typeRoots?: string[]` on an interface with an index signature | none | no: 0.1 refuses index signatures |
| FunctionDeclaration.typeParameters | 7 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked12-function-type-parameters-absent.a | yes |
| InterfaceDeclaration.typeParameters | 7 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked12-interface-type-parameters*.a | yes |
| JSDoc.comment | 7 | `readonly comment?: string \| NodeArray<JSDocComment>` | ranked12-jsdoc-comment*.a | yes |
| JsonSourceFile.statements | 7 | `readonly statements: NodeArray<JsonObjectExpressionStatement>` | ranked12-json-statements*.a | yes |
| NodeWithTypeArguments.typeArguments | 7 | `readonly typeArguments?: NodeArray<TypeNode>` | ranked12-type-arguments*.a | yes |
| TypeParameterDeclaration.modifiers | 7 | `readonly modifiers?: NodeArray<Modifier>` | ranked12-type-parameter-modifiers*.a | yes |
| UnionOrIntersectionTypeNode.types | 7 | cast to the alias `UnionTypeNode \| IntersectionTypeNode` | none | no: union-target cast, lane 4 |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts.
JSDocComment is modeled as JSDocText | JSDocLink (tsc adds JSDocLinkCode and
JSDocLinkPlain, same shape). JSDoc.comment is a string-or-array union: the probes read
it as a string, as an array narrowed with `typeof`, and refuse a number and a
malformed array element. TypeParameterDeclaration.modifiers is `NodeArray<Modifier>`,
so a decorator refuses at its kind read. ParsedCommandLine is modeled without
`options: CompilerOptions`, whose declaration carries an index signature 0.1 refuses;
the full tsc declaration would not compile.

Named runtime refusal pinned (Node succeeds): ranked12-command-errors-push, `element
write failed: <array write> expected object, found uncertified source element
contract`; Diagnostic's `start` and `length` are `number | undefined` and the
original elements hold undefined.

Mutants and what they showed:

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-union-member-mask (runtime kinds mask also admits numbers) | ranked12-jsdoc-comment-wrong | survived: same refusal |
| native-union-member-mask-and-selection (mask widened and the emitted member selection disabled) | ranked12-jsdoc-comment-wrong | a later narrowing check refuses with a different message |
| javascript-union-member-mask (kinds mask also admits numbers) | ranked12-jsdoc-comment-wrong | a second member check refuses with a different message |

Native checks a nullable union read twice: `adamic_object_nullish_view` tests the
member kinds mask, then `nullishMemberSelection` (internal/native/view_nullish.go)
emits a kind test over the union's members. Disabling the mask alone leaves the
selection to refuse with the identical message, so that mutant survives; the
surviving log is kept. Disabling both is caught. Neither single check is the only
defense for this union.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 110 / 2,516 | 224 / 673 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The array remainder includes five union-target pairs lane 4 owns (67 reads) and three
index-signature pairs 0.1 refuses (26 reads). Tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked12' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked12-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

33 PASS lines, gofmt clean, vet passes.
