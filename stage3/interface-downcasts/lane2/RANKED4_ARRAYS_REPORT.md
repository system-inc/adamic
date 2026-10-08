Held the next five ranked array contracts to Node in both backends and fixed what they exposed.
Commit: follows merge 8a0cef02 (views-integration 6a7f1bf3); the pushed SHA accompanies the handoff.
Validation: 33 probes in TestCheckedViewRanked4ArrayContracts, sanitized/release native, JavaScript, leak checks on every program that finishes.
Mutants: six execution catches (graph object size, write type name, record write pairs in both backends, scalar write storage in both backends) plus one lowering crash witness.
Limits: array writes into a slot that held undefined, and pushes of records outside the flat scalar subset, stay named runtime refusals; nine lower and five native graph tests fail identically on the integration tip.

| Candidate pair | Reads | tsc 6.0.3 declaration | Fixtures |
| --- | ---: | --- | --- |
| CaseBlock.clauses | 26 | `readonly clauses: NodeArray<CaseOrDefaultClause>` | ranked4-case*.a |
| ParsedCommandLine.fileNames | 26 | `fileNames: string[]` | ranked4-files*.a |
| DiagnosticMessageChain.next | 26 | `next?: DiagnosticMessageChain[]` | ranked4-chain*.a |
| FlowLabel.antecedent | 25 | `antecedent: FlowNode[] \| undefined` | ranked4-label*.a |
| CommaListExpression.elements | 25 | `readonly elements: NodeArray<Expression>` | ranked4-comma*.a |

Declarations were read from microsoft/TypeScript at the census source commit
050880ce (src/compiler/types.ts). The census reaches these pairs by a FlowNode
to FlowLabel cast, a Node to CaseBlock cast, a type predicate on
`ParsedCommandLine | Diagnostic`, and as union members (`string |
DiagnosticMessageChain` and the NamedExports/TupleTypeNode/.../CommaListExpression
union). The fixtures model each root with a synthetic `kind` tag, as earlier
batches did; FlowNode is modeled as `FlowStart | FlowLabel`, tagged, while tsc's
FlowNode members share an untagged `flags: FlowFlags`. Mutable arrays, recursive
elements and a required `| undefined` array are new in this group.

What the fixtures cover: array kind, union element tags (CaseOrDefaultClause),
a nested element field (CaseClause.expression), lazy later elements, element
kind for `string[]` reads and join, map callbacks over element fields, the
recursive chain read two levels deep, absent and present-undefined optional
arrays, and writes through the mutable views: push, index set and whole-array
assignment for `string[]` and DiagnosticMessageChain[], checked against the
original array's element contract (a number-holding array refuses a string
push; a chain array whose original messageText was number refuses a well-formed
chain).

Defects found and fixed:

- `internal/lower/graph_flow.go`: an array literal stored in a union-typed slot
  (`antecedent: [start]` in a `FlowNode | FlowNode[] | undefined` field) called
  GetTypeArguments on the union and the compiler panicked (nil dereference in
  cycleFinder.graphFlows) once the program had a push. The new
  `arrayLiteralElement` joins the element types of the slot's array members.
  Crash witness: restoring the old call panics again on ranked4-label-push. The
  join itself has no witness: a mutant that seeds no elements for union slots
  passes every ranked4 fixture, and a separate cycle probe (outer.items =
  [inner], inner.peer = outer) compiles to byte-identical C with and without it,
  because the union field is already seeded whole. It is kept because it only
  adds seeds, the conservative direction for graph adoption.
- `internal/native/graph_regions.go`: graph adoption of an object literal copied
  `count * (sizeof(adamic_value) + 2)` bytes after the header. Objects now carry
  an aligned `unsigned int` contract per slot after those bytes, so the copy was
  short and the literal's contract writes overflowed the new storage. ASan
  reported a 4-byte heap-buffer-overflow in 12 of 33 fixtures, any program with a
  graph type; release builds printed the right answers. Adoption now uses
  `adamic_object_size`. Mutant: the old size, caught by ranked4-comma under ASan.
- `internal/native/emit_statements.go`: the checked view write named only scalar
  types, so an array refused into a slot holding undefined said `is not a ;
  expected , found nullish` natively while JavaScript said `array`. Native now
  uses the JavaScript backend's names. Mutant: the old map, caught by
  ranked4-label-assign.

Merged codex/views-integration 6a7f1bf3 for lane 1's nullish member reads.
Before it, `FlowNode[] | undefined` (no question mark) holding undefined was
refused at the read (`found nullish`) while Node printed `absent`. The only
conflict was in docs/checked-views-plan.md; both sections are kept. The merge
changed two earlier pins from `is not a` to `matches no member of` for nullable
arrays; all three backends print the new text, and the pins were updated.

Named refusals pinned in both backends (Node succeeds; Adamic exits 70):

- ranked4-label-assign: assigning an array through FlowLabel into a slot that
  was made holding undefined. The original slot is `FlowNode | FlowNode[] |
  undefined`, a mixed union stage 0 can't store; the write refuses rather than
  change the slot's representation.
- ranked4-label-push: pushing a FlowStart into antecedent refuses with
  `uncertified source element contract`. FlowNode records (unknown `node`,
  nested antecedent) are outside the flat scalar reference-write subset.

| Mutant | Witness |
| --- | --- |
| native-graph-object-size | ranked4-comma: ASan heap-buffer-overflow |
| native-write-type-name | ranked4-label-assign: empty type names in the diagnostic |
| native-record-write-pairs | ranked4-chain-set-wrong: write admitted, later read fails with a different diagnostic |
| javascript-record-write-pairs | ranked4-chain-set-wrong: same |
| native-scalar-write-storage | ranked4-files-push-wrong: string stored in a number array, exit and output differ |
| javascript-scalar-write-storage | ranked4-files-set-wrong: write admitted, later read fails with a different diagnostic |
| cycle-union-array-literal (crash witness) | ranked4-label-push: lowering panics again |

Every mutant was executed and restored in finally. Build failure, clang failure
or frontend rejection is not accepted as a runtime mutant witness; the crash
witness is listed apart. A map_set.c iterator adoption has the same short object
size; no witness could reach it (a cyclic Map probe first hits the inherited
adamic_map_set use-after-free), so it is reported, not changed.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 22 / 1,428 | 312 / 1,761 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 |

These are overlapping candidate obligations from read-census/summary.json.
Representative probes receive no production pair credit for element, consumer
or intrinsic reads.

Inherited failures, identical on views-integration 6a7f1bf3 (logs in
ranked4-array-logs/): native TestGraphRegionsMillion, TestGraphClosureEnvironment
(clang: closure code signature), TestGraphLazyRegions, TestGraphRegionsRuntime
(leak), TestGraphContainerBoundary (use-after-free in adamic_map_set); lower
TestCensusMarkerResultIsAssignable, TestCensusMarkerZeroCallIsNotAssumedSafe,
TestOverloadedShorthandFunctionValueStaysNotYet,
TestCensusPredicateMarkerKeepsProofBoundaries, TestCensusOverloadRelation,
TestPhantomArrayCastsAreErased, TestPhantomArrayRequiredCastsAreErased,
TestNestedFunctionGapsAreLoud, TestPhantomArrayProofs (missing
adamic/cycle-capable refusal). The lower package also needs `npm ci` in
stage3/api for @types/node; cloud/setup.sh does not do it.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Arrays|NativeArrays|NodeArrayRecords|ArraySearch|RankedArrayContracts|RankedParserArrayContracts|RankedArrayUnionContracts|MutableArrayUnionRefusal|OptionalDeclarations|GenericArrayCast|ArrayReferenceWrites|Ranked4ArrayContracts)$' -count=1 -timeout 20m
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -count=1 -timeout 20m
python3 stage3/interface-downcasts/lane2/run-ranked4-array-mutants.py
```

Combined lane 2 array oracle: PASS, 24.6s. ir and javascript pass; lower and
native fail only with the inherited tests above. Full repository gate was not
run. The next ranked contracts are ResolvedType.constructSignatures (25 reads),
ClassDeclaration.modifiers, TupleTypeNode.elements and
ExpressionWithTypeArguments.typeArguments (23 each), JSDoc.tags and
FunctionLikeDeclaration.parameters (22 each).
