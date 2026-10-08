Built direct-write and fresh-construction presence guards and a recorded implements representation relation.
Commits: integration d335b9b1; direct-write batch 58e743bf; construction batch recorded in git history on codex/stricter-optional-writes.
Focused loader/lower/command tests pass; runtime Node/native/JavaScript proofs pass; production schedules 50, emits 0.
Numeric, reference and class absent-write mutants stop at the guard, exit 70; fresh-construction absence mutant stops at the guard; required-field absence mutant disagrees with Node.
Not covered: 17 other optional-site contracts, whole-program lowering, full repository gate, indexed/catch/JSON work.

The loader admits exactOptionalPropertyTypes-only TS2412 diagnostics for direct
property assignments and exact-only TS2420 implements relations. Ordinary project
errors and all other option errors remain errors. Sites with joint indexed/optional
attribution remain outside this unit. Pending contracts are never reported as
emitted checks. The production probe distinguishes scheduled contracts from checks
in a successfully lowered program.

A direct store holds its receiver once, evaluates the value, stores it using the
existing presence representation, then checks own presence. A missing descriptor
still stops loudly in the existing write machinery. Numeric, reference and class
fixtures observe in, keys, hasOwn and undefined reads against independent Node.
Separate absent-store mutants for each representation compile with sanitizers and
stop at the new guard with exit 70. The readiness pass preserves this guard.

A class implements relation is recorded after the instance storage has lowered.
There is no additional value check: its required undefined field is present,
exactly as ruled. A required-field absence mutant compiles and exits 0, producing
false for in and hasOwn; independent Node catches that disagreement. Interface
alias Object.keys currently refuses during lowering, so the fixture observes keys
and hasOwn through the same class instance, and in/read through its optional view.
A separate exact interface assignment remains an unconverted TS2375 relation.

`c`, `js` and build input paths accept `--explain-checks`; emitted checks are
printed to stderr after successful lowering. Types-only checking does not claim
that a guard was emitted.

Production input is generated with stage3/apply.sh from ledger revision
3f0926c0a55a7b5f64f037b1745e0e984e08c8be, followed by pinned upstream npm ci.
Using current adaptations instead changes 14 input files and yields 168 sites;
that tree cannot substantiate claims about the original 67.
The exact historical input yields all 171 sites, including all 67 exact-only
optional sites. The direct batch schedules 24 TS2412 writes and one TS2420 relation. Fresh data literals add 25 scheduled sites: 13 TS2375, 7 TS2379, 3 TS2345, 2 TS2322. Spreads, accessor literals and other structural/callback contracts remain errors.
It retains 90 ordinary production diagnostics and 121 unconverted option errors.
Consequently **zero of the 67 have compiled into whole-program emitted checks**.
The successful runtime fixtures emit three direct checks, one construction check and record one relation.
This distinction is deliberate and is a block on the requested completion claim.

Commands (full output retained in evidence or named temporary logs):

```
go test ./internal/load ./internal/lower ./cmd/adamic -run 'TestProduction|TestProject|TestOptionalWidening|TestWASIRequestSelection|TestRequestNativeStaysCommand' -count=1
go test ./internal/oracle -run '^TestOptional(LiteralGuard|WriteGuard|ImplementsRepresentation)$' -v -count=1
go test ./internal/oracle -run 'TestOptionalWriteGuard|TestOptionalField' -count=1
go run ./stage3/optional-writes/probe.go /tmp/optional-writes-ledger-tree/src/compiler/*.ts
go run ./stage3/stricter-options/probe.go --production /tmp/optional-writes-ledger-tree/src/compiler/*.ts
python stage3/stricter-options/validate.py /tmp/optional-ledger-loader.json /tmp/optional-writes-ledger-tree --allow-project-errors
```

Planning estimate previously given: October 10 UTC for all 67, first half one
working day. This remains provisional; the whole-program checker blocks mean it
is not a verified delivery date. Do not count the scheduled batch as completed
whole-program checks.

Fresh construction checks hold the newly allocated object, then require every
explicit data property to be own-present before returning it to its target
contract. A sanitizer-enabled mutant marks the allocation's slots absent; the
construction guard catches it with exit 70. The loader test separately proves
that acceptance records a pending contract without claiming an emitted check.
