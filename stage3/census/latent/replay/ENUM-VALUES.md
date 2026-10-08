# enumNeverValue unit

Added one concrete contextual specialization of a generic function value.
Receiver blocker and replay merge remain 64246ed0 and 1e8413ac.
Focused Node differential and boundary tests pass; counts adds one row.
Six independent mutants failed; the identity mutant disagreed with Node in both backends.
No verified table-site unlock is claimed; multi-specialization values remain NotYet.

The new helper is separate from generic.go and expression.go's only edit replaces
its generic-value rejection with helper dispatch. No other expression.go function
changed. The implementation infers binders from a single concrete contextual
signature, proves exact instantiated parameters and result, then uses existing
lowerFunction and functionValue calling/ownership conventions. It rechecks
instantiated mutations using the existing proof. An unresolved phantom binder,
polymorphic context, differing arity or signature, and a second distinct value
specialization remain NotYet. The latter prevents different forwarders from
changing source function identity. Direct generic calls keep their existing path.

The generic_function_value.a fixture uses a defaulted generic callback, explicit
number and string contexts, repeated function identity, and a runtime-built
string. Its own generic_function_value_test.go registers it. The source Node,
JavaScript backend, sanitized native (ASan/UBSan and leak check), and release
native all agree on stdout: "7 11 19 true\nAdamic\n", exit 0.

## Validation

Every test invocation wrote to a log. Commands sourced
/workspace/adamic-tools/env.sh. Final focused commands were:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestGenericFunctionValueBoundaries|TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_value.a' -count=1 -timeout 10m > /tmp/notyet-generic-final.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/notyet-generic-counts.log 2>&1
go vet ./internal/lower > /tmp/notyet-generic-vet.log 2>&1
```

Focused tests: lower 0.659s, oracle 0.473s, both exit 0. The initial oracle
run took 12.076s and also passed. Counts exits 0, 27.718s, adds only the new
fixture row, preserving every existing row. Formatting and git diff --check
print nothing. No whole package tests or full gate ran.

The temporary mutant runner /tmp/notyet-generic-mutants.py restores the source
in a finally block. ADAMIC_GATE_UNCACHED=1 python3 /tmp/notyet-generic-mutants.py
wrote /tmp/notyet-generic-mutants-final.log and individual mutant logs:

| Mutant | Catcher | Outcome |
| --- | --- | --- |
| Remove specialization reuse | Node comparison in fixture | Both backends print false instead of true; exit 1 |
| Allow multiple value specializations | boundaries/identity | Lowering incorrectly succeeds; exit 1 |
| Drop exact parameter proof | boundaries/parameter | Lowering incorrectly succeeds; exit 1 |
| Drop exact result proof | boundaries/result | Lowering incorrectly succeeds; exit 1 |
| Accept extra contextual parameters | boundaries/arity | Lowering incorrectly succeeds; exit 1 |
| Infer an unused phantom binder from another binder | boundaries/phantom | Lowering incorrectly succeeds; exit 1 |

The parameter witness uses two compatible object types so a downstream slotless
refusal cannot mask it. Its earlier union-parameter version reached a subsequent
NotYet and was replaced. All final mutants failed without build errors or panics.
Nil/malformed checker inputs and the existing recursion-depth guard were not
independently mutated. The polymorphic-context test proves the refusal boundary;
its guard is also backed by the unresolved-binder gate, not an independent kill.

## Replay and table accounting

The table pin is 57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7.
Its roots/raw.csv has 22 generic-value observations, deduplicating exact
(kind, where, reason, text) to the requested 17 root sites.
Both requested replays ran before and after this change with the exact original
reason using the command in THIS-OUTSIDE.md:

- core.ts:220:112 selects contains at 220:1 and stops earlier at 220:62,
  NotYet: a value of type T. Replay exit 1 before and after.
- core.ts:780:53 selects insertSorted at 768:1 and stops earlier at 770:5,
  NotYet: a value of type T. Replay exit 1 before and after.

This does not prove an unlock in those units or the census's instantiated
ancestor contexts. Verified coverage count is 0/17; fixture-level support is
proven separately. Logs are /tmp/notyet-generic-first.log,
/tmp/notyet-generic-second.log and /tmp/notyet-generic-{first,second}-after.log.

The next reasons' baseline replays reproduced the exact narrowed-object-union
stop at checker.ts:37210:16 and tsbuildPublic.ts:265:50, both exit 0.
The former also reports a dependency's unknown-type stop at core.ts:1750:25.
The fs example at tracing.ts:95:19 fails to reproduce (exit 1), stopping earlier
at tracing.ts:41:9 with NotYet: a value of type any. This base contains no
production node:fs.openSync lowering. These baseline logs are
/tmp/notyet-union-{first,second}-before.log and /tmp/notyet-fs-before.log.
