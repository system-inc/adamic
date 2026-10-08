# ecmascript/decorators syntax helpers

Claim: `b9fa10f4b945fb05e48410b8e90200fd530285b4`, triage rank 25.

| Go symbol | Helper |
| --- | --- |
| CallName | call_name.a |
| Of | of.a |
| HasDecoratorInSet | has_decorator_in_set.a |

`nodes.a` preserves Go node identity and nil decorator lists; `projection.a`
adapts parser nodes for consuming rules. Each contract has its own file.
`main.a` replays captured inputs, serializing names by UTF-16 code unit and
lists by identity, including nil versus an empty list.

The exact oracle is cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`.
`testdata/capture.py` overlays transparent wrappers around unchanged Go helper
bodies, then runs every upstream test in base and core, the complete live
consumer packages. Actual calls include calls nested inside Go helpers.
`testdata/coverage.json` pins the source hash and records counts; compressed
fixtures contain the original Go AST facts, calls and returned bytes.
Live captures: CallName 301, Of 510, HasDecoratorInSet 172 (983 total).
Controls: CallName 73, Of 132, HasDecoratorInSet 66 (271 total).
All 1,254 observations are compared on each of the three backends.
The proof rule capture contains 54 unique upstream cases.
The recorder also captures typed harness cases. Controls cover nil, nondecorators,
bare decorators, qualified and parenthesized calls, type arguments, Unicode,
modifier order and membership including an empty string.

`helpers_test.go` checks every fixture on source Node, emitted JavaScript and
ASan/UBSan native. Each of the three mutants compiles and runs on all backends,
then must differ from the Go answer; crashes or compilation failures do not count.

Inspected partial bundles on `codex/typeaware-wave-25` (`decorator_shape.a`),
`-26` (`decorator_scope.a`) and `-27` (`parity_decorators.a`). Their direct-child
filtering and identifier guards inform the projection, but none retained the
three exact shared contracts. No partial helper implementation was copied.
The proof rule adapts `codex/lint-wave1-11` and uses the existing shared imports
package for BindingsOf; it introduces no unported helper dependency.

The frozen syntax blocker cohort is these three helpers. IsNullableType and
BooleanOption are checker/parity helpers outside that cohort and remain unported
here. No runtime regex compiler or unsupported language construct is required.

Run with the repository tool environment:

    python3 stage1/cohere/lint/helpers/decorators/testdata/capture.py
    go test ./stage1/cohere/lint/helpers/decorators -count=1 -v
    python3 stage1/cohere/lint/rules/base-security-require-context-access/testdata/run_selected.py

Then run every helpers package and the full lint package, setting
ADAMIC_TYPESCRIPT_SOURCE to the pinned compiler checkout for its corpus gate.

The finished unit includes the 54/54 green proof rule and merges parser recovery
from `origin/lint-batch/wave2-03` (`86c5d841`, merge `67c4c034a`). All shared helper
packages passed their fixture and mutant suites before the merge; all decorator
fixtures and mutants passed again afterward. The rule's owned runner documents
the merged batch's temporary unused-import overlay. No parser patch is included.
