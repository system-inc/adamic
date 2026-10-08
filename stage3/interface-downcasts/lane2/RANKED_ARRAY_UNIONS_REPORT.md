Built readonly array-union contracts and held the next five ranked array field pairs to Node in both backends.
Commit: follows 69c398621c76738e9482b087c94375508997b0a3; the pushed SHA accompanies the handoff.
Validation: fifteen runtime probes plus one exact frontend refusal, sanitized/release native, JavaScript and positive native leak checks.
Mutants: nine execution catches for array kind, element kind/tag, union element merging, long diagnostics and eager scanning.
Limits: mutable array unions stay named lazy read refusals; full tsc schemas and exact production reachability remain unmeasured; date remains October 10, 12:00 MDT.

| Candidate pair | Reads | Fixture |
| --- | ---: | --- |
| BindingPattern.elements | 35 | ranked3-pattern.a |
| UnionOrIntersectionType.types | 35 | ranked3-types.a |
| SignatureDeclaration.parameters | 31 | ranked3-signature.a |
| ClassLikeDeclaration.members | 29 | ranked3-class.a |
| TypeReferenceNode.typeArguments | 27 | ranked3-reference.a |

This adds 157 candidate read obligations. The first pair previously refused
array use, and the JavaScript wrong-array path panicked during emission because
an array union reached the tagged object-union emitter. Readonly alternatives
now share array storage recognition while retaining the union of their declared
element types. Array field reads check presence/readiness/array kind and do not
scan elements. Selected elements use existing kind and tagged object-union
checks, then scalar payload reads remain checked. Dropping all but the first
element contract incorrectly accepts an invalid element tag; the executed
union-element-merge mutant detects this.

Mutable array unions are unsupported metadata until their field is read.
`ranked3-mutable-array-unread.a` prints pattern in all backends. Node prints 1
for the read fixture; lowering instead refuses at 23:18, naming field elements
and unsupported mutable array union contract. The diagnostic and adaptation
text are pinned exactly. This is a frontend policy refusal, not a claimed
runtime mutant witness or newly implemented mutable-union write contract.

The long BindingPattern array-union type name exposed a second bug: native
field-kind diagnostics allocated room for one expected-type name but printed
it twice. snprintf returned the full length, and panic read beyond the allocated
buffer. Native now budgets both occurrences. The wrong-array fixture pins the
same intended exit-70 diagnostic in native and JavaScript. Restoring the old
allocation size produces ASan heap-buffer-overflow, caught by that fixture.

Other probes cover optional typeArguments present/absent, the omitted binding
alternative, malformed object/number element, invalid element tag, and a later
invalid tag/pos left untouched while reading length and the first element.
Minimal interfaces model scalar pos/id payloads and synthetic tagged roots;
they do not reproduce every production member or TypeScript's numeric SyntaxKind
hierarchy. This holds the candidate array-field contract family, not the complete
production schemas or the 2,936 cast sites. The element/consumer ledger receives
no production pair credit from these representative probes.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 17 / 1,300 | 317 / 1,889 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 |

These are overlapping candidate obligations. Exact reachability needs checker-clean
production IR. New named shared hooks are listed under Lane 2 in the plan.
No callable file, IR node or native metadata field was added in this batch.
Reference writes remain the flat required scalar-record subset from e540f4d7.

| Mutant | Witness |
| --- | --- |
| native-array-kind | wrong-array value causes ASan invalid reference access |
| javascript-array-kind | wrong-array read prints undefined |
| native-element-kind | numeric element followed as object causes ASan failure |
| javascript-element-kind | skips the named element failure and reaches a different field failure |
| native-element-tag | invalid tag prints 1 |
| javascript-element-tag | invalid tag prints 1 |
| union-element-merge | first-only element contract accepts invalid tag and prints 1 |
| native-long-diagnostic | restored undersized message causes ASan buffer overflow |
| native-eager-elements | reading the array scans the untouched later pos and exits 70 |

Every listed mutant was executed and restored in finally. A first mutation-script
attempt stopped on an ambiguous JavaScript selector after restoring its completed
mutants; the corrected selector changes only the indexed-read hook. Build failure,
compiler panic or frontend rejection is not accepted as mutant evidence.
Raw logs retain initial refusals, the diagnostic overflow, pin development,
the selector correction and final results.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedView(Arrays|NativeArrays|NodeArrayRecords|ArraySearch|RankedArrayContracts|RankedParserArrayContracts|RankedArrayUnionContracts|MutableArrayUnionRefusal|OptionalDeclarations|GenericArrayCast|ArrayReferenceWrites)$' -count=1 -timeout 15m
go test ./internal/ir ./internal/lower ./internal/native ./internal/javascript -run 'Test.*View|TestLazyView|TestSharedArrayContractAdapter' -count=1 -timeout 15m
python3 stage3/interface-downcasts/lane2/run-ranked-array-union-mutants.py
```

The next ranked contracts are CaseBlock.clauses, ParsedCommandLine.fileNames,
DiagnosticMessageChain.next (26 reads each), FlowLabel.antecedent and
CommaListExpression.elements (25 each). Full repository gate was not run.

Final restored combined array oracle gate: PASS, 34.916s. Affected package
selection passes: lowering 12.561s, native 29.912s, JavaScript 2.501s; IR builds
with no matching tests. The nine-mutant script completes successfully. No
production file remains mutated. Evidence is in ranked-array-union-logs/.
