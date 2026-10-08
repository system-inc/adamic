Held the five highest-read remaining array contracts to Node, sanitized/release native, emitted JavaScript and leak checks.
Commits: follows 3b1263b650b374183da1cd879b29a73dfb10135e; the new pushed SHA is in the handoff.
Validation: twelve probes; positive controls, optional absence, lazy reads, and five exact wrong-value refusals.
Mutants: native/JavaScript array kind, JavaScript descendant kind, and native eager scanning are caught during execution.
Limits: representative candidate obligations only; reference writes remain next; estimate October 10, 12:00 MDT.

The date moved because the original estimate omitted the optional-array descriptor
failure after integration and the NodeArray ancestry, owned-property storage and
original-slot write certificate work. Integration and the environment restart added
validation time. This was an estimation error, not a new dependency on another lane.
The current rule permits lane-owned named shared hooks; no handoff is awaited.

This group adds source oracles for existing mechanisms, rather than claiming new
compiler support from a test alone. The next five unheld array candidates are:

| Candidate type / field | Reads | Representative fixture |
| --- | ---: | --- |
| UnionType.types | 107 | ranked-union-types.a |
| Signature.typeParameters | 90 | ranked-signature-type-parameters.a |
| SourceFile.statements | 72 | ranked-source-statements.a |
| Signature.parameters | 71 | ranked-signature-parameters.a |
| IntersectionType.types | 70 | ranked-intersection-types.a |

These 410 reads use small representative interfaces, not unchanged complete tsc
interfaces. Type, TypeParameter and Symbol carry representative scalar payloads;
NodeArray<Statement> uses the native own-property and instantiated ancestry hooks.
The source optional typeParameters property is also exercised absent. A later Type
with a malformed id is untouched when the first element and length are read.
The source Node output is pinned for every probe, including unsafe casts. Those
negative casts produce undefined or wrong on Node; checked implementations instead
stop with exit 70, naming the field/element and expected versus actual kind.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 7 / 884 | 327 / 2,305 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 |

Families overlap. This is a ranked candidate fixture ledger, not exact reachability,
not a statement that prior consumer mechanisms lack tests, and not a count of the
2,936 production cast sites unlocked. Checker-clean production IR remains unavailable.

The initial proposed fixture logs exposed string-only console typing; numeric
arguments were corrected to template strings, preserving Node output. That run is
recorded as a failure. The corrected first oracle passes in 2.939s. The four mutation
logs require executed output/refusal mismatch and reject build failures, compiler
panics, frontend NotYet, or JavaScript syntax failure as evidence. All production
files are restored in finally. Positive native runs also pass leaksUncached.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRankedArrayContracts$' -count=1 -timeout 10m
python3 stage3/interface-downcasts/lane2/run-ranked-array-mutants.py
```

No compiler hook was added in this fixture checkpoint. Existing hooks remain
listed under Lane 2 in docs/checked-views-plan.md; callables are unchanged. The
full repository gate was not rerun for fixture-only additions. The next array
contract is VariableDeclarationList.declarations, 63 reads. Reference writes
need the original element contract and incoming allocation certificate, independent
of the current pointer storage kind; the blanket refusal will not simply be removed.

Final restored twelve-probe oracle passes in 3.097s. No production file remains mutated.
