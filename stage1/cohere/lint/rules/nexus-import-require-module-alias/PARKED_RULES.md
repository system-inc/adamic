# Rules outside this landing branch

The complete implementations and prior evidence remain in commit
2ce7c57035b02996a3e757e4da640a28d636dd15 on this branch's history.
The constructor rule descriptor is removed from this landing tip. The shared fix in
4f18a05c9 allows no-lonely-if to be restored and tested with its original budget witness.

- `@typescript-eslint/no-unnecessary-parameter-property-assignment`: original
  upstream `constructor(public { a }: { a: string })` fails in the stage 1
  parser with `expected CloseParenToken, got OpenBraceToken at 33`.
  Exact reproducer: `parked/constructor-parser.ts.txt`.
- Resolved shared blocker, `no-lonely-if`: 4f18a05c9 serializes the ten-pass
  rejection and unconverged rules rather than panicking. Its descriptor and
  unchanged eleven-level witness are restored. The original reproducer is also
  retained at `parked/fix-budget.ts.txt`.

Neither reproducer is deleted or weakened. Nine owned rule descriptors
remain registered. These archived raw inputs are not active owned witnesses.
