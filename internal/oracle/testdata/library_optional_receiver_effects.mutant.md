# Optional receiver eager-argument witness

Control: `TestOptionalReceiverArgumentEffects`, also the fixture's native,
JavaScript, and WASI oracle families. The receiver logs four times and the
argument/callback factories log only on the two present-receiver paths.
The final line is `4 2`.

Production-source mutant: immediately after optionalIntrinsic creates `body`
with the receiver declaration, hoist MapDelete's key out of its conditional:

```go
if call, ok := lowered.(ir.MapDelete); ok {
    argumentLocal := len(l.result.Locals)
    l.result.Locals = append(l.result.Locals, ir.Local{
        Name: "mutant_eager_argument", Type: call.Key.Type(),
        Function: function, ExpressionAssigned: true,
    })
    body = append(body, ir.Declare{Local: argumentLocal, Value: call.Key})
    call.Key = ir.Read{Local: argumentLocal, Of: call.Key.Type()}
    lowered = call
}
```

This retains the saved receiver and its guard, but evaluates the argument even
when the receiver is undefined. A Go source overlay planted this exact change;
the working tree's implementation stayed unchanged. With oracle caching disabled:

```
ADAMIC_GATE_UNCACHED=1 go test -overlay /tmp/optional-mutant-overlay.json -json ./internal/oracle -run '^TestOptionalReceiverArgumentEffects$' -count=1 -timeout=15m
```

Observed red: `TestOptionalReceiverArgumentEffects` failed on both native and
JavaScript with `stdout differs`. Each mutant adds `argument` between the absent
receiver's `receiver` and `undefined` lines and ends with `4 3`. Node ends with
`4 2`; all runs exit zero with empty stderr. The mutant was not committed.
