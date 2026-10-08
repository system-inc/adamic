Held the seventh ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows b2332701 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 36 probes in TestCheckedViewRanked7ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes) plus one pinned frontier in TestCheckedViewRanked7Frontiers.
Mutants: two execution catches (field numeric literal unions in both backends).
Limits: pushes of non-flat records (DiagnosticRelatedInformation, DiagnosticWithLocation) stay named runtime refusals; tsc's `diagnostic.relatedInformation = []` through an element view is a loud NotYet.

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| ClassDeclaration.heritageClauses | 18 | `readonly heritageClauses?: NodeArray<HeritageClause>` (ClassLikeDeclarationBase) | ranked7-heritage*.a | yes |
| SetAccessorDeclaration.parameters | 18 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked7-set-parameters*.a | yes |
| Type.aliasTypeArguments | 18 | `aliasTypeArguments?: readonly Type[]` | ranked7-alias*.a, ranked7-template-types.a | yes |
| ConstructorDeclaration.parameters | 17 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked7-constructor-parameters*.a | yes |
| HeritageClause.types | 17 | `readonly types: NodeArray<ExpressionWithTypeArguments>` | ranked7-heritage*.a | yes |
| Diagnostic.relatedInformation | 16 | `relatedInformation?: DiagnosticRelatedInformation[]` | ranked7-related*.a | reads only |
| ParsedCommandLine.projectReferences | 16 | `projectReferences?: readonly ProjectReference[]` | ranked7-references*.a | yes |
| SourceFile.bindDiagnostics | 16 | `bindDiagnostics: DiagnosticWithLocation[]` | ranked7-bind*.a | reads only |
| TemplateLiteralType.types | 16 | `types: readonly Type[]` | ranked7-template-types*.a | yes |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts.
The census reaches the parameters pairs as members of HasType, ClassDeclaration as a
member of ObjectTypeDeclaration, Type.aliasTypeArguments through
`indexedAccessType.objectType`, and Diagnostic.relatedInformation through
`diagnostics[result]`. The fixtures model HasType as three members (SetAccessor,
Constructor, Property) narrowed by kind; Type untagged with `flags`; Diagnostic reached
as an element of `ParsedCommandLine.errors: Diagnostic[]`, tsc's own field.
DiagnosticRelatedInformation keeps `file: SourceFile | undefined`, `start` and `length`
as `number | undefined` and `messageText: string | DiagnosticMessageChain`; the mixed
messageText union is never read, and lazy checking leaves it unchecked.

Passing a DiagnosticWithLocation where a Diagnostic is expected is refused by 0.1
(`file` is writable as `SourceFile | undefined` through the wider type and read as
SourceFile through the narrower). tsc relies on that unsound covariance; the fixtures
reach Diagnostic through `errors` instead, and the refusal is not counted as a gap.

What the probes cover beyond earlier groups: a numeric literal union field
(`token: 96 | 119`) on an array element, an optional boolean (`circular`) and optional
string (`originalPath`) on array elements, an optional array nested in an element of an
optional array (`aliasTypeArguments![0]!.aliasTypeArguments`), arrays narrowed out of a
union by kind before the read, some/filter/reduce/for...of over viewed elements, a
required object field on an element (`file`), and whole-array assignment of a mutable
`DiagnosticWithLocation[]`.

Pinned frontiers and refusals (Node succeeds in each):

- ranked7-related-assign, before lowering: `diagnostic.relatedInformation = []` on a
  Diagnostic read from a viewed element is `stage 0 can't lower a checked write without
  a reifiable source-slot type certificate yet`. tsc's addRelatedInfo does this.
- ranked7-related-push, exit 70: `found uncertified source element contract`; the
  original elements hold undefined-typed fields.
- ranked7-bind-push, exit 70: same refusal for a DiagnosticWithLocation, which holds a
  SourceFile. tsc's binder pushes bindDiagnostics this way.

Reads of relatedInformation and bindDiagnostics are credited; their writes are not
certified, and the report counts reads only.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-field-literal (view_fields.go literal test always true) | ranked7-heritage-wrong-token | exit 0 printing `5` |
| javascript-field-literal (no allowed literals passed) | ranked7-heritage-wrong-token | exit 0 printing `5` |

The other checks these probes rest on were proven by the fifth and sixth groups'
mutants at the same sites.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 42 / 1,818 | 292 / 1,371 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

Overlapping candidate obligations from read-census/summary.json; tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked7' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked7-array-mutants.py
gofmt -l internal/oracle; go vet ./internal/oracle
```

37 PASS lines (36 probes and the frontier), gofmt clean, vet passes. No compiler,
runtime or backend file changed, so the lane 2 suite and packages were not rerun for
this group; the sixth group's runs stand.
