Held the fifth ranked group of array contracts to Node in both backends; no compiler change was needed.
Commit: follows 5c609c01 on codex/views-arrays-callables-parser; the pushed SHA accompanies the handoff.
Validation: 40 probes in TestCheckedViewRanked5ArrayContracts (sanitized and release native, JavaScript, leak checks on every program that finishes) plus TestCheckedViewRanked5UnionCastFrontier.
Mutants: eight execution catches (array kind, element kind, union element tag and nullable member kinds, each in both backends).
Limits: a direct cast to the FunctionLikeDeclaration union stays a frontend refusal (lane 4's union-target admission), so that pair is not credited; tuples moved to lane 4c and leave this report's counts.

| Candidate pair | Reads | tsc declaration (050880ce) | Fixtures | Credited |
| --- | ---: | --- | --- | --- |
| ResolvedType.constructSignatures | 25 | `constructSignatures: readonly Signature[]` | ranked5-signatures*.a | yes |
| ClassDeclaration.modifiers | 23 | `readonly modifiers?: NodeArray<ModifierLike>` | ranked5-modifiers*.a | yes |
| ExpressionWithTypeArguments.typeArguments | 23 | `readonly typeArguments?: NodeArray<TypeNode>` (from NodeWithTypeArguments) | ranked5-heritage*.a | yes |
| TupleTypeNode.elements | 23 | `readonly elements: NodeArray<TypeNode \| NamedTupleMember>` | ranked5-tuple*.a | yes |
| JSDoc.tags | 22 | `readonly tags?: NodeArray<JSDocTag>` | ranked5-jsdoc*.a | yes |
| FunctionLikeDeclaration.parameters | 22 | `readonly parameters: NodeArray<ParameterDeclaration>` (SignatureDeclarationBase), read through the seven-member union | ranked5-function*.a | no |

Declarations were read from microsoft/TypeScript at the census source commit
050880ce59e30b356b686bd3144efe24f875ebc8, src/compiler/types.ts. TupleTypeNode.elements
is an array contract in the census (its family is "array contracts", not "tuples"),
so it stays in lane 2 after the tuple family moved to lane 4c.

The fixtures model each root with a synthetic `kind` tag, as earlier batches did.
Signature keeps tsc's mutable fields (`declaration?`, `parameters: readonly Symbol[]`,
`minArgumentCount`) and has no tag; ResolvedType keeps `callSignatures` and
`constructSignatures` as mutable fields holding readonly arrays, so whole-array
assignment through the view is legal, as in tsc's setStructuredTypeMembers.
ModifierLike is `ModifierToken | Decorator` with tags; NamedTupleMember is narrowed
by a type predicate, as tsc's isNamedTupleMember does. JSDocTag.comment is modeled
as `string | undefined`; tsc's `string | NodeArray<JSDocComment>` is a mixed union
outside lane 2. FunctionLikeDeclaration is a union of three of tsc's seven members.

What the probes cover: array kind (a number, a string and the array-like object
`{length: 1}`), element kind (a number or string where a node is expected), union
element tags (`static` is not a ModifierLike), nested element fields (decorator
expression text, named tuple member name, parameter name, signature parameter
escapedName two arrays deep), lazy later elements (an unread malformed second
signature and type argument), a malformed element reached after a good one
(heritage-wrong-late prints `5` first), absent optional arrays, an optional
declaration of union type inside an array element, intrinsics over viewed
elements (map, filter, some, join, for...of), and whole-array assignment through
the view: an empty array, a record array, and two writes whose new elements break
the original alias's element contract. Those last two are admitted at the write
and refused at the alias's next read, in all three backends, rather than read in
the wrong representation.

FunctionLikeDeclaration.parameters: the census reaches it through
`(node as FunctionLikeDeclaration)` in utilities.ts and checker.ts. Stage 0 refuses
a cast whose target is a union of interfaces from a non-union source
(cast_proof.go); routing union targets to the shared view entry is lane 4's
handoff in docs/checked-views-plan.md. The read itself is held: member casts
(FunctionDeclaration, MethodDeclaration, ArrowFunction) upcast to the union and
read `parameters` through it, including a mixed `FunctionLikeDeclaration[]`. The
direct cast is pinned as a named frontend refusal by
TestCheckedViewRanked5UnionCastFrontier, never compiled. The pair receives no
credit until that admission exists.

| Mutant | Witness | How it failed |
| --- | --- | --- |
| native-array-kind (object.c returns any slot for wanted 5) | ranked5-signatures-wrong-array | ASan SEGV reading a number as an array |
| javascript-array-kind (any object passes the array check) | ranked5-tuple-wrong-array | exit 0 printing `1` for the array-like object |
| native-element-kind (view_arrays.c skips the element kind) | ranked5-tuple-wrong-element | UBSan misaligned shape access |
| javascript-element-kind (element check never panics) | ranked5-jsdoc-wrong-element | a different, later diagnostic |
| native-union-element-tag (tag test disabled) | ranked5-modifiers-wrong-tag | exit 0 printing `1` |
| javascript-union-element-tag (allowed-literal check disabled) | ranked5-modifiers-wrong-tag | exit 0 printing `1` |
| native-nullable-member-kinds (member kinds unchecked) | ranked5-heritage-wrong-array | exit 0 printing `4611686018427388000`, a number's bits read as a length |
| javascript-nullable-member-kinds (member kinds unchecked) | ranked5-modifiers-wrong-array | exit 0 printing `6`, a string's length |

Every mutant was executed by run-ranked5-array-mutants.py and restored in a
finally block; the script rejects a build, clang, syntax or lowering failure as a
witness. The native-nullable-member-kinds result is the silent miscompile that
check exists to prevent.

Observed, not fixed (lane 1's readiness code, outside this family): for null in a
required field whose source slot is typed `null`, native says `found null` and
JavaScript says `found nullish`, for scalar, object and array fields alike. Repro:
`interface A extends Base { readonly kind: 'a'; readonly n: number; }` read through
a cast from `{kind: 'a', n: null}`. Native's object.c names kind 12 `null`;
JavaScript's adamicViewField says `null` only for absent or undefined-member
reads. Both exit 70 with the same stdout. These fixtures use numbers instead of
null so the oracle pins only agreeing text.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 27 / 1,544 | 307 / 1,645 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |

These are overlapping candidate obligations from read-census/summary.json. The
tuple family (6 pairs / 9 reads) moved to lane 4c and is no longer counted here.
Representative probes receive no production pair credit for element, consumer or
intrinsic reads.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Arrays|NativeArrays|NodeArrayRecords|ArraySearch|RankedArrayContracts|RankedParserArrayContracts|RankedArrayUnionContracts|MutableArrayUnionRefusal|OptionalDeclarations|GenericArrayCast|ArrayReferenceWrites|Ranked4ArrayContracts|Ranked5ArrayContracts|Ranked5UnionCastFrontier)$|^TestCountsAreRecorded$' -count=1 -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked5-array-mutants.py
gofmt -l cmd internal; go vet ./internal/oracle
```

Every listed checked view test passes (55.6s). TestCountsAreRecorded fails on
eight graph_regions fixtures (`free(): invalid pointer` in the counted build); the
same eight fail identically at e4535b0e, before the fourth ranked group, so they
are inherited. gofmt is clean and vet passes. The full repository gate was not run;
no compiler, runtime or backend file changed in this group.
