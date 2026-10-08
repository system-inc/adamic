Held the ninth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows dd198e0d on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 37 probes in TestCheckedViewRanked9ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes).
Mutants: two execution catches (checked element reads inside includes/indexOf searches, in both backends).
Limits: reads only for the families below; no new write certificate.

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures |
| --- | ---: | --- | --- |
| ArrowFunction.parameters | 13 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked9-arrow-parameters*.a |
| ImportDeclaration.modifiers | 13 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked9-import-modifiers*.a |
| NamedImports.elements | 13 | `readonly elements: NodeArray<ImportSpecifier>` | ranked9-named-imports*.a |
| NodeBuilderContext.typeStack | 13 | `typeStack: number[]` (checker.ts) | ranked9-type-stack*.a |
| ResolvedType.indexInfos | 13 | `indexInfos: readonly IndexInfo[]` | ranked9-index-infos*.a |
| FunctionDeclaration.modifiers | 12 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked9-function-modifiers*.a |
| GetAccessorDeclaration.parameters | 12 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked9-get-parameters.a |
| IndexSignatureDeclaration.parameters | 12 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked9-index-parameters*.a |
| MethodDeclaration.modifiers | 12 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked9-method-modifiers.a |
| ModuleBlock.statements | 12 | `readonly statements: NodeArray<Statement>` | ranked9-module-block*.a |
| NewExpression.arguments | 12 | `readonly arguments?: NodeArray<Expression>` | ranked9-new-arguments*.a |
| PropertyDeclaration.modifiers | 12 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked9-property-modifiers*.a |
| ResolvedType.properties | 12 | `properties: Symbol[]` | ranked9-properties*.a |
| TemplateExpression.templateSpans | 12 | `readonly templateSpans: NodeArray<TemplateSpan>` | ranked9-template-spans*.a |

All fourteen are credited. Declarations were read from microsoft/TypeScript 050880ce,
src/compiler/types.ts, and NodeBuilderContext from src/compiler/checker.ts, whose
symbolToDeclarationsWorker pushes `type.id` and `-1` onto `context.typeStack` and pops
both. HasType is modeled with six members (Arrow, GetAccessor, IndexSignature,
Function, Method, Property); Statement as ImportDeclaration | VariableStatement; the
template literal pieces as TemplateMiddle | TemplateTail; ImportSpecifier keeps
`propertyName?: Identifier | StringLiteral` and `isTypeOnly: boolean`; IndexInfo keeps
`keyType`, `type`, `isReadonly` and `declaration?: IndexSignatureDeclaration`.

What the probes cover beyond earlier groups: push, pop, includes and lastIndexOf on a
mutable `number[]` through the view; find over a viewed `Symbol[]`; a push of a flat
Symbol record admitted, and one whose original elements disagree refused at the
write; a whole-array assignment followed by a push; modifiers read through a
parameter of a union type that spans HasType and ImportDeclaration; an optional object
field holding an array (`namedBindings!.elements`); a union-typed object field on
each element (`literal: TemplateMiddle | TemplateTail`); and an optional object field
on an element of a readonly record array (`IndexInfo.declaration`).

`ranked9-type-stack-wrong-push` refuses with `expected number, found heap pointers`:
the write check sees physical storage, and a reference array cannot say which logical
type it holds, so it names the storage. Both backends say the same.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-search-elements (search reads elements unchecked) | ranked9-type-stack-includes-wrong | exit 0 printing `false` |
| javascript-search-elements (search skips the element check) | ranked9-type-stack-includes-wrong | exit 0 printing `false` |

Observed: the mutants print `false`, which is also Node's stdout. They are caught
because the oracle pins the refusal the violated `number[]` contract requires,
not because the output differs from Node's.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 65 / 2,120 | 269 / 1,069 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

Overlapping candidate obligations from read-census/summary.json; tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked9' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked9-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

37 PASS lines, gofmt clean, vet passes. No compiler, runtime or backend file changed.
