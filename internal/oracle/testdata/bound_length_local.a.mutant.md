# Bound local length reads runtime metadata

`TestBoundLengthLocalAgrees/native` goes red when the runtime `source_length`
assignments for BoundClosure and BoundMethod are removed. The bound adapters
keep their emitted zero, printing:

```
0 0 0 0 0
0 0
0 0
```

Node and the fixed head print:

```
3 2 1 0 0
1 0
1 0
```

The mutant was run using a Go source overlay, leaving the production files intact.
The target is captured in the adapter's first environment cell; the fixed count
includes arguments beyond the target's source length.

A second tested mutant assigns `len(adapter.Parameters)` instead of reading the
captured target's runtime length. `TestBoundLengthLocalAgrees/native` then prints
`3 2` for the default-parameter row and `2 1` for the typed-arrow row; Node prints
`1 0` for both. The typed arrow's actual one-parameter closure must win over the
variable's declared two-parameter signature.

The JavaScript counterpart was also tested: replacing the BoundClosure
constructor's runtime length with zero turns
`TestBoundLengthLocalAgrees/javascript` red with the same all-zero output.

The final fixture keeps the original-function rows. Method observations remain refused: a reassigned callable can hide an entry-backed bound method whose runtime length is unavailable. The erased-rebind probe printed 2 in Node and 0 on the first implementation; TestBoundLengthLocalKeepsRefusals/erased_rebind now pins the stop.
