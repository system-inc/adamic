# Skip the yielded-item judgment (#2p41tb1)

In `container_fit.go`, replace the return inside `if containerYieldResultPair(value, target)`
with `return ""`. Leave the other walk branches unchanged.

`go test ./internal/oracle -run '^TestYieldResultUnwrapRefused/hazard$' -count=1`
is red: `TestYieldResultUnwrapRefused/hazard` receives nil instead of the exact
container-fit NotYet. The fixture lowers under the mutant. Node prints `true`;
the yielded FakeMap is an ir.Object, while the target reads its field as ir.Map.
The explicit generator result keeps the operand's own slot Source, so this pin
reaches the unwrap rather than stopping first at `yield elem`.
