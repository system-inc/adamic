Held the eighth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows 5a7f9db7 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 37 probes in TestCheckedViewRanked8ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes).
Mutants: four execution catches (required-field presence and record write pairs, each in both backends).
Limits: a push of a Signature (it holds an array) stays a named runtime refusal; Signature.compositeSignatures is credited for reads only.

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| JsxElement.children | 15 | `readonly children: NodeArray<JsxChild>` | ranked8-jsx*.a | yes |
| MethodDeclaration.parameters | 15 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked8-method-parameters*.a | yes |
| Signature.compositeSignatures | 15 | `compositeSignatures?: Signature[]` | ranked8-composite*.a | reads only |
| Bundle.sourceFiles | 14 | `readonly sourceFiles: readonly SourceFile[]` | ranked8-bundle*.a | yes |
| EnumDeclaration.members | 14 | `readonly members: NodeArray<EnumMember>` | ranked8-enum*.a | yes |
| FunctionDeclaration.parameters | 14 | `readonly parameters: NodeArray<ParameterDeclaration>` | ranked8-function-parameters*.a | yes |
| GenericType.typeParameters | 14 | `typeParameters: TypeParameter[] \| undefined` (InterfaceType) | ranked8-generic*.a | yes |
| ParameterDeclaration.modifiers | 14 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked8-parameter-modifiers*.a | yes |
| SourceFile.imports | 14 | `imports: readonly StringLiteralLike[]` | ranked8-imports*.a | yes |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts.
JsxChild is modeled as JsxText, JsxExpression, JsxElement and JsxFragment (tsc adds
JsxSelfClosingElement), so children recurse through two members; the recursive
`childText` walks them with map, as tsc's emitters do. HasType has four members here
(Method, Function, Property, Parameter). EnumMember.name is `Identifier |
StringLiteral` (tsc's PropertyName is wider). StringLiteralLike is tsc's exact pair.
Signature.compositeSignatures is reached as the census does, through
`resolved.callSignatures[0]`. TypeParameter is a flat record (`flags`, optional
`isThisType`), so GenericType.typeParameters is mutable and certified for writes.

What the probes cover beyond earlier groups: a recursive union element type
(`NodeArray<JsxChild>` inside JsxElement and JsxFragment), with a malformed text two
levels down; an optional object field on elements (`EnumMember.initializer`); an
element union by tag inside an object field (`name.kind`); a required
`T[] | undefined` field holding undefined (prints `none`) and missing (refuses with
`is not initialized ... found missing`, as the required declaration promises; Node
prints `none`); arrays of objects holding arrays (`Bundle.sourceFiles[0].imports`);
recursive optional arrays (`compositeSignatures![0]!.compositeSignatures`); and
pushes through a view: a flat TypeParameter record is admitted against the original
element contract, a record whose field type disagrees with the original elements is
refused at the write (`expected { flags: string; }, found { flags: number; }`), and a
Signature push is refused as uncertified.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-required-presence (a missing slot passes the presence check) | ranked8-generic-missing | a different diagnostic (`found unsupported representation`) |
| javascript-required-presence (own-property test removed) | ranked8-generic-missing | exit 0 printing `none` |
| native-record-write-pairs (incompatible pair admitted) | ranked8-generic-push-wrong | refused later, at the original alias's read |
| javascript-record-write-pairs (incompatible pair admitted) | ranked8-generic-push-wrong | refused later, at the original alias's read |

Observed: with the write-pair guard disabled, the bad push is still refused at
`raw.typeParameters[1]!.flags`, because reads of the original record are checked.
The guard moves the failure to the write, where the diagnostic names the
contracts; it is not the only defense against a silent wrong answer here.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 51 / 1,947 | 283 / 1,242 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

Overlapping candidate obligations from read-census/summary.json; tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRanked8' -count=1 -v -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked8-array-mutants.py
gofmt -l internal/oracle
```

37 PASS lines, gofmt clean. No compiler, runtime or backend file changed.
