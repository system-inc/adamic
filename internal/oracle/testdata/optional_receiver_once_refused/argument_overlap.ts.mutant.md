# Optional receiver argument-overlap witness

Control: `TestOptionalReceiverArgumentOverlapRefuses`. `Set.size` is the real
library property instantiating the `a.b?.has(a.b.c)` overlap: an argument contains
an expression syntactically equal to the receiver. The intentional syntax rule
keeps the words `an optional intrinsic without exactly one explicit receiver operand`.
Main currently admits this size variant, so this pin also records the deliberate
admission tightening required by that syntax rule.

Production-source mutant: remove optionalIntrinsic's loop over call arguments
that invokes optionalIntrinsicArgumentRepeats. Keep the scoped binding, both
use-count checks, and every receiver guard.

Run with a Go source overlay and `ADAMIC_GATE_UNCACHED=1`:

```
go test -overlay /tmp/apple-once/mutant-b.json -json -count=1 -timeout 4h -run '^TestOptionalReceiverArgumentOverlapRefuses$' ./internal/oracle
```

Observed red: `TestOptionalReceiverArgumentOverlapRefuses`, reporting
`want the argument/receiver overlap refusal, got <nil>`. The mutant lowers the
fixture. This would be semantically sound: JavaScript re-reads a.b in the
argument. Its admission is left to Lower, as requested. The mutant is not in
the production tree.
