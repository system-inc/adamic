Held the next five most-read remaining array contracts to Node in sanitized/release native and JavaScript.
Commit: follows e540f4d7de11090fd82053f1671ed2ab5df2063f; pushed SHA accompanies the handoff.
Validation: fourteen probes, including five named wrong-array refusals, lazy object reads, open numeric enums and narrower enum members.
Mutants: native/JavaScript array-kind, native/JavaScript enum-member checks and native eager element scanning; execution evidence retained.
Limits: candidate fixture obligations only; complete production TypeScript schemas and exact cast-site reachability are unmeasured; estimate remains October 10, 12:00 MDT.

| Candidate pair | Reads | Fixture |
| --- | ---: | --- |
| VariableDeclarationList.declarations | 63 | ranked2-variables.a |
| CallExpression.arguments | 62 | ranked2-call.a |
| ObjectLiteralExpression.properties | 49 | ranked2-object.a |
| TupleType.elementFlags | 43 | ranked2-tuple.a |
| ArrayLiteralExpression.elements | 42 | ranked2-array.a |

This group holds 259 additional candidate read obligations for array field
contracts. The NodeArray fixtures retain real scalar own-field declarations and
use minimal element interfaces exposing numeric pos. They do not reproduce all
VariableDeclaration, Expression or ObjectLiteralElementLike members/unions.
The tagged roots are representative synthetic admissions. This is existing
compiler support held to source oracles, not a claim that full tsc lowers.
TupleType.elementFlags is an ordinary array field; it does not implement tuple
position/optional/rest contracts.

| Family | Candidate pairs / reads | Cumulative ranked obligations held | Candidate remainder |
| --- | ---: | ---: | ---: |
| Array contracts | 334 / 3,189 | 12 / 1,143 | 322 / 2,046 |
| Element or consumer reads | 251 / 1,602 | 0 / 0 | 251 / 1,602 |
| Intrinsic contracts | 179 / 794 | 0 / 0 | 179 / 794 |
| Own array fields | 30 / 72 | 3 / 26 | 27 / 46 |
| Tuples | 6 / 9 | 0 / 0 | 6 / 9 |

Families overlap. The frozen candidate inventory in lazy-array-priority.json
records each held pair; none of these numbers claims an unlocked count among
the 2,936 production downcasts. Exact reachability waits on checker-clean IR.
This fixture group introduces no new shared/compiler hook. The preceding
reference-write group's named hooks are listed under Lane 2 in the plan.

Numeric enums follow the existing enums.go doctrine: the whole numeric
declaration admits numbers outside its named members. A member-specific type
is narrower. Initial logs retain the mistaken expectation that 7 must fail an
open enum read; that oracle was corrected after reading the existing doctrine.
The final positive open-enum probe prints 7. A string element instead fails
with expected ElementFlags/found string; Optional read as Required fails with
expected ElementFlags.Required/found number 2. Both failures are exit 70,
with their exact messages pinned in both backends. No finite-members-only
rule was introduced for whole numeric enums.

The lazy probe has a later malformed pos field. Array field and length reads
and the first valid element remain successful. Its eager-scanning mutant
fails during execution, rather than being accepted as a compiler rejection.
All successful native controls also pass leaksUncached.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewRankedParserArrayContracts$' -count=1 -timeout 10m
python3 stage3/interface-downcasts/lane2/run-ranked-parser-array-mutants.py
```

Reference writes remain the flat required scalar-record subset from e540f4d7.
The next candidate contracts are BindingPattern.elements (35 reads),
UnionOrIntersectionType.types (35), SignatureDeclaration.parameters (31),
and additional expression/statement NodeArray fields. Callable files are unchanged.

Final restored combined array oracle gate: PASS, 21.230s. All five mutants
are caught during execution; native skipped kind checking causes ASan SEGV,
JavaScript prints undefined, both skipped enum-member guards print 2, and
eager scanning exits 70 on the untouched later pos field. No production source
remains mutated. Raw logs are in ranked-parser-array-logs/. The full repository
gate was not rerun for this fixture-only group; the preceding compiler group
ran the affected package selection, whose final logs are retained.
