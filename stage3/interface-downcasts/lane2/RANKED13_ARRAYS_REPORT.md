Held the thirteenth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows 6b3e88a0 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 28 probes in TestCheckedViewRanked13ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes).
Mutants: two execution catches (an array element where an object element is declared, in both backends).
Limits: five pairs on this stretch are not lane 2's (four union-alias casts for lane 4, one index signature).

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| ClassExpression.typeParameters | 6 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked13-class-expression-type-parameters*.a | yes |
| ClassLikeDeclaration.heritageClauses | 6 | cast to the alias `ClassDeclaration \| ClassExpression` | none | no: lane 4 |
| ConditionalRoot.inferTypeParameters | 6 | `inferTypeParameters?: TypeParameter[]` | ranked13-conditional-root*.a | yes |
| ConditionalRoot.outerTypeParameters | 6 | `outerTypeParameters?: TypeParameter[]` | ranked13-conditional-root*.a | yes |
| DeclarationWithTypeParameterChildren.typeParameters | 6 | cast to a union alias | none | no: lane 4 |
| FunctionExpression.modifiers | 6 | `readonly modifiers?: NodeArray<Modifier>` | ranked13-function-expression-modifiers.a | yes |
| FunctionTypeNode.typeParameters | 6 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked13-function-type-type-parameters.a | yes |
| ImportTypeNode.typeArguments | 6 | `readonly typeArguments?: NodeArray<TypeNode>` | ranked13-import-type-arguments.a | yes |
| JSDocImportTag.comment | 6 | `readonly comment?: string \| NodeArray<JSDocComment>` | ranked13-jsdoc-import-comment.a | yes |
| JSDocTypedefTag.comment | 6 | `readonly comment?: string \| NodeArray<JSDocComment>` | ranked13-jsdoc-typedef-comment*.a | yes |
| JsxOpeningLikeElement.typeArguments | 6 | cast to a union alias | none | no: lane 4 |
| MapLike<string[]>.<dynamic-key> | 6 | `[index: string]: T` | none | no: 0.1 refuses index signatures |
| MappedTypeNode.members | 6 | `readonly members?: NodeArray<TypeElement>` | ranked13-mapped-members*.a | yes |
| MethodSignature.parameters | 6 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked13-method-signature*.a | yes |
| MethodSignature.typeParameters | 6 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked13-method-signature.a | yes |
| SourceFile.commentDirectives | 6 | `commentDirectives?: CommentDirective[]` | ranked13-comment-directives*.a | yes |
| SourceFile.parseDiagnostics | 6 | `parseDiagnostics: DiagnosticWithLocation[]` | ranked13-parse-diagnostics*.a | yes |
| TaggedTemplateExpression.typeArguments | 6 | `readonly typeArguments?: NodeArray<TypeNode>` | ranked13-tagged-template-arguments*.a | yes |
| TypeQueryNode.typeArguments | 6 | `readonly typeArguments?: NodeArray<TypeNode>` | ranked13-type-query-arguments.a | yes |
| AnonymousType.aliasTypeArguments | 5 | `aliasTypeArguments?: readonly Type[]` | ranked13-anonymous-alias-arguments*.a | yes |
| CaseOrDefaultClause.statements | 5 | cast to the alias `CaseClause \| DefaultClause` | none | no: lane 4 |
| ConstructorDeclaration.typeParameters | 5 | `readonly typeParameters?: NodeArray<TypeParameterDeclaration>` | ranked13-constructor-type-parameters-absent.a | yes |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts.
ConditionalRoot keeps `instantiations?: Map<string, Type>` unread and is reached as
`ConditionalType.root`, as the census's `type.root.inferTypeParameters` does.
CommentDirective keeps `type: CommentDirectiveType`, a const enum as in tsc.
DiagnosticWithLocation is modeled without its `file`, so a parseDiagnostics push is a
flat record push and is admitted against the original elements' contract.

Observed: ranked13-comment-directives-open-enum stores 7 in a `CommentDirectiveType`
field (members 0 and 1) and Adamic prints 7, as Node does. That is the documented
rule in internal/lower/view_contracts.go, "A whole numeric enum admits numbers
outside its declared members", matching TypeScript's numeric enums. Single-member
enum literal types stay closed (ranked2-tuple-wrong-member).

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-array-element-as-object (an array element reads as object) | ranked13-anonymous-alias-arguments-wrong-array | UBSan misaligned shape access |
| javascript-array-element-as-object (object check admits arrays) | ranked13-anonymous-alias-arguments-wrong-array | a different, later diagnostic |

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 127 / 2,616 | 207 / 573 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

The array remainder includes nine union-target pairs lane 4 owns (90 reads) and four
index-signature pairs 0.1 refuses (32 reads). Tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked13' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked13-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

28 PASS lines, gofmt clean, vet passes.
