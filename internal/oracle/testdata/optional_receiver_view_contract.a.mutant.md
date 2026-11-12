# Optional receiver binding witness

Control: `TestOptionalReceiverViewContractAgrees`, with native, sanitized,
JavaScript and WASI rows. The identity cast registers the recursive array view;
main's reflect comparison refuses the first find because its two property reads
carry different view contracts. The identity cast also lowers in 0.1, so the fixture uses .a.

Production-source mutant: remove the bound-read return at the top of
`lowering.expression`, and disable optionalIntrinsic's `uses == 0` refusal so the
mutant can execute. Keep the saved receiver declaration and the builtin's own
receiver lowering. Disabling the counter guard isolates the runtime proof from
that earlier structural safety check.

Run with a Go source overlay and `ADAMIC_GATE_UNCACHED=1`:

```
go test -overlay /tmp/apple-once/mutant-a.json -json -count=1 -timeout 4h -run '^TestOptionalReceiverViewContractAgrees$/(native|javascript)$' ./internal/oracle
```

Observed red: `TestOptionalReceiverViewContractAgrees/native` and
`TestOptionalReceiverViewContractAgrees/javascript`, both `stdout differs`.
Node calls the receiver three times; the mutant calls it five times, twice on
each present-receiver path and once on the absent path. Both exit zero with
empty stderr. The mutant is not in the production tree.
