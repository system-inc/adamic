Held the sixth ranked group of array contracts to Node in both backends and fixed one backend disagreement.
Commit: follows c94b2a7b on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 37 probes in TestCheckedViewRanked6ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes) plus two pinned frontiers in TestCheckedViewRanked6Frontiers.
Mutants: five execution catches (undefined element check in both backends, its native name, checked join in both backends).
Limits: the `ClassDeclaration | ClassExpression` cast is lane 4's union-target admission and is not credited; JSDocArray production with its own jsDocCache field is a loud NotYet; pushing a JSDoc record stays a named runtime refusal.

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| ResolvedType.callSignatures | 22 | `callSignatures: readonly Signature[]` | ranked6-calls*.a | yes |
| TemplateLiteralType.texts | 22 | `texts: readonly string[]` | ranked6-template*.a | yes |
| ClassDeclaration \| ClassExpression.members | 21 | `readonly members: NodeArray<ClassElement>` (ClassLikeDeclarationBase), cast as `parent as ClassDeclaration \| ClassExpression` | ranked6-class-like*.a | no |
| TupleType.labeledElementDeclarations | 21 | `labeledElementDeclarations?: readonly (NamedTupleMember \| ParameterDeclaration \| undefined)[]` | ranked6-labels*.a, ranked6-element-flags.a | yes |
| ClassDeclaration.members | 19 | `readonly members: NodeArray<ClassElement>` | ranked6-class-members*.a | yes |
| HasJSDoc.jsDoc | 19 | `jsDoc?: JSDocArray` (JSDocContainer), `JSDocArray extends Array<JSDoc> { jsDocCache?: readonly JSDocTag[] }` | ranked6-jsdoc-parent*.a | yes |
| JsxAttributes.properties | 19 | `readonly properties: NodeArray<JsxAttributeLike>` | ranked6-jsx*.a | yes |

Declarations were read from microsoft/TypeScript 050880ce, src/compiler/types.ts; the
two union casts were read from src/compiler/checker.ts (lines 38934 and 39030,
`const node = parent as ClassDeclaration | ClassExpression`). Roots are modeled with
synthetic `kind` tags. HasJSDoc is tsc's union of about ninety declarations; the
fixtures use three members (ClassDeclaration, ClassExpression, FunctionDeclaration)
and reach it as the census does, through `JSDoc.parent`. TemplateLiteralType and
ResolvedType keep tsc's mutable fields holding readonly arrays.

What the probes cover beyond the fifth group: a field whose type is an object union
(`JSDoc.parent: HasJSDoc`) read and narrowed before its array is read; an element
type that admits undefined (`labeledElementDeclarations` with a hole) next to one
that doesn't (`JsxAttributes.properties` holding undefined refuses at that element's
read); a field-level literal union on a non-union element (`ClassElement.kind:
'property' | 'method'`), which is checked only when read, so `class-members-wrong-tag`
prints its first line and refuses at the kind read, as lazy checking promises;
`readonly number[]` reduce; string join and map over `readonly string[]`; and
whole-array assignment of a `readonly string[]` field, including one that breaks the
original alias's element contract and is refused at that alias's next read.

Defect found and fixed:

- `internal/native/runtime/view_arrays.c`: a present NULL reference in a reference
  array (undefined; null is the adamic_null object) refused with `found nullish`
  natively and `found undefined` in JavaScript. The earlier pin `array-undefined`
  (TestCheckedViewArrays) and JavaScript both say `undefined`, so native now does.
  Both backends exited 70 with the same stdout before; only stderr differed.
  Mutant: the old name, caught by ranked6-jsx-hole.

Frontiers pinned exactly, before lowering, never compiled (Node runs both):

- ranked6-class-like-union-cast: `node as ClassDeclaration | ClassExpression` from
  Base is refused by cast_proof.go. The read is held through member casts and a
  parameter of the union type (`memberNames`, after tsc's checkUnusedClassMembers).
- ranked6-jsdoc-parent-cache: `Object.assign([doc], {jsDocCache: [tag]})` is
  `stage 0 can't lower array own-field production with unsupported field jsDocCache
  yet`. tsc's parser stores a plain JSDoc[]; jsDocCache is written later, an own
  array field pair outside this group.

Named runtime refusal pinned in both backends (Node succeeds, Adamic exits 70):
ranked6-jsdoc-parent-push pushes the viewed JSDoc into its owner's jsDoc array and
refuses with `uncertified incoming record contract`. JSDoc carries `parent: HasJSDoc`
and an optional comment, outside the flat scalar reference-write subset.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-undefined-element (undefined admitted) | ranked6-jsx-hole | exit 0 printing `true` |
| native-undefined-element-name (old `nullish`) | ranked6-jsx-hole | stderr differs |
| javascript-undefined-element (undefined admitted) | ranked6-jsx-hole | exit 0 printing `true` |
| native-join-elements (emitter reads join elements unchecked) | ranked6-template-join-wrong | ASan SEGV |
| javascript-join-elements (join skips the element check) | ranked6-template-join-wrong | exit 0 printing `1\|2` |

A first native join mutant, disabling the checked path inside the runtime's
adamic_view_array_string, survived: string joins check each element in emitted code
(view_arrays.go) before adamic_array_join and never reach that runtime path. The
mutant was moved to the emitter, where it is caught. Which programs reach
adamic_view_array_string's checked path is not established by this group. The other
checks these probes rest on (array kind, element kind, object union tags, nullable
member kinds) were proven by the fifth group's mutants at the same sites.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 33 / 1,666 | 301 / 1,523 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

Overlapping candidate obligations from read-census/summary.json; tuples are lane 4c's.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Arrays|NativeArrays|NodeArrayRecords|ArraySearch|RankedArrayContracts|RankedParserArrayContracts|RankedArrayUnionContracts|MutableArrayUnionRefusal|OptionalDeclarations|GenericArrayCast|ArrayReferenceWrites|Ranked4ArrayContracts|Ranked5ArrayContracts|Ranked5UnionCastFrontier|Ranked6ArrayContracts|Ranked6Frontiers)$' -count=1 -v -timeout 20m
go test ./internal/native ./internal/javascript -count=1 -timeout 30m
python3 stage3/interface-downcasts/lane2/run-ranked6-array-mutants.py
```

The lane 2 oracle passes (248 PASS lines, 32.2s). javascript passes. native fails
only the five inherited graph tests named in RANKED4_ARRAYS_REPORT.md
(TestGraphRegionsMillion, TestGraphLazyRegions, TestGraphClosureEnvironment,
TestGraphContainerBoundary, TestGraphRegionsRuntime). gofmt is clean. The full
repository gate was not run.
