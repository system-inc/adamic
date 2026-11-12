# Optional receiver argument reread witness

Control: `TestOptionalReceiverArgumentOverlapAgrees`, held to Node on native,
sanitized native with leak checking, JavaScript and WASI.

The absent optional field skips its argument, leaving the counter at zero.
The present receiver is the first Set returned by Changing.b, containing 1.
The argument increments its counter and reads b again, receiving a fresh Set
containing 1 and 2. Its size is 2, which the first Set does not contain.
Node prints:

```
undefined 0 0
false 1 2
true
```

The final row is the plain `a.b?.has(a.b.size)` agreement. The first row's final
zero denotes no Changing getter reads; the present row reports a.reads itself.
The getter's return type is non-nullable. The absent row uses an optional data
field and an argument with only a counter bump, avoiding the separate existing
refusal of narrowed accessor rereads. All three rows lower as .a.

Lower's production-source mutant: replace the receiver AST pointer comparison
at the top of lowering.expression with structural syntax equality, comparing
kinds, resolved identifier symbols and children while ignoring source positions.
Keep the call scope check. The argument's a.b then also returns the saved Read.

For the runtime witness, also bypass optionalIntrinsic's `uses > 1` refusal;
otherwise that structural guard catches the mutant before a backend runs. Keep
the zero-use check. The structural comparator exists only in the source overlay.

```
ADAMIC_GATE_UNCACHED=1 go test -overlay /tmp/apple-overlap/mutant.json -json -count=1 -timeout 4h -run '^TestOptionalReceiverArgumentOverlapAgrees$/(native|javascript)$' ./internal/oracle
```

Observed red: `TestOptionalReceiverArgumentOverlapAgrees/native` and
`TestOptionalReceiverArgumentOverlapAgrees/javascript`, both `stdout differs`.
The present row becomes `true 1 1`: the argument incorrectly uses the first
Set's size, and the getter is read only once. Node and mutant exit zero with
empty stderr. The mutant is not in the production tree.
