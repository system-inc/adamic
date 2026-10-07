# Rules outside this landing branch

The complete implementations and prior evidence remain in commit
2ce7c57035b02996a3e757e4da640a28d636dd15 on this branch's history.
Their descriptor directories are removed from this landing tip, as requested.

- `@typescript-eslint/no-unnecessary-parameter-property-assignment`: original
  upstream `constructor(public { a }: { a: string })` fails in the stage 1
  parser with `expected CloseParenToken, got OpenBraceToken at 33`.
  Exact reproducer: `parked/constructor-parser.ts.txt`.
- `no-lonely-if`: eleven nested else-if fixes exhaust Go's ten-pass budget,
  returning `Converged:false`; the shared oracle panics `fix failed` before
  comparison. Exact reproducer: `parked/fix-budget.ts.txt`.

Neither reproducer is deleted or weakened. Eight other owned rule descriptors
remain registered. These archived raw inputs are not active owned witnesses.
