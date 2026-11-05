# Mutants for a method read only to test for it (#4q1b2er)

Each mutant is one exact line edit, run on a Linux box against the test it names, then put back from
a saved copy. The oracle rows run with `ADAMIC_GATE_UNCACHED=1`. The lower tests are in
internal/lower/method_presence_test.go, and the oracle's is internal/oracle/method_presence_test.go.

| Mutant | Edit | Red |
|---|---|---|
| Every consumer admitted | lower/method_presence.go, `methodPresenceTest`: `\|\| !l.alwaysPresentMethod(node) \|\| !presenceConsumer(node) {` made `\|\| !l.alwaysPresentMethod(node) {` | TestMethodReadsThatKeepTheValueAreRefused, all eight rows, `\|\|` among them: each stops as not yet at property's backstop instead of refused, since no consumer hook takes it. TestAMethodReadAsAValueIsRefused's six rows go red with them |
| The receiver not evaluated | `methodPresence`: `return methodPresenceRead{effects: []ir.Statement{ir.Evaluate{Value: value}}}, true, nil` made `return methodPresenceRead{}, true, nil` | TestMethodPresenceAgreesWithNode, receiver_effects.a, native, sanitized and javascript: the getter never runs for a test, so its count prints 0 where Node prints 1. The other five stay green: their receivers are reads with no effects, and optional_receiver.a's ?. test reads its receiver through its presence, not through the effects this drops |
